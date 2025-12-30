package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/evaan/Claret/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func FrontendHandler(c *gin.Context, db *gorm.DB, rdb *redis.Client) {
	semesterStr := util.GetParamOrQuery(c, "semester")
	var semester int
	if semesterStr == "" {
		err := db.Raw("SELECT id FROM semesters WHERE latest").Scan(&semester).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		semesterStr = strconv.Itoa(semester)
	} else {
		var err error
		semester, err = strconv.Atoi(semesterStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	response := util.FrontendAPIResponse{
		Courses:    make([]util.CourseFrontendAPI, 0),
		Subjects:   make([]util.Subject, 0),
		Times:      make([]util.CourseTimeFrontendAPI, 0),
		Seatings:   make([]util.CourseSeating, 0),
		Professors: make([]util.ProfessorRatingAPI, 0),
	}

	ctx := context.Background()
	cacheKey := "frontend:" + semesterStr
	val, err := rdb.Get(ctx, cacheKey).Result()
	if err == nil {
		if err := json.Unmarshal([]byte(val), &response); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else {
		if err != redis.Nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var courses []util.CourseFrontendAPI

		err := db.Raw(`
			SELECT
				c.id,
				c.name,
				c.crn,
				c.section,
				c.credits,
				c.campus,
				c.subject_id,
				c.type
			FROM courses c
			WHERE c.semester_id = ?
		`, semester).Scan(&courses).Error

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var instructorRows []struct {
			CourseID   string `gorm:"column:course_id"`
			Instructor string `gorm:"column:professor_name"`
		}

		err = db.Raw(`
			SELECT
				c.id AS course_id,
				ci.professor_name
			FROM courses c
			JOIN course_instructors ci ON ci.course_key = c.key
			WHERE c.semester_id = ?
		`, semester).Scan(&instructorRows).Error

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		instructorMap := make(map[string][]string)

		for _, row := range instructorRows {
			instructorMap[row.CourseID] =
				append(instructorMap[row.CourseID], row.Instructor)
		}

		for i := range courses {
			if instructors, ok := instructorMap[courses[i].ID]; ok {
				courses[i].Instructors = instructors
			} else {
				courses[i].Instructors = []string{}
			}
		}

		response.Courses = courses

		err = db.Raw(`
			SELECT DISTINCT s.*
			FROM subjects s
			JOIN courses c ON s.id = c.subject_id
			WHERE c.semester_id = ?
		`, semester).Scan(&response.Subjects).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		err = db.Raw(`
			SELECT
				ct.start_time,
				ct.end_time,
				ct.days,
				ct.location,
				ct.date_range,
				ct.type,
				c.crn
			FROM course_times ct
			JOIN courses c ON c.key = ct.course_key
			WHERE ct.course_key LIKE ? || '%'
		`, strconv.Itoa(semester)).Scan(&response.Times).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		err = db.Raw(`SELECT * FROM professor_ratings`).Scan(&response.Professors).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		cacheData, err := json.Marshal(response)
		if err == nil {
			_ = rdb.Set(ctx, cacheKey, cacheData, 24*time.Hour).Err()
		}
	}

	var cursor uint64
	for {
		keys, newCursor, err := rdb.Scan(
			ctx,
			cursor,
			"seats:"+semesterStr+":*",
			1000,
		).Result()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if len(keys) > 0 {
			vals, err := rdb.MGet(ctx, keys...).Result()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			for _, v := range vals {
				if v == nil {
					continue
				}
				var seating util.CourseSeating
				if err := json.Unmarshal([]byte(v.(string)), &seating); err == nil {
					response.Seatings = append(response.Seatings, seating)
				}
			}
		}

		cursor = newCursor
		if cursor == 0 {
			break
		}
	}

	c.JSON(http.StatusOK, response)
}

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
		Courses:  make([]util.CourseFrontendAPI, 0),
		Subjects: make([]util.Subject, 0),
		Times:    make([]util.CourseTimeFrontendAPI, 0),
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

		err = db.Raw(`
			SELECT
				c.id,
				c.name,
				c.crn,
				c.section,
				c.credits,
				c.campus,
				c.type,
				c.subject_id,
				COALESCE(STRING_AGG(DISTINCT ci.professor_name, ', '), '') AS instructor
			FROM courses c
			LEFT JOIN course_instructors ci ON ci.course_key = c.key
			WHERE c.semester_id = ?
			GROUP BY
				c.id,
				c.name,
				c.crn,
				c.section,
				c.credits,
				c.campus,
				c.type,
				c.subject_id
		`, semester).Scan(&response.Courses).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

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
				start_time,
				end_time,
				days,
				location,
				date_range,
				type
			FROM course_times
			WHERE course_key LIKE ? || '%'
		`, strconv.Itoa(semester)).Scan(&response.Times).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		cacheData, err := json.Marshal(response)
		if err == nil {
			_ = rdb.Set(ctx, cacheKey, cacheData, 24*time.Hour).Err()
		}
	}

	c.JSON(http.StatusOK, response)
}

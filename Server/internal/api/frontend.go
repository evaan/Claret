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

// CoursesHandler godoc
// @Summary Frontend
// @Description Returns all data (courses, prof ratings, subjects, seats, times, and exams) for a specified semester (or latest for no semester)
// @Accept json
// @Produce json
// @Param semester query string false "Semester ID (i.e. 202401)"
// @Success 200 {object} util.FrontendAPIResponse
// @Failure 400 {object} util.ErrorResponse
// @Failure 500 {object} util.ErrorResponse
// @Router /frontend [get]
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
		Profs:    make([]util.ProfessorRatingAPI, 0),
		Subjects: make([]util.Subject, 0),
		Times:    make([]util.CourseTimeFrontendAPI, 0),
		Exams:    make([]util.ExamTimeAPI, 0),
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

		err = db.Raw(`SELECT c.id, c.name, c.crn, c.section, c.credits, c.campus, c.date_range, c.subject_id, c.semester_id, c.comment, c.levels, c.registration_dates, c.types,
			STRING_AGG(DISTINCT ci.professor_name, ', ') AS instructor FROM courses c LEFT JOIN course_instructors ci ON ci.course_key = c.key
			WHERE c.semester_id = ? GROUP BY c.id, c.name, c.crn, c.section, c.credits, c.campus, c.date_range, c.subject_id, c.semester_id, c.comment,
			c.levels, c.registration_dates, c.types;`, semester).Scan(&response.Courses).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		err = db.Raw(`SELECT DISTINCT pr.* FROM professor_ratings pr
			JOIN course_instructors ci ON pr.professor_name = ci.professor_name
			JOIN courses c ON ci.course_key = c.key
			WHERE c.semester_id = ?`, semester).Scan(&response.Profs).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		err = db.Raw(`SELECT DISTINCT s.*
			FROM subjects s
			JOIN courses c ON s.id = c.subject_id
			WHERE c.semester_id = ?`, semester).Scan(&response.Subjects).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		err = db.Raw(`SELECT DISTINCT
			start_time, end_time, days, location, date_range,
			type, course_crn, semester_id FROM course_times WHERE semester_id = ?`, semester).Scan(&response.Times).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		err = db.Raw(`SELECT *
			FROM exam_times et
			JOIN courses c ON et.course_key = c.key
			WHERE c.semester_id = ?`, semester).Scan(&response.Exams).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		cacheData, err := json.Marshal(response)
		if err == nil {
			_ = rdb.Set(ctx, cacheKey, cacheData, 24*time.Hour).Err()
		}
	}

	response.Seatings = make([]util.CourseSeatingResponse, 0)
	var cursor uint64
	for {
		var keys []string
		keys, cursor, err = rdb.Scan(ctx, cursor, "seats:"+semesterStr+":*", 0).Result()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		for _, key := range keys {
			val, err := rdb.Get(ctx, key).Result()
			if err == redis.Nil {
				continue
			} else if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			var seating util.CourseSeating
			if err := json.Unmarshal([]byte(val), &seating); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			response.Seatings = append(response.Seatings, util.CourseSeatingResponse{
				CRN:      seating.CRN,
				Seats:    seating.Seats,
				Waitlist: seating.Waitlist,
			})
		}

		if cursor == 0 {
			break
		}
	}

	c.JSON(http.StatusOK, response)
}

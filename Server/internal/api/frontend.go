package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/evaan/Claret/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// FrontendHandler returns section metadata and the latest known seat snapshots.
// @Summary Get schedule builder data
// @Description Returns courses, subjects, meetings, professor ratings, and timestamped seat snapshots. Seat snapshots fall back to PostgreSQL when the live cache is unavailable.
// @Tags Frontend
// @Produce json
// @Param semester query int false "Semester ID (defaults to the latest semester)"
// @Success 200 {object} util.FrontendAPIResponse
// @Failure 400 {object} util.ErrorResponse
// @Failure 500 {object} util.ErrorResponse
// @Router /frontend [get]
func FrontendHandler(c *gin.Context, db *gorm.DB, rdb *redis.Client) {
	db = db.WithContext(c.Request.Context())
	semester, ok := semesterParam(c, db, util.GetParamOrQuery(c, "semester"), true)
	if !ok {
		return
	}
	semesterStr := strconv.Itoa(semester)

	response := util.FrontendAPIResponse{
		Courses:    make([]util.CourseFrontendAPI, 0),
		Subjects:   make([]util.Subject, 0),
		Times:      make([]util.CourseTimeFrontendAPI, 0),
		Seatings:   make([]util.CourseSeating, 0),
		Professors: make([]util.ProfessorRatingAPI, 0),
	}

	ctx := c.Request.Context()
	cacheKey := "frontend:v2:" + semesterStr
	val, err := rdb.Get(ctx, cacheKey).Result()
	cacheHit := err == nil && json.Unmarshal([]byte(val), &response) == nil
	if !cacheHit {
		courses := make([]util.CourseFrontendAPI, 0)

		err := db.Raw(`
			SELECT
				c.id,
				c.name,
				c.crn,
				c.section,
				c.credits,
				c.campus,
				c.subject_id,
				c.type,
				c.instructional_method,
				c.part_of_term,
				c.is_section_linked,
				c.link_identifier,
				c.cross_list,
				c.credits_min,
				c.credits_max,
				c.prerequisites,
				c.corequisites,
				c.linked_sections
			FROM courses c
			WHERE c.semester_id = ?
		`, semester).Scan(&courses).Error

		if err != nil {
			internalError(c, err)
			return
		}

		var instructorRows []struct {
			CRN        string `gorm:"column:crn"`
			Instructor string `gorm:"column:professor_name"`
		}

		err = db.Raw(`
			SELECT
				c.crn,
				ci.professor_name
			FROM courses c
			JOIN course_instructors ci ON ci.course_key = c.key
			WHERE c.semester_id = ?
		`, semester).Scan(&instructorRows).Error

		if err != nil {
			internalError(c, err)
			return
		}

		instructorMap := make(map[string][]string)

		for _, row := range instructorRows {
			instructorMap[row.CRN] =
				util.Unique(append(instructorMap[row.CRN], row.Instructor))
		}

		for i := range courses {
			if instructors, ok := instructorMap[courses[i].CRN]; ok {
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
			internalError(c, err)
			return
		}

		err = db.Raw(`
			SELECT
				ct.start_time,
				ct.end_time,
				ct.days,
				ct.location,
				ct.date_range,
				ct.start_date,
				ct.end_date,
				ct.building_name,
				ct.type,
				c.crn
			FROM course_times ct
			JOIN courses c ON c.key = ct.course_key
			WHERE c.semester_id = ?
		`, semester).Scan(&response.Times).Error
		if err != nil {
			internalError(c, err)
			return
		}

		cacheData, err := json.Marshal(response)
		if err == nil {
			if err := rdb.Set(ctx, cacheKey, cacheData, time.Hour).Err(); err != nil {
				log.Printf("Unable to cache frontend metadata: %v", err)
			}
		}
	}

	// RMP refreshes after course scraping; don't hide new ratings in an old
	// course metadata cache (including caches created before this change).
	response.Professors = make([]util.ProfessorRatingAPI, 0)
	if err := db.Raw(`SELECT * FROM professor_ratings`).Scan(&response.Professors).Error; err != nil {
		internalError(c, err)
		return
	}

	// Seat snapshots are deliberately outside the metadata cache.
	// Redis overlays newer values; PostgreSQL provides the last known snapshot.
	response.Seatings = make([]util.CourseSeating, 0)
	if err := db.Where("semester = ?", semester).Find(&response.Seatings).Error; err != nil {
		internalError(c, err)
		return
	}
	seatings := make(map[string]util.CourseSeating, len(response.Seatings))
	for _, seating := range response.Seatings {
		seatings[seating.CRN] = seating
	}
	// Fetch only keys for known courses, in bounded batches; never SCAN the
	// whole Redis keyspace for every public request.
	for start := 0; start < len(response.Courses); start += 500 {
		end := min(start+500, len(response.Courses))
		keys := make([]string, 0, end-start)
		for _, course := range response.Courses[start:end] {
			keys = append(keys, "seats:"+semesterStr+":"+course.CRN)
		}
		if len(keys) > 0 {
			vals, err := rdb.MGet(ctx, keys...).Result()
			if err != nil {
				break
			}

			for _, v := range vals {
				value, ok := v.(string)
				if !ok {
					continue
				}
				var seating util.CourseSeating
				if err := json.Unmarshal([]byte(value), &seating); err == nil && seating.Semester == semester {
					previous, exists := seatings[seating.CRN]
					if !exists || seating.UpdatedAt.After(previous.UpdatedAt) {
						seatings[seating.CRN] = seating
					}
				}
			}
		}
	}
	response.Seatings = response.Seatings[:0]
	for _, seating := range seatings {
		response.Seatings = append(response.Seatings, seating)
	}

	c.JSON(http.StatusOK, response)
}

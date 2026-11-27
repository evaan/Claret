package api

import (
	"net/http"

	"github.com/evaan/Claret/internal/util"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ScheduleHandler returns only the sections needed to display a shared schedule.
// @Summary Get schedule-only data
// @Description Returns courses and meetings for the requested CRNs, without catalog subjects, seat counts, or ratings. An empty CRN list returns empty arrays.
// @Produce json
// @Param semester query int true "Semester ID"
// @Param crns query string false "Comma-separated CRNs (up to 100)"
// @Success 200 {object} util.ScheduleAPIResponse
// @Failure 400 {object} util.ErrorResponse
// @Failure 500 {object} util.ErrorResponse
// @Router /schedule [get]
func ScheduleHandler(c *gin.Context, db *gorm.DB) {
	db = db.WithContext(c.Request.Context())
	crns, ok := crnParams(c, c.Query("crns"), false)
	if !ok {
		return
	}
	semester, ok := semesterParam(c, db, c.Query("semester"), false)
	if !ok {
		return
	}

	response := util.ScheduleAPIResponse{
		Courses: make([]util.ScheduleCourseAPI, 0),
		Times:   make([]util.CourseTimeFrontendAPI, 0),
	}
	if len(crns) == 0 {
		c.JSON(http.StatusOK, response)
		return
	}
	err := db.Table("courses").
		Select("id, name, crn, section, credits, campus, subject_id, type, instructional_method").
		Where("semester_id = ? AND crn IN ?", semester, crns).
		Order("id, section, crn").
		Scan(&response.Courses).Error
	if err != nil {
		internalError(c, err)
		return
	}
	err = db.Table("course_times AS ct").
		Select("ct.start_time, ct.end_time, ct.days, ct.location, ct.date_range, ct.start_date, ct.end_date, ct.building_name, ct.type, c.crn").
		Joins("JOIN courses c ON c.key = ct.course_key").
		Where("c.semester_id = ? AND c.crn IN ?", semester, crns).
		Order("c.crn, ct.id").
		Scan(&response.Times).Error
	if err != nil {
		internalError(c, err)
		return
	}
	c.JSON(http.StatusOK, response)
}

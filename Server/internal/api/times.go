package api

import (
	"net/http"

	"github.com/evaan/Claret/internal/util"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// TimesHandler godoc
// @Summary Get Course Times
// @Description Returns all times from course
// @Accept json
// @Produce json
// @Param semester query string true "Semester ID (i.e. 202401)"
// @Param crn query string true "Course Registration Number (i.e. 40983)"
// @Success 200 {array} util.CourseTimeAPI
// @Failure 400 {object} util.ErrorResponse
// @Failure 500 {object} util.ErrorResponse
// @Router /times [get]
func TimesHandler(c *gin.Context, db *gorm.DB) {
	semester := util.GetParamOrQuery(c, "semester")
	crn := util.GetParamOrQuery(c, "crn")
	if crn == "" || semester == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "semester and crn are required parameters",
		})
		return
	}

	times := make([]util.CourseTimeAPI, 0)
	err := db.Raw(`SELECT
		ct.start_time, ct.end_time, ct.days, ct.location, ct.date_range,
		ct.type, ct.course_key, ct.course_crn, ct.semester_id,
		STRING_AGG(DISTINCT ci.professor_name, ', ' ORDER BY ci.professor_name) AS professor_names
		FROM course_times ct
		LEFT JOIN course_instructors ci ON ct.course_key = ci.course_key
		WHERE ct.semester_id = ? AND ct.course_crn = ?
		GROUP BY
		ct.start_time, ct.end_time, ct.days, ct.location, ct.date_range,
		ct.type, ct.course_key, ct.course_crn, ct.semester_id`, semester, crn).Scan(&times).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, times)
}

package api

import (
	"net/http"
	"strings"

	"github.com/evaan/Claret/internal/util"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CoursesHandler godoc
// @Summary Get all courses
// @Description Returns all courses for a specified semester
// @Accept json
// @Produce json
// @Param semester query string true "Semester ID (i.e. 202401)"
// @Param id query string false "Course ID (i.e. ECE 3400)"
// @Param crn query string false "Course Registration Number (i.e. 40983)"
// @Success 200 {array} util.CourseAPI
// @Failure 400 {object} util.ErrorResponse
// @Failure 500 {object} util.ErrorResponse
// @Router /courses [get]
func CoursesHandler(c *gin.Context, db *gorm.DB) {
	db = db.WithContext(c.Request.Context())
	semester := util.GetParamOrQuery(c, "semester")
	if semester == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "semester is a required parameter",
		})
		return
	}

	id := util.GetParamOrQuery(c, "id")
	crn := strings.TrimSpace(c.Query("crn"))

	courses := make([]util.CourseAPI, 0)
	err := db.Raw(`SELECT c.*,
		COALESCE((SELECT STRING_AGG(DISTINCT ci.professor_name, ', ' ORDER BY ci.professor_name)
			FROM course_instructors ci WHERE ci.course_key = c.key), '') AS instructor
		FROM courses c WHERE c.semester_id = ?
		AND (? = '' OR c.id LIKE '%' || ? || '%') AND (? = '' OR c.crn = ?)
		ORDER BY c.id, c.section, c.crn`, semester, id, id, crn, crn).Scan(&courses).Error
	if err != nil {
		internalError(c, err)
		return
	}

	c.JSON(http.StatusOK, courses)
}

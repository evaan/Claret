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
	err := db.Raw("SELECT * FROM courses WHERE semester_id = ? AND (? = '' OR id LIKE '%' || ? || '%') AND (? = '' OR crn = ?)", semester, id, id, crn, crn).Scan(&courses).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, courses)
}

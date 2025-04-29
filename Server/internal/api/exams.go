package api

import (
	"net/http"

	"github.com/evaan/Claret/internal/util"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ExamsHandler godoc
// @Summary Get all exams
// @Description Returns all exams for a specified semester
// @Accept json
// @Produce json
// @Param semester query string true "Semester ID (i.e. 202401)"
// @Success 200 {array} util.ExamTimeAPI
// @Failure 400 {object} util.ErrorResponse
// @Failure 500 {object} util.ErrorResponse
// @Router /exams [get]
func ExamsHander(c *gin.Context, db *gorm.DB) {
	semester := util.GetParamOrQuery(c, "semester")
	if semester == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "semester is a required parameter",
		})
		return
	}

	exams := make([]util.ExamTimeAPI, 0)
	err := db.Raw(`SELECT *
		FROM exam_times et JOIN courses c ON et.course_key = c.key
		WHERE c.semester_id = ?`, semester).Scan(&exams).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, exams)
}

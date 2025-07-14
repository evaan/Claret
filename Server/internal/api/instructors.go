package api

import (
	"net/http"

	"github.com/evaan/Claret/internal/util"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// InstructorsHandler godoc
// @Summary Get Course Instructors
// @Description Returns all instructors and instructor ratings from a course
// @Accept json
// @Produce json
// @Param semester query string true "Semester ID (i.e. 202401)"
// @Param crn query string true "Course Registration Number (i.e. 40983)"
// @Success 200 {array} util.ProfessorRatingAPI
// @Failure 400 {object} util.ErrorResponse
// @Failure 500 {object} util.ErrorResponse
// @Router /instructors [get]
func InstructorsHandler(c *gin.Context, db *gorm.DB) {
	semester := util.GetParamOrQuery(c, "semester")
	crn := util.GetParamOrQuery(c, "crn")
	if crn == "" || semester == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "semester and crn are required parameters",
		})
		return
	}

	ratings := make([]util.ProfessorRatingAPI, 0)
	err := db.Raw("SELECT ci.professor_name, pr.rating, pr.difficulty, pr.would_retake, pr.rating_count FROM course_instructors ci LEFT JOIN professor_ratings pr ON ci.professor_name = pr.professor_name WHERE ci.course_key = ?", semester+crn).Scan(&ratings).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ratings)
}

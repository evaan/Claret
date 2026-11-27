package api

import (
	"net/http"

	"github.com/evaan/Claret/internal/util"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RmpHandler godoc
// @Summary Get Course Instructors
// @Description Returns all instructor ratings
// @Accept json
// @Produce json
// @Param name query string false "Instructor Name"
// @Success 200 {array} util.ProfessorRatingAPI
// @Failure 400 {object} util.ErrorResponse
// @Failure 500 {object} util.ErrorResponse
// @Router /rmp [get]
func RmpHandler(c *gin.Context, db *gorm.DB) {
	db = db.WithContext(c.Request.Context())
	ratings := make([]util.ProfessorRatingAPI, 0)

	name := util.GetParamOrQuery(c, "name")

	err := db.Raw("SELECT * FROM professor_ratings WHERE professor_name LIKE ?", "%"+name+"%").Scan(&ratings).Error
	if err != nil {
		internalError(c, err)
		return
	}

	c.JSON(http.StatusOK, ratings)
}

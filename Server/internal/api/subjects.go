package api

import (
	"net/http"

	"github.com/evaan/Claret/internal/util"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SubjectsHandler godoc
// @Summary Get Subjects
// @Description Returns all subjects
// @Accept json
// @Produce json
// @Param semester query string false "Semester ID (i.e. 202401)"
// @Success 200 {array} util.Subject
// @Failure 500 {object} util.ErrorResponse
// @Router /subjects [get]
func SubjectsHandler(c *gin.Context, db *gorm.DB) {
	subjects := make([]util.Subject, 0)

	err := db.Raw(`SELECT DISTINCT s.*
		FROM subjects s
		JOIN courses c ON s.id = c.subject_id
		WHERE (? = '' OR c.semester_id = ?)`, util.GetParamOrQuery(c, "semester")).Scan(&subjects).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, subjects)
}

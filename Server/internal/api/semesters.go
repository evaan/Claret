package api

import (
	"net/http"

	"github.com/evaan/Claret/internal/util"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SemestersHandler godoc
// @Summary Get Semesters
// @Description Returns all semesters
// @Accept json
// @Produce json
// @Param semester query string false "Semester ID (i.e. 202401)"
// @Success 200 {array} util.Semester
// @Failure 500 {object} util.ErrorResponse
// @Router /semester [get]
func SemestersHandler(c *gin.Context, db *gorm.DB) {
	semesters := make([]util.Semester, 0)

	err := db.Raw("SELECT * FROM semesters WHERE (? = 'false' OR latest = true) AND (? = 'false' OR mi = true) AND (? = 'false' OR medicine = true)", util.GetParamOrQueryWithDefault(c, "latest", "false"), util.GetParamOrQueryWithDefault(c, "mi", "false"), util.GetParamOrQueryWithDefault(c, "medicine", "false")).Scan(&semesters).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, semesters)
}

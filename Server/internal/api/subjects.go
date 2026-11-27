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
	db = db.WithContext(c.Request.Context())
	subjects := make([]util.Subject, 0)
	query := db.Table("subjects s").Distinct("s.*").Joins("JOIN courses c ON s.id = c.subject_id")
	if raw := util.GetParamOrQuery(c, "semester"); raw != "" {
		semester, ok := semesterParam(c, db, raw, false)
		if !ok {
			return
		}
		query = query.Where("c.semester_id = ?", semester)
	}
	err := query.Scan(&subjects).Error
	if err != nil {
		internalError(c, err)
		return
	}

	c.JSON(http.StatusOK, subjects)
}

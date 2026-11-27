package api

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/evaan/Claret/internal/util"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func semesterParam(c *gin.Context, db *gorm.DB, raw string, latest bool) (int, bool) {
	raw = strings.TrimSpace(raw)
	query := db.WithContext(c.Request.Context()).Model(&util.Semester{})
	if raw == "" && latest {
		query = query.Where("latest").Order("id DESC")
	} else {
		semester, err := strconv.Atoi(raw)
		if err != nil || semester <= 0 || len(raw) > 10 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "a valid semester is required"})
			return 0, false
		}
		query = query.Where("id = ?", semester)
	}
	var semester util.Semester
	if err := query.First(&semester).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "semester not found"})
		} else {
			internalError(c, err)
		}
		return 0, false
	}
	return semester.ID, true
}

func crnParams(c *gin.Context, raw string, required bool) ([]string, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" && !required {
		return []string{}, true
	}
	if len(raw) > 1100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at most 100 CRNs may be requested"})
		return nil, false
	}
	parts := strings.Split(raw, ",")
	if len(parts) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at most 100 CRNs may be requested"})
		return nil, false
	}
	crns := make([]string, 0, len(parts))
	seen := make(map[string]bool, len(parts))
	for _, part := range parts {
		crn := strings.TrimSpace(part)
		if crn == "" || len(crn) > 10 || strings.IndexFunc(crn, func(r rune) bool {
			return r < '0' || r > '9'
		}) >= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "crns must be a comma-separated list of numeric CRNs"})
			return nil, false
		}
		if !seen[crn] {
			crns = append(crns, crn)
			seen[crn] = true
		}
	}
	return crns, true
}

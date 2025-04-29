package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/evaan/Claret/internal/scrapers"
	"github.com/evaan/Claret/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// SeatsHandler godoc
// @Summary Get Course Seats
// @Description Returns seats from a specified course, may take a few seconds if seats not cached
// @Accept json
// @Produce json
// @Param semester query string true "Semester ID (i.e. 202401)"
// @Param crn query string true "Course Registration Number (i.e. 40983)"
// @Success 200 {object} util.CourseSeating
// @Failure 400 {object} util.ErrorResponse
// @Failure 500 {object} util.ErrorResponse
// @Router /seats [get]
func SeatsHandler(c *gin.Context, db *gorm.DB, rdb *redis.Client) {
	semester := util.GetParamOrQuery(c, "semester")
	crn := util.GetParamOrQuery(c, "crn")
	if semester == "" || crn == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "semester and crn are required parameters",
		})
		return
	}

	ctx := context.Background()

	val, err := rdb.Get(ctx, "seats:"+semester+":"+crn).Result()
	if err != nil && err != redis.Nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	} else if err != redis.Nil {
		response := util.CourseSeating{}
		if err := json.Unmarshal([]byte(val), &response); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, response)
		return
	}

	response := scrapers.GetSeats(semester, crn)

	c.JSON(http.StatusOK, response)

	responseJson, err := json.Marshal(response)
	if err != nil {
		fmt.Println("Error marshalling data to put into Redis database:", err)
		return
	}
	err = rdb.Set(ctx, "seats:"+semester+":"+crn, responseJson, time.Hour).Err()
	if err != nil {
		fmt.Println("Error adding data to Redis database:", err)
	}
}

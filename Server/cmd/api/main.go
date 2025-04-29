package api

import (
	"net/http"

	_ "github.com/evaan/Claret/docs"
	"github.com/evaan/Claret/internal/api"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

// @title Claret API
// @version 1.0
// @description The API for Claret, a tool for Memorial University students.
// @host api.claretformun.com
// @BasePath /
func StartAPI(db *gorm.DB, rdb *redis.Client, enableLimit bool, limit string) {
	r := gin.Default()

	// TODO: maybe analytics for active users? redis?

	if enableLimit {
		r.Use(api.RateLimiterMiddleware(limit))
	}

	r.Use(cors.Default())

	r.SetTrustedProxies([]string{
		"127.0.0.1",

		// cloudflare ipv4
		"173.245.48.0/20",
		"103.21.244.0/22",
		"103.22.200.0/22",
		"103.31.4.0/22",
		"141.101.64.0/18",
		"108.162.192.0/18",
		"190.93.240.0/20",
		"188.114.96.0/20",
		"197.234.240.0/22",
		"198.41.128.0/17",
		"162.158.0.0/15",
		"104.16.0.0/13",
		"104.24.0.0/14",
		"172.64.0.0/13",
		"131.0.72.0/22",
	})

	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/swagger/index.html")
	})
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"alive": true,
		})
	})

	r.GET("/semesters", func(c *gin.Context) {
		api.SemestersHandler(c, db)
	})

	r.GET("/courses", func(c *gin.Context) {
		api.CoursesHandler(c, db)
	})

	r.GET("/courses/:semester", func(c *gin.Context) {
		api.CoursesHandler(c, db)
	})

	r.GET("/courses/:semester/:id", func(c *gin.Context) {
		api.CoursesHandler(c, db)
	})

	r.GET("/times", func(c *gin.Context) {
		api.TimesHandler(c, db)
	})

	r.GET("/times/:semester/:crn", func(c *gin.Context) {
		api.TimesHandler(c, db)
	})

	r.GET("/instructors", func(c *gin.Context) {
		api.InstructorsHandler(c, db)
	})

	r.GET("/instructors/:semester/:crn", func(c *gin.Context) {
		api.InstructorsHandler(c, db)
	})

	r.GET("/rmp", func(c *gin.Context) {
		api.RmpHandler(c, db)
	})

	r.GET("/rmp/:name", func(c *gin.Context) {
		api.RmpHandler(c, db)
	})

	r.GET("/subjects", func(c *gin.Context) {
		api.SubjectsHandler(c, db)
	})

	r.GET("/subjects/:semester", func(c *gin.Context) {
		api.SubjectsHandler(c, db)
	})

	r.GET("/exams/:semester", func(c *gin.Context) {
		api.ExamsHander(c, db)
	})

	r.GET("/exams", func(c *gin.Context) {
		api.ExamsHander(c, db)
	})

	r.GET("/frontend", func(c *gin.Context) {
		api.FrontendHandler(c, db, rdb)
	})

	r.GET("/frontend/:semester", func(c *gin.Context) {
		api.FrontendHandler(c, db, rdb)
	})

	r.GET("/seats/:semester/:crn", func(c *gin.Context) {
		api.SeatsHandler(c, db, rdb)
	})

	r.GET("/seats", func(c *gin.Context) {
		api.SeatsHandler(c, db, rdb)
	})

	r.GET("/claret.ics", func(c *gin.Context) {
		api.ICalHandler(c, db)
	})

	// TODO: seating

	r.Run()
}

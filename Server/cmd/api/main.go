package api

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

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
func StartAPI(db *gorm.DB, rdb *redis.Client, enableLimit bool, limit string) error {
	r := gin.Default()
	r.Use(api.RequestTimeoutMiddleware(10 * time.Second))

	if enableLimit {
		middleware, err := api.RateLimiterMiddleware(limit)
		if err != nil {
			return err
		}
		r.Use(middleware)
	}

	r.Use(cors.Default())

	// Trust only the actual ingress hops, never arbitrary Docker/private ranges.
	var proxies []string
	for _, proxy := range strings.Split(os.Getenv("API_TRUSTED_PROXIES"), ",") {
		if proxy = strings.TrimSpace(proxy); proxy != "" {
			proxies = append(proxies, proxy)
		}
	}
	if err := r.SetTrustedProxies(proxies); err != nil {
		return fmt.Errorf("invalid API_TRUSTED_PROXIES: %w", err)
	}
	r.RemoteIPHeaders = []string{"X-Forwarded-For"}

	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/swagger/index.html")
	})
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.GET("/health", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		sqlDB, err := db.DB()
		if err != nil || sqlDB.PingContext(ctx) != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"alive": false})
			return
		}
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

	r.GET("/schedule", func(c *gin.Context) {
		api.ScheduleHandler(c, db)
	})

	r.GET("/claret.ics", func(c *gin.Context) {
		api.ICalHandler(c, db)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	server := &http.Server{
		Addr:              net.JoinHostPort("", port),
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    16 << 10,
	}
	return server.ListenAndServe()
}

package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"time"
	_ "time/tzdata"

	"github.com/evaan/Claret/cmd/api"
	"github.com/evaan/Claret/cmd/scrapers"
	"github.com/evaan/Claret/internal/util"
	_ "github.com/joho/godotenv/autoload"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	if len(os.Args) == 2 && os.Args[1] == "healthcheck" {
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		client := &http.Client{Timeout: 3 * time.Second}
		resp, err := client.Get("http://" + net.JoinHostPort("127.0.0.1", port) + "/health")
		if err != nil {
			os.Exit(1)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			os.Exit(1)
		}
		return
	}
	logger := log.Default()

	logger.Println("👋 Claret")

	db, err := gorm.Open(postgres.Open(os.Getenv("POSTGRES_URL")), &gorm.Config{})
	if err != nil {
		logger.Fatal(err)
	}
	logger.Println("💿 Connected to PostgreSQL Database!")

	sqlDB, err := db.DB()
	if err != nil {
		logger.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	defer sqlDB.Close()

	if os.Getenv("AUTO_MIGRATE") != "false" {
		if err := db.AutoMigrate(
			&util.Semester{}, &util.Subject{}, &util.Course{}, &util.CourseSeating{},
			&util.CourseTime{}, &util.Professor{}, &util.CourseInstructor{},
			&util.ProfessorRating{}, &util.ExamTime{},
		); err != nil {
			logger.Fatal("Database migration failed: ", err)
		}
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:                  os.Getenv("REDIS_URL"),
		Username:              os.Getenv("REDIS_USERNAME"),
		Password:              os.Getenv("REDIS_PASSWORD"),
		DB:                    util.GetEnvAsInt("REDIS_CACHE_DB"),
		DialTimeout:           time.Second,
		ReadTimeout:           time.Second,
		WriteTimeout:          time.Second,
		ContextTimeoutEnabled: true,
		MaxRetries:            -1,
	})
	defer rdb.Close()

	if util.GetEnvAsBool("SCRAPER_ENABLED") {
		scrapers.Entrypoint(db, os.Getenv("SCRAPER_WEBHOOK_URL"), util.GetEnvAsBool("SCRAPER_ALL"), rdb)
	}

	if util.GetEnvAsBool("API_ENABLED") {
		logger.Println("💿 Connected to Redis Database!")
		if err := api.StartAPI(db, rdb, util.GetEnvAsBool("API_RATE_LIMIT_ENABLED"), os.Getenv("API_RATE_LIMIT")); err != nil {
			logger.Fatal("API stopped: ", err)
		}
	}

	select {}
}

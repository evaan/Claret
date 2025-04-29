package main

import (
	"log"
	"os"

	"github.com/evaan/Claret/cmd/api"
	"github.com/evaan/Claret/cmd/scrapers"
	"github.com/evaan/Claret/internal/util"
	_ "github.com/joho/godotenv/autoload"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	logger := log.Default()

	logger.Println("👋 Claret")

	db, err := gorm.Open(postgres.Open(os.Getenv("POSTGRES_URL")), &gorm.Config{})
	if err != nil {
		logger.Fatal(err)
	}
	logger.Println("💿 Connected to PostgreSQL Database!")

	db.AutoMigrate(&util.Semester{})
	db.AutoMigrate(&util.Subject{})
	db.AutoMigrate(&util.Course{})
	db.AutoMigrate(&util.CourseTime{})
	db.AutoMigrate(&util.Professor{})
	db.AutoMigrate(&util.CourseInstructor{})
	db.AutoMigrate(&util.ProfessorRating{})
	db.AutoMigrate(&util.ExamTime{})

	if util.GetEnvAsBool("SCRAPER_ENABLED") {
		go scrapers.Entrypoint(db, os.Getenv("API_WEBHOOK_URL"), util.GetEnvAsBool("SCRAPER_ALL"))
	}

	if util.GetEnvAsBool("API_ENABLED") {
		rdb := redis.NewClient(&redis.Options{
			Addr:     os.Getenv("REDIS_URL"),
			Username: os.Getenv("REDIS_USERNAME"),
			Password: os.Getenv("REDIS_PASSWORD"),
			DB:       util.GetEnvAsInt("REDIS_CACHE_DB"),
		})
		logger.Println("💿 Connected to Redis Database!")
		go api.StartAPI(db, rdb, util.GetEnvAsBool("API_RATE_LIMIT_ENABLED"), os.Getenv("API_RATE_LIMIT"))
	}

	select {}
}

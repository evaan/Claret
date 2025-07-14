package scrapers

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/evaan/Claret/internal/scrapers"
	"github.com/evaan/Claret/internal/util"
	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

func Scrape(db *gorm.DB, webhookUrl string, scrapeAll bool, rdb *redis.Client) {
	startTime := time.Now()
	coursesScraped := 0
	logger := log.Default()
	ctx := context.Background()

	logger.Println("⭐ Scraping Started!")

	var profs []string

	for _, semester := range scrapers.GetSemesters(logger) {
		if semester.ViewOnly && !scrapeAll {
			continue
		}
		if db.Where("id = ?", semester.ID).Find(&util.Semester{}).RowsAffected > 0 {
			if semester.ViewOnly {
				continue
			}
			db.Delete(&semester)
		}
		logger.Println("📝 Processing Semester: " + semester.Name + " (" + strconv.Itoa(semester.ID) + ")")
		db.Create(&semester)
		for _, subject := range scrapers.GetSubjects(logger, semester.ID) {
			db.FirstOrCreate(&subject)
			logger.Println(" 📝 Processing " + subject.Name + " (" + subject.ID + ")")
			courses, courseTimes, professors, courseInstructors := scrapers.GetCourses(logger, semester.ID, subject.ID)
			for _, course := range courses {
				db.Create(&course)
				// db.Create(&util.CourseSeating{CourseKey: course.Key, Capacity: 0, Available: 0, Scraped: "Never"})
				coursesScraped++
			}
			for _, time := range courseTimes {
				db.Create(&time)
			}
			for _, professor := range professors {
				db.FirstOrCreate(&professor)
				profs = append(profs, professor.Name)
			}
			for _, instructor := range courseInstructors {
				db.Create(&instructor)
			}
		}
		logger.Println(" 📝 Processing Exams")
		for _, exam := range scrapers.GetExams(semester.ID) {
			db.Save(&exam)
		}
		rdb.Del(ctx, "frontend:"+strconv.Itoa(semester.ID))
	}

	logger.Println("⭐ RMP Scraping Started!")

	profRatings := scrapers.RMP(logger, profs)
	for _, rating := range profRatings {
		db.Save(&rating)
	}

	scrapingTime := time.Since(startTime)

	logger.Println("✅ Scrape Complete in " + fmt.Sprintf("%02d:%02d", int(scrapingTime.Minutes()), int(scrapingTime.Seconds())%60) + "!")
	logger.Printf("🚀 Courses scraped: %d", coursesScraped)

	if webhookUrl != "" {
		logger.Println("🔔 Sending message to Discord")
		params := fmt.Sprintf(`{"username":"Claret Scraper","embeds":[{"author":{"name":"Claret Scraper Report","url":"https://claretformun.com"},"timestamp":"%s","color":65280,"fields":[{"name":"Scraping Time","value":"%s"},{"name":"Courses Scraped","value":"%d"}]}]}`, time.Now().Format(time.RFC3339), fmt.Sprintf("%02d:%02d", int(scrapingTime.Minutes()), int(scrapingTime.Seconds())%60), coursesScraped)
		r, err := http.NewRequest("POST", os.Getenv("SCRAPER_WEBHOOK_URL"), bytes.NewBuffer([]byte(params)))
		if err != nil {
			panic(err)
		}
		r.Header.Add("Content-Type", "application/json")
		client := &http.Client{}
		res, err := client.Do(r)
		if err != nil {
			panic(err)
		}
		defer res.Body.Close()
	}
}

func Entrypoint(db *gorm.DB, webhookURL string, scrapeAll bool, rdb *redis.Client) {
	c := cron.New()

	Scrape(db, webhookURL, scrapeAll, rdb)

	c.AddFunc("30 4 * * 1", func() { Scrape(db, webhookURL, scrapeAll, rdb) })
	c.Start()
}

package scrapers

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
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
	jsession, err := scrapers.GetJsession()
	if err != nil {
		log.Fatalln(err)
	}

	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}
	baseUrl, _ := url.Parse("https://self-service.mun.ca")
	jar.SetCookies(baseUrl, []*http.Cookie{
		{Name: "JSESSIONID", Value: jsession, Path: "/StudentRegistrationSsb"},
	})

	startTime := time.Now()
	coursesScraped := 0
	logger := log.Default()
	ctx := context.Background()

	logger.Println("⭐ Scraping Started!")

	// var profs []string

	semesters, err := scrapers.GetSemesters()
	if err != nil {
		logger.Fatal(err)
	}

	for _, semester := range semesters {
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
		subjects, err := scrapers.GetSubjects(semester)
		if err != nil {
			logger.Fatal(err)
		}
		for _, subject := range subjects {
			db.Save(&subject)
		}
		courses, err := scrapers.GetCourses(client, semester)
		if err != nil {
			logger.Fatal(err)
		}
		for _, course := range courses {
			db.Create(&course)
			coursesScraped++
			meetingTimes, instructors, err := scrapers.GetMeetingTimes(semester, course.CRN)
			if err != nil {
				logger.Println(err)
				continue
			}
			for _, meetingTime := range meetingTimes {
				db.Create(&meetingTime)
			}
			for _, instructor := range instructors {
				db.Save(&util.Professor{Name: instructor})
				db.Save(&util.CourseInstructor{
					ProfessorName: instructor,
					CourseKey:     course.Key,
				})
			}
		}
		// logger.Println(" 📝 Processing Exams")
		// for _, exam := range scrapers.GetExams(semester.ID) {
		// 	db.Save(&exam)
		// }
		rdb.Del(ctx, "frontend:"+strconv.Itoa(semester.ID))
	}

	// logger.Println("⭐ RMP Scraping Started!")

	// profRatings := scrapers.RMP(logger, profs)
	// for _, rating := range profRatings {
	// 	db.Save(&rating)
	// }

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

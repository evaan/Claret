package scrapers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/evaan/Claret/internal/scrapers"
	"github.com/evaan/Claret/internal/util"
	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func Scrape(db *gorm.DB, webhookUrl string, scrapeAll bool, rdb *redis.Client) {
	hour := time.Now().Hour()
	if hour == 3 || hour == 4 {
		return
	}

	jsession, err := scrapers.GetJsession()
	if err != nil {
		log.Printf("Unable to start catalog scrape: %v", err)
		return
	}

	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar, Timeout: 30 * time.Second}
	baseUrl, _ := url.Parse("https://self-service.mun.ca")
	jar.SetCookies(baseUrl, []*http.Cookie{
		{Name: "JSESSIONID", Value: jsession, Path: "/StudentRegistrationSsb"},
	})

	startTime := time.Now()
	coursesScraped := 0
	logger := log.Default()
	ctx := context.Background()

	logger.Println("⭐ Scraping Started!")

	semesters, err := scrapers.GetSemesters()
	if err != nil {
		logger.Printf("Unable to fetch semesters: %v", err)
		return
	}

	for _, semester := range semesters {
		if semester.ViewOnly && !scrapeAll {
			continue
		}
		count, err := scrapeSemester(db, client, semester)
		if err != nil {
			logger.Printf("Keeping last good catalog for %d: %v", semester.ID, err)
			continue
		}
		coursesScraped += count
		cacheCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		if err := rdb.Del(cacheCtx, "frontend:v2:"+strconv.Itoa(semester.ID)).Err(); err != nil {
			logger.Printf("Unable to invalidate catalog cache for %d: %v", semester.ID, err)
		}
		cancel()
	}

	logger.Println("⭐ RMP Scraping Started!")

	var profs []string
	if err := db.Model(&util.Professor{}).Pluck("name", &profs).Error; err != nil {
		logger.Printf("Unable to load instructor names for RMP: %v", err)
	}
	profRatings := scrapers.RMP(logger, profs)
	for _, rating := range profRatings {
		if err := db.Omit("Professor").Save(&rating).Error; err != nil {
			logger.Printf("Unable to save RMP rating for %q: %v", rating.ProfessorName, err)
		}
	}

	scrapingTime := time.Since(startTime)

	logger.Println("✅ Scrape Completed in " + fmt.Sprintf("%02d:%02d", int(scrapingTime.Minutes()), int(scrapingTime.Seconds())%60) + "!")
	logger.Printf("🚀 Courses scraped: %d", coursesScraped)

	if webhookUrl != "" {
		logger.Println("🔔 Sending message to Discord")
		if err := util.SendWebhook(webhookUrl, map[string]string{
			"username": "Claret Scraper",
			"content":  fmt.Sprintf("Scrape completed in %s; %d courses updated.", scrapingTime.Round(time.Second), coursesScraped),
		}); err != nil {
			logger.Printf("Unable to send scraper report: %v", err)
		}
	}
}

// Only persist snapshots whose course already exists; retain newer snapshots.
func saveSeatSnapshots(db *gorm.DB, semester int, seatings []util.CourseSeating) error {
	var keys []string
	if err := db.Model(&util.Course{}).Where("semester_id = ?", semester).Pluck("key", &keys).Error; err != nil {
		return err
	}
	known := make(map[string]bool, len(keys))
	for _, key := range keys {
		known[key] = true
	}
	snapshots := make([]util.CourseSeating, 0, len(seatings))
	for _, seating := range seatings {
		if known[seating.CourseKey] {
			snapshots = append(snapshots, seating)
		}
	}
	if len(snapshots) == 0 {
		return nil
	}
	return db.Omit("Course").Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "semester"}, {Name: "crn"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"course_key", "seats", "max_seats", "waitlist", "max_waitlist", "seats_available",
			"wait_count", "open_section", "cross_list_capacity", "cross_list_count", "cross_list_available", "updated_at",
		}),
		// A full scrape may finish after a newer, ten-minute seat refresh.
		Where: clause.Where{Exprs: []clause.Expression{
			clause.Expr{SQL: "EXCLUDED.updated_at >= course_seatings.updated_at"},
		}},
	}).CreateInBatches(&snapshots, 500).Error
}

func ScrapeSeats(db *gorm.DB, rdb *redis.Client) {
	hour := time.Now().Hour()
	if hour == 3 || hour == 4 {
		return
	}

	logger := log.Default()
	ctx := context.Background()
	startTime := time.Now()

	logger.Println("🪑 Seat Scraping Started!")

	jsession, err := scrapers.GetJsession()
	if err != nil {
		logger.Println(err)
		return
	}

	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar, Timeout: 30 * time.Second}
	baseUrl, _ := url.Parse("https://self-service.mun.ca")
	jar.SetCookies(baseUrl, []*http.Cookie{
		{Name: "JSESSIONID", Value: jsession, Path: "/StudentRegistrationSsb"},
	})

	semesters, err := scrapers.GetSemesters()
	if err != nil {
		logger.Printf("Unable to fetch semesters for seat refresh: %v", err)
		return
	}

	for _, semester := range semesters {
		if !semester.ViewOnly {
			_, seatings, err := scrapers.GetCourses(client, semester)
			if err != nil {
				logger.Println(err)
				continue
			}
			dbCtx, dbCancel := context.WithTimeout(ctx, 30*time.Second)
			err = saveSeatSnapshots(db.WithContext(dbCtx), semester.ID, seatings)
			dbCancel()
			if err != nil {
				logger.Println("Unable to persist seat snapshots: ", err)
				continue
			}
			// Pipeline cache writes so a large catalog doesn't require one
			// Redis round trip per course.
			cacheCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			_, err = rdb.Pipelined(cacheCtx, func(pipe redis.Pipeliner) error {
				for _, seating := range seatings {
					seatingJson, err := json.Marshal(seating)
					if err != nil {
						return err
					}
					pipe.Set(cacheCtx, "seats:"+strconv.Itoa(seating.Semester)+":"+seating.CRN, seatingJson, 10*time.Minute)
				}
				return nil
			})
			cancel()
			if err != nil {
				logger.Printf("Unable to cache seats: %v", err)
			}
		}
	}

	scrapingTime := time.Since(startTime)
	logger.Println("🪑 Seat Scraping Completed in " + fmt.Sprintf("%02d:%02d", int(scrapingTime.Minutes()), int(scrapingTime.Seconds())%60) + "!")
}

func Entrypoint(db *gorm.DB, webhookURL string, scrapeAll bool, rdb *redis.Client) {
	// Share a lock across startup and both cron jobs; never pile up scrapes.
	var running sync.Mutex
	guard := func(job func(), wait bool) func() {
		return func() {
			if wait {
				// Don't lose a weekly catalog run when seats start at :30 too.
				running.Lock()
			} else if !running.TryLock() {
				log.Println("Skipping scraper job: another scrape is still running")
				return
			}
			defer running.Unlock()
			defer func() {
				if recovered := recover(); recovered != nil {
					log.Printf("Scraper job failed: %v", recovered)
				}
			}()
			job()
		}
	}
	full := guard(func() { Scrape(db, webhookURL, scrapeAll, rdb) }, true)
	seats := guard(func() { ScrapeSeats(db, rdb) }, false)
	c := cron.New()
	if _, err := c.AddFunc("*/10 * * * *", seats); err != nil {
		log.Printf("Unable to schedule seats: %v", err)
		return
	}
	// Banner maintenance blocks 03:00–04:59; the old 04:30 job always skipped.
	if _, err := c.AddFunc("30 5 * * 1", full); err != nil {
		log.Printf("Unable to schedule catalog: %v", err)
		return
	}
	c.Start()
	go full()
}

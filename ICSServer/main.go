package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Global
var db *sql.DB
var logger *log.Logger
var banner_tz *time.Location
var day_of_week_map map[string]time.Weekday

func main() {
	// Setup Logger
	logger = log.Default()

	logger.Println("👋 Claret ICS Server")

	day_of_week_map = map[string]time.Weekday{
		"M": time.Monday,
		"T": time.Tuesday,
		"W": time.Wednesday,
		"R": time.Thursday,
		"F": time.Friday,
		"S": time.Saturday,
		"U": time.Sunday,
	}

	// Load Configuration from Enviroment Variables
	DB_URL := os.Getenv("DB_URL")
	if DB_URL == "" {
		logger.Fatal("No DB_URL in Enviroment Variables")
	}

	PORT := os.Getenv("PORT")
	if PORT == "" {
		logger.Fatal("No PORT in Enviroment Variables")
	}

	BANNER_IANA_TZ := os.Getenv("BANNER_TZ")
	if BANNER_IANA_TZ == "" {
		BANNER_IANA_TZ := os.Getenv("TZ")
		if BANNER_IANA_TZ == "" {
			logger.Fatal("No TZ or BANNER_TZ in Enviroment Variables")
		}
	}

	var err error
	banner_tz, err = time.LoadLocation(BANNER_IANA_TZ)
	if err != nil {
		logger.Fatal(err)
	}
	logger.Println("🕒 Banner Time Zone:", banner_tz.String())

	// Connect to Database
	db, err = sql.Open("pgx", DB_URL) // db is global, maybe change?
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Close()

	// Check Connection to Database
	pingErr := db.Ping()
	if pingErr != nil {
		logger.Fatal(pingErr)
	}
	logger.Println("💿 Connected to Database!")

	// Register HTTP Handlers
	http.HandleFunc("/health", health)
	http.HandleFunc("/feed.ics", ics)

	// Start HTTP Server
	logger.Println("✅ ICS Server running server on port", PORT)
	http.ListenAndServe(":"+PORT, nil)
}

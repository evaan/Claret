package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/joho/godotenv/autoload"
)

var db *sql.DB
var logger *log.Logger
var err error
var loc *time.Location

func main() {
	logger = log.Default()
	logger.Println("👋 Claret API")

	DB_URL := os.Getenv("DB_URL")
	if DB_URL == "" {
		logger.Fatal("DB_URL is not defined in environment variables")
	}

	PORT := os.Getenv("PORT")
	if PORT == "" {
		logger.Fatal("PORT is not defined in environment variables")
	}

	db, err = sql.Open("pgx", DB_URL)
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		logger.Fatal(err)
	}
	logger.Println("💿 Connected to Database!")

	loc, err = time.LoadLocation("America/St_Johns")
	if err != nil {
		logger.Fatal(err)
	}

	http.HandleFunc("/", index)
	http.HandleFunc("/all", all)
	http.HandleFunc("/subjects", subjects)
	http.HandleFunc("/semesters", semesters)
	http.HandleFunc("/courses", courses)
	http.HandleFunc("/times", times)
	http.HandleFunc("/seating", seating)
	http.HandleFunc("/rmp", rmp)
	http.HandleFunc("/exams", exams)
	http.HandleFunc("/engi", engi)

	logger.Println("✅ API running server on port", PORT)
	http.ListenAndServe(":"+PORT, nil)
}

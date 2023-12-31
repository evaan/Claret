package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var db *sql.DB
var logger *log.Logger
var err error

type Course struct {
	crn        string
	id         string
	name       string
	section    string
	dateRange  any
	courseType any
	instructor any
	subject    string
	campus     string
	comment    any
	credits    int
	semester   int
}

type Time struct {
	crn       string
	days      string
	startTime string
	endTime   string
	location  string
}

func all(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT * FROM courses")
	if err != nil {
		logger.Fatal(err)
	}
	defer rows.Close()

	output := make(map[string][]any)

	for rows.Next() {
		var course Course
		tmp := make(map[string]any)

		err := rows.Scan(&course.crn, &course.id, &course.name, &course.section, &course.dateRange, &course.courseType, &course.instructor, &course.subject, &course.campus, &course.comment, &course.credits, &course.semester)
		if err != nil {
			logger.Fatal(err)
		}

		tmp["crn"] = course.crn
		tmp["id"] = course.id
		tmp["name"] = course.name
		tmp["section"] = course.section
		tmp["dateRange"] = course.dateRange
		tmp["type"] = course.courseType
		tmp["instructor"] = course.instructor
		tmp["subject"] = course.subject
		tmp["campus"] = course.campus
		tmp["comment"] = course.comment
		tmp["credits"] = course.credits
		tmp["semester"] = course.semester

		output["courses"] = append(output["courses"], tmp)
	}

	rows, err = db.Query("SELECT times.crn, times.days, times.\"startTime\", times.\"endTime\", times.location FROM times")
	if err != nil {
		logger.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var time Time
		tmp := make(map[string]any)

		err := rows.Scan(&time.crn, &time.days, &time.startTime, &time.endTime, &time.location)
		if err != nil {
			logger.Fatal(err)
		}

		tmp["crn"] = time.crn
		tmp["days"] = time.days
		tmp["startTime"] = time.startTime
		tmp["endTime"] = time.endTime
		tmp["location"] = time.location

		output["times"] = append(output["times"], tmp)
	}

	jsonString, err := json.Marshal(output)
	if err != nil {
		logger.Fatal(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(jsonString))
}

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

	http.HandleFunc("/all", all)

	logger.Println("✅ API running server on port", PORT)
	http.ListenAndServe(":"+PORT, nil)
}

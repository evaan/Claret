package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/exec"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var db *sql.DB
var logger *log.Logger
var err error

type Subject struct {
	Name         string `json:"name"`
	FriendlyName string `json:"friendlyName"`
}

type Course struct {
	Crn        string `json:"crn"`
	Id         string `json:"id"`
	Name       string `json:"name"`
	Section    string `json:"section"`
	DateRange  any    `json:"dateRange"`
	CourseType any    `json:"type"`
	Instructor any    `json:"instructor"`
	Subject    string `json:"subject"`
	Campus     string `json:"campus"`
	Comment    any    `json:"comment"`
	Credits    int    `json:"credits"`
	Semester   int    `json:"semester"`
}

type Time struct {
	Crn       string `json:"crn"`
	Days      string `json:"days"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
	Location  string `json:"location"`
}

type Seating struct {
	Crn       string `json:"crn"`
	Available string `json:"available"`
	Max       string `json:"max"`
	Waitlist  string `json:"waitlist"`
	Checked   string `json:"checked"`
}

func all(w http.ResponseWriter, r *http.Request) {
	output := make(map[string][]any)

	subjects, err := db.Query("SELECT * FROM subjects")
	if err != nil {
		logger.Fatal(err)
	}
	defer subjects.Close()

	for subjects.Next() {
		var subject Subject

		err := subjects.Scan(&subject.Name, &subject.FriendlyName)
		if err != nil {
			logger.Fatal(err)
		}

		output["subjects"] = append(output["subjects"], subject)
	}

	courses, err := db.Query("SELECT * FROM courses")
	if err != nil {
		logger.Fatal(err)
	}
	defer courses.Close()

	for courses.Next() {
		var course Course

		err := courses.Scan(&course.Crn, &course.Id, &course.Name, &course.Section, &course.DateRange, &course.CourseType, &course.Instructor, &course.Subject, &course.Campus, &course.Comment, &course.Credits, &course.Semester)
		if err != nil {
			logger.Fatal(err)
		}

		output["courses"] = append(output["courses"], course)
	}

	times, err := db.Query("SELECT times.crn, times.days, times.\"startTime\", times.\"endTime\", times.location FROM times")
	if err != nil {
		logger.Fatal(err)
	}
	defer times.Close()

	for times.Next() {
		var time Time

		err := times.Scan(&time.Crn, &time.Days, &time.StartTime, &time.EndTime, &time.Location)
		if err != nil {
			logger.Fatal(err)
		}

		output["times"] = append(output["times"], time)
	}

	seatings, err := db.Query("SELECT * FROM seatings")
	if err != nil {
		logger.Fatal(err)
	}
	defer seatings.Close()

	for seatings.Next() {
		var seating Seating

		err := seatings.Scan(&seating.Crn, &seating.Available, &seating.Max, &seating.Waitlist, &seating.Checked)
		if err != nil {
			logger.Fatal(err)
		}

		output["seatings"] = append(output["seatings"], seating)
	}

	jsonString, err := json.Marshal(output)
	if err != nil {
		logger.Fatal(err)
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(jsonString))
}

func subjects(w http.ResponseWriter, r *http.Request) {
	var output []Subject

	subjects, err := db.Query("SELECT * FROM subjects")
	if err != nil {
		logger.Fatal(err)
	}
	defer subjects.Close()

	for subjects.Next() {
		var subject Subject

		err := subjects.Scan(&subject.Name, &subject.FriendlyName)
		if err != nil {
			logger.Fatal(err)
		}

		output = append(output, subject)
	}

	jsonString, err := json.Marshal(output)
	if err != nil {
		logger.Fatal(err)
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(jsonString))
}

func courses(w http.ResponseWriter, r *http.Request) {
	var output []Course

	courses, err := db.Query("SELECT * FROM courses WHERE courses.crn LIKE $1", "%"+r.URL.Query().Get("crn")+"%")
	if err != nil {
		logger.Fatal(err)
	}
	defer courses.Close()

	for courses.Next() {
		var course Course

		err := courses.Scan(&course.Crn, &course.Id, &course.Name, &course.Section, &course.DateRange, &course.CourseType, &course.Instructor, &course.Subject, &course.Campus, &course.Comment, &course.Credits, &course.Semester)
		if err != nil {
			logger.Fatal(err)
		}

		output = append(output, course)
	}

	course, err := json.Marshal(output)
	if err != nil {
		logger.Fatal(err)
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(course))
}

func times(w http.ResponseWriter, r *http.Request) {
	var output []Time

	times, err := db.Query("SELECT times.crn, times.days, times.\"startTime\", times.\"endTime\", times.location FROM times WHERE times.crn = $1", r.URL.Query().Get("crn"))
	if err != nil {
		logger.Fatal(err)
	}
	defer times.Close()

	for times.Next() {
		var time Time

		err := times.Scan(&time.Crn, &time.Days, &time.StartTime, &time.EndTime, &time.Location)
		if err != nil {
			logger.Fatal(err)
		}

		output = append(output, time)
	}

	course, err := json.Marshal(output)
	if err != nil {
		logger.Fatal(err)
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(course))
}

func seating(w http.ResponseWriter, r *http.Request) {
	var output []Seating

	cmd := exec.Command("python3", "../Scraper/SeatingScrape.py", r.URL.Query().Get("crn"), r.URL.Query().Get("semester"))
	cmd.Run()

	seatings, err := db.Query("SELECT * FROM seatings WHERE seatings.crn = $1", r.URL.Query().Get("crn"))
	if err != nil {
		logger.Fatal(err)
	}

	for seatings.Next() {
		var seating Seating

		err := seatings.Scan(&seating.Crn, &seating.Available, &seating.Max, &seating.Waitlist, &seating.Checked)
		if err != nil {
			logger.Fatal(err)
		}

		output = append(output, seating)
	}

	jsonString, err := json.Marshal(output)
	if err != nil {
		logger.Fatal(err)
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
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
	http.HandleFunc("/subjects", subjects)
	http.HandleFunc("/courses", courses)
	http.HandleFunc("/times", times)
	http.HandleFunc("/seating", seating)

	logger.Println("✅ API running server on port", PORT)
	http.ListenAndServe(":"+PORT, nil)
}

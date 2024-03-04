package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/joho/godotenv/autoload"
	"github.com/lestrrat-go/strftime"
)

var db *sql.DB
var logger *log.Logger
var err error
var loc *time.Location

//TODO MAKE IT SO WHEN THERES AN ERROR IT DOESNT CAUSE NULL POINTER

type Semester struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Latest   bool   `json:"latest"`
	ViewOnly bool   `json:"viewOnly"`
	Medical  bool   `json:"medical"`
	MI       bool   `json:"mi"`
}

type Subject struct {
	Name         string `json:"name"`
	FriendlyName string `json:"friendlyName"`
}

type Course struct {
	Crn         string `json:"crn"`
	Id          string `json:"id"`
	Name        string `json:"name"`
	Section     string `json:"section"`
	DateRange   any    `json:"dateRange"`
	CourseType  any    `json:"type"`
	Instructor  any    `json:"instructor"`
	SubjectFull string `json:"subjectFull"`
	Subject     string `json:"subject"`
	Campus      string `json:"campus"`
	Comment     any    `json:"comment"`
	Credits     int    `json:"credits"`
	Semester    int    `json:"semester"`
	Level       string `json:"level"`
	Identifier  string `json:"identifier"`
}

type Time struct {
	Identifier string `json:"identifier"`
	Crn        string `json:"crn"`
	Days       string `json:"days"`
	StartTime  string `json:"startTime"`
	EndTime    string `json:"endTime"`
	Location   string `json:"location"`
	Type       string `json:"courseType"`
	Semester   int    `json:"semester"`
}

type Seating struct {
	Crn        string `json:"crn"`
	Available  string `json:"available"`
	Max        string `json:"max"`
	Waitlist   any    `json:"waitlist"`
	Checked    string `json:"checked"`
	Identifier string `json:"identifier"`
	Semester   int    `json:"semester"`
}

func all(w http.ResponseWriter, r *http.Request) {
	output := make(map[string][]any)

	var subjects *sql.Rows
	var err error

	if r.URL.Query().Get("semester") != "" {
		subjects, err = db.Query("SELECT DISTINCT subject, \"subjectFull\" FROM courses WHERE semester = $1", r.URL.Query().Get("semester"))
	} else {
		subjects, err = db.Query("SELECT DISTINCT subject, \"subjectFull\" FROM courses")
	}
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	defer subjects.Close()

	for subjects.Next() {
		var subject Subject

		err := subjects.Scan(&subject.Name, &subject.FriendlyName)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		output["subjects"] = append(output["subjects"], subject)
	}

	var courses *sql.Rows

	if r.URL.Query().Get("semester") != "" {
		courses, err = db.Query("SELECT * FROM courses WHERE semester = $1", r.URL.Query().Get("semester"))
	} else {
		courses, err = db.Query("SELECT * FROM courses")
	}

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	defer courses.Close()

	for courses.Next() {
		var course Course

		err := courses.Scan(&course.Crn, &course.Id, &course.Name, &course.Section, &course.DateRange, &course.CourseType, &course.Instructor, &course.Subject, &course.SubjectFull, &course.Campus, &course.Comment, &course.Credits, &course.Semester, &course.Level, &course.Identifier)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		output["courses"] = append(output["courses"], course)
	}

	var times *sql.Rows

	if r.URL.Query().Get("semester") != "" {
		times, err = db.Query("SELECT times.crn, times.days, times.\"startTime\", times.\"endTime\", times.location, times.type, times.identifier FROM times WHERE semester = $1", r.URL.Query().Get("semester"))
	} else {
		times, err = db.Query("SELECT times.crn, times.days, times.\"startTime\", times.\"endTime\", times.location, times.type, times.identifier FROM times")
	}

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	defer times.Close()

	for times.Next() {
		var time Time

		err := times.Scan(&time.Crn, &time.Days, &time.StartTime, &time.EndTime, &time.Location, &time.Type, &time.Identifier)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		output["times"] = append(output["times"], time)
	}

	var seatings *sql.Rows

	if r.URL.Query().Get("semester") != "" {
		seatings, err = db.Query("SELECT * FROM seatings WHERE semester = $1", r.URL.Query().Get("semester"))
	} else {
		seatings, err = db.Query("SELECT * FROM seatings")
	}

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	defer seatings.Close()

	for seatings.Next() {
		var seating Seating

		err := seatings.Scan(&seating.Identifier, &seating.Crn, &seating.Available, &seating.Max, &seating.Waitlist, &seating.Checked, &seating.Semester)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
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

	var subjects *sql.Rows
	var err error

	if r.URL.Query().Get("semester") != "" {
		subjects, err = db.Query("SELECT DISTINCT subject, \"subjectFull\" FROM courses WHERE semester = $1", r.URL.Query().Get("semester"))
	} else {
		subjects, err = db.Query("SELECT DISTINCT subject, \"subjectFull\" FROM courses")
	}

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	defer subjects.Close()

	for subjects.Next() {
		var subject Subject

		err := subjects.Scan(&subject.Name, &subject.FriendlyName)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		output = append(output, subject)
	}

	jsonString, err := json.Marshal(output)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(jsonString))
}

func semesters(w http.ResponseWriter, _ *http.Request) {
	var output []Semester

	semesters, err := db.Query("SELECT id, name, latest, \"viewOnly\", medical, mi FROM semesters")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	defer semesters.Close()

	for semesters.Next() {
		var semester Semester

		err := semesters.Scan(&semester.ID, &semester.Name, &semester.Latest, &semester.ViewOnly, &semester.Medical, &semester.MI)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		output = append(output, semester)
	}

	jsonString, err := json.Marshal(output)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(jsonString))
}

func courses(w http.ResponseWriter, r *http.Request) {
	var output []Course

	if r.URL.Query().Get("semester") == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("CRN was not provided, please add ?semester={semester} in your URL."))
		return
	}

	courses, err := db.Query("SELECT * FROM courses WHERE courses.crn LIKE $1", "%"+r.URL.Query().Get("crn")+"%")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	defer courses.Close()

	for courses.Next() {
		var course Course

		err := courses.Scan(&course.Crn, &course.Id, &course.Name, &course.Section, &course.DateRange, &course.CourseType, &course.Instructor, &course.Subject, &course.SubjectFull, &course.Campus, &course.Comment, &course.Credits, &course.Semester, &course.Level, &course.Identifier)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		output = append(output, course)
	}

	course, err := json.Marshal(output)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(course))
}

func times(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var output []Time

	if r.URL.Query().Get("crn") == "" || r.URL.Query().Get("semester") == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("CRN was not provided, please add ?crn={crn}&semester={semester} in your URL."))
		return
	}

	times, err := db.Query("SELECT times.crn, times.days, times.\"startTime\", times.\"endTime\", times.location, times.type, times.identifier FROM times WHERE times.crn = $1 AND times.semester = $2", r.URL.Query().Get("crn"), r.URL.Query().Get("semester"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	defer times.Close()

	for times.Next() {
		var time Time

		err := times.Scan(&time.Crn, &time.Days, &time.StartTime, &time.EndTime, &time.Location, &time.Type, &time.Identifier)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		output = append(output, time)
	}

	course, err := json.Marshal(output)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(course))
}

func seating(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.URL.Query().Get("crn") == "" || r.URL.Query().Get("semester") == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("CRN was not provided, please add ?crn={crn}&semester={semester} in your URL."))
		return
	}

	var checked string
	var jsonString []byte

	var semester string
	err := db.QueryRow("SELECT courses.semester FROM courses WHERE courses.identifier = $1", r.URL.Query().Get("semester")+r.URL.Query().Get("crn")).Scan(&semester)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Course could not be found, double-check your CRN and try again." + r.URL.Query().Get("semester") + r.URL.Query().Get("crn")))
		return
	}

	exists := true

	time1, err := time.ParseInLocation("2006-01-02T15:04", checked, loc)
	if err != nil {
		time1 = time.Now().Add(time.Duration(-6) * time.Minute)
	}

	if !time1.After(time.Now().Add(-5 * time.Minute)) {
		c := colly.NewCollector()

		var cells []string

		if exists {
			c.OnHTML("caption", func(e *colly.HTMLElement) {
				if e.Text == "Registration Availability" {
					e.DOM.Parent().Find("td.dddefault").Each(func(i int, s *goquery.Selection) {
						cells = append(cells, s.Text())
					})
				}
			})

			c.OnHTML("span.errortext", func(e *colly.HTMLElement) {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Course could not be found, double-check your CRN and try again."))
				exists = false
			})

			c.Visit("https://selfservice.mun.ca/direct/bwckschd.p_disp_detail_sched?term_in=" + semester + "&crn_in=" + r.URL.Query().Get("crn"))
			c.Wait()

			if !exists {
				return
			}

			var output []Seating
			var seating Seating

			seating.Crn = r.URL.Query().Get("crn")
			seating.Identifier = r.URL.Query().Get("semester") + r.URL.Query().Get("crn")
			if len(cells) != 0 {
				seating.Available = cells[2]
				seating.Max = cells[0]
				if len(cells) >= 6 {
					seating.Waitlist = cells[4]
				} else {
					seating.Waitlist = nil
				}
			}
			checkedTime, err := strftime.Format("%Y-%m-%dT%H:%M", time.Now())
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(err.Error()))
				return
			}
			seating.Checked = checkedTime

			jsonString, err = json.Marshal(append(output, seating))
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(err.Error()))
				return
			}

			_, err = db.Exec(`UPDATE seatings
			SET available = $2, max = $3, waitlist = $4, checked = $5
			WHERE identifier = $1;`, seating.Identifier, seating.Available, seating.Max, seating.Waitlist, seating.Checked)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(err.Error()))
				return
			}
		}
	} else {
		var seating Seating
		var output []Seating

		err := db.QueryRow("SELECT * FROM seatings WHERE seatings.identifier = $1", r.URL.Query().Get("semester")+r.URL.Query().Get("crn")).Scan(&seating.Crn, &seating.Available, &seating.Max, &seating.Waitlist, &seating.Checked, &seating.Identifier)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		jsonString, err = json.Marshal(append(output, seating))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
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

	loc, err = time.LoadLocation("America/St_Johns")
	if err != nil {
		logger.Fatal(err)
	}

	http.HandleFunc("/all", all)
	http.HandleFunc("/subjects", subjects)
	http.HandleFunc("/semesters", semesters)
	http.HandleFunc("/courses", courses)
	http.HandleFunc("/times", times)
	http.HandleFunc("/seating", seating)

	logger.Println("✅ API running server on port", PORT)
	http.ListenAndServe(":"+PORT, nil)
}

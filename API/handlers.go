package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"github.com/lestrrat-go/strftime"
)

//TODO probably isnt an awful idea to further split this up

func all(w http.ResponseWriter, r *http.Request) {
	output := make(map[string][]any)

	var subjects *sql.Rows
	var err error

	//TODO: maybe make it cache at one point through cloudflare tasks or something?
	if r.URL.Query().Get("semester") == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Semester was not provided, please add ?semester={semester} in your URL."))
		return
	}

	subjects, err = db.Query("SELECT DISTINCT subject, \"subjectFull\" FROM courses WHERE semester = $1", r.URL.Query().Get("semester"))
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
		courses, err = db.Query("SELECT crn, id, name, section, \"dateRange\", type, instructor, subject, \"subjectFull\", campus, comment, credits, semester, level, identifier FROM courses WHERE semester = $1", r.URL.Query().Get("semester"))
	} else {
		courses, err = db.Query("SELECT crn, id, name, section, \"dateRange\", type, instructor, subject, \"subjectFull\", campus, comment, credits, semester, level, identifier FROM courses")
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
		seatings, err = db.Query("SELECT identifier, crn, available, max, waitlist, checked, semester FROM seatings WHERE semester = $1", r.URL.Query().Get("semester"))
	} else {
		seatings, err = db.Query("SELECT identifier, crn, available, max, waitlist, checked, semester FROM seatings")
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

	profs, err := db.Query("SELECT DISTINCT p.name, p.rating, p.id, p.difficulty, p.rating_count, p.would_retake FROM professors p JOIN prof_and_semesters ps ON p.name = ps.name AND ps.semester = $1", r.URL.Query().Get("semester"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	defer profs.Close()

	for profs.Next() {
		var prof Professor

		err := profs.Scan(&prof.Name, &prof.Rating, &prof.Id, &prof.Difficulty, &prof.RatingCount, &prof.WouldRetake)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		output["profs"] = append(output["profs"], prof)

	}

	if r.URL.Query().Get("semester") == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Semester was not provided, please add ?semester={semester} in your URL."))
		return
	}

	examTimes, err := db.Query("SELECT DISTINCT e.crn, e.location, e.time, c.id, c.section FROM exam_times e JOIN courses c ON c.crn = e.crn AND c.semester = e.semester WHERE e.semester = $1", r.URL.Query().Get("semester"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	defer examTimes.Close()

	for examTimes.Next() {
		var examTime ExamTime

		err := examTimes.Scan(&examTime.Crn, &examTime.Location, &examTime.Time, &examTime.Id, &examTime.Section)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		output["exams"] = append(output["exams"], examTime)
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
		w.Write([]byte("Semester was not provided, please add ?semester={semester} in your URL."))
		return
	}

	courses, err := db.Query("SELECT crn, id, name, section, \"dateRange\", type, instructor, subject, \"subjectFull\", campus, comment, credits, semester, level, identifier FROM courses WHERE courses.semester = $1 AND courses.crn LIKE $2", r.URL.Query().Get("semester"), "%"+r.URL.Query().Get("id")+"%")
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

func rmp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var output []Professor

	profs, err := db.Query("SELECT name, rating, id, difficulty, rating_count, would_retake FROM professors WHERE LOWER(name) LIKE LOWER($1)", "%"+r.URL.Query().Get("name")+"%")

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	defer profs.Close()

	for profs.Next() {
		var prof Professor

		err := profs.Scan(&prof.Name, &prof.Rating, &prof.Id, &prof.Difficulty, &prof.RatingCount, &prof.WouldRetake)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		output = append(output, prof)
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
			checkedTime, err := strftime.Format("%Y-%m-%dT%H:%M", time.Now().UTC())
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

		err := db.QueryRow("SELECT identifier, crn, available, max, waitlist, checked, semester FROM seatings WHERE seatings.identifier = $1", r.URL.Query().Get("semester")+r.URL.Query().Get("crn")).Scan(&seating.Crn, &seating.Available, &seating.Max, &seating.Waitlist, &seating.Checked, &seating.Identifier)
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

func exams(w http.ResponseWriter, r *http.Request) {
	var output []ExamTime

	if r.URL.Query().Get("semester") == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Semester was not provided, please add ?semester={semester} in your URL."))
		return
	}

	examTimes, err := db.Query("SELECT DISTINCT e.crn, e.location, e.time, c.id, c.section FROM exam_times e JOIN courses c ON c.crn = e.crn AND c.semester = e.semester WHERE e.semester = $1 AND ($2 = '' OR c.crn = $2)", r.URL.Query().Get("semester"), r.URL.Query().Get("crn"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	defer examTimes.Close()

	for examTimes.Next() {
		var examTime ExamTime

		err := examTimes.Scan(&examTime.Crn, &examTime.Location, &examTime.Time, &examTime.Id, &examTime.Section)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		output = append(output, examTime)
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

func index(w http.ResponseWriter, _ *http.Request) {
	w.Write([]byte("<p>values in square brackets are optional, while braces are mandatory</p><p><strong>/all?semester=[semester]</strong> - get <strong>all</strong> data of a semester, or if no semester is provided return <strong>every</strong> semester combined (will be >60MB of raw JSON)</p><p><strong>/subjects?semester=[semester]</strong> - return a list of all subjects from a semester, or all semesters if none is provided</p><p><strong>/semesters</strong> - return a list of all semesters</p><p><strong>/courses?semester={semester}&id=[id]</strong> - return a list of all courses from a semester that contains crn, if no crn is provided it will return all courses</p><p><strong>/times?semester={semester}&crn={crn}</strong> - return a list of all times for a certain course slot</p><p><strong>/seating?semester={semester}&crn={crn}</strong> - scrapes muns course offering for seatings, then returns them</p><p><strong>/rmp?name=[name]</strong> - returns all mun rate my prof ratings, or a search for a specific name</p><p><strong>/exams?semester={semester}</strong> - returns all final exams from specified semester"))
}

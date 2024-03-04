package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	_ "github.com/joho/godotenv/autoload"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Semester struct {
	ID       int    `gorm:"not null"`
	Name     string `gorm:"not null"`
	Latest   bool   `gorm:"not null"`
	ViewOnly bool   `gorm:"column:viewOnly;not null"`
	Medical  bool   `gorm:"not null"`
	MI       bool   `gorm:"not null"`
}

type Subject struct {
	Name         string `gorm:"primaryKey;not null"`
	FriendlyName string `gorm:"not null"`
}

type Course struct {
	CRN         string  `gorm:"primaryKey"`
	Id          string  `gorm:"not null"`
	Name        string  `gorm:"not null"`
	Section     string  `gorm:"not null"`
	DateRange   *string `gorm:"column:dateRange"`
	Type        *string
	Instructor  *string
	Subject     string `gorm:"column:subject;not null"`
	SubjectFull string `gorm:"column:subjectFull;not null"`
	Campus      string `gorm:"not null"`
	Comment     *string
	Credits     int      `gorm:"not null"`
	SemesterID  int      `gorm:"column:semester;not null"`
	Semester    Semester `gorm:"constraint:OnDelete:CASCADE;"`
	Level       string   `gorm:"not null"`
}

type CourseTime struct {
	ID        int    `gorm:"primaryKey;autoIncrement"`
	CourseCRN string `gorm:"column:crn"`
	Course    Course `gorm:"constraint:OnDelete:CASCADE;"`
	Days      string `gorm:"not null"`
	StartTime string `gorm:"column:startTime;not null"`
	EndTime   string `gorm:"column:endTime;not null"`
	Location  string `gorm:"not null"`
	Type      string `gorm:"not null"`
	Semester  int    `gorm:"not null"`
}

func (CourseTime) TableName() string {
	return "times"
}

type Seating struct {
	Crn       string `gorm:"primaryKey"`
	Available int    `gorm:"not null"`
	Max       int    `gorm:"not null"`
	Waitlist  int
	Checked   string `gorm:"not null"`
}

var db *gorm.DB
var logger *log.Logger
var replaceMap map[string]string
var existingCourses []string

func first(s string, _ bool) string { return s }
func Ternary[T any](b bool, t, f T) T {
	if b {
		return t
	}
	return f
}
func parseTime(t string) string {
	startTime, err := time.Parse("3:04 pm", t)
	if err != nil {
		logger.Fatal(err)
	}
	return startTime.Format("15:04")
}

func getSemesters() []Semester {
	c := colly.NewCollector()

	var semesters []Semester
	foundLatest := false

	c.OnHTML("select[name=p_term]", func(e *colly.HTMLElement) {
		e.DOM.Find("option").Each(func(i int, s *goquery.Selection) {
			if s.Text() != "None" {
				output, err := strconv.Atoi(first(s.Attr("value")))
				if err != nil {
					logger.Fatal(err)
				}
				semesters = append(semesters, Semester{output, s.Text(), !foundLatest && !strings.Contains(s.Text(), "M"), strings.Contains(s.Text(), "(View only)"), strings.Contains(s.Text(), "Medicine"), !strings.Contains(s.Text(), "Medicine") && strings.Contains(s.Text(), "M")})
				if !foundLatest && !strings.Contains(s.Text(), "M") {
					foundLatest = true
				}
			}
		})
	})

	c.Visit("https://selfservice.mun.ca/direct/bwckschd.p_disp_dyn_sched")
	c.Wait()

	return semesters
}

func processSemester(semester int) []Subject {
	var subjects []Subject

	c := colly.NewCollector()

	c.OnHTML("select[name=sel_subj]", func(e *colly.HTMLElement) {
		e.DOM.Find("option").EachWithBreak(func(i int, s *goquery.Selection) bool {
			if s.Text() != "All" {
				subjects = append(subjects, Subject{Name: first(s.Attr("value")), FriendlyName: s.Text()})
			}
			return true
		})
	})

	params := []byte("p_calling_proc=bwckschd.p_disp_dyn_sched&p_term=" + strconv.Itoa(semester))
	err := c.PostRaw("https://selfservice.mun.ca/direct/bwckgens.p_proc_term_date", params)
	if err != nil {
		logger.Fatal(err)
	}
	c.Wait()

	return subjects
}

func processCourse(title []string, body []string, semester int, subject string) {
	var campus string
	var credits int
	var comment *string
	var timeStartLine int
	var commentEndLine int
	var level string

	for i, line := range body {
		if strings.Contains(line, "Campus") {
			campus = line[:len(line)-7]
		}
		if strings.Contains(line, "Credits") {
			var err error
			credits, err = strconv.Atoi(string(strings.TrimSpace(line)[0]))
			if err != nil {
				logger.Fatal(err)
			}
		}
		if line == "Scheduled Meeting Times" {
			timeStartLine = i
		}
		if strings.HasPrefix(line, "Associated") {
			commentEndLine = i
		}
		if strings.HasPrefix(line, "Levels:") {
			level = strings.TrimSpace(line[8:])
		}
	}

	if commentEndLine != 0 {
		jointStrings := strings.Replace(strings.Join(body[0:commentEndLine], ""), "\n", "", -1)
		comment = &jointStrings
	}

	instructor := strings.TrimPrefix(body[len(body)-1], "(P)")

	var types []string

	if timeStartLine != 0 {
		for i := 0; i <= len(body[timeStartLine+8:])/7-1; i++ {
			if !slices.Contains(types, body[timeStartLine+8:][7*i+5]) {
				types = append(types, body[timeStartLine+8:][7*i+5])
			}
		}
	}

	var typesStr = strings.Join(types, ", ")

	if timeStartLine != 0 {
		db.Save(&Course{
			Name:        strings.Join(title[:len(title)-3], " - "),
			Id:          title[len(title)-2],
			CRN:         title[len(title)-3],
			Section:     title[len(title)-1],
			DateRange:   &body[len(body)-3],
			Type:        &typesStr,
			Instructor:  &instructor,
			Subject:     strings.Split(title[len(title)-2], " ")[0],
			SubjectFull: subject,
			Campus:      campus,
			Comment:     comment,
			Credits:     credits,
			SemesterID:  semester,
			Level:       level,
		})
	} else {
		db.Save(&Course{
			Name:        strings.Join(title[:len(title)-3], " - "),
			Id:          title[len(title)-2],
			CRN:         title[len(title)-3],
			Section:     title[len(title)-1],
			Subject:     strings.Split(title[len(title)-2], " ")[0],
			SubjectFull: subject,
			Campus:      campus,
			Comment:     comment,
			Credits:     credits,
			SemesterID:  semester,
			Level:       level,
		})
	}

	existingCourses = append(existingCourses, title[len(title)-3])

	addTimesToDB := func() {
		if timeStartLine != 0 {
			times := body[timeStartLine+8:]

			for i := 0; i <= len(times)/7-1; i++ {
				location := times[3+(i*7)]
				for from, to := range replaceMap {
					location = strings.Replace(location, from, to, 1)
				}

				if times[1+(i*7)] == "TBA" {
					db.Save(&CourseTime{
						CourseCRN: title[len(title)-3],
						StartTime: "TBA",
						EndTime:   "TBA",
						Days:      times[2+(i*7)],
						Location:  location,
						Semester:  semester,
					})
				} else {
					db.Save(&CourseTime{
						CourseCRN: title[len(title)-3],
						StartTime: parseTime(strings.Split(times[1+(i*7)], " - ")[0]),
						EndTime:   Ternary(times[1+(i*7)] == "TBA", "TBA", parseTime(strings.Split(times[1+(i*7)], " - ")[1])),
						Days:      times[2+(i*7)],
						Location:  location,
						Type:      times[5+(i*7)],
						Semester:  semester,
					})
				}
			}
		}
	}

	var oldCourseTimeCount int64
	db.Model(&CourseTime{}).Where("crn = ?", title[len(title)-3]).Count(&oldCourseTimeCount)

	addTimesToDB()

	var newCourseTimeCount int64
	db.Model(&CourseTime{}).Where("crn = ?", title[len(title)-3]).Count(&oldCourseTimeCount)

	if oldCourseTimeCount != newCourseTimeCount {
		db.Where("crn = ?", title[len(title)-3]).Delete(&CourseTime{})
	}

	addTimesToDB()

	db.Save(&Seating{Crn: title[len(title)-3], Available: 0, Max: 0, Waitlist: 0, Checked: "Never"})
}

func processSubject(subject Subject, semester int, course string) {
	c := colly.NewCollector()

	var courses []*goquery.Selection

	c.OnHTML("th.ddtitle", func(e *colly.HTMLElement) {
		courses = append(courses, e.DOM)
	})

	params := []byte("term_in=" + strconv.Itoa(semester) + "&sel_subj=dummy&sel_day=dummy&sel_schd=dummy&sel_insm=dummy&sel_camp=dummy&sel_levl=dummy&sel_sess=dummy&sel_instr=dummy&sel_ptrm=dummy&sel_attr=dummy&sel_subj=" + subject.Name + "&sel_crse=" + course + "&sel_title=&sel_schd=%25&sel_insm=%25&sel_from_cred=&sel_to_cred=&sel_camp=%25&sel_levl=%25&sel_ptrm=%25&sel_instr=%25&sel_sess=%25&sel_attr=%25&begin_hh=0&begin_mi=0&begin_ap=a&end_hh=0&end_mi=0&end_ap=a")
	err := c.PostRaw("https://selfservice.mun.ca/direct/bwckschd.p_get_crse_unsec", params)
	if err != nil {
		logger.Fatal(err)
	}
	c.Wait()

	if len(courses) == 101 && course == "" {
		for i := 1; i <= 9; i++ {
			processSubject(subject, semester, strconv.Itoa(i))
		}
		return
	}

	for _, course := range courses {
		tmp := strings.Split(course.Parent().Next().Text(), "\n")
		var body []string
		for _, line := range tmp {
			if len(line) > 0 {
				body = append(body, line)
			}
		}
		processCourse(strings.Split(course.Text(), " - "), body, semester, subject.FriendlyName)
	}
}

func removeNonExistingCourses() {
	logger.Println("🗑️ Removing non-existent courses...")
	rows, err := db.Raw("select crn, semester from courses").Rows()
	if err != nil {
		logger.Fatal(err)
	}
	for rows.Next() {
		var crn string
		var semester int
		rows.Scan(&crn, &semester)
		if !slices.Contains(existingCourses, crn) {
			exists := true
			c := colly.NewCollector()
			c.OnHTML("span.errortext", func(e *colly.HTMLElement) {
				exists = false
			})
			c.Visit("https://selfservice.mun.ca/direct/bwckschd.p_disp_detail_sched?term_in=" + strconv.Itoa(semester) + "&crn_in=" + crn)
			c.Wait()
			if !exists {
				db.Where("crn = ?", crn).Delete(&Course{})
			}
		}
	}
	logger.Println("✅ Complete!")
}

func scrape() {
	startTime := time.Now()
	existingCourses = nil

	logger.Println("⭐ Scraping Started!")

	for _, semester := range getSemesters() {
		//if course has already been scraped and its view only (is not going to be changed), dont scrape it
		if !semester.ViewOnly || db.Where("id = ?", semester.ID).Find(&Semester{}).RowsAffected == 0 {
			db.Save(&semester)
			logger.Println("📝 Processing Semester: " + semester.Name + " (" + strconv.Itoa(semester.ID) + ")")
			for _, subject := range processSemester(semester.ID) {
				logger.Println("	📝 Processing " + subject.FriendlyName + " (" + subject.Name + ")")
				processSubject(subject, semester.ID, "")

			}
		}
	}

	scrapingTime := time.Since(startTime)

	if os.Getenv("WEBHOOK_URL") != "" {
		logger.Println("🔔 Sending message to Discord")
		params := fmt.Sprintf(`{"username":"Claret Scraper","embeds":[{"author":{"name":"Claret Scraper Report","url":"https://claretformun.com"},"timestamp":"%s","color":65280,"fields":[{"name":"Scraping Time","value":"%s"},{"name":"Courses Scraped","value":"%d"}]}]}`, time.Now().Format(time.RFC3339), fmt.Sprintf("%02d:%02d", int(scrapingTime.Minutes()), int(scrapingTime.Seconds())%60), len(existingCourses))
		r, err := http.NewRequest("POST", os.Getenv("WEBHOOK_URL"), bytes.NewBuffer([]byte(params)))
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

	logger.Println("✅ Scrape Complete in " + fmt.Sprintf("%02d:%02d", int(scrapingTime.Minutes()), int(scrapingTime.Seconds())%60) + "!")
	logger.Printf("Courses scraped: %d", len(existingCourses))
}

func main() {
	logger = log.Default()
	logger.Println("👋 Claret Scraper")

	DB_URL := os.Getenv("DB_URL")
	if DB_URL == "" {
		logger.Fatal("DB_URL is not defined in environment variables")
	}

	replaceMap = map[string]string{
		"Arts and Administration Bldg": "A",
		"Henrietta Harvey Bldg":        "HH",
		"Business Administration Bldg": "BN",
		"INCO Innovation Centre":       "IIC",
		"Biotechnology Bldg":           "BT",
		"St. John's College":           "J",
		"Chemistry - Physics Bldg":     "C",
		"Core Science Facility":        "CSF",
		"M. O. Morgan Bldg":            "MU",
		"Computing Services":           "CS",
		"Physical Education Bldg":      "PE",
		"G. A. Hickman Bldg":           "ED",
		"Queen's College":              "QC",
		"Queen Elizabeth II Library":   "L",
		"S. J. Carew Bldg.":            "EN",
		"Science Bldg":                 "S",
		"Alexander Murray Bldg":        "ER",
		"Health Sciences Centre":       "H",
		"Coughlan College":             "CL",
		"Marine Institute":             "MI",
		"Center for Nursing Studies":   "N",
		"Arts and Science (SWGC)":      "AS",
		"Fine Arts (SWGC)":             "FA",
		"Forest Centre":                "FC",
		"Library/Computing (SWGC)":     "LC",
		"Western Memorial Hospital":    "WMH",
		"\u00A0":                       "N/A",
	}

	var err error

	db, err = gorm.Open(postgres.Open(DB_URL), &gorm.Config{})
	if err != nil {
		logger.Fatal(err)
	}
	logger.Println("💿 Connected to Database!")

	// migrate schemas
	db.AutoMigrate(&Semester{})
	db.AutoMigrate(&Course{})
	db.AutoMigrate(&CourseTime{})
	db.AutoMigrate(&Seating{})
	logger.Println("💾 Migrated Schemas!")

	scrape()
}

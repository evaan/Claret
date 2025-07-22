package scrapers

import (
	"log"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/evaan/Claret/internal/util"
	"github.com/gocolly/colly/v2"
)

func ParseCourse(logger *log.Logger, e *colly.HTMLElement, semester int, subject string) (util.Course, []util.CourseTime, []util.Professor, []util.CourseInstructor) {
	title := e.Text
	body := e.DOM.Parent().Next().Text()

	semesterStr := strconv.Itoa(semester)

	var course util.Course

	course.SemesterID = semester
	course.SubjectID = subject

	segments := strings.Split(title, " - ")

	n := len(segments)
	course.Section = segments[n-1]
	course.ID = segments[n-2]
	course.CRN = segments[n-3]
	course.Key = semesterStr + course.CRN
	course.Name = strings.Join(segments[:n-3], " - ")

	comment := ""
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		commentDone := false
		if line != "" {
			if !commentDone && !strings.HasPrefix(line, "Associated Term") {
				comment += line
			} else if !commentDone {
				commentDone = true
			}
			if strings.HasPrefix(line, "Levels:") {
				course.Levels = strings.TrimSpace(strings.TrimPrefix(line, "Levels:"))
			} else if strings.HasPrefix(line, "Registration Dates:") {
				course.RegistrationDates = strings.TrimSpace(strings.TrimPrefix(line, "Registration Dates: "))
			}
			if strings.HasSuffix(line, "Credits") { // get credits from course
				course.Credits = util.GetCredits(logger, line)
			} else if strings.HasSuffix(line, "Campus") { // get course campus
				course.Campus = strings.TrimSuffix(line, " Campus")
			}
		}
	}

	course.Comment = &comment

	var schedule []string

	e.DOM.Parent().Next().Find("table.datadisplaytable").First().Find("td.dddefault").Each(func(i int, sel *goquery.Selection) {
		schedule = append(schedule, strings.TrimSpace(strings.TrimPrefix(sel.Text(), "(P)")))
	})

	var courseTimes []util.CourseTime
	var professors []util.Professor
	var courseInstructors []util.CourseInstructor

	types := make([]string, 0)

	// iterate through course time table
	for i := range len(schedule) / 7 {
		var courseTime util.CourseTime
		courseTime.CourseKey = course.Key
		// get time, some courses are TBA rather than "12:00 am - 12:01 am" because banner is an amazing piece of software
		if schedule[i*7+1] == "TBA" {
			courseTime.StartTime = "00:00"
			courseTime.EndTime = "00:01"
			courseTime.CourseCRN = course.CRN
			courseTime.SemesterID = course.SemesterID
		} else {
			times := strings.Split(schedule[i*7+1], " - ")
			startTime, err := time.Parse("3:04 pm", times[0])
			if err != nil {
				logger.Printf("Error scraping course time: %s\n", err.Error())
				util.SendErrorToWebhook(os.Getenv("SCRAPER_WEBHOOK_URL"), err)
				continue
			}
			courseTime.StartTime = startTime.Format("15:04")
			endTime, err := time.Parse("3:04 pm", times[1])
			if err != nil {
				logger.Printf("Error scraping course time: %s\n", err.Error())
				util.SendErrorToWebhook(os.Getenv("SCRAPER_WEBHOOK_URL"), err)
				continue
			}
			courseTime.EndTime = endTime.Format("15:04")
			courseTime.CourseCRN = course.CRN
			courseTime.SemesterID = course.SemesterID
		}
		if schedule[i*7+2] != "" {
			days := schedule[i*7+2]
			courseTime.Days = &days
		}
		courseTime.Location = util.ReplaceBuildingName(schedule[i*7+3])
		courseTime.DateRange = schedule[i*7+4]
		if course.DateRange == nil {
			dateRange := courseTime.DateRange
			course.DateRange = &dateRange
		}
		courseTime.Type = schedule[i*7+5]
		if !slices.Contains(types, schedule[i*7+5]) {
			types = append(types, schedule[i*7+5])
		}
		instructors := schedule[i*7+6]
		if instructors != "TBA" {
			for _, instructor := range strings.Split(instructors, ", ") {
				professors = append(professors, util.Professor{Name: instructor})
				courseInstructors = append(courseInstructors, util.CourseInstructor{CourseKey: course.Key, ProfessorName: instructor})
			}
		}

		courseTimes = append(courseTimes, courseTime)
	}

	if len(types) == 0 {
		course.Types = "No Activity"
	} else {
		course.Types = strings.Join(types, ", ")
	}

	return course, courseTimes, professors, courseInstructors
}

func GetCoursesWithCourse(logger *log.Logger, semester int, subject string, course string) ([]util.Course, []util.CourseTime, []util.Professor, []util.CourseInstructor) {
	c := colly.NewCollector()

	semesterStr := strconv.Itoa(semester)

	var courses []util.Course
	var courseTimes []util.CourseTime
	var professors []util.Professor
	var courseInstructors []util.CourseInstructor

	c.OnHTML("th.ddtitle", func(e *colly.HTMLElement) {
		course, times, profs, instructors := ParseCourse(logger, e, semester, subject)
		courses = append(courses, course)
		courseTimes = append(courseTimes, times...)
		professors = append(professors, profs...)
		courseInstructors = append(courseInstructors, instructors...)
	})

	err := c.PostRaw("https://selfservice.mun.ca/direct/bwckschd.p_get_crse_unsec", util.MapToBytes(map[string]any{
		"term_in":       semesterStr,
		"sel_subj":      []string{"dummy", subject},
		"sel_day":       "dummy",
		"sel_schd":      []string{"dummy", "%"},
		"sel_insm":      []string{"dummy", "%"},
		"sel_camp":      []string{"dummy", "%"},
		"sel_levl":      []string{"dummy", "%"},
		"sel_sess":      []string{"dummy", "%"},
		"sel_instr":     []string{"dummy", "%"},
		"sel_ptrm":      []string{"dummy", "%"},
		"sel_attr":      []string{"dummy", "%"},
		"sel_crse":      course,
		"sel_title":     "",
		"sel_from_cred": "",
		"sel_to_cred":   "",
		"begin_hh":      "0",
		"begin_mi":      "0",
		"begin_ap":      "a",
		"end_hh":        "0",
		"end_mi":        "0",
		"end_ap":        "a",
	}))
	if err != nil {
		logger.Printf("Error getting courses: %s\n", err.Error())
	}

	c.Wait()

	return courses, courseTimes, util.Unique(professors), courseInstructors
}

func GetCourses(logger *log.Logger, semester int, subject string) ([]util.Course, []util.CourseTime, []util.Professor, []util.CourseInstructor) {
	c := colly.NewCollector()

	semesterStr := strconv.Itoa(semester)

	var courses []util.Course
	var courseTimes []util.CourseTime
	var professors []util.Professor
	var courseInstructors []util.CourseInstructor

	c.OnHTML("th.ddtitle", func(e *colly.HTMLElement) {
		course, times, profs, instructors := ParseCourse(logger, e, semester, subject)
		courses = append(courses, course)
		courseTimes = append(courseTimes, times...)
		professors = append(professors, profs...)
		courseInstructors = append(courseInstructors, instructors...)
	})

	err := c.PostRaw("https://selfservice.mun.ca/direct/bwckschd.p_get_crse_unsec", util.MapToBytes(map[string]any{
		"term_in":       semesterStr,
		"sel_subj":      []string{"dummy", subject},
		"sel_day":       "dummy",
		"sel_schd":      []string{"dummy", "%"},
		"sel_insm":      []string{"dummy", "%"},
		"sel_camp":      []string{"dummy", "%"},
		"sel_levl":      []string{"dummy", "%"},
		"sel_sess":      []string{"dummy", "%"},
		"sel_instr":     []string{"dummy", "%"},
		"sel_ptrm":      []string{"dummy", "%"},
		"sel_attr":      []string{"dummy", "%"},
		"sel_crse":      "",
		"sel_title":     "",
		"sel_from_cred": "",
		"sel_to_cred":   "",
		"begin_hh":      "0",
		"begin_mi":      "0",
		"begin_ap":      "a",
		"end_hh":        "0",
		"end_mi":        "0",
		"end_ap":        "a",
	}))
	if err != nil {
		logger.Printf("Error getting courses: %s\n", err.Error())
	}

	c.Wait()

	if len(courses) >= 101 {
		courses = nil
		courseTimes = nil
		professors = nil
		courseInstructors = nil
		for i := 0; i <= 9; i++ {
			courses1, times, profs, instructors := GetCoursesWithCourse(logger, semester, subject, strconv.Itoa(i))
			courses = append(courses, courses1...)
			courseTimes = append(courseTimes, times...)
			professors = append(professors, profs...)
			courseInstructors = append(courseInstructors, instructors...)
		}
	}

	return courses, courseTimes, util.Unique(professors), courseInstructors
}

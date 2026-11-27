package api

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/evaan/Claret/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// ICalHandler godoc
// @Summary Get iCal Calendar
// @Description Returns an iCal file containing all schedule items for selected courses
// @Accept json
// @Produce text/calendar
// @Param semester query string true "Semester ID (i.e. 202401)"
// @Param crn query string true "Course Registration Numbers seperated by commas (i.e. 40983,40984)"
// @Success 200 {object} string
// @Failure 400 {object} util.ErrorResponse
// @Failure 500 {object} util.ErrorResponse
// @Router /claret.ics [get]
func ICalHandler(c *gin.Context, db *gorm.DB) {
	db = db.WithContext(c.Request.Context())
	crns, ok := crnParams(c, c.Query("crns"), true)
	if !ok {
		return
	}
	semester, ok := semesterParam(c, db, c.Query("semester"), false)
	if !ok {
		return
	}

	builder := strings.Builder{}
	builder.WriteString("BEGIN:VCALENDAR\nVERSION:2.0\nPRODID:-//claretformun.com//Claret Schedule Builder/EN\nCALSCALE:GREGORIAN\nMETHOD:PUBLISH\nX-WR-CALNAME:MUN Courses (via Claret)\nX-WR-TIMEZONE:America/St_Johns\n")

	courseTimes := []util.CourseTimeICal{}
	err := db.Raw(`SELECT ct.course_key,c.crn AS course_crn,c.semester_id,ct.start_time,ct.end_time,ct.days,ct.date_range,ct.location,ct.type,c.id AS course_id,c.name AS course_name,COALESCE(STRING_AGG(DISTINCT ci.professor_name, ', ' ORDER BY ci.professor_name), '') AS instructor_names FROM course_times ct JOIN courses c ON c.key = ct.course_key LEFT JOIN course_instructors ci ON ci.course_key = ct.course_key WHERE c.crn = ANY(?) AND c.semester_id = ? AND NOT (ct.start_time = '00:00' AND ct.end_time = '00:01') AND ct.date_range <> '' AND ct.days IS NOT NULL GROUP BY ct.course_key,c.crn,c.semester_id,ct.start_time,ct.end_time,ct.days,ct.date_range,ct.location,ct.type,c.id,c.name`, pq.Array(crns), semester).Scan(&courseTimes).Error
	if err != nil {
		internalError(c, err)
		return
	}

	re := regexp.MustCompile(`(\d+)(st|nd|rd|th)`)
	loc, err := time.LoadLocation("America/St_Johns")
	if err != nil {
		internalError(c, err)
		return
	}

	for _, ct := range courseTimes {
		if ct.Days == nil || ct.DateRange == "" {
			continue
		}
		dates := strings.Split(re.ReplaceAllString(ct.DateRange, ""), " - ")
		if len(dates) != 2 {
			continue
		}
		startDate, err := time.ParseInLocation("01/02/2006", strings.TrimSpace(dates[0]), loc)
		if err != nil {
			fmt.Println("Error parsing start date:", err)
			continue
		}
		endDate, err := time.ParseInLocation("01/02/2006", strings.TrimSpace(dates[1]), loc)
		if err != nil {
			fmt.Println("Error parsing end date:", err)
			continue
		}
		startTime, err := time.Parse("15:04", ct.StartTime)
		if err != nil {
			fmt.Println("Error parsing start time:", err)
			continue
		}
		endTime, err := time.Parse("15:04", ct.EndTime)
		if err != nil {
			fmt.Println("Error parsing end time:", err)
			continue
		}

		firstClass := util.EarliestClassDate(startDate, *ct.Days)
		lastClass := util.LatestClassDate(endDate, *ct.Days)
		if firstClass.IsZero() || lastClass.IsZero() || firstClass.After(lastClass) {
			continue
		}
		uid := sha256.Sum256([]byte(strings.Join([]string{ct.CourseKey, ct.StartTime, ct.EndTime, *ct.Days, ct.DateRange, ct.Location, ct.Type}, "\x00")))
		builder.WriteString(fmt.Sprintf(
			"BEGIN:VEVENT\nUID:%s@claretformun.com\nDTSTAMP:%s\nDTSTART;TZID=America/St_Johns:%s\nDTEND;TZID=America/St_Johns:%s\nRRULE:FREQ=WEEKLY;BYDAY=%s;UNTIL=%s\nSUMMARY:%s\nLOCATION:%s\nDESCRIPTION:%s\nEND:VEVENT\n",
			fmt.Sprintf("%x", uid[:16]),
			time.Now().UTC().Format("20060102T150405Z"),
			time.Date(firstClass.Year(), firstClass.Month(), firstClass.Day(), startTime.Hour(), startTime.Minute(), 0, 0, loc).Format("20060102T150405"),
			time.Date(firstClass.Year(), firstClass.Month(), firstClass.Day(), endTime.Hour(), endTime.Minute(), 0, 0, loc).Format("20060102T150405"),
			util.ICalRepeatDates(*ct.Days),
			time.Date(lastClass.Year(), lastClass.Month(), lastClass.Day(), 23, 59, 59, 0, loc).UTC().Format("20060102T150405Z"),
			icalText(ct.CourseID+" - "+ct.CourseName),
			icalText(ct.Location),
			icalText("Instructor(s): "+ct.InstructorNames+" - Generated with Claret"),
		))
	}

	builder.WriteString("END:VCALENDAR\n")
	c.Header("Content-Disposition", `attachment; filename="claret.ics"`)
	c.Data(http.StatusOK, "text/calendar; charset=utf-8", []byte(foldICal(builder.String())))
}

func icalText(text string) string {
	return strings.NewReplacer("\\", "\\\\", "\r\n", "\\n", "\r", "\\n", "\n", "\\n", ";", "\\;", ",", "\\,").Replace(text)
}

// RFC 5545 limits content lines to 75 octets, with CRLF and UTF-8-safe folding.
func foldICal(calendar string) string {
	var out strings.Builder
	for _, line := range strings.Split(strings.TrimSuffix(calendar, "\n"), "\n") {
		width := 0
		for _, char := range line {
			value := string(char)
			if width+len(value) > 75 {
				out.WriteString("\r\n ")
				width = 1
			}
			out.WriteString(value)
			width += len(value)
		}
		out.WriteString("\r\n")
	}
	return out.String()
}

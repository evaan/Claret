package api

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
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
	semesterStr := c.Query("semester")
	crns := c.Query("crns")
	if semesterStr == "" || crns == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "semester and crns are required parameters",
		})
		return
	}

	semester, err := strconv.Atoi(semesterStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	var builder strings.Builder

	// vcalendar header
	builder.WriteString("BEGIN:VCALENDAR\nVERSION:2.0\nPRODID:-//claretformun.com//Claret Schedule Builder/EN\nCALSCALE:GREGORIAN\nMETHOD:PUBLISH\nX-WR-CALNAME:MUN Courses (via Claret)\nX-WR-TIMEZONE:America/St_Johns\n")

	courseTimes := make([]util.CourseTimeICal, 0)
	err = db.Raw(`
		SELECT
			ct.course_key, ct.course_crn, ct.semester_id, ct.start_time, ct.end_time, ct.days, ct.date_range, ct.location, c.id AS course_id, c.name AS course_name,
			STRING_AGG(DISTINCT ci.professor_name, ', ' ORDER BY ci.professor_name) AS instructor_names 
		FROM course_times ct
		JOIN courses c ON c.crn = ct.course_crn AND c.semester_id = ct.semester_id
		LEFT JOIN course_instructors ci ON ci.course_key = ct.course_key
		WHERE ct.course_crn = ANY (?) AND ct.semester_id = ?
			AND NOT (ct.start_time = '00:00' AND ct.end_time = '00:01')
			AND ct.date_range != ''
			AND ct.days IS NOT NULL
		GROUP BY
			ct.course_key, ct.course_crn, ct.semester_id, ct.start_time, ct.end_time,
			ct.days, ct.date_range, ct.location, c.id, c.name`, pq.Array(strings.Split(crns, ",")), semester).Scan(&courseTimes).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	re := regexp.MustCompile(`(\d+)(st|nd|rd|th)`)
	nst, err := time.LoadLocation("America/St_Johns")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	for _, courseTime := range courseTimes {
		dateRange := strings.Split(re.ReplaceAllString(courseTime.DateRange, ""), " - ")
		if len(dateRange) != 2 {
			continue // there is no reason it shouldn't be structured like this
		}
		startDate, err := time.Parse("Jan 2, 2006", dateRange[0])
		if err != nil {
			fmt.Println("Error parsing time:", err)
			continue
		}
		endDate, err := time.Parse("Jan 2, 2006", dateRange[1])
		if err != nil {
			fmt.Println("Error parsing time:", err)
			continue
		}
		startTime, err := time.Parse("15:04", courseTime.StartTime)
		if err != nil {
			fmt.Println("Error parsing time:", err)
			continue
		}
		endTime, err := time.Parse("15:04", courseTime.EndTime)
		if err != nil {
			fmt.Println("Error parsing time:", err)
			continue
		}
		firstTime := util.EarliestClassDate(startDate, *courseTime.Days)
		lastTime := util.LatestClassDate(endDate, *courseTime.Days)
		builder.WriteString(fmt.Sprintf(
			"BEGIN:VEVENT\nUID:%s@claretformun.com\nDTSTAMP:%s\nDTSTART;TZID=America/St_Johns:%s\nDTEND;TZID=America/St_Johns:%s\nRRULE:FREQ=WEEKLY;BYDAY=%s;UNTIL=%s\nSUMMARY:%s\nLOCATION:%s\nDESCRIPTION:%s\nEND:VEVENT\n",
			courseTime.CourseKey,
			time.Now().UTC().Format("20060102T150405Z"),
			time.Date(firstTime.Year(), firstTime.Month(), firstTime.Day(), startTime.Hour(), startTime.Minute(), 0, 0, nst).Format("20060102T150405"),
			time.Date(firstTime.Year(), firstTime.Month(), firstTime.Day(), endTime.Hour(), endTime.Minute(), 0, 0, nst).Format("20060102T150405"),
			util.ICalRepeatDates(*courseTime.Days),
			time.Date(lastTime.Year(), lastTime.Month(), lastTime.Day(), 23, 59, 59, 0, nst).Format("20060102T150405Z"),
			courseTime.CourseID+" - "+courseTime.CourseName,
			courseTime.Location,
			"Instructor(s): "+courseTime.InstructorNames+" - Generated with Claret",
		))
	}

	builder.WriteString("END:VCALENDAR\n")

	c.String(http.StatusOK, builder.String())
}

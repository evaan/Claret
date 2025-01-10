package main

import (
	"bytes"
	"database/sql"
	"net/http"
	"strings"
	"time"
)

func health(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Ok"))
}

func ics(w http.ResponseWriter, r *http.Request) {
	var rows *sql.Rows
	var err error

	query_crn := r.URL.Query().Get("crn")
	query_semester := r.URL.Query().Get("semester")

	if query_crn == "" {
		rows, err = db.Query("SELECT courses.semester, times.crn, courses.id, courses.name, courses.\"dateRange\", times.days, times.\"startTime\", times.\"endTime\", times.location FROM times JOIN courses ON times.crn = courses.crn")
	} else {
		query_crn_split := strings.Split(query_crn, ",")
		for i := range query_crn_split { query_crn_split[i] = query_semester + query_crn_split[i] }
		rows, err = db.Query(`select courses.id, courses.name, courses."dateRange", times.days, times."startTime", times."endTime", times.location from courses join times on courses.identifier = times.identifier where courses.identifier = any($1);`, query_crn_split)
	}

	if err != nil {
		logger.Println(err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	defer rows.Close()

	out_buf := new(bytes.Buffer)

	// start calendar
	out_buf.Write([]byte("BEGIN:VCALENDAR\r\n"))
	out_buf.Write([]byte("VERSION:2.0\r\n"))

	// PRODID: This value is required and should be in the form of a FPI
	// (Formal Product Identifier), as defined in ISO.9070.1991.
	// https://en.wikipedia.org/wiki/Formal_Public_Identifier
	out_buf.Write([]byte("PRODID:+//IDN evanvokey.com//Claret ICS Server//EN\r\n"))

	// iCal Method (values not specified in the RFC, for feeds use PUBLISH)
	out_buf.Write([]byte("METHOD:PUBLISH\r\n"))

	// Calendar Name & Description
	cal_name := "Courses (via Claret)" // TODO: Include semester names (ex: Fall 2023)
	out_buf.Write([]byte("X-WR-CALNAME:" + cal_name + "\r\n"))

	for rows.Next() {
		var (
			id         string
			name       string
			date_range string
			days       string
			start      string
			end        string
			location   string
		)

		err := rows.Scan(&id, &name, &date_range, &days, &start, &end, &location)
		if err != nil {
			logger.Println(err)

		}

		day_of_week := strings.Split(days, "")
		start_date, end_reccurence_date, err := dateRangeParse(date_range)
		if err != nil {
			// TODO: If error parsing date range, use some default semester range
			// For now, skip entry
			continue
		}

		// TODO: multiple days a week can be expressed in RRULE, so this loop could be removed
		for _, day := range day_of_week {
			ICAL_DATE_TIME_LOCAL_FORM := "20060102T150405"
			first_event_date := next_weekday(start_date, day_of_week_map[day])
			start_time, err := time.Parse("15:04", start)
			if err != nil {
				continue // if parsing time fails (such as TBA), skip entry
			}

			end_time, err := time.Parse("15:04", end)
			if err != nil {
				continue // if parsing time fails (such as TBA), skip entry
			}
			dt_start := time.Date(first_event_date.Year(), first_event_date.Month(), first_event_date.Day(), start_time.Hour(), start_time.Minute(), start_time.Second(), start_time.Nanosecond(), first_event_date.Location())
			dt_end := time.Date(first_event_date.Year(), first_event_date.Month(), first_event_date.Day(), end_time.Hour(), end_time.Minute(), end_time.Second(), end_time.Nanosecond(), first_event_date.Location())

			out_buf.Write([]byte("BEGIN:VEVENT\r\n"))
			out_buf.Write([]byte("UID:" + day + dt_start.Format(ICAL_DATE_TIME_LOCAL_FORM) + dt_end.Format(ICAL_DATE_TIME_LOCAL_FORM) + "@claret-cal-uid.evanvokey.com" + "\r\n"))
			out_buf.Write([]byte("SUMMARY:" + id + " - " + name + "\r\n"))
			out_buf.Write([]byte("DESCRIPTION:No Event Description" + "\r\n"))
			out_buf.Write([]byte("LOCATION:" + location + "\r\n"))

			out_buf.Write([]byte("DTSTART;TZID=/" + dt_start.Location().String() + ":" + dt_start.Format(ICAL_DATE_TIME_LOCAL_FORM) + "\r\n"))
			out_buf.Write([]byte("DTEND;TZID=/" + dt_end.Location().String() + ":" + dt_end.Format(ICAL_DATE_TIME_LOCAL_FORM) + "\r\n"))

			out_buf.Write([]byte("RRULE:FREQ=WEEKLY;UNTIL=" + end_reccurence_date.UTC().Format(ICAL_DATE_TIME_LOCAL_FORM) + "Z	\r\n"))
			out_buf.Write([]byte("DTSTAMP;TZID=" + time.Now().Location().String() + ":" + time.Now().Format(ICAL_DATE_TIME_LOCAL_FORM) + "\r\n"))

			out_buf.Write([]byte("END:VEVENT\r\n"))
		}
	}

	out_buf.Write([]byte("END:VCALENDAR\r\n"))

	lineFoldBytes(out_buf, w)
	w.Header().Add("content-type", "text/calendar")

}

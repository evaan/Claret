package main

import (
	"bytes"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"
)

func lineFoldBytes(in *bytes.Buffer, out http.ResponseWriter) {

	in_lines := strings.Split(in.String(), "\r\n")

	for _, line := range in_lines {
		var left, right int
		for left, right = 0, 74; right < len(line); left, right = right, right+74 {
			for !utf8.RuneStart(line[right]) {
				right--
			}

			if left != 0 {
				out.Write([]byte("\t"))
			}

			out.Write([]byte(line[left:right]))
			out.Write([]byte("\r\n"))
		}
		if left != 0 {
			out.Write([]byte("\t"))
		}

		out.Write([]byte(line[left:]))
		out.Write([]byte("\r\n"))
	}
}

func dateRangeParse(dr string) (start_time time.Time, end_time time.Time, err error) {
	x := strings.Split(dr, " - ")
	date_range_start := x[0]
	date_range_end := x[1]

	date_range_form := "Jan 02, 2006"

	start_time, err = time.ParseInLocation(date_range_form, date_range_start, banner_tz)
	if err != nil {
		return
	}
	end_time, err = time.ParseInLocation(date_range_form, date_range_end, banner_tz)
	return
}

func next_weekday(given_date_time time.Time, weekday time.Weekday) time.Time {
	weekday_delta := (7 + int(weekday-given_date_time.Weekday())) % 7
	result := given_date_time.AddDate(0, 0, weekday_delta)
	return result
}

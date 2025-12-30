package scrapers

import (
	"encoding/json"
	"errors"
	"html"
	"io"
	"net/http"
	"slices"
	"strconv"

	"github.com/evaan/Claret/internal/util"
)

func GetCourses(client *http.Client, semester util.Semester) ([]util.Course, []util.CourseSeating, error) {
	err := SaveTerm(client, semester)
	if err != nil {
		return nil, nil, err
	}

	err = SendSearch(client, semester)
	if err != nil {
		return nil, nil, err
	}

	courses := make([]util.Course, 0)
	seatings := make([]util.CourseSeating, 0)
	items := 1
	offset := 0

	for items > len(courses) {
		resp, err := client.Get("https://self-service.mun.ca/StudentRegistrationSsb/ssb/searchResults/searchResults?txt_subject=&txt_term=" + strconv.Itoa(semester.ID) + "&uniqueSessionId=claret&pageOffset=" + strconv.Itoa(offset) + "&pageMaxSize=500&sortColumn=subjectDescription&sortDirection=asc")
		if err != nil {
			return nil, nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return nil, nil, errors.New("Recieved status code " + strconv.Itoa(resp.StatusCode) + " when requesting courses")
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, nil, err
		}

		var searchResults struct {
			Data []struct {
				CourseReferenceNumber   string  `json:"courseReferenceNumber"`
				Subject                 string  `json:"subject"`
				CourseNumber            string  `json:"courseNumber"`
				CourseTitle             string  `json:"courseTitle"`
				CreditHours             float32 `json:"creditHours"`
				CampusDescription       string  `json:"campusDescription"`
				SequenceNumber          string  `json:"sequenceNumber"`
				ScheduleTypeDescription string  `json:"scheduleTypeDescription"`
				MaximumEnrollment       int     `json:"maximumEnrollment"`
				Enrollment              int     `json:"enrollment"`
				WaitCapacity            int     `json:"waitCapacity"`
				WaitAvailable           int     `json:"waitAvailable"`
			} `json:"data"`
			SectionsFetchedCount int `json:"sectionsFetchedCount"`
		}

		err = json.Unmarshal(body, &searchResults)
		if err != nil {
			return nil, nil, err
		}

		if items != searchResults.SectionsFetchedCount {
			items = searchResults.SectionsFetchedCount
		}
		offset += len(searchResults.Data)

		for _, courseData := range searchResults.Data {
			courses = append(courses, util.Course{
				Key:        strconv.Itoa(semester.ID) + courseData.CourseReferenceNumber,
				ID:         courseData.Subject + " " + courseData.CourseNumber,
				Name:       html.UnescapeString(courseData.CourseTitle),
				CRN:        html.UnescapeString(courseData.CourseReferenceNumber),
				Section:    html.UnescapeString(courseData.SequenceNumber),
				Credits:    courseData.CreditHours,
				SubjectID:  html.UnescapeString(courseData.Subject),
				SemesterID: semester.ID,
				Type:       html.UnescapeString(courseData.ScheduleTypeDescription),
				Campus:     html.UnescapeString(courseData.CampusDescription),
			})
			seatings = append(seatings, util.CourseSeating{
				Semester:    semester.ID,
				CRN:         courseData.CourseReferenceNumber,
				Seats:       courseData.Enrollment,
				MaxSeats:    courseData.MaximumEnrollment,
				Waitlist:    courseData.WaitAvailable,
				MaxWaitlist: courseData.WaitCapacity,
			})
		}
	}

	return courses, seatings, nil
}

func GetMeetingTimes(semester util.Semester, crn string) ([]util.CourseTime, []string, error) {
	resp, err := http.Get("https://self-service.mun.ca/StudentRegistrationSsb/ssb/searchResults/getFacultyMeetingTimes?term=" + strconv.Itoa(semester.ID) + "&courseReferenceNumber=" + crn)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, nil, errors.New("Recieved status code " + strconv.Itoa(resp.StatusCode) + " when requesting course info")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	var meetingTimes struct {
		Fmt []struct {
			Faculty []struct {
				DisplayName string `json:"displayName"`
			} `json:"faculty"`
			MeetingTime struct {
				BeginTime              string `json:"beginTime"`
				EndTime                string `json:"endTime"`
				StartDate              string `json:"startDate"`
				EndDate                string `json:"endDate"`
				Building               string `json:"building"`
				Room                   string `json:"room"`
				MeetingTypeDescription string `json:"meetingTypeDescription"`
				Monday                 bool   `json:"monday"`
				Tuesday                bool   `json:"tuesday"`
				Wednesday              bool   `json:"wednesday"`
				Thursday               bool   `json:"thursday"`
				Friday                 bool   `json:"friday"`
				Saturday               bool   `json:"saturday"`
				Sunday                 bool   `json:"sunday"`
			} `json:"meetingTime"`
		} `json:"fmt"`
	}

	err = json.Unmarshal(body, &meetingTimes)
	if err != nil {
		return nil, nil, err
	}

	instructors := make([]string, 0)
	courseTimes := make([]util.CourseTime, 0)

	for _, meetingTime := range meetingTimes.Fmt {
		for _, instructor := range meetingTime.Faculty {
			if !slices.Contains(instructors, instructor.DisplayName) {
				instructors = append(instructors, html.UnescapeString(instructor.DisplayName))
			}
		}
		if len(meetingTime.MeetingTime.BeginTime) == 0 || len(meetingTime.MeetingTime.EndTime) == 0 {
			continue
		}
		var days string
		if meetingTime.MeetingTime.Monday {
			days += "M"
		}
		if meetingTime.MeetingTime.Tuesday {
			days += "T"
		}
		if meetingTime.MeetingTime.Wednesday {
			days += "W"
		}
		if meetingTime.MeetingTime.Thursday {
			days += "R"
		}
		if meetingTime.MeetingTime.Friday {
			days += "F"
		}
		if meetingTime.MeetingTime.Saturday {
			days += "S"
		}
		if meetingTime.MeetingTime.Sunday {
			days += "U"
		}
		courseTimes = append(courseTimes, util.CourseTime{
			StartTime: meetingTime.MeetingTime.BeginTime[:2] + ":" + meetingTime.MeetingTime.BeginTime[2:],
			EndTime:   meetingTime.MeetingTime.EndTime[:2] + ":" + meetingTime.MeetingTime.EndTime[2:],
			Days:      &days,
			Location:  meetingTime.MeetingTime.Building + " " + meetingTime.MeetingTime.Room,
			DateRange: meetingTime.MeetingTime.StartDate + " - " + meetingTime.MeetingTime.EndDate,
			Type:      meetingTime.MeetingTime.MeetingTypeDescription,
			CourseKey: strconv.Itoa(semester.ID) + crn,
		})
	}

	return courseTimes, instructors, nil
}

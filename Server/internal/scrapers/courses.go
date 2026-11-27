package scrapers

import (
	"encoding/json"
	"errors"
	"html"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

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
		if resp.StatusCode != 200 {
			resp.Body.Close()
			return nil, nil, errors.New("Recieved status code " + strconv.Itoa(resp.StatusCode) + " when requesting courses")
		}

		body, err := util.ReadBody(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, nil, err
		}

		var searchResults struct {
			Data []struct {
				CourseReferenceNumber          string   `json:"courseReferenceNumber"`
				Subject                        string   `json:"subject"`
				CourseNumber                   string   `json:"courseNumber"`
				CourseTitle                    string   `json:"courseTitle"`
				CreditHours                    float32  `json:"creditHours"`
				CampusDescription              string   `json:"campusDescription"`
				SequenceNumber                 string   `json:"sequenceNumber"`
				ScheduleTypeDescription        string   `json:"scheduleTypeDescription"`
				MaximumEnrollment              int      `json:"maximumEnrollment"`
				Enrollment                     int      `json:"enrollment"`
				WaitCapacity                   int      `json:"waitCapacity"`
				WaitAvailable                  int      `json:"waitAvailable"`
				InstructionalMethodDescription string   `json:"instructionalMethodDescription"`
				PartOfTerm                     string   `json:"partOfTerm"`
				IsSectionLinked                bool     `json:"isSectionLinked"`
				LinkIdentifier                 string   `json:"linkIdentifier"`
				CrossList                      string   `json:"crossList"`
				CreditHourLow                  *float32 `json:"creditHourLow"`
				CreditHourHigh                 *float32 `json:"creditHourHigh"`
				SeatsAvailable                 *int     `json:"seatsAvailable"`
				WaitCount                      *int     `json:"waitCount"`
				OpenSection                    *bool    `json:"openSection"`
				CrossListCapacity              *int     `json:"crossListCapacity"`
				CrossListCount                 *int     `json:"crossListCount"`
				CrossListAvailable             *int     `json:"crossListAvailable"`
			} `json:"data"`
			Success              bool `json:"success"`
			SectionsFetchedCount int  `json:"sectionsFetchedCount"`
		}

		err = json.Unmarshal(body, &searchResults)
		if err != nil {
			return nil, nil, err
		}
		if !searchResults.Success {
			return nil, nil, errors.New("Self-Service course search was not successful")
		}
		if searchResults.SectionsFetchedCount < 0 || searchResults.SectionsFetchedCount > 20000 || offset+len(searchResults.Data) > 20000 {
			return nil, nil, errors.New("Self-Service course count exceeds safety limit")
		}

		if items != searchResults.SectionsFetchedCount {
			items = searchResults.SectionsFetchedCount
		}
		if len(searchResults.Data) == 0 && offset < items {
			return nil, nil, errors.New("Self-Service returned an empty page before all courses were fetched")
		}
		offset += len(searchResults.Data)

		updatedAt := time.Now().UTC()
		for _, courseData := range searchResults.Data {
			courses = append(courses, util.Course{
				CourseDetails: util.CourseDetails{
					InstructionalMethod: html.UnescapeString(courseData.InstructionalMethodDescription),
					PartOfTerm:          html.UnescapeString(courseData.PartOfTerm),
					IsSectionLinked:     courseData.IsSectionLinked,
					LinkIdentifier:      html.UnescapeString(courseData.LinkIdentifier),
					CrossList:           html.UnescapeString(courseData.CrossList),
					CreditsMin:          courseData.CreditHourLow,
					CreditsMax:          courseData.CreditHourHigh,
				},
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
				Semester:           semester.ID,
				CRN:                courseData.CourseReferenceNumber,
				Seats:              courseData.Enrollment,
				MaxSeats:           courseData.MaximumEnrollment,
				Waitlist:           courseData.WaitAvailable,
				MaxWaitlist:        courseData.WaitCapacity,
				CourseKey:          strconv.Itoa(semester.ID) + courseData.CourseReferenceNumber,
				SeatsAvailable:     courseData.SeatsAvailable,
				WaitCount:          courseData.WaitCount,
				OpenSection:        courseData.OpenSection,
				CrossListCapacity:  courseData.CrossListCapacity,
				CrossListCount:     courseData.CrossListCount,
				CrossListAvailable: courseData.CrossListAvailable,
				UpdatedAt:          updatedAt,
			})
		}
	}

	if len(courses) != items {
		return nil, nil, errors.New("Self-Service course count did not match the fetched catalog")
	}
	return courses, seatings, nil
}

func GetMeetingTimes(semester util.Semester, crn string) ([]util.CourseTime, []string, error) {
	return getMeetingTimes(&http.Client{Timeout: 30 * time.Second}, semester, crn)
}

func getMeetingTimes(client *http.Client, semester util.Semester, crn string) ([]util.CourseTime, []string, error) {
	resp, err := client.Get("https://self-service.mun.ca/StudentRegistrationSsb/ssb/searchResults/getFacultyMeetingTimes?term=" + strconv.Itoa(semester.ID) + "&courseReferenceNumber=" + crn)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, nil, errors.New("Recieved status code " + strconv.Itoa(resp.StatusCode) + " when requesting course info")
	}

	body, err := util.ReadBody(resp.Body)
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
				BuildingDescription    string `json:"buildingDescription"`
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
	if meetingTimes.Fmt == nil {
		return nil, nil, errors.New("Self-Service returned no meeting-times payload")
	}

	instructors := make([]string, 0)
	courseTimes := make([]util.CourseTime, 0)

	for _, meetingTime := range meetingTimes.Fmt {
		for _, instructor := range meetingTime.Faculty {
			name := html.UnescapeString(instructor.DisplayName)
			if name != "" && !slices.Contains(instructors, name) {
				instructors = append(instructors, name)
			}
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
			MeetingDetails: util.MeetingDetails{
				StartDate:    meetingTime.MeetingTime.StartDate,
				EndDate:      meetingTime.MeetingTime.EndDate,
				BuildingName: html.UnescapeString(meetingTime.MeetingTime.BuildingDescription),
			},
			StartTime: bannerTime(meetingTime.MeetingTime.BeginTime),
			EndTime:   bannerTime(meetingTime.MeetingTime.EndTime),
			Days:      &days,
			Location:  strings.TrimSpace(html.UnescapeString(meetingTime.MeetingTime.Building + " " + meetingTime.MeetingTime.Room)),
			DateRange: meetingTime.MeetingTime.StartDate + " - " + meetingTime.MeetingTime.EndDate,
			Type:      html.UnescapeString(meetingTime.MeetingTime.MeetingTypeDescription),
			CourseKey: strconv.Itoa(semester.ID) + crn,
		})
	}

	return courseTimes, instructors, nil
}

func bannerTime(value string) string {
	if len(value) != 4 {
		return "TBA"
	}
	parsed, err := time.Parse("1504", value)
	if err != nil {
		return "TBA"
	}
	return parsed.Format("15:04")
}

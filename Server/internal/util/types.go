package util

import "time"

type Semester struct {
	ID       int    `gorm:"primaryKey" json:"id"`
	Name     string `gorm:"not null" json:"name"`
	Latest   bool   `gorm:"not null" json:"latest"`
	Medicine bool   `gorm:"not null" json:"medicine"`
	MI       bool   `gorm:"not null" json:"mi"`
	ViewOnly bool   `gorm:"not null" json:"viewOnly"`
}

type Subject struct {
	ID   string `gorm:"primaryKey;not null" json:"code"`
	Name string `gorm:"not null" json:"description"`
}

type Course struct {
	CourseDetails
	Key        string   `gorm:"primaryKey"`
	ID         string   `gorm:"not null"`
	Name       string   `gorm:"not null"`
	CRN        string   `gorm:"not null"`
	Section    string   `gorm:"not null"`
	Credits    float32  `gorm:"not null"` // Self-Service creditHours for this section, including credit-bearing labs.
	Campus     string   `gorm:"not null"`
	Type       string   `gorm:"not null"`
	SubjectID  string   `gorm:"not null"`
	Subject    Subject  `gorm:"constraint:OnDelete:CASCADE;"`
	SemesterID int      `gorm:"not null"`
	Semester   Semester `gorm:"constraint:OnDelete:CASCADE;"`
}

// CourseDetails contains section metadata supplied by Memorial Self-Service.
// CreditsMin/Max preserve Self-Service's course-level creditHourLow/High.
// They must not override a section's Credits: lectures and zero-credit labs can share the same bounds.
type CourseDetails struct {
	InstructionalMethod string              `json:"instructionalMethod"`
	PartOfTerm          string              `json:"partOfTerm"`
	IsSectionLinked     bool                `gorm:"not null;default:false" json:"isSectionLinked"`
	LinkIdentifier      string              `json:"linkIdentifier"`
	CrossList           string              `json:"crossList"`
	CreditsMin          *float32            `json:"creditsMin"`
	CreditsMax          *float32            `json:"creditsMax"`
	Prerequisites       *CourseRequirements `gorm:"serializer:json;type:jsonb" json:"prerequisites"`
	Corequisites        *CourseRequirements `gorm:"serializer:json;type:jsonb" json:"corequisites"`
	LinkedSections      *LinkedSections     `gorm:"serializer:json;type:jsonb" json:"linkedSections"`
}

// LinkedSections contains alternative registration combinations from Banner.
// Every section in one option is required; options are alternatives, not cumulative.
// Nil means unfetched/failed; an empty Options slice is an explicit no-links response.
type LinkedSections struct {
	Options []LinkedSectionOption `json:"options"`
}

type LinkedSectionOption struct {
	Sections []LinkedSection `json:"sections"`
}

type LinkedSection struct {
	CRN     string `json:"crn"`
	Section string `json:"section"`
	Type    string `json:"type"`
	Title   string `json:"title"`
}

// CourseRequirements preserves Banner's requirement text and table expressions.
// A nil pointer means the requirements have not been fetched successfully.
type CourseRequirements struct {
	Text   string             `json:"text"`
	Tables []RequirementTable `json:"tables"`
}

type RequirementTable struct {
	Headers []string   `json:"headers"`
	Rows    [][]string `json:"rows"`
}

type CourseSeating struct {
	Semester           int       `gorm:"primaryKey" json:"semester"`
	CRN                string    `gorm:"primaryKey" json:"crn"`
	CourseKey          string    `gorm:"not null;uniqueIndex" json:"-"`
	Course             Course    `gorm:"constraint:OnDelete:CASCADE;" json:"-"`
	Seats              int       `json:"seats"`
	MaxSeats           int       `json:"maxSeats"`
	Waitlist           int       `json:"waitlist"`
	MaxWaitlist        int       `json:"maxWaitlist"`
	SeatsAvailable     *int      `json:"seatsAvailable"`
	WaitCount          *int      `json:"waitCount"`
	OpenSection        *bool     `json:"openSection"`
	CrossListCapacity  *int      `json:"crossListCapacity"`
	CrossListCount     *int      `json:"crossListCount"`
	CrossListAvailable *int      `json:"crossListAvailable"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

type CourseAPI struct {
	CourseDetails
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	CRN        string  `json:"crn"`
	Section    string  `json:"section"`
	Credits    float32 `json:"credits"`
	Campus     string  `json:"campus"`
	SubjectId  string  `json:"subject"`
	Instructor string  `json:"instructor"`
	Type       string  `json:"type"`
}

type CourseFrontendAPI struct {
	CourseDetails
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	CRN         string   `json:"crn"`
	Section     string   `json:"section"`
	Credits     float32  `json:"credits"`
	Campus      string   `json:"campus"`
	SubjectId   string   `json:"subject"`
	Instructors []string `json:"instructors"`
	Type        string   `json:"type"`
}

type CourseTime struct {
	MeetingDetails
	ID        int    `gorm:"primaryKey;autoIncrement"`
	StartTime string `gorm:"not null"`
	EndTime   string `gorm:"not null"`
	Days      *string
	Location  string `gorm:"not null"`
	DateRange string `gorm:"not null"`
	Type      string `gorm:"not null"`
	CourseKey string `gorm:"not null"`
	Course    Course `gorm:"constraint:OnDelete:CASCADE;"`
}

type MeetingDetails struct {
	StartDate    string `json:"startDate"`
	EndDate      string `json:"endDate"`
	BuildingName string `json:"buildingName"`
}

type CourseTimeAPI struct {
	MeetingDetails
	StartTime      string  `json:"startTime"`
	EndTime        string  `json:"endTime"`
	Days           *string `json:"days"`
	Location       string  `json:"location"`
	DateRange      string  `json:"dateRange"`
	Type           string  `json:"type"`
	CourseCRN      string  `json:"crn"`
	ProfessorNames *string `json:"professorNames"`
}

type CourseTimeICal struct {
	StartTime       string  `gorm:"column:start_time"`
	EndTime         string  `gorm:"column:end_time"`
	Days            *string `gorm:"column:days"`
	Location        string  `gorm:"column:location"`
	DateRange       string  `gorm:"column:date_range"`
	Type            string  `gorm:"column:type"`
	CourseKey       string  `gorm:"column:course_key"`
	CourseCRN       string  `gorm:"column:course_crn"`
	SemesterID      int     `gorm:"column:semester_id"`
	CourseID        string  `gorm:"column:course_id"`
	CourseName      string  `gorm:"column:course_name"`
	InstructorNames string  `gorm:"column:instructor_names"`
}

type CourseTimeFrontendAPI struct {
	MeetingDetails
	StartTime string  `json:"startTime"`
	EndTime   string  `json:"endTime"`
	Days      *string `json:"days"`
	Location  string  `json:"location"`
	DateRange string  `json:"dateRange"`
	Type      string  `json:"type"`
	CRN       string  `json:"crn"`
}

type Professor struct {
	Name string `gorm:"primaryKey"`
}

type CourseInstructor struct {
	ID            int       `gorm:"primaryKey;autoIncrement"`
	ProfessorName string    `gorm:"not null"`
	Professor     Professor `gorm:"constraint:OnDelete:CASCADE;"`
	CourseKey     string    `gorm:"not null"`
	Course        Course    `gorm:"constraint:OnDelete:CASCADE;"`
}

type CourseInstructorAPI struct {
	ProfessorName string `json:"name"`
	CRN           string `json:"crn"`
}

type ProfessorRating struct {
	ProfessorName string    `gorm:"primaryKey;not null" json:"name"`
	Professor     Professor `gorm:"constraint:OnDelete:CASCADE;"`
	Rating        float64   `gorm:"not null" json:"rating"`
	ID            int       `gorm:"not null" json:"id"`
	Difficulty    float64   `gorm:"not null" json:"difficulty"`
	RatingCount   int       `gorm:"not null" json:"ratings"`
	WouldRetake   float64   `gorm:"not null" json:"wouldRetake"`
}

type ProfessorRatingAPI struct {
	ProfessorName string  `json:"name"`
	Rating        float64 `json:"rating"`
	ID            int     `json:"id"`
	Difficulty    float64 `json:"difficulty"`
	RatingCount   int     `json:"ratings"`
	WouldRetake   float64 `json:"wouldRetake"`
}

type FrontendAPIResponse struct {
	Courses    []CourseFrontendAPI     `json:"courses"`
	Subjects   []Subject               `json:"subjects"`
	Times      []CourseTimeFrontendAPI `json:"times"`
	Seatings   []CourseSeating         `json:"seatings"`
	Professors []ProfessorRatingAPI    `json:"profs"`
}

type ScheduleAPIResponse struct {
	Courses []ScheduleCourseAPI     `json:"courses"`
	Times   []CourseTimeFrontendAPI `json:"times"`
}

type ScheduleCourseAPI struct {
	ID                  string  `json:"id"`
	Name                string  `json:"name"`
	CRN                 string  `json:"crn"`
	Section             string  `json:"section"`
	Credits             float32 `json:"credits"`
	Campus              string  `json:"campus"`
	SubjectID           string  `json:"subject"`
	Type                string  `json:"type"`
	InstructionalMethod string  `json:"instructionalMethod"`
}

type ExamTime struct {
	CRN        string   `gorm:"not null"`
	SemesterID int      `gorm:"column:semester;not null"`
	Semester   Semester `gorm:"constraint:OnDelete:CASCADE;"`
	CourseKey  string   `gorm:"primaryKey;not null"`
	Course     Course   `gorm:"constraint:OnDelete:CASCADE;"`
	Time       string   `gorm:"not null"`
	Location   string   `gorm:"not null"`
}

type ExamTimeAPI struct {
	CRN      string `json:"crn"`
	Time     string `json:"time"`
	Location string `json:"location"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

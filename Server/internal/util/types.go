package util

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
	Key        string   `gorm:"primaryKey"`
	ID         string   `gorm:"not null"`
	Name       string   `gorm:"not null"`
	CRN        string   `gorm:"not null"`
	Section    string   `gorm:"not null"`
	Credits    float32  `gorm:"not null"`
	Campus     string   `gorm:"not null"`
	Type       string   `gorm:"not null"`
	SubjectID  string   `gorm:"not null"`
	Subject    Subject  `gorm:"constraint:OnDelete:CASCADE;"`
	SemesterID int      `gorm:"not null"`
	Semester   Semester `gorm:"constraint:OnDelete:CASCADE;"`
}

type CourseSeating struct {
	Semester string      `json:"semester"`
	CRN      string      `json:"crn"`
	Seats    SeatingInfo `json:"seats"`
	Waitlist SeatingInfo `json:"waitlist"`
}

type CourseSeatingResponse struct {
	CRN      string      `json:"crn"`
	Seats    SeatingInfo `json:"seats"`
	Waitlist SeatingInfo `json:"waitlist"`
}

type SeatingInfo struct {
	Capacity  int `json:"capacity"`
	Actual    int `json:"actual"`
	Remaining int `json:"remaining"`
}

type CourseAPI struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	CRN        string  `json:"crn"`
	Section    string  `json:"section"`
	Credits    float32 `json:"credits"`
	Campus     string  `json:"campus"`
	SubjectId  string  `json:"subject"`
	Instructor string  `json:"instructor"`
}

type CourseFrontendAPI struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	CRN        string  `json:"crn"`
	Section    string  `json:"section"`
	Credits    float32 `json:"credits"`
	Campus     string  `json:"campus"`
	SubjectId  string  `json:"subject"`
	Instructor string  `gorm:"column:instructor" json:"instructor"`
	Type       string  `json:"type"`
}

type CourseTime struct {
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

type CourseTimeAPI struct {
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
	ID              int
	StartTime       string
	EndTime         string
	Days            *string
	Location        string
	DateRange       string
	Type            string
	CourseKey       string
	CourseCRN       string
	SemesterID      int
	CourseID        string
	CourseName      string
	InstructorNames string
}

type CourseTimeFrontendAPI struct {
	StartTime string  `json:"startTime"`
	EndTime   string  `json:"endTime"`
	Days      *string `json:"days"`
	Location  string  `json:"location"`
	DateRange string  `json:"dateRange"`
	Type      string  `json:"type"`
	CourseKey string  `json:"key"`
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
	Courses  []CourseFrontendAPI     `json:"courses"`
	Subjects []Subject               `json:"subjects"`
	Times    []CourseTimeFrontendAPI `json:"times"`
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

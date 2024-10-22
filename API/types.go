package main

type Semester struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Latest   bool   `json:"latest"`
	ViewOnly bool   `json:"viewOnly"`
	Medical  bool   `json:"medical"`
	MI       bool   `json:"mi"`
}

type Subject struct {
	Name         string `json:"name"`
	FriendlyName string `json:"friendlyName"`
}

type Course struct {
	Crn         string `json:"crn"`
	Id          string `json:"id"`
	Name        string `json:"name"`
	Section     string `json:"section"`
	DateRange   any    `json:"dateRange"`
	CourseType  any    `json:"type"`
	Instructor  any    `json:"instructor"`
	SubjectFull string `json:"subjectFull"`
	Subject     string `json:"subject"`
	Campus      string `json:"campus"`
	Comment     any    `json:"comment"`
	Credits     int    `json:"credits"`
	Semester    int    `json:"semester"`
	Level       string `json:"level"`
	Identifier  string `json:"identifier"`
}

type Time struct {
	Identifier string `json:"identifier"`
	Crn        string `json:"crn"`
	Days       string `json:"days"`
	StartTime  string `json:"startTime"`
	EndTime    string `json:"endTime"`
	Location   string `json:"location"`
	Type       string `json:"courseType"`
	Semester   int    `json:"semester"`
}

type Seating struct {
	Crn        string `json:"crn"`
	Available  string `json:"available"`
	Max        string `json:"max"`
	Waitlist   any    `json:"waitlist"`
	Checked    string `json:"checked"`
	Identifier string `json:"identifier"`
	Semester   int    `json:"semester"`
}

type Professor struct {
	Name        string  `json:"name"`
	Rating      float64 `json:"rating"`
	Id          int     `json:"id"`
	Difficulty  float64 `json:"difficulty"`
	RatingCount int     `json:"ratings"`
	WouldRetake float64 `json:"wouldRetake"`
}

type ExamTime struct {
	Id       string `json:"id"`
	Section  string `json:"section"`
	Crn      string `json:"crn"`
	Time     string `json:"time"`
	Location string `json:"location"`
}

type EngSeats struct {
	Id         int    `json:"id"`
	Subject    string `json:"subject"`
	Name       string `json:"name"`
	Course     string `json:"course"`
	Section    string `json:"section"`
	Registered int    `json:"registered"`
	Date       string `json:"date"`
	Semester   int    `json:"semester"`
}

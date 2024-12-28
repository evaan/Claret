package main

type Semester struct {
	ID       int    `gorm:"not null"`
	Name     string `gorm:"not null"`
	Latest   bool   `gorm:"not null"`
	ViewOnly bool   `gorm:"column:viewOnly;not null"`
	Medical  bool   `gorm:"not null"`
	MI       bool   `gorm:"not null"`
	Scraped  bool   `gorm:"not null"`
}

type Subject struct {
	Name         string `gorm:"primaryKey;not null"`
	FriendlyName string `gorm:"not null"`
}

type Course struct {
	CRN         string  `gorm:"not null"`
	Id          string  `gorm:"not null"`
	Name        string  `gorm:"not null"`
	Section     string  `gorm:"not null"`
	DateRange   *string `gorm:"column:dateRange"`
	Type        *string
	Instructor  *string
	Subject     string `gorm:"column:subject;not null"`
	SubjectFull string `gorm:"column:subjectFull;not null"`
	Campus      string `gorm:"not null"`
	Comment     *string
	Credits     int      `gorm:"not null"`
	SemesterID  int      `gorm:"column:semester;not null"`
	Semester    Semester `gorm:"constraint:OnDelete:CASCADE;"`
	Level       string   `gorm:"not null"`
	Identifier  string   `gorm:"primaryKey"`
}

type CourseTime struct {
	ID               int      `gorm:"primaryKey;autoIncrement"`
	CRN              string   `gorm:"not null"`
	Days             string   `gorm:"not null"`
	StartTime        string   `gorm:"column:startTime;not null"`
	EndTime          string   `gorm:"column:endTime;not null"`
	Location         string   `gorm:"not null"`
	Type             string   `gorm:"not null"`
	SemesterID       int      `gorm:"column:semester;not null"`
	Semester         Semester `gorm:"constraint:OnDelete:CASCADE;"`
	CourseIdentifier string   `gorm:"column:identifier"`
	Course           Course   `gorm:"constraint:OnDelete:CASCADE;"`
}

type ProfAndSemester struct {
	ID         int      `gorm:"primaryKey;autoIncrement"`
	Name       string   `gorm:"not null"`
	SemesterID int      `gorm:"column:semester;not null"`
	Semester   Semester `gorm:"constraint:OnDelete:CASCADE;"`
}

func (CourseTime) TableName() string {
	return "times"
}

type Seating struct {
	Identifier string `gorm:"primaryKey"`
	Crn        string `gorm:"not null"`
	Available  int    `gorm:"not null"`
	Max        int    `gorm:"not null"`
	Waitlist   int
	Checked    string   `gorm:"not null"`
	SemesterID int      `gorm:"column:semester;not null"`
	Semester   Semester `gorm:"constraint:OnDelete:CASCADE;"`
}

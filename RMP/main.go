package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"

	_ "github.com/joho/godotenv/autoload"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

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

type School struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type Professor struct {
	Name        string  `gorm:"primaryKey,not null"`
	Rating      float64 `gorm:"not null"`
	Id          int     `gorm:"not null"`
	Difficulty  float64 `gorm:"not null"`
	RatingCount int     `gorm:"not null"`
	WouldRetake float64 `gorm:"not null"`
}

type RateMyProfNode struct {
	Difficulty       float64 `json:"avgDifficulty"`
	Rating           float64 `json:"avgRating"`
	Department       string  `json:"department"`
	FirstName        string  `json:"firstName"`
	ID               string  `json:"id"`
	Saved            bool    `json:"isSaved"`
	LastName         string  `json:"lastName"`
	LegacyID         int     `json:"legacyId"`
	Ratings          int     `json:"numRatings"`
	RetakePercentage float64 `json:"wouldTakeAgainPercent"`
}

type Edge struct {
	Professor RateMyProfNode `json:"node"`
}

type Teachers struct {
	Edges []Edge `json:"edges"`
}

type Search struct {
	Teachers Teachers `json:"teachers"`
}

type Data struct {
	Search Search `json:"search"`
}

type Root struct {
	Data Data `json:"data"`
}

type MatchedProf struct {
	MunName  string
	RmpProf  Professor
	Distance float64
}

var db *gorm.DB

func main() {
	jsonData, err := os.ReadFile("rmp.json")
	if err != nil {
		fmt.Println("Error reading JSON file:", err)
		return
	}

	db, _ = gorm.Open(postgres.Open(os.Getenv("DB_URL")), &gorm.Config{})

	db.AutoMigrate(&Professor{})

	var root Root
	err = json.Unmarshal(jsonData, &root)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return
	}

	var profs []string
	munProfDepts := make(map[string][]string)
	rows, err := db.Raw("select instructor, \"subjectFull\" from courses").Rows()
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		var prof sql.NullString
		var department string
		rows.Scan(&prof, &department)
		if prof.Valid && (prof.String != "TBA" && !slices.Contains(profs, prof.String)) {
			munProfDepts[prof.String] = append(munProfDepts[prof.String], department)
			profs = append(profs, strings.Split(prof.String, ", ")...)
		}
	}

	var munNames []string
	var rmpProfs []string

	for _, edge := range root.Data.Search.Teachers.Edges {
		if edge.Professor.Ratings > 0 {
			prof := edge.Professor
			matchedProf := closestName(profs, prof.FirstName+" "+prof.LastName)
			if matchedProf.Name != "" {
				if !slices.Contains(munNames, matchedProf.Name) {
					munNames = append(munNames, matchedProf.Name)
				}
				rmpProfs = append(rmpProfs, prof.FirstName+" "+prof.LastName)
			}
		}
	}

	for _, prof := range munNames {
		matchedProf := closestName(rmpProfs, prof)
		fmt.Printf("MUN Name: %s RMP Name: %s Distance: %f\n", prof, matchedProf.Name, matchedProf.Distance)

	}
}

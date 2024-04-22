package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"strings"

	_ "github.com/joho/godotenv/autoload"
)

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

func rmp() {
	logger.Println("⭐ RMP Scraping Started!")

	body := []byte(`{"query":"query TeacherSearchResultsPageQuery(\n  $query: TeacherSearchQuery!\n  $schoolID: ID\n  $includeSchoolFilter: Boolean!\n) {\n  search: newSearch {\n    ...TeacherSearchPagination_search_1ZLmLD\n  }\n  school: node(id: $schoolID) @include(if: $includeSchoolFilter) {\n    __typename\n    ... on School {\n      name\n    }\n    id\n  }\n}\n\nfragment TeacherSearchPagination_search_1ZLmLD on newSearch {\n  teachers(query: $query, first: 999999, after: \"\") {\n    didFallback\n    edges {\n      cursor\n      node {\n        ...TeacherCard_teacher\n        id\n        __typename\n      }\n    }\n    pageInfo {\n      hasNextPage\n      endCursor\n    }\n    resultCount\n    filters {\n      field\n      options {\n        value\n        id\n      }\n    }\n  }\n}\n\nfragment TeacherCard_teacher on Teacher {\n  id\n  legacyId\n  avgRating\n  numRatings\n  ...CardFeedback_teacher\n  ...CardSchool_teacher\n  ...CardName_teacher\n  ...TeacherBookmark_teacher\n}\n\nfragment CardFeedback_teacher on Teacher {\n  wouldTakeAgainPercent\n  avgDifficulty\n}\n\nfragment CardSchool_teacher on Teacher {\n  department\n  school {\n    name\n    id\n  }\n}\n\nfragment CardName_teacher on Teacher {\n  firstName\n  lastName\n}\n\nfragment TeacherBookmark_teacher on Teacher {\n  id\n  isSaved\n}\n","variables":{"query":{"text":"","schoolID":"U2Nob29sLTE0NDE=","fallback":true,"departmentID":null},"schoolID":"U2Nob29sLTE0NDE=","includeSchoolFilter":true}}`)

	req, err := http.NewRequest("POST", "https://www.ratemyprofessors.com/graphql", bytes.NewBuffer(body))
	req.Header.Add("Authorization", "Basic dGVzdDp0ZXN0")
	if err != nil {
		logger.Fatal(err)
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		logger.Println("❌ RMP Failed scraping due to: " + err.Error())
		return
	}
	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		logger.Panic(err)
	}

	file, _ := os.Create("rmp.json")
	file.Write(resBody)
	file.Close()

	var root Root
	err = json.Unmarshal(resBody, &root)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return
	}

	var profs []string
	rows, err := db.Raw("select distinct instructor from courses").Rows()
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		var prof sql.NullString
		rows.Scan(&prof)
		if prof.Valid && (prof.String != "TBA" && !slices.Contains(profs, prof.String)) {
			profs = append(profs, strings.Split(prof.String, ", ")...)
		}
	}

	var munNames []string
	var rmpProfs []string
	rmpMap := make(map[string]Professor)

	for _, edge := range root.Data.Search.Teachers.Edges {
		if edge.Professor.Ratings > 0 {
			prof := edge.Professor
			matchedProf := closestName(profs, prof.FirstName+" "+prof.LastName)
			if matchedProf.Name != "" {
				if !slices.Contains(munNames, matchedProf.Name) {
					munNames = append(munNames, matchedProf.Name)
				}
				rmpProfs = append(rmpProfs, prof.FirstName+" "+prof.LastName)
				rmpMap[prof.FirstName+" "+prof.LastName] = Professor{prof.FirstName + " " + prof.LastName, prof.Rating, prof.LegacyID, prof.Difficulty, prof.Ratings, prof.RetakePercentage}
			}
		}
	}

	for _, prof := range munNames {
		matchedProf := closestName(rmpProfs, prof)
		matchedProf1 := rmpMap[matchedProf.Name]
		matchedProf1.Name = prof
		db.Save(&matchedProf1)
	}

	logger.Println("✅ RMP Scrape Complete!")
}

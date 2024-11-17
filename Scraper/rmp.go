package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

type PageInfo struct {
	Cursor      string `json:"endCursor"`
	HasNextPage bool   `json:"hasNextPage"`
}

type Teachers struct {
	Edges []Edge   `json:"edges"`
	Info  PageInfo `json:"pageInfo"`
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

	cursor := "null"
	hasNextPage := true

	for hasNextPage {
		body := []byte(fmt.Sprintf(`{
			"query": "query TeacherSearchPaginationQuery($cursor: String, $query: TeacherSearchQuery!) { search: newSearch { teachers(query: $query, first: 1000, after: $cursor) { didFallback edges { cursor node { ...TeacherCard_teacher id __typename } } pageInfo { hasNextPage endCursor } resultCount filters { field options { value id } } } } } fragment TeacherCard_teacher on Teacher { id legacyId avgRating numRatings ...CardFeedback_teacher ...CardSchool_teacher ...CardName_teacher ...TeacherBookmark_teacher } fragment CardFeedback_teacher on Teacher { wouldTakeAgainPercent avgDifficulty } fragment CardSchool_teacher on Teacher { department school { name id } } fragment CardName_teacher on Teacher { firstName lastName } fragment TeacherBookmark_teacher on Teacher { id isSaved }",
			"variables": {
				"cursor": "%s",
				"query": {
				"text": "",
				"schoolID": "U2Nob29sLTE0NDE=",
				"fallback": true
				}
			}
		}`, cursor))

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

		cursor = root.Data.Search.Teachers.Info.Cursor
		hasNextPage = root.Data.Search.Teachers.Info.HasNextPage
	}

	logger.Println("✅ RMP Scrape Complete!")
}

package scrapers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/evaan/Claret/internal/util"
)

type Node struct {
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
	Professor Node `json:"node"`
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
	RmpProf  util.ProfessorRating
	Distance float64
}

func RMP(logger *log.Logger, professors []string) []util.ProfessorRating {
	cursor := "null"
	hasNextPage := true

	var finalRatings []util.ProfessorRating

	seen := make(map[string]struct{})
	var munProfs []string
	for _, p := range professors {
		p = strings.TrimSpace(p)
		if p != "" {
			if _, ok := seen[p]; !ok {
				seen[p] = struct{}{}
				munProfs = append(munProfs, p)
			}
		}
	}

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
		if err != nil {
			logger.Printf("Error scraping RMP: %s\n", err.Error())
			util.SendErrorToWebhook(os.Getenv("SCRAPER_WEBHOOK_URL"), err)
			return finalRatings
		}
		req.Header.Add("Authorization", "Basic dGVzdDp0ZXN0")

		client := &http.Client{}
		res, err := client.Do(req)
		if err != nil {
			logger.Println("❌ RMP failed to fetch:", err)
			return finalRatings
		}
		defer res.Body.Close()

		resBody, err := io.ReadAll(res.Body)
		if err != nil {
			logger.Panic(err)
		}

		var root Root
		if err := json.Unmarshal(resBody, &root); err != nil {
			logger.Println("❌ Error unmarshalling JSON:", err)
			return finalRatings
		}

		for _, edge := range root.Data.Search.Teachers.Edges {
			prof := edge.Professor
			if prof.Ratings == 0 {
				continue
			}

			rmpName := strings.TrimSpace(prof.FirstName + " " + prof.LastName)
			match := util.ClosestName(munProfs, rmpName)
			if match.Name == "" {
				continue
			}

			finalRatings = append(finalRatings, util.ProfessorRating{
				ProfessorName: match.Name,
				Rating:        prof.Rating,
				ID:            prof.LegacyID,
				Difficulty:    prof.Difficulty,
				RatingCount:   prof.Ratings,
				WouldRetake:   prof.RetakePercentage,
			})
		}

		cursor = root.Data.Search.Teachers.Info.Cursor
		hasNextPage = root.Data.Search.Teachers.Info.HasNextPage
	}

	return finalRatings
}

package scrapers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/evaan/Claret/internal/util"
)

const rmpSchoolID = "U2Nob29sLTE0NDE="

const rmpQuery = `query TeacherSearchPaginationQuery($cursor: String, $query: TeacherSearchQuery!) {
	search: newSearch {
		teachers(query: $query, first: 100, after: $cursor) {
			edges { node {
				legacyId firstName lastName avgRating avgDifficulty numRatings wouldTakeAgainPercent
				school { id }
			} }
			pageInfo { hasNextPage endCursor }
		}
	}
}`

type rmpTeacher struct {
	FirstName        string  `json:"firstName"`
	LastName         string  `json:"lastName"`
	LegacyID         int     `json:"legacyId"`
	Rating           float64 `json:"avgRating"`
	Difficulty       float64 `json:"avgDifficulty"`
	Ratings          int     `json:"numRatings"`
	RetakePercentage float64 `json:"wouldTakeAgainPercent"`
	School           struct {
		ID string `json:"id"`
	} `json:"school"`
}

func RMP(logger *log.Logger, professors []string) []util.ProfessorRating {
	teachers, err := fetchRMP(&http.Client{Timeout: 30 * time.Second})
	if err != nil {
		logger.Printf("❌ RMP scrape failed; existing ratings retained: %v", err)
		return nil
	}
	ratings := matchRMP(professors, teachers)
	logger.Printf("⭐ RMP: fetched %d rated Memorial profiles, matched %d instructor names", len(teachers), len(ratings))
	return ratings
}

func fetchRMP(client *http.Client) ([]rmpTeacher, error) {
	var cursor *string // GraphQL expects null, not the literal string "null".
	seenCursors := make(map[string]bool)
	seenIDs := make(map[int]bool)
	teachers := make([]rmpTeacher, 0)
	for {
		body, err := json.Marshal(map[string]any{
			"query": rmpQuery,
			"variables": map[string]any{
				"cursor": cursor,
				"query":  map[string]any{"text": "", "schoolID": rmpSchoolID, "fallback": false},
			},
		})
		if err != nil {
			return nil, err
		}
		req, err := http.NewRequest(http.MethodPost, "https://www.ratemyprofessors.com/graphql", bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Basic dGVzdDp0ZXN0")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", "Mozilla/5.0")
		req.Header.Set("Origin", "https://www.ratemyprofessors.com")
		req.Header.Set("Referer", "https://www.ratemyprofessors.com/")
		res, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("fetch page: %w", err)
		}
		data, err := io.ReadAll(io.LimitReader(res.Body, 8<<20))
		res.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("read page: %w", err)
		}
		if res.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("upstream HTTP %d", res.StatusCode)
		}
		var root struct {
			Errors []struct{ Message string } `json:"errors"`
			Data   struct {
				Search struct {
					Teachers *struct {
						Edges []struct {
							Node rmpTeacher `json:"node"`
						} `json:"edges"`
						PageInfo struct {
							HasNextPage bool   `json:"hasNextPage"`
							EndCursor   string `json:"endCursor"`
						} `json:"pageInfo"`
					} `json:"teachers"`
				} `json:"search"`
			} `json:"data"`
		}
		if err := json.Unmarshal(data, &root); err != nil {
			return nil, fmt.Errorf("decode page: %w", err)
		}
		if len(root.Errors) > 0 {
			return nil, fmt.Errorf("GraphQL: %s", root.Errors[0].Message)
		}
		page := root.Data.Search.Teachers
		if page == nil {
			return nil, fmt.Errorf("response is missing teacher results")
		}
		for _, edge := range page.Edges {
			teacher := edge.Node
			if teacher.School.ID != rmpSchoolID || teacher.LegacyID <= 0 || teacher.Ratings <= 0 || seenIDs[teacher.LegacyID] {
				continue
			}
			seenIDs[teacher.LegacyID] = true
			teachers = append(teachers, teacher)
		}
		if !page.PageInfo.HasNextPage {
			return teachers, nil
		}
		next := page.PageInfo.EndCursor
		if len(page.Edges) == 0 || next == "" || seenCursors[next] {
			return nil, fmt.Errorf("pagination did not advance")
		}
		seenCursors[next] = true
		cursor = &next
	}
}

func matchRMP(professors []string, teachers []rmpTeacher) []util.ProfessorRating {
	names := make([]string, len(teachers))
	byName := make(map[string]rmpTeacher, len(teachers))
	for i, teacher := range teachers {
		names[i] = strings.TrimSpace(teacher.FirstName + " " + teacher.LastName)
		byName[names[i]] = teacher
	}
	ratings := make([]util.ProfessorRating, 0)
	seen := make(map[string]bool)
	for _, name := range professors {
		if strings.TrimSpace(name) == "" || seen[name] {
			continue
		}
		seen[name] = true
		match := util.ClosestName(names, name)
		if match.Name == "" {
			continue
		}
		teacher := byName[match.Name]
		ratings = append(ratings, util.ProfessorRating{
			ProfessorName: name, ID: teacher.LegacyID, Rating: teacher.Rating,
			Difficulty: teacher.Difficulty, RatingCount: teacher.Ratings,
			WouldRetake: teacher.RetakePercentage,
		})
	}
	return ratings
}

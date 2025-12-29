package scrapers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/evaan/Claret/internal/util"
)

func GetSemesters() ([]util.Semester, error) {
	resp, err := http.Get("https://self-service.mun.ca/StudentRegistrationSsb/ssb/classSearch/getTerms?searchTerm=&offset=1&max=0")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New("Recieved status code " + strconv.Itoa(resp.StatusCode) + " when requesting semesters")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var semestersRes []struct {
		Code        string `json:"code"`
		Description string `json:"description"`
	}

	err = json.Unmarshal(body, &semestersRes)
	if err != nil {
		return nil, err
	}

	foundLatest := false
	var semesters []util.Semester

	for _, semester := range semestersRes {
		semesterInt, err := strconv.Atoi(semester.Code)
		if err != nil {
			return nil, err
		}
		semesters = append(semesters, util.Semester{
			ID:       semesterInt,
			Name:     strings.Replace(semester.Description, " (View only)", "", 1),
			Latest:   !foundLatest && !strings.Contains(semester.Description, "M"),
			Medicine: strings.Contains(semester.Description, "Medicine"),
			MI:       strings.Contains(semester.Description, "MI"),
			ViewOnly: strings.Contains(semester.Description, "(View only)"),
		})
		if !foundLatest && !strings.Contains(semester.Description, "M") {
			foundLatest = true
		}
	}

	return semesters, nil
}

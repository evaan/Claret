package scrapers

import (
	"encoding/json"
	"fmt"
	"html"
	"strconv"

	"github.com/evaan/Claret/internal/util"
)

func GetSubjects(semester util.Semester) ([]util.Subject, error) {
	resp, err := upstreamClient.Get("https://self-service.mun.ca/StudentRegistrationSsb/ssb/classSearch/get_subject?searchTerm=&term=" + strconv.Itoa(semester.ID) + "&offset=1&max=500")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("subjects returned status %d", resp.StatusCode)
	}
	body, err := util.ReadBody(resp.Body)
	if err != nil {
		return nil, err
	}

	var subjects []util.Subject

	err = json.Unmarshal(body, &subjects)
	if err != nil {
		return nil, err
	}

	for i := range subjects {
		subjects[i].Name = html.UnescapeString(subjects[i].Name)
	}

	return subjects, nil
}

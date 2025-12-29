package scrapers

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/evaan/Claret/internal/util"
)

func GetSubjects(semester util.Semester) ([]util.Subject, error) {
	resp, err := http.Get("https://self-service.mun.ca/StudentRegistrationSsb/ssb/classSearch/get_subject?searchTerm=&term=" + strconv.Itoa(semester.ID) + "&offset=1&max=500")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var subjects []util.Subject

	err = json.Unmarshal(body, &subjects)
	if err != nil {
		return nil, err
	}

	return subjects, nil
}

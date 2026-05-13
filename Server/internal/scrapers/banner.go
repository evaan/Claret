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

func GetJsession() (string, error) {
	resp, err := http.Get("https://self-service.mun.ca/StudentRegistrationSsb")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", errors.New("Recieved status code " + strconv.Itoa(resp.StatusCode) + " when requesting JSESSION")
	}

	for _, c := range resp.Cookies() {
		if c.Name == "JSESSIONID" {
			return c.Value, nil
		}
	}

	return "", errors.New("No JSESSION found")
}

func SaveTerm(client *http.Client, semester util.Semester) error {
	resp, err := client.Get("https://self-service.mun.ca/StudentRegistrationSsb/ssb/term/saveTerm?mode=search&term=" + strconv.Itoa(semester.ID) + "&uniqueSessionId=claret")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.New("Recieved status code " + strconv.Itoa(resp.StatusCode) + " when saving term")
	}

	return nil
}

func SendSearch(client *http.Client, semester util.Semester) error {
	reqBody := "term=" + strconv.Itoa(semester.ID) + "&studyPath=&studyPathText=&startDatepicker=&endDatepicker=&uniqueSessionId=claret"

	resp, err := client.Post("https://self-service.mun.ca/StudentRegistrationSsb/ssb/term/search?mode=search", "application/x-www-form-urlencoded", strings.NewReader(reqBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.New("Recieved status code " + strconv.Itoa(resp.StatusCode) + " when sending search request")
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var respJson struct {
		RegAllowed          bool     `json:"regAllowed"`
		StudentEligFailures []string `json:"studentEligFailures"`
		FwdURL              string   `json:"fwdURL"`
	}

	err = json.Unmarshal(respBody, &respJson)
	if err != nil {
		return err
	}

	if len(respJson.StudentEligFailures) > 0 {
		return errors.New("Unexpected search response: " + strings.Join(respJson.StudentEligFailures, ", "))
	}

	return nil
}

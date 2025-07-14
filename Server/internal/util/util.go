package util

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func MapToBytes(data map[string]any) []byte {
	if len(data) == 0 {
		return []byte{}
	}

	var b bytes.Buffer

	for key, value := range data {
		escapedKey := url.QueryEscape(key)

		switch v := value.(type) {
		case string:
			b.WriteString(escapedKey)
			b.WriteByte('=')
			b.WriteString(url.QueryEscape(v))
			b.WriteByte('&')
		case []string:
			for _, item := range v {
				b.WriteString(escapedKey)
				b.WriteByte('=')
				b.WriteString(url.QueryEscape(item))
				b.WriteByte('&')
			}
		}
	}

	buf := b.Bytes()

	return buf[:len(buf)-1]
}

var buildingCodeMap = map[string]string{
	"Arts and Administration Bldg": "A",
	"Henrietta Harvey Bldg":        "HH",
	"Business Administration Bldg": "BN",
	"INCO Innovation Centre":       "IIC",
	"Biotechnology Bldg":           "BT",
	"St. John's College":           "J",
	"Chemistry - Physics Bldg":     "C",
	"Core Science Facility":        "CSF",
	"M. O. Morgan Bldg":            "MU",
	"Computing Services":           "CS",
	"Physical Education Bldg":      "PE",
	"G. A. Hickman Bldg":           "ED",
	"Queen's College":              "QC",
	"Queen Elizabeth II Library":   "L",
	"S. J. Carew Bldg.":            "EN",
	"Science Bldg":                 "S",
	"Alexander Murray Bldg":        "ER",
	"Health Sciences Centre":       "H",
	"Coughlan College":             "CL",
	"Marine Institute":             "MI",
	"Center for Nursing Studies":   "N",
	"Arts and Science (SWGC)":      "AS",
	"Fine Arts (SWGC)":             "FA",
	"Forest Centre":                "FC",
	"Library/Computing (SWGC)":     "LC",
	"Western Memorial Hospital":    "WMH",
}

func ReplaceBuildingName(location string) string {
	for longName, shortCode := range buildingCodeMap {
		if strings.Contains(location, longName) {
			return strings.Replace(location, longName, shortCode, 1)
		}
	}
	return location
}

func GetCredits(logger *log.Logger, line string) int {
	var credits []int
	for _, segment := range strings.Split(strings.TrimSuffix(line, " Credits"), " ") {
		if segment != "OR" && strings.TrimSpace(segment) != "" {
			credit, err := strconv.ParseFloat(segment, 64)
			if err != nil {
				logger.Printf("Error parsing credits: %s\n", err.Error())
				SendErrorToWebhook(os.Getenv("SCRAPER_WEBHOOK_URL"), err)
				continue
			}
			credits = append(credits, int(credit))
		}
	}
	return slices.Max(credits)
}

func Unique[T comparable](input []T) []T {
	seen := make(map[T]struct{})
	result := make([]T, 0, len(input))
	for _, v := range input {
		if _, exists := seen[v]; !exists {
			seen[v] = struct{}{}
			result = append(result, v)
		}
	}
	return result
}

func GetEnvAsBool(key string) bool {
	output, err := strconv.ParseBool(os.Getenv(key))
	if err != nil {
		fmt.Println("Error parsing environment variable:", err)
		return false
	}
	return output
}

func GetEnvAsInt(key string) int {
	output, err := strconv.Atoi(os.Getenv(key))
	if err != nil {
		fmt.Println("Error parsing environment variable:", err)
		return 0
	}
	return output
}

func GetParamOrQuery(c *gin.Context, key string) string {
	if query := strings.TrimSpace(c.Query(key)); query != "" {
		return query
	}
	return strings.TrimSpace(c.Param(key))
}

func GetParamOrQueryWithDefault(c *gin.Context, key string, fallback string) string {
	if query := strings.TrimSpace(c.Query(key)); query != "" {
		return query
	} else if param := strings.TrimSpace(c.Param(key)); param != "" {
		return param
	} else {
		return fallback
	}
}

func SendErrorToWebhook(webhookUrl string, err error) {
	params := fmt.Sprintf(`{"username":"Claret Scraper","embeds":[{"author":{"name":"Claret Scraper Error","url":"https://claretformun.com"},"timestamp":"%s","color":16711680,"fields":[{"name":"Error","value":"%s"}]}]}`, time.Now().Format(time.RFC3339), err.Error())
	r, err := http.NewRequest("POST", os.Getenv("SCRAPER_WEBHOOK_URL"), bytes.NewBuffer([]byte(params)))
	if err != nil {
		panic(err)
	}
	r.Header.Add("Content-Type", "application/json")
	client := &http.Client{}
	res, err := client.Do(r)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
}

var runeToWeekday = map[rune]time.Weekday{
	'M': time.Monday,
	'T': time.Tuesday,
	'W': time.Wednesday,
	'R': time.Thursday,
	'F': time.Friday,
	'S': time.Saturday,
	'U': time.Sunday,
}

var runeToICal = map[rune]string{
	'M': "MO",
	'T': "TU",
	'W': "WE",
	'R': "TH",
	'F': "FR",
	'S': "SA",
	'U': "SU",
}

func EarliestClassDate(start time.Time, compact string) time.Time {
	var earliest time.Time

	for _, char := range strings.ToUpper(compact) {
		targetDay, ok := runeToWeekday[char]
		if !ok {
			continue
		}

		offset := (int(targetDay) - int(start.Weekday()) + 7) % 7
		match := start.AddDate(0, 0, offset)

		if earliest.IsZero() || match.Before(earliest) {
			earliest = match
		}
	}

	return earliest
}

func LatestClassDate(start time.Time, compact string) time.Time {
	var latest time.Time

	for _, char := range strings.ToUpper(compact) {
		targetDay, ok := runeToWeekday[char]
		if !ok {
			continue
		}

		offset := (int(targetDay) - int(start.Weekday()) + 7) % 7
		match := start.AddDate(0, 0, -offset)

		if latest.IsZero() || match.After(latest) {
			latest = match
		}
	}

	return latest
}

func ICalRepeatDates(input string) string {
	var output []string

	for _, char := range strings.ToUpper(input) {
		if val, ok := runeToICal[char]; ok {
			output = append(output, val)
		}
	}

	return strings.Join(output, ",")
}

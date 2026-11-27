package scrapers

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/evaan/Claret/internal/util"
)

// These are section-specific, so fetch them during the full scrape, not the
// ten-minute seat refresh. Use the existing search session for both requests.
func GetPrerequisites(client *http.Client, semester util.Semester, crn string) (*util.CourseRequirements, error) {
	return getRequirements(client, semester, crn, "getSectionPrerequisites", "preReqs")
}

func GetCorequisites(client *http.Client, semester util.Semester, crn string) (*util.CourseRequirements, error) {
	return getRequirements(client, semester, crn, "getCorequisites", "coReqs")
}

func getRequirements(client *http.Client, semester util.Semester, crn, endpoint, sectionID string) (*util.CourseRequirements, error) {
	form := url.Values{
		"term":                  {strconv.Itoa(semester.ID)},
		"courseReferenceNumber": {crn},
		"first":                 {"first"},
	}
	resp, err := client.PostForm("https://self-service.mun.ca/StudentRegistrationSsb/ssb/searchResults/"+endpoint, form)
	if err != nil {
		return nil, fmt.Errorf("%s for %d/%s: %w", endpoint, semester.ID, crn, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s for %d/%s: status %d", endpoint, semester.ID, crn, resp.StatusCode)
	}
	body, err := util.ReadBody(resp.Body)
	if err != nil {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	section := doc.Find(`section[aria-labelledby="` + sectionID + `"]`)
	if section.Length() != 1 {
		return nil, fmt.Errorf("%s for %d/%s: missing requirements section", endpoint, semester.ID, crn)
	}
	// Hidden fields repeat the current course's title/number; they are not
	// requirements. Never store or render upstream HTML.
	section.Find("script, style, .notvisible, [hidden], [aria-hidden='true'], h3").Remove()
	result := &util.CourseRequirements{Tables: make([]util.RequirementTable, 0)}
	section.Find("table").Each(func(_ int, table *goquery.Selection) {
		parsed := util.RequirementTable{Headers: make([]string, 0), Rows: make([][]string, 0)}
		table.Find("thead tr").First().Find("th, td").Each(func(_ int, cell *goquery.Selection) {
			parsed.Headers = append(parsed.Headers, requirementText(cell.Text()))
		})
		table.Find("tr").Each(func(_ int, row *goquery.Selection) {
			if row.ParentsFiltered("thead").Length() > 0 {
				return
			}
			cells := make([]string, 0)
			row.ChildrenFiltered("td, th").Each(func(_ int, cell *goquery.Selection) {
				cells = append(cells, requirementText(cell.Text()))
			})
			if len(cells) > 0 {
				parsed.Rows = append(parsed.Rows, cells)
			}
		})
		if len(parsed.Rows) > 0 {
			result.Tables = append(result.Tables, parsed)
		}
	})
	section.Find("table").Remove()
	result.Text = requirementText(section.Text())
	if result.Text == "" && len(result.Tables) == 0 {
		return nil, fmt.Errorf("%s for %d/%s: empty requirements response", endpoint, semester.ID, crn)
	}
	return result, nil
}

func requirementText(text string) string {
	return strings.Join(strings.Fields(text), " ")
}

package scrapers

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/evaan/Claret/internal/util"
)

var linkedCRN = regexp.MustCompile(`^\d+$`)

// GetLinkedSections preserves Banner's options: all rows in one option must be
// registered together. It is only used during full scrapes of linked sections.
func GetLinkedSections(client *http.Client, semester util.Semester, crn string) (*util.LinkedSections, error) {
	fail := func(reason string) (*util.LinkedSections, error) {
		return nil, fmt.Errorf("getLinkedSections for %d/%s: %s", semester.ID, crn, reason)
	}
	resp, err := client.PostForm("https://self-service.mun.ca/StudentRegistrationSsb/ssb/searchResults/getLinkedSections", url.Values{
		"term": {strconv.Itoa(semester.ID)}, "courseReferenceNumber": {crn}, "first": {"first"},
	})
	if err != nil {
		return fail(err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fail(fmt.Sprintf("status %d", resp.StatusCode))
	}
	body, err := util.ReadBody(resp.Body)
	if err != nil {
		return fail(err.Error())
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return fail(err.Error())
	}
	section := doc.Find(`section[aria-labelledby="linked"]`)
	if section.Length() != 1 {
		return fail("missing linked sections")
	}
	section.Find("script, style, .notvisible, [hidden], [aria-hidden='true']").Remove()
	result := &util.LinkedSections{Options: make([]util.LinkedSectionOption, 0)}
	if section.Find("table").Length() == 0 &&
		strings.EqualFold(requirementText(section.Text()), "No linked course information available.") {
		return result, nil
	}
	invalid := false
	section.Find("table").Each(func(_ int, table *goquery.Selection) {
		pending := false
		columnsValid := false
		table.ChildrenFiltered("thead, tbody").Each(func(_ int, group *goquery.Selection) {
			if group.Is("thead") {
				if group.Find(".linkedSectionsAddAll").Length() > 0 {
					if pending || group.Find(".linkedSectionsAddAll").Length() != 1 {
						invalid = true
					}
					pending = true
					columnsValid = false
				}
				if headers := group.Find("th"); headers.Length() > 0 {
					want := []string{"Title", "Schedule Type", "Section", "CRN"}
					columnsValid = headers.Length() == len(want)
					headers.Each(func(i int, header *goquery.Selection) {
						if i >= len(want) || !strings.EqualFold(requirementText(header.Text()), want[i]) {
							columnsValid = false
						}
					})
				}
				return
			}
			if !pending || !columnsValid {
				invalid = true
			}
			pending = false
			option := util.LinkedSectionOption{Sections: make([]util.LinkedSection, 0)}
			group.ChildrenFiltered("tr").Each(func(_ int, row *goquery.Selection) {
				cells := row.ChildrenFiltered("td")
				if cells.Length() != 4 {
					invalid = true
					return
				}
				entry := util.LinkedSection{
					Title: requirementText(cells.Eq(0).Text()), Type: requirementText(cells.Eq(1).Text()),
					Section: requirementText(cells.Eq(2).Text()), CRN: requirementText(cells.Eq(3).Text()),
				}
				if !linkedCRN.MatchString(entry.CRN) || entry.Section == "" || entry.Type == "" || entry.Title == "" {
					invalid = true
					return
				}
				option.Sections = append(option.Sections, entry)
			})
			if len(option.Sections) == 0 {
				invalid = true
			}
			result.Options = append(result.Options, option)
		})
		if pending {
			invalid = true
		}
	})
	// Never use a partially parsed option to decide that a schedule is complete.
	if invalid || len(result.Options) == 0 {
		return fail("unrecognized or empty linked-section options")
	}
	return result, nil
}

package scrapers

import (
	"log"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/evaan/Claret/internal/util"
	"github.com/gocolly/colly/v2"
)

func GetSemesters(logger *log.Logger) []util.Semester {
	c := colly.NewCollector()

	var semesters []util.Semester
	foundLatest := false

	c.OnHTML("select[name=p_term]", func(e *colly.HTMLElement) {
		e.DOM.Find("option").Each(func(i int, s *goquery.Selection) {
			name := s.Text()
			if name != "None" {
				id, exists := s.Attr("value")
				if exists {
					idInt, err := strconv.Atoi(id)
					if err != nil {
						logger.Printf("Error scraping semester: %s\n", err.Error())
					} else {
						semesters = append(semesters, util.Semester{
							ID:       idInt,
							Name:     strings.Replace(s.Text(), " (View only)", "", 1),
							Latest:   !foundLatest && !strings.Contains(name, "M"),
							Medicine: strings.Contains(name, "Medicine"),
							MI:       strings.Contains(name, "MI"),
							ViewOnly: strings.Contains(name, "(View only)"),
						})
						if !foundLatest && !strings.Contains(s.Text(), "M") {
							foundLatest = true
						}
					}
				}
			}
		})
	})

	c.Visit("https://selfservice.mun.ca/direct/bwckschd.p_disp_dyn_sched")

	c.Wait()

	return semesters
}

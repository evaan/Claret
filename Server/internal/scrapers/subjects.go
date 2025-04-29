package scrapers

import (
	"log"
	"strconv"

	"github.com/PuerkitoBio/goquery"
	"github.com/evaan/Claret/internal/util"
	"github.com/gocolly/colly/v2"
)

func GetSubjects(logger *log.Logger, semester int) []util.Subject {
	var subjects []util.Subject

	c := colly.NewCollector()

	c.OnHTML("select[name=sel_subj]", func(e *colly.HTMLElement) {
		e.DOM.Find("option").Each(func(i int, s *goquery.Selection) {
			if s.Text() != "All" {
				id, exists := s.Attr("value")
				if exists {
					subjects = append(subjects, util.Subject{
						ID:   id,
						Name: s.Text(),
					})
				}
			}
		})
	})

	err := c.PostRaw("https://selfservice.mun.ca/direct/bwckgens.p_proc_term_date", util.MapToBytes(map[string]any{
		"p_calling_proc": "bwckschd.p_disp_dyn_sched",
		"p_term":         strconv.Itoa(semester),
	}))
	if err != nil {
		logger.Printf("Error getting subjects: %s\n", err.Error())
	}

	c.Wait()

	return subjects
}

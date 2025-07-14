package scrapers

import (
	"strconv"

	"github.com/PuerkitoBio/goquery"
	"github.com/evaan/Claret/internal/util"
	"github.com/gocolly/colly/v2"
)

func GetExams(semester int) []util.ExamTime {
	c := colly.NewCollector()

	exams := make([]util.ExamTime, 0)

	c.OnHTML("table.bordertable", func(e *colly.HTMLElement) {
		e.DOM.Find("tr").Each(func(i int, s *goquery.Selection) {
			selection := s.Find("td.dbdefault")
			if len(selection.Nodes) == 0 {
				return
			}
			crn := selection.Next().Next().Next().First().Text()
			exams = append(exams, util.ExamTime{
				CRN:        crn,
				SemesterID: semester,
				CourseKey:  strconv.Itoa(semester) + crn,
				Time:       selection.Last().Prev().Text(),
				Location:   selection.Last().Text(),
			})
		})
	})

	c.Visit("https://selfservice.mun.ca/direct/swkgexm.P_Query_Exam?p_term_code=" + strconv.Itoa(semester) + "&p_title=")
	c.Wait()

	return exams
}

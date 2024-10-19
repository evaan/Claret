package main

import (
	"strconv"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
)

type ExamTime struct {
	Crn        string   `gorm:"notNull"`
	SemesterID int      `gorm:"column:semester;not null"`
	Semester   Semester `gorm:"constraint:OnDelete:CASCADE;"`
	Identifier string   `gorm:"primaryKey"`
	Time       string   `gorm:"notNull"`
	Location   string   `gorm:"notNull"`
}

func exams(semester int) {
	c := colly.NewCollector()

	c.OnHTML("table.bordertable", func(e *colly.HTMLElement) {
		e.DOM.Find("tr").Each(func(i int, s *goquery.Selection) {
			selection := s.Find("td.dbdefault")
			if len(selection.Nodes) == 0 {
				return
			}
			crn := selection.Next().Next().Next().First().Text()
			db.Save(&ExamTime{
				Crn:        crn,
				SemesterID: semester,
				Identifier: crn + strconv.Itoa(semester),
				Time:       selection.Last().Prev().Text(),
				Location:   selection.Last().Text(),
			})
		})
	})

	c.Visit("https://selfservice.mun.ca/direct/swkgexm.P_Query_Exam?p_term_code=" + strconv.Itoa(semester) + "&p_title=")
	c.Wait()
}

package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
)

type EngSeats struct {
	Id         int    `gorm:"autoIncrement"`
	Subject    string `gorm:"notNull"`
	Name       string `gorm:"notNull"`
	Course     string `gorm:"notNull"`
	Section    string `gorm:"notNull"`
	Registered int    `gorm:"notNull"`
	Date       string `gorm:"notNull"`
}

func engSeating(semester int, crn string, subject string, id string, section string, name string) {
	if db.Where("course = ? AND section = ? AND date = ?", id, section, fmt.Sprintf("%d-%d-%d", time.Now().Day(), time.Now().Month(), time.Now().Year())).Find(&EngSeats{}).RowsAffected > 0 {
		return
	}

	c := colly.NewCollector()

	var cells []string

	c.OnHTML("caption", func(e *colly.HTMLElement) {
		if e.Text == "Registration Availability" {
			e.DOM.Parent().Find("td.dddefault").Each(func(i int, s *goquery.Selection) {
				cells = append(cells, s.Text())
			})
		}
	})

	c.Visit("https://selfservice.mun.ca/direct/bwckschd.p_disp_detail_sched?term_in=" + strconv.Itoa(semester) + "&crn_in=" + crn)
	c.Wait()

	if len(cells) <= 0 {
		return
	}

	max, err := strconv.Atoi(cells[0])
	if err != nil {
		logger.Fatal(err)
	}
	available, err := strconv.Atoi(cells[2])
	if err != nil {
		logger.Fatal(err)
	}

	db.Save(&EngSeats{
		Subject:    subject,
		Name:       name,
		Course:     id,
		Section:    section,
		Registered: max - available,
		Date:       fmt.Sprintf("%d-%d-%d", time.Now().Day(), time.Now().Month(), time.Now().Year()),
	})
}

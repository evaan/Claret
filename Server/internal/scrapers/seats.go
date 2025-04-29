package scrapers

import (
	"fmt"
	"strconv"

	"github.com/evaan/Claret/internal/util"
	"github.com/gocolly/colly/v2"
)

func GetSeats(semester string, crn string) util.CourseSeating {
	seating := util.CourseSeating{Semester: semester, CRN: crn}

	c := colly.NewCollector()

	c.OnHTML("table.datadisplaytable[summary=\"This layout table is used to present the seating numbers.\"] > tbody", func(e *colly.HTMLElement) {
		e.ForEach("td.dddefault", func(i int, item *colly.HTMLElement) {
			itemInt, err := strconv.Atoi(item.Text)
			if err != nil {
				fmt.Println("Error parsing seat table item", err)
				return
			}
			switch i {
			case 0:
				seating.Seats.Capacity = itemInt
			case 1:
				seating.Seats.Actual = itemInt
			case 2:
				seating.Seats.Remaining = itemInt
			case 3:
				seating.Waitlist.Capacity = itemInt
			case 4:
				seating.Waitlist.Actual = itemInt
			case 5:
				seating.Waitlist.Remaining = itemInt
			}
		})
	})

	c.Visit("https://selfservice.mun.ca/direct/bwckschd.p_disp_detail_sched?term_in=" + semester + "&crn_in=" + crn)

	return seating
}

package scrapers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/evaan/Claret/internal/scrapers"
	"github.com/evaan/Claret/internal/util"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Fetch everything before opening a transaction. A failed fetch or write must
// never delete the last usable semester, meetings, or seat snapshots.
func scrapeSemester(db *gorm.DB, client *http.Client, semester util.Semester) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	var previous []util.Course
	err := db.WithContext(ctx).Where("semester_id = ?", semester.ID).Find(&previous).Error
	cancel()
	if err != nil {
		return 0, err
	}
	if semester.ViewOnly && len(previous) > 0 {
		return 0, nil
	}
	log.Printf("Processing semester %s (%d)", semester.Name, semester.ID)
	subjects, err := scrapers.GetSubjects(semester)
	if err != nil {
		return 0, err
	}
	courses, seatings, err := scrapers.GetCourses(client, semester)
	if err != nil {
		return 0, err
	}
	if len(subjects) == 0 || len(courses) == 0 {
		return 0, fmt.Errorf("refusing an empty catalog replacement")
	}
	old := make(map[string]util.Course, len(previous))
	for _, course := range previous {
		old[course.Key] = course
	}
	keys := make([]string, 0, len(courses))
	seen := make(map[string]bool, len(courses))
	var meetings []util.CourseTime
	var instructors []util.CourseInstructor
	professors := make(map[string]util.Professor)
	for i := range courses {
		course := &courses[i]
		if course.Key == "" || course.CRN == "" || course.SubjectID == "" || course.SemesterID != semester.ID || seen[course.Key] {
			return 0, fmt.Errorf("invalid or duplicate course in upstream catalog")
		}
		seen[course.Key] = true
		keys = append(keys, course.Key)
		course.Prerequisites, err = scrapers.GetPrerequisites(client, semester, course.CRN)
		if err != nil {
			log.Printf("Preserving prerequisites for %s: %v", course.Key, err)
			course.Prerequisites = old[course.Key].Prerequisites
		}
		course.Corequisites, err = scrapers.GetCorequisites(client, semester, course.CRN)
		if err != nil {
			log.Printf("Preserving corequisites for %s: %v", course.Key, err)
			course.Corequisites = old[course.Key].Corequisites
		}
		if course.IsSectionLinked {
			course.LinkedSections, err = scrapers.GetLinkedSections(client, semester, course.CRN)
			if err != nil {
				log.Printf("Preserving linked sections for %s: %v", course.Key, err)
				course.LinkedSections = old[course.Key].LinkedSections
			}
		}
		times, names, err := scrapers.GetMeetingTimes(semester, course.CRN)
		if err != nil {
			return 0, fmt.Errorf("meetings for %s: %w", course.Key, err)
		}
		meetings = append(meetings, times...)
		for _, name := range util.Unique(names) {
			professors[name] = util.Professor{Name: name}
			instructors = append(instructors, util.CourseInstructor{ProfessorName: name, CourseKey: course.Key})
		}
	}
	ctx, cancel = context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if semester.Latest {
			if err := tx.Model(&util.Semester{}).Where("latest AND id <> ?", semester.ID).Update("latest", false).Error; err != nil {
				return err
			}
		}
		upsert := clause.OnConflict{UpdateAll: true}
		if err := tx.Clauses(upsert).Create(&semester).Error; err != nil {
			return err
		}
		if err := tx.Clauses(upsert).CreateInBatches(&subjects, 200).Error; err != nil {
			return err
		}
		if err := tx.Omit(clause.Associations).Clauses(upsert).CreateInBatches(&courses, 200).Error; err != nil {
			return err
		}
		// Upserts retain existing seat snapshots and exams for retained courses.
		if err := tx.Where("semester_id = ? AND key NOT IN ?", semester.ID, keys).Delete(&util.Course{}).Error; err != nil {
			return err
		}
		if err := tx.Where("course_key IN ?", keys).Delete(&util.CourseTime{}).Error; err != nil {
			return err
		}
		if err := tx.Where("course_key IN ?", keys).Delete(&util.CourseInstructor{}).Error; err != nil {
			return err
		}
		if len(meetings) > 0 {
			if err := tx.Omit(clause.Associations).CreateInBatches(&meetings, 500).Error; err != nil {
				return err
			}
		}
		profs := make([]util.Professor, 0, len(professors))
		for _, prof := range professors {
			profs = append(profs, prof)
		}
		if len(profs) > 0 {
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(&profs, 500).Error; err != nil {
				return err
			}
			if err := tx.Omit(clause.Associations).CreateInBatches(&instructors, 500).Error; err != nil {
				return err
			}
		}
		return saveSeatSnapshots(tx, semester.ID, seatings)
	})
	if err != nil {
		return 0, err
	}
	return len(courses), nil
}

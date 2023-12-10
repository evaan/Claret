import React from "react"
import FullCalendar from "@fullcalendar/react"
import timeGridPlugin from '@fullcalendar/timegrid'
import { useAtom } from "jotai"
import { selectedCoursesAtom, timesAtom } from "../api/atoms"
import { Course, Time } from "../api/types"
import moment from "moment"

export default function Schedule() {
    const [selectedCourses] = useAtom(selectedCoursesAtom);
    const [times] = useAtom(timesAtom);

    const courseTimes: {title: string, start: string, end?: string}[] = []

    let NATimes = 8;

    selectedCourses.forEach((course: Course) => {
        const courseTimes1 = times.filter((time: Time) => time.crn == course.crn)
        courseTimes1.forEach((time: Time) => {
            if ((time.startTime == "00:00" && time.endTime == "00:01") || time.startTime == "TBA" || time.startTime == "TBA") {
                courseTimes.push({title: `${course.id}-${course.section} - ${course.type}`, start: moment().startOf("week").add('days',  6).toDate().toISOString().split("T")[0]+"T"+NATimes.toString().padStart(2, "0")+":00"});
                NATimes++;
            } else {
                if (time.days.includes("M")) courseTimes.push({title: `${course.id}-${course.section} - ${course.type}`, start: moment().startOf("week").add('days',  1).toDate().toISOString().split("T")[0]+"T"+time.startTime, end: moment().startOf("week").add('days',  1).toDate().toISOString().split("T")[0]+"T"+time.endTime});
                if (time.days.includes("T")) courseTimes.push({title: `${course.id}-${course.section} - ${course.type}`, start: moment().startOf("week").add('days',  2).toDate().toISOString().split("T")[0]+"T"+time.startTime, end: moment().startOf("week").add('days',  2).toDate().toISOString().split("T")[0]+"T"+time.endTime});
                if (time.days.includes("W")) courseTimes.push({title: `${course.id}-${course.section} - ${course.type}`, start: moment().startOf("week").add('days',  3).toDate().toISOString().split("T")[0]+"T"+time.startTime, end: moment().startOf("week").add('days',  3).toDate().toISOString().split("T")[0]+"T"+time.endTime});
                if (time.days.includes("R")) courseTimes.push({title: `${course.id}-${course.section} - ${course.type}`, start: moment().startOf("week").add('days',  4).toDate().toISOString().split("T")[0]+"T"+time.startTime, end: moment().startOf("week").add('days',  4).toDate().toISOString().split("T")[0]+"T"+time.endTime});
                if (time.days.includes("F")) courseTimes.push({title: `${course.id}-${course.section} - ${course.type}`, start: moment().startOf("week").add('days',  5).toDate().toISOString().split("T")[0]+"T"+time.startTime, end: moment().startOf("week").add('days',  5).toDate().toISOString().split("T")[0]+"T"+time.endTime});

            }
        })
    });

    return (
        <FullCalendar plugins={[timeGridPlugin]} headerToolbar={{left: "", center: "", right: ""}} allDaySlot={false} hiddenDays={[0]} slotMinTime="07:00" nowIndicator={false}
        expandRows={true} events={courseTimes} contentHeight="auto" />
    )
}
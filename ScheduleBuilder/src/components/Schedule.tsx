import React from "react";
import FullCalendar from "@fullcalendar/react";
import timeGridPlugin from "@fullcalendar/timegrid";
import { useAtom } from "jotai";
import { selectedCoursesAtom, timesAtom } from "../api/atoms";
import { Course, Time } from "../api/types";
import * as Moment from "moment";
import { extendMoment } from "moment-range";
import { Accordion, Button } from "react-bootstrap";
const moment = extendMoment(Moment);

export default function Schedule() {
    const [selectedCourses, setSelectedCourses] = useAtom(selectedCoursesAtom);
    const [times] = useAtom(timesAtom);

    let credits = 0;
    let overlapping = false;
    const courseTimes: {title: string, start: string, end?: string}[] = [];
    let NATimes = 7;


    selectedCourses.forEach((course: Course) => {
        credits += course.credits;
        const courseTimes1 = times.filter((time: Time) => time.crn == course.crn);
        courseTimes1.forEach((time: Time) => {
            if ((time.startTime == "00:00" && time.endTime == "00:01") || time.startTime == "TBA" || time.startTime == "TBA") {
                courseTimes.push({title: `${course.id}-${course.section} - ${course.type}`, start: moment().startOf("week").add("days",  6).toDate().toISOString().split("T")[0]+"T"+NATimes.toString().padStart(2, "0")+":00"});
                NATimes++;
            } else {
                if (time.days.includes("M")) courseTimes.push({title: `${course.id}-${course.section} - ${course.type}`, start: moment().startOf("week").add("days",  1).toDate().toISOString().split("T")[0]+"T"+time.startTime, end: moment().startOf("week").add("days",  1).toDate().toISOString().split("T")[0]+"T"+time.endTime});
                if (time.days.includes("T")) courseTimes.push({title: `${course.id}-${course.section} - ${course.type}`, start: moment().startOf("week").add("days",  2).toDate().toISOString().split("T")[0]+"T"+time.startTime, end: moment().startOf("week").add("days",  2).toDate().toISOString().split("T")[0]+"T"+time.endTime});
                if (time.days.includes("W")) courseTimes.push({title: `${course.id}-${course.section} - ${course.type}`, start: moment().startOf("week").add("days",  3).toDate().toISOString().split("T")[0]+"T"+time.startTime, end: moment().startOf("week").add("days",  3).toDate().toISOString().split("T")[0]+"T"+time.endTime});
                if (time.days.includes("R")) courseTimes.push({title: `${course.id}-${course.section} - ${course.type}`, start: moment().startOf("week").add("days",  4).toDate().toISOString().split("T")[0]+"T"+time.startTime, end: moment().startOf("week").add("days",  4).toDate().toISOString().split("T")[0]+"T"+time.endTime});
                if (time.days.includes("F")) courseTimes.push({title: `${course.id}-${course.section} - ${course.type}`, start: moment().startOf("week").add("days",  5).toDate().toISOString().split("T")[0]+"T"+time.startTime, end: moment().startOf("week").add("days",  5).toDate().toISOString().split("T")[0]+"T"+time.endTime});
            }
        });
    });

    for (const time of courseTimes) {
        for (const time1 of courseTimes) {
            if (moment.range(moment(time.start), moment(time.end)).overlaps(moment.range(moment(time1.start), moment(time1.end))) && (time !== time1)) {
                overlapping = true;
                break;
            }
            else overlapping = false;
        }
    }

    return (
        <div>
            <FullCalendar plugins={[timeGridPlugin]} headerToolbar={{left: "", center: "", right: ""}} allDaySlot={false} hiddenDays={[0]} slotMinTime="07:00" nowIndicator={false}
            expandRows={true} events={courseTimes} contentHeight="auto" dayHeaderFormat={{ weekday: "long" }} dayHeaderContent={function(arg) {return(arg.text == "Saturday" ? "Others" : arg.text);}} />
            {overlapping && <div className="p-3 my-2 bg-warning text-dark rounded">Warning: You have overlapping courses.</div>}
            {credits > 15 && <div className="p-3 my-2 bg-warning text-dark rounded">Warning: Without explicit permission, MUN does not allow registration for more than 15 credit hours.</div>}
            <Accordion className="mt-2">
                <Accordion.Item eventKey="overview">
                    <Accordion.Header>Overview</Accordion.Header>
                    <Accordion.Body>
                        <p><strong>Total Credit Hours:</strong> {credits}</p>
                        <p><strong>Selected Courses:</strong></p>
                        {selectedCourses.map((course: Course) => (
                            <div key={course.crn} style={{display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: "4px"}}>
                                <p>{course.id} - {course.crn} - {course.type} - {course.instructor}</p>
                                <Button variant="danger" onClick={() => setSelectedCourses(selectedCourses.filter((course1: Course) => course1 !== course))}>Remove Course</Button>
                            </div>
                        ))}
                    </Accordion.Body>
                </Accordion.Item>
            </Accordion>
        </div>
    );
}
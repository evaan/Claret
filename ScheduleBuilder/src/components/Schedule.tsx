import React from "react";
import FullCalendar from "@fullcalendar/react";
import timeGridPlugin from "@fullcalendar/timegrid";
import { useAtom } from "jotai";
import { selectedCoursesAtom, timesAtom } from "../api/atoms";
import { Course, Time } from "../api/types";
import * as Moment from "moment";
import { extendMoment } from "moment-range";
import { Accordion, Button } from "react-bootstrap";
import { ICalButton } from "./ICalButton";
const moment = extendMoment(Moment);

export default function Schedule() {
    const [selectedCourses, setSelectedCourses] = useAtom(selectedCoursesAtom);
    const [times] = useAtom(timesAtom);

    let credits = 0;
    let overlapping = false;
    const courseTimes: {title: string, start: string, end?: string}[] = [];

    const startTimes: number[] = [];
    const endTimes: number[] = [];
    let NACourses = 0;
    for (const course of selectedCourses) {
        const courseTimes = times.filter((time: Time) => time.crn == course.crn);
        courseTimes.forEach((time: Time) => {
            if (time.startTime == "00:00" || time.endTime == "00:01") {NACourses++; return;}
            startTimes.push(moment(time.startTime, "HH:mm").hour());
            endTimes.push(moment(time.endTime, "HH:mm").hour()+1);
        });
    }
    let NAStartTime = startTimes.length == 0 ? 9 : Math.min(...startTimes);

    const min: number = Math.min(...startTimes);
    const max: number = Math.max(...endTimes);

    selectedCourses.forEach((course: Course) => {
        credits += course.credits;
        const courseTimes1 = times.filter((time: Time) => time.crn == course.crn);
        courseTimes1.forEach((time: Time) => {
            if ((time.startTime == "00:00" && time.endTime == "00:01") || time.startTime == "TBA" || time.startTime == "TBA") {
                courseTimes.push({title: `${course.id}-${course.section} - ${time.location}`, start: moment().startOf("week").add("days",  6).toDate().toISOString().split("T")[0]+"T"+NAStartTime.toString().padStart(2, "0")+":00"});
                NAStartTime++;
            } else {
                if (time.days.includes("M")) courseTimes.push({title: `${course.id}-${course.section} - ${time.location}`, start: moment().startOf("week").add("days",  1).toDate().toISOString().split("T")[0]+"T"+time.startTime, end: moment().startOf("week").add("days",  1).toDate().toISOString().split("T")[0]+"T"+time.endTime});
                if (time.days.includes("T")) courseTimes.push({title: `${course.id}-${course.section} - ${time.location}`, start: moment().startOf("week").add("days",  2).toDate().toISOString().split("T")[0]+"T"+time.startTime, end: moment().startOf("week").add("days",  2).toDate().toISOString().split("T")[0]+"T"+time.endTime});
                if (time.days.includes("W")) courseTimes.push({title: `${course.id}-${course.section} - ${time.location}`, start: moment().startOf("week").add("days",  3).toDate().toISOString().split("T")[0]+"T"+time.startTime, end: moment().startOf("week").add("days",  3).toDate().toISOString().split("T")[0]+"T"+time.endTime});
                if (time.days.includes("R")) courseTimes.push({title: `${course.id}-${course.section} - ${time.location}`, start: moment().startOf("week").add("days",  4).toDate().toISOString().split("T")[0]+"T"+time.startTime, end: moment().startOf("week").add("days",  4).toDate().toISOString().split("T")[0]+"T"+time.endTime});
                if (time.days.includes("F")) courseTimes.push({title: `${course.id}-${course.section} - ${time.location}`, start: moment().startOf("week").add("days",  5).toDate().toISOString().split("T")[0]+"T"+time.startTime, end: moment().startOf("week").add("days",  5).toDate().toISOString().split("T")[0]+"T"+time.endTime});
            }
        });
    });

    overlapCheck:
    for (const time of courseTimes) {
        for (const time1 of courseTimes) {
            if (moment.range(moment(time.start), moment(time.end)).overlaps(moment.range(moment(time1.start), moment(time1.end))) && (time !== time1)) {
                overlapping = true;
                break overlapCheck;
            }
            else overlapping = false;
        }
    }
    
    return (
        <div>
            <FullCalendar plugins={[timeGridPlugin]} headerToolbar={{left: "", center: "", right: ""}} allDaySlot={false} hiddenDays={[0]} nowIndicator={false}
            expandRows={true} events={courseTimes} height="auto" dayHeaderFormat={{ weekday: "long" }} dayHeaderContent={(arg) => arg.text == "Saturday" ? "Others" : arg.text} 
            slotDuration={max-min > 12 ? "00:30:00" : "00:15:00"} slotMinTime={`${startTimes.length == 0 ? 9 : min}:00`} slotMaxTime={`${Math.max((endTimes.length == 0 ? 17 : max), (startTimes.length == 0 ? 9 : min)+NACourses)}:00`} />
            {overlapping && <div className="p-3 my-2 bg-warning text-dark rounded">Warning: You have overlapping courses.</div>}
            {credits > 15 && <div className="p-3 my-2 bg-warning text-dark rounded">Warning: Without explicit permission, MUN does not allow registration for more than 15 credit hours.</div>}
            <Accordion className="mt-2" defaultActiveKey="overview">
                <Accordion.Item eventKey="overview">
                    <Accordion.Header>Overview</Accordion.Header>
                    <Accordion.Body>
                        <p><strong>Total Credit Hours:</strong> {credits}</p>
                        <p><strong>Selected Courses:</strong></p>
                        {selectedCourses.map((course: Course) => (
                            <div key={course.crn} style={{display: "flex", justifyContent: "space-between", alignItems: "center", marginTop: "4px"}} className="border border-secondary rounded p-2">
                                <p style={{paddingLeft: "8px"}}>{course.id} - {course.crn} - {course.type} - {course.instructor}</p>
                                <Button variant="text" style={{paddingLeft: "8px", paddingRight: "8px", paddingTop: "4px", paddingBottom: "4px"}} onClick={() => setSelectedCourses(selectedCourses.filter((course1: Course) => course1 !== course))}>&#10006;</Button>
                            </div>
                        ))}
                        <ICalButton/>
                    </Accordion.Body>
                </Accordion.Item>
            </Accordion>
        </div>
    );
}
import React from "react";
import FullCalendar from "@fullcalendar/react";
import timeGridPlugin from "@fullcalendar/timegrid";
import { useAtom } from "jotai";
import { selectedCoursesAtom, timesAtom } from "../api/atoms";
import { Course, Time } from "../api/types";
import * as Moment from "moment";
import { extendMoment } from "moment-range";
import { Accordion, Button } from "react-bootstrap";
import ICalModal from "./ICalModal";
const moment = extendMoment(Moment);

export default function Schedule() {
    const [selectedCourses, setSelectedCourses] = useAtom(selectedCoursesAtom);
    const [times] = useAtom(timesAtom);

    const [modalOpen, setModalOpen] = React.useState<boolean>(false);
    const closeModal = () => setModalOpen(false);

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

    let otherDay = false;
    function dayOfWeekName(day: string) {
        if (day == "Sunday") {
            if (!otherDay) otherDay = true;
            else return "Others";
        }
        if (day == "Saturday" && !weekend) return "Others";
        return day;
    }

    let weekend = false;

    selectedCourses.forEach((course: Course) => {
        credits += course.credits;
        times.filter((time: Time) => time.crn == course.crn).forEach((time: Time) => {
            if (time.days.includes("M")) courseTimes.push({title: `${course.id}-${course.section} - ${time.location}`, start: moment().startOf("week").add(1, "days").toDate().toISOString().split("T")[0]+"T"+time.startTime, end: moment().startOf("week").add(1, "days").toDate().toISOString().split("T")[0]+"T"+time.endTime});
            if (time.days.includes("T")) courseTimes.push({title: `${course.id}-${course.section} - ${time.location}`, start: moment().startOf("week").add(2, "days").toDate().toISOString().split("T")[0]+"T"+time.startTime, end: moment().startOf("week").add(2, "days").toDate().toISOString().split("T")[0]+"T"+time.endTime});
            if (time.days.includes("W")) courseTimes.push({title: `${course.id}-${course.section} - ${time.location}`, start: moment().startOf("week").add(3, "days").toDate().toISOString().split("T")[0]+"T"+time.startTime, end: moment().startOf("week").add(3, "days").toDate().toISOString().split("T")[0]+"T"+time.endTime});
            if (time.days.includes("R")) courseTimes.push({title: `${course.id}-${course.section} - ${time.location}`, start: moment().startOf("week").add(4, "days").toDate().toISOString().split("T")[0]+"T"+time.startTime, end: moment().startOf("week").add(4, "days").toDate().toISOString().split("T")[0]+"T"+time.endTime});
            if (time.days.includes("F")) courseTimes.push({title: `${course.id}-${course.section} - ${time.location}`, start: moment().startOf("week").add(5, "days").toDate().toISOString().split("T")[0]+"T"+time.startTime, end: moment().startOf("week").add(5, "days").toDate().toISOString().split("T")[0]+"T"+time.endTime});
            if (time.days.includes("S")) {
                weekend = true;
                courseTimes.push({title: `${course.id}-${course.section} - ${time.location}`, start: moment().startOf("week").add(6, "days").toDate().toISOString().split("T")[0]+"T"+time.startTime, end: moment().startOf("week").add(6, "days").toDate().toISOString().split("T")[0]+"T"+time.endTime});
            }
            if (time.days.includes("U")) {
                weekend = true;
                courseTimes.push({title: `${course.id}-${course.section} - ${time.location}`, start: moment().startOf("week").add("days").toDate().toISOString().split("T")[0]+"T"+time.startTime, end: moment().startOf("week").add("days").toDate().toISOString().split("T")[0]+"T"+time.endTime});
            }
        });
    });

    //another loop because it may not bring other courses to the others tab
    selectedCourses.forEach((course: Course) => {
        times.filter((time: Time) => time.crn == course.crn).forEach((time: Time) => {
            if ((time.startTime == "00:00" && time.endTime == "00:01") || time.startTime == "TBA" || time.startTime == "TBA") {
                courseTimes.push({title: `${course.id}-${course.section} - ${time.location}`, start: moment().startOf("week").add(weekend ? 7 : 6, "days").toDate().toISOString().split("T")[0]+"T"+NAStartTime.toString().padStart(2, "0")+":00"});
                NAStartTime++;
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
    
    const [isCopied, setIsCopied] = React.useState(false);

    const copySharingURL = () => {
        let sharingURL = "https://claretformun.com?crns=";
        selectedCourses.forEach((course: Course) => {
            sharingURL += course.crn + ",";
        });
        navigator.clipboard.writeText(sharingURL.slice(0, -1));
        setIsCopied(true);
        setTimeout(() => {setIsCopied(false);}, 1000);
    }

    return (
        <div>
            <FullCalendar plugins={[timeGridPlugin]} headerToolbar={{left: "", center: "", right: ""}} allDaySlot={false} nowIndicator={false}
            expandRows={true} events={courseTimes} height="auto" dayHeaderFormat={{ weekday: "long" }} dayHeaderContent={(arg) => {return dayOfWeekName(arg.text);}} 
            slotDuration={max-min > 12 ? "00:30:00" : "00:15:00"} slotMinTime={`${startTimes.length == 0 ? 9 : min}:00`} slotMaxTime={`${Math.max((endTimes.length == 0 ? 17 : max), (startTimes.length == 0 ? 9 : min)+NACourses)}:00`} 
            initialView="timeGrid" visibleRange={{start: moment().startOf("week").add(weekend ? 0 : 1, "days").format("YYYY-MM-DD"), end: moment().endOf("week").add(weekend ? 2 : 1, "days").format("YYYY-MM-DD")}}/>
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
                    </Accordion.Body>
                </Accordion.Item>
                <Accordion.Item eventKey="sharing">
                    <Accordion.Header>Sharing/Exporting</Accordion.Header>
                    <Accordion.Body>
                        <Button className="mt-3 w-100" onClick={() => setModalOpen(true)} disabled={selectedCourses.length == 0}>
                            {selectedCourses.length == 0 ? "Subscribe to Calendar (No courses selected)" : "Subscribe to Calendar"}
                        </Button>
                        <ICalModal isOpen={modalOpen} onHide={closeModal}/>
                        <Button className="mt-2 w-100" onClick={copySharingURL} disabled={selectedCourses.length == 0}>
                        {selectedCourses.length == 0 ? "Copy Claret link to clipboard (No courses selected)" : isCopied ? "Copied!" : "Copy Claret link to clipboard"}
                        </Button>
                    </Accordion.Body>
                </Accordion.Item>
            </Accordion>
        </div>
    );
}
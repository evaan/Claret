import React from "react";
import FullCalendar from "@fullcalendar/react";
import timeGridPlugin from "@fullcalendar/timegrid";
import { useAtom } from "jotai";
import {
  selectedCoursesAtom,
  selectedSemesterAtom,
  timesAtom,
} from "../api/atoms";
import { Course, Time } from "../api/types";
import Moment from "moment";
import { extendMoment } from "moment-range";
import { Accordion, Button } from "react-bootstrap";
import ClearModal from "./ClearModal";
import SectionModal from "./SectionModal";
import ICalModal from "./ICalModal";
import ExamModal from "./ExamModal";
const moment = extendMoment(Moment);

export function SectionButton1(props: { section: Course }) {
  const [modalOpen, setModalOpen] = React.useState<boolean>(false);

  const closeModal = () => setModalOpen(false);

  return (
    <div>
      <Button variant="link" style={{ width: "100%", paddingLeft: "8px" }} onClick={() => setModalOpen(true)}>
        {props.section.id} - {props.section.crn} - {props.section.type} - {props.section.instructor}
      </Button>
      <SectionModal isOpen={modalOpen} onHide={closeModal} section={props.section} />
    </div>
  );
}

export default function Schedule() {
  const [selectedCourses, setSelectedCourses] = useAtom(selectedCoursesAtom);
  const [times] = useAtom(timesAtom);

  const [calModalOpen, setCalModalOpen] = React.useState<boolean>(false);
  const closeCalModal = () => setCalModalOpen(false);

  const [examModalOpen, setExamModalOpen] = React.useState<boolean>(false);
  const closeExamModal = () => setExamModalOpen(false);

  const [clearModalOpen, setClearModalOpen] = React.useState<boolean>(false);
  const closeClearModal = () => setClearModalOpen(false);

  const [selectedSemester] = useAtom(selectedSemesterAtom);

  let credits = 0;
  let overlapping = false;
  let courseTimes: { title: string; start: string; end?: string }[] = [];

  const startTimes: number[] = [];
  const endTimes: number[] = [];
  let weekend = false;

  React.useEffect(() => {
    courseTimes = [];
  }, [selectedSemester]);

  function dayOfWeekName(day: string) {
    if (day === "Sunday") {
      return "Others";
    }
    if (day === "Saturday" && !weekend) return "Others";
    return day;
  }

  selectedCourses.forEach((course: Course) => {
    credits += course.credits;
    times
      .filter((time: Time) => time.crn === course.crn && time.days !== null)
      .forEach((time: Time) => {
        if (
          (time.startTime === "00:00" && time.endTime === "00:01") ||
          time.startTime === "TBA"
        ) {
          return;
        }
        startTimes.push(moment(time.startTime, "HH:mm").hour());
        endTimes.push(moment(time.endTime, "HH:mm").hour() + 1);

        const addEvent = (dayOffset: number) => {
          courseTimes.push({
            title: `${course.id}-${course.section} - ${time.location}`,
            start: moment().startOf("week").add(dayOffset, "days").format("YYYY-MM-DD") + "T" + time.startTime,
            end: moment().startOf("week").add(dayOffset, "days").format("YYYY-MM-DD") + "T" + time.endTime,
          });
        };

        if (time.days.includes("M")) addEvent(1);
        if (time.days.includes("T")) addEvent(2);
        if (time.days.includes("W")) addEvent(3);
        if (time.days.includes("R")) addEvent(4);
        if (time.days.includes("F")) addEvent(5);
        if (time.days.includes("S")) {
          weekend = true;
          addEvent(6);
        }
        if (time.days.includes("U")) {
          weekend = true;
          addEvent(7);
        }
      });
  });

  const min = startTimes.length > 0 ? Math.min(...startTimes) : 9;
  const max = endTimes.length > 0 ? Math.max(...endTimes) : 17;

  const othersStartingHour = startTimes.length > 0 ? Math.min(...startTimes) : 9;
  let othersHourCursor = othersStartingHour;

  selectedCourses.forEach((course: Course) => {
    times
      .filter((time: Time) =>
        time.crn === course.crn &&
        ((time.startTime === "00:00" && time.endTime === "00:01") || time.startTime === "TBA")
      )
      .forEach((time: Time) => {
        courseTimes.push({
          title: `${course.id}-${course.section} - ${time.location}`,
          start: moment().startOf("week").add(weekend ? 7 : 6, "days").format("YYYY-MM-DD") + "T" + othersHourCursor.toString().padStart(2, "0") + ":00",
          end: moment().startOf("week").add(weekend ? 7 : 6, "days").format("YYYY-MM-DD") + "T" + (othersHourCursor + 1).toString().padStart(2, "0") + ":00",
        });
        othersHourCursor++;
      });
  });

  overlapCheck: for (const time of courseTimes) {
    for (const time1 of courseTimes) {
      if (
        moment.range(moment(time.start), moment(time.end)).overlaps(moment.range(moment(time1.start), moment(time1.end))) &&
        time !== time1
      ) {
        overlapping = true;
        break overlapCheck;
      } else overlapping = false;
    }
  }

  const [isCopied, setIsCopied] = React.useState(false);

  const copySharingURL = () => {
    navigator.clipboard.writeText(window.location.href);
    setIsCopied(true);
    setTimeout(() => {
      setIsCopied(false);
    }, 1000);
  };

  function removeCourse(course: Course) {
    setSelectedCourses(selectedCourses.filter((course1: Course) => course !== course1));
    const params = new URLSearchParams(window.location.search);
    let crns = "";
    selectedCourses.forEach((course1: Course) => {
      if (course.crn !== course1.crn) crns += course1.crn + ",";
    });
    params.set("crns", crns);
    window.history.replaceState(null, "", `?${params}`);
  }

  return (
    <div>
      <FullCalendar
        plugins={[timeGridPlugin]}
        headerToolbar={{ left: "", center: "", right: "" }}
        allDaySlot={false}
        nowIndicator={false}
        expandRows={true}
        events={courseTimes}
        height="auto"
        dayHeaderFormat={{ weekday: "long" }}
        dayHeaderContent={(arg) => dayOfWeekName(arg.text)}
        slotDuration={max - min > 12 ? "00:30:00" : "00:15:00"}
        slotMinTime={`${min}:00`}
        slotMaxTime={`${Math.max(max, min + courseTimes.length)}:00`}
        initialView="timeGrid"
        visibleRange={{
          start: moment().startOf("week").add(weekend ? 0 : 1, "days").format("YYYY-MM-DD"),
          end: moment().endOf("week").add(weekend ? 2 : 1, "days").format("YYYY-MM-DD"),
        }}
        eventColor="#A8415B"
      />
      {overlapping && (
        <div className="bg-warning text-dark my-2 rounded p-3">
          Warning: You have overlapping courses.
        </div>
      )}
      {credits > 15 && (
        <div className="bg-warning text-dark my-2 rounded p-3">
          Warning: Without explicit permission, MUN does not allow registration for more than 15 credit hours.
        </div>
      )}
      <Accordion className="mt-2" defaultActiveKey="overview">
        <Accordion.Item eventKey="overview">
          <Accordion.Header>Overview</Accordion.Header>
          <Accordion.Body>
            <p>
              <strong>Total Credit Hours:</strong> {credits}
            </p>
            <p>
              <strong>Selected Courses:</strong>
            </p>
            {selectedCourses.map((course: Course) => (
              <div key={course.crn} style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginTop: "4px" }} className="border-secondary rounded border p-2">
                <SectionButton1 section={course} />
                <Button variant="text" style={{ padding: "4px 8px" }} onClick={() => removeCourse(course)}>
                  &#10006;
                </Button>
              </div>
            ))}
            <hr />
            <Button variant="outline-danger" style={{ width: "100%" }} onClick={() => setClearModalOpen(true)}>
              Clear Courses
            </Button>
            <ClearModal isOpen={clearModalOpen} onHide={closeClearModal} />
          </Accordion.Body>
        </Accordion.Item>
        <Accordion.Item eventKey="sharing">
          <Accordion.Header>Sharing/Exporting</Accordion.Header>
          <Accordion.Body>
            <Button className="w-100" onClick={() => setCalModalOpen(true)} disabled={selectedCourses.length === 0}>
              {selectedCourses.length === 0 ? "Subscribe to Calendar (No courses selected)" : "Subscribe to Calendar"}
            </Button>
            <ICalModal isOpen={calModalOpen} onHide={closeCalModal} />
            <Button className="w-100 mt-2" onClick={copySharingURL} disabled={selectedCourses.length === 0}>
              {selectedCourses.length === 0 ? "Copy Claret link to clipboard (No courses selected)" : isCopied ? "Copied!" : "Copy Claret link to clipboard"}
            </Button>
            <ExamModal isOpen={examModalOpen} onHide={closeExamModal} />
            <Button className="w-100 mt-2" onClick={() => setExamModalOpen(true)} disabled={selectedCourses.length === 0}>
              {selectedCourses.length === 0 ? "View final exam schedule (no courses selected)" : "View final exam schedule"}
            </Button>
          </Accordion.Body>
        </Accordion.Item>
      </Accordion>
    </div>
  );
}
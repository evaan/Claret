import React from "react";
import FullCalendar from "@fullcalendar/react";
import timeGridPlugin from "@fullcalendar/timegrid";
import { useAtom } from "jotai";
import {
  selectedCoursesAtom,
  selectedSemesterAtom,
  timesAtom,
} from "../api/atoms";
import { Course } from "../api/types";
import moment from "moment";
import { Accordion, Button } from "react-bootstrap";
import ClearModal from "./ClearModal";
import SectionModal from "./SectionModal";
import ICalModal from "./ICalModal";
import ExamModal from "./ExamModal";
import { buildSchedule } from "../api/schedule";
import { totalCredits, uniqueCourseCount } from "../api/functions";
import { LinkedSectionWarnings } from "./LinkedSections";

function renderEventText(text: string) {
  return text.split(/(\s[-–—]\s)/).map((part, index) => index % 2
    ? <React.Fragment key={index}><span className="schedule-event-separator">{part}</span><br className="schedule-event-break" /></React.Fragment>
    : part);
}

export function SectionButton1(props: { section: Course }) {
  const [modalOpen, setModalOpen] = React.useState<boolean>(false);

  const closeModal = () => setModalOpen(false);

  return (
    <div>
      <Button variant="link" className="selected-course-link" onClick={() => setModalOpen(true)}>
        <strong>{props.section.id} · {props.section.section}</strong>
        <span>{props.section.name}</span>
        <small>{props.section.type} · CRN {props.section.crn}</small>
      </Button>
      <SectionModal isOpen={modalOpen} onHide={closeModal} section={props.section} />
    </div>
  );
}

export default function Schedule(props: { standalone?: boolean }) {
  const [selectedCourses, setSelectedCourses] = useAtom(selectedCoursesAtom);
  const [times] = useAtom(timesAtom);
  const [semester] = useAtom(selectedSemesterAtom);
  const layout = React.useMemo(() => buildSchedule(selectedCourses, times), [selectedCourses, times]);
  const calendarRef = React.useRef<FullCalendar>(null);
  const calendarContainerRef = React.useRef<HTMLDivElement>(null);
  const [fitHeight, setFitHeight] = React.useState<number>();
  const [fullWeekdays, setFullWeekdays] = React.useState(false);

  React.useLayoutEffect(() => {
    const container = calendarContainerRef.current;
    if (!container) return;
    let width = 0;
    let frame = 0;
    let disposed = false;
    const textContext = document.createElement("canvas").getContext("2d");
    const resizeCalendar = () => {
      if (container.getBoundingClientRect().width <= 0) return;
      const mobile = window.matchMedia("(max-width: 767px)").matches;
      const viewportHeight = window.visualViewport?.height ?? window.innerHeight;
      const top = container.getBoundingClientRect().top + window.scrollY;
      const switcher = props.standalone ? null : document.querySelector(".mobile-view-switcher");
      const page = container.closest(".schedule-page");
      const bottom = props.standalone && page
        ? parseFloat(window.getComputedStyle(page).paddingBottom)
        : (switcher?.getBoundingClientRect().height ?? 0) + 12;
      setFitHeight(mobile || props.standalone ? Math.max(120, Math.floor(viewportHeight - top - bottom)) : undefined);
      calendarRef.current?.getApi().updateSize();
      const headings = container.querySelectorAll<HTMLElement>(".fc-col-header-cell");
      const label = headings[0]?.querySelector<HTMLElement>(".fc-col-header-cell-cushion");
      if (label && textContext) {
        const style = window.getComputedStyle(label);
        textContext.font = style.font;
        const requiredWidth = Math.max(...moment.weekdays().map(day => textContext.measureText(day).width))
          + parseFloat(style.paddingLeft) + parseFloat(style.paddingRight) + 2;
        setFullWeekdays(!mobile && [...headings].every(heading => heading.getBoundingClientRect().width >= requiredWidth));
      }
    };
    const scheduleResize = () => {
      cancelAnimationFrame(frame);
      frame = requestAnimationFrame(resizeCalendar);
    };
    const observer = new ResizeObserver(([entry]) => {
      if (entry.contentRect.width === width) return;
      width = entry.contentRect.width;
      cancelAnimationFrame(frame);
      if (width <= 0) return;
      scheduleResize();
    });
    observer.observe(container);
    window.addEventListener("resize", scheduleResize);
    window.visualViewport?.addEventListener("resize", scheduleResize);
    void document.fonts.ready.then(() => { if (!disposed) scheduleResize(); });
    resizeCalendar();
    return () => {
      disposed = true;
      observer.disconnect();
      cancelAnimationFrame(frame);
      window.removeEventListener("resize", scheduleResize);
      window.visualViewport?.removeEventListener("resize", scheduleResize);
    };
  }, [props.standalone, layout.days.length]);

  const [calModalOpen, setCalModalOpen] = React.useState<boolean>(false);
  const closeCalModal = () => setCalModalOpen(false);

  const [examModalOpen, setExamModalOpen] = React.useState<boolean>(false);
  const closeExamModal = () => setExamModalOpen(false);

  const [clearModalOpen, setClearModalOpen] = React.useState<boolean>(false);
  const closeClearModal = () => setClearModalOpen(false);

  const credits = totalCredits(selectedCourses);
  const { min, max, conflictingEvents, calendarMinWidth } = layout;
  // Leave room for larger labels and headers that wrap on narrow screens.
  const fitSlotMinutes = fitHeight == null ? undefined
    : [30, 60, 120, 240].find(minutes => Math.ceil((max - min) * 60 / minutes) * (props.standalone ? 24 : 20) <= fitHeight - 60) ?? 240;

  const [isCopied, setIsCopied] = React.useState(false);
  const [preparingPDF, setPreparingPDF] = React.useState(false);
  const [pdfFailed, setPDFFailed] = React.useState(false);
  const [pdfBlocked, setPDFBlocked] = React.useState(false);
  const preparedPDF = React.useRef<{ courses: Course[]; times: typeof times; semester: string; url: string }>();
  const pdfURLs = React.useRef<string[]>([]);
  React.useEffect(() => () => {
    pdfURLs.current.forEach(url => URL.revokeObjectURL(url));
  }, []);
  const scheduleURL = new URL(window.location.href);
  scheduleURL.searchParams.set("view", "schedule");
  scheduleURL.searchParams.set("crns", selectedCourses.map(course => course.crn).join(","));

  async function openPDF() {
    setPDFFailed(false);
    setPDFBlocked(false);
    const openTab = (url: string) => {
      const tab = window.open(url, "_blank");
      if (tab) tab.opener = null;
      else setPDFBlocked(true);
    };
    const semesterName = semester?.name ?? "Schedule";
    const cached = preparedPDF.current;
    if (cached && cached.courses === selectedCourses && cached.times === times && cached.semester === semesterName) {
      openTab(cached.url);
      return;
    }
    setPreparingPDF(true);
    try {
      const { createSchedulePDF } = await import("../api/schedulePDF");
      const url = URL.createObjectURL(createSchedulePDF(selectedCourses, times, semesterName).output("blob"));
      pdfURLs.current.push(url);
      preparedPDF.current = { courses: selectedCourses, times, semester: semesterName, url };
      openTab(url);
    } catch {
      setPDFFailed(true);
    } finally {
      setPreparingPDF(false);
    }
  }

  const copySharingURL = () => {
    navigator.clipboard.writeText(window.location.href);
    setIsCopied(true);
    setTimeout(() => {
      setIsCopied(false);
    }, 1000);
  };

  function removeCourse(course: Course) {
    const remaining = selectedCourses.filter((course1: Course) => course.crn !== course1.crn);
    setSelectedCourses(remaining);
    const params = new URLSearchParams(window.location.search);
    params.set("crns", remaining.map(course1 => course1.crn).join(","));
    window.history.replaceState(null, "", `?${params}`);
  }

  const calendar = (
      <div className="calendar-scroll" aria-label="Weekly schedule" tabIndex={0}>
        <div ref={calendarContainerRef} style={{ "--calendar-min-width": `${calendarMinWidth}px` } as React.CSSProperties}>
          <FullCalendar
            ref={calendarRef}
            plugins={[timeGridPlugin]}
            headerToolbar={false}
            allDaySlot={false}
            nowIndicator={false}
            expandRows={fitHeight != null}
            eventMinHeight={fitHeight != null ? 0 : 15}
            slotEventOverlap={false}
            events={layout.events}
            eventContent={event => <div
              className="schedule-event-content"
              title={`${event.event.title}${event.event.extendedProps.unscheduled ? "" : `\n${event.timeText}`}${event.event.classNames.includes("has-conflict") ? "\nTime conflict" : ""}`}
            >
              {!event.event.extendedProps.unscheduled && <div className="fc-event-time">{renderEventText(event.timeText)}</div>}
              <div className="fc-event-title-container"><div className="fc-event-title fc-sticky">{renderEventText(event.event.title)}</div></div>
            </div>}
            height={fitHeight ?? "auto"}
            dayHeaderFormat={{ weekday: fullWeekdays ? "long" : "short" }}
            dayHeaderContent={arg => moment(arg.date).isSame(layout.unscheduledDate, "day")
              ? <span title="Courses without a scheduled meeting time">N/A</span> : <span>{arg.text}</span>}
            slotDuration={fitSlotMinutes == null ? (max - min > 12 ? "00:30:00" : "00:15:00") : { minutes: fitSlotMinutes }}
            slotLabelInterval={fitSlotMinutes == null ? undefined : { minutes: Math.max(60, fitSlotMinutes) }}
            slotMinTime={`${min}:00`}
            slotMaxTime={`${max}:00`}
            initialView="timeGrid"
            visibleRange={layout.range}
            eventColor="#A8415B"
          />
        </div>
      </div>
  );
  if (props.standalone) return calendar;

  return (
    <div>
      {calendar}
      <LinkedSectionWarnings courses={selectedCourses} />
      {conflictingEvents.size > 0 && (
        <div className="bg-warning text-dark my-2 rounded p-3" role="status">
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
          <Accordion.Header>Selected courses</Accordion.Header>
          <Accordion.Body>
            <p className="d-flex flex-wrap gap-2">
              <span><strong>Courses:</strong> {uniqueCourseCount(selectedCourses)}</span>
              <span><strong>Credits:</strong> {credits}</span>
            </p>
            {selectedCourses.map((course: Course) => (
              <div key={course.crn} className="selected-course">
                <SectionButton1 section={course} />
                <Button variant="link" aria-label={`Remove ${course.id} section ${course.section}`} onClick={() => removeCourse(course)}>
                  &#10006;
                </Button>
              </div>
            ))}
            <Button variant="outline-danger" className="w-100 mt-3" disabled={selectedCourses.length === 0} onClick={() => setClearModalOpen(true)}>
              Clear Courses
            </Button>
            <ClearModal isOpen={clearModalOpen} onHide={closeClearModal} />
          </Accordion.Body>
        </Accordion.Item>
        <Accordion.Item eventKey="sharing">
          <Accordion.Header>Share schedule</Accordion.Header>
          <Accordion.Body>
            <Button className="w-100" onClick={() => setCalModalOpen(true)} disabled={selectedCourses.length === 0}>
              Subscribe to calendar
            </Button>
            <ICalModal isOpen={calModalOpen} onHide={closeCalModal} />
            <Button className="w-100 mt-2" onClick={copySharingURL} disabled={selectedCourses.length === 0}>
              {isCopied ? "Copied!" : "Copy schedule link"}
            </Button>
            <Button as="a" className="w-100 mt-2" href={scheduleURL.href} target="_blank" rel="noopener noreferrer"
              disabled={selectedCourses.length === 0}>
              Open schedule only
            </Button>
            <Button className="w-100 mt-2" onClick={openPDF}
              disabled={selectedCourses.length === 0 || preparingPDF} aria-busy={preparingPDF}>
              {preparingPDF ? "Preparing PDF..." : "Open schedule PDF"}
            </Button>
            {pdfFailed && <p className="detail-muted mt-2" role="alert">The PDF could not be created. Please try again.</p>}
            {pdfBlocked && <p className="detail-muted mt-2" role="alert">Your browser blocked the new tab. Click again to open the PDF.</p>}
            <ExamModal isOpen={examModalOpen} onHide={closeExamModal} />
            {/* <Button className="w-100 mt-2" onClick={() => setExamModalOpen(true)} disabled={selectedCourses.length === 0}>
              {selectedCourses.length === 0 ? "View final exam schedule (no courses selected)" : "View final exam schedule"}
            </Button> */}
          </Accordion.Body>
        </Accordion.Item>
      </Accordion>
    </div>
  );
}

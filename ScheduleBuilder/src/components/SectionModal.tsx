import { Button, Modal } from "react-bootstrap";
import { Course } from "../api/types";
import { useAtom } from "jotai";
import { profsAtom, seatingAtom, selectedCoursesAtom, timesAtom } from "../api/atoms";
import { availableSeats, formatCredits, formatInstructorName, isOnline, isUnscheduled } from "../api/functions";
import moment from "moment";
import { RequirementsDropdown } from "./Requirements";

const dayNames: Record<string, string> = {
    M: "Monday", T: "Tuesday", W: "Wednesday", R: "Thursday",
    F: "Friday", S: "Saturday", U: "Sunday",
};

export default function SectionModal(props: {isOpen: boolean; onHide: () => void; section: Course}) {
    const [times] = useAtom(timesAtom);
    const [profs] = useAtom(profsAtom);
    const [selectedCourses, setSelectedCourses] = useAtom(selectedCoursesAtom);
    const [seatings] = useAtom(seatingAtom);
    const section = props.section;
    const selected = selectedCourses.some(course => course.crn === section.crn);
    const seating = seatings.find(seating => seating.crn === section.crn);
    const online = isOnline(section);
    const meetings = times.filter(time => time.crn === section.crn && (!online || !isUnscheduled(time)));
    const updatedAt = seating?.updatedAt ? new Date(seating.updatedAt) : null;
    const hasTimestamp = updatedAt !== null && !Number.isNaN(updatedAt.getTime()) && updatedAt.getFullYear() > 2000;

    function toggleCourse() {
        const nextCourses = selected
            ? selectedCourses.filter(course => course.crn !== section.crn)
            : [...selectedCourses, section];
        setSelectedCourses(nextCourses);
        const params = new URLSearchParams(window.location.search);
        params.set("crns", nextCourses.map(course => course.crn).join(","));
        window.history.replaceState(null, "", `?${params}`);
        props.onHide();
    }

    return (
        <Modal show={props.isOpen} onHide={props.onHide} centered scrollable className="section-modal">
            <Modal.Header closeButton>
                <Modal.Title>
                    {section.id} - {section.name}
                    <small className="detail-muted d-block">Section {section.section} · CRN {section.crn}</small>
                </Modal.Title>
            </Modal.Header>
            <Modal.Body>
                <dl className="section-details">
                    <div><dt>Credits</dt><dd>{formatCredits(section)}</dd></div>
                    {section.type && <div><dt>Type</dt><dd>{section.type}</dd></div>}
                    {section.instructionalMethod && <div><dt>Delivery</dt><dd>{section.instructionalMethod}</dd></div>}
                    {section.campus && <div><dt>Campus</dt><dd>{section.campus}</dd></div>}
                    {section.partOfTerm && <div><dt>Part of term</dt><dd>{section.partOfTerm}</dd></div>}
                    {seating ? <>
                        <div><dt>Seats available</dt><dd className={
                            availableSeats(seating) <= 0 ? "seats-unavailable"
                                : availableSeats(seating) <= seating.maxSeats * 0.25 ? "text-warning" : ""
                        }>
                            {Math.max(0, availableSeats(seating))} / {seating.maxSeats}
                        </dd></div>
                        <div><dt>Waitlist spaces</dt><dd>{seating.maxWaitlist > 0 ? `${Math.max(0, seating.waitlist)} / ${seating.maxWaitlist}` : "No waitlist offered"}</dd></div>
                        {seating.crossListCapacity != null && <div><dt>Shared seats</dt><dd>
                            {seating.crossListAvailable ?? "Unknown"} / {seating.crossListCapacity}
                            {seating.crossListCount != null && ` (${seating.crossListCount} enrolled)`}
                        </dd></div>}
                        <div><dt>Updated</dt><dd className="detail-muted">
                            {hasTimestamp ? moment(updatedAt).format("h:mm A") : "Unavailable"}
                        </dd></div>
                    </> : <div><dt>Seats available</dt><dd className="detail-muted">Unavailable</dd></div>}

                    <div><dt>Instructors</dt><dd>
                        {(section.instructors ?? []).length > 0 ? <ul className="detail-list">
                            {[...new Set(section.instructors)].map(instructor => {
                                const prof = profs.find(prof => prof.name === instructor);
                                const displayName = formatInstructorName(instructor);
                                const href = prof
                                    ? `https://www.ratemyprofessors.com/professor/${prof.id}`
                                    : `https://www.ratemyprofessors.com/search/professors/1441?q=${encodeURIComponent(displayName.replace(/\s[A-Za-z]\.\s/g, " "))}`;
                                return <li key={instructor}>
                                    {displayName}
                                    {instructor !== "TBA" && <> · <a href={href} rel="noreferrer" target="_blank">
                                        {prof ? `RMP ${prof.rating}/5 (${prof.ratings} ratings)` : "Search RateMyProfessors"}
                                    </a></>}
                                </li>;
                            })}
                        </ul> : <span className="detail-muted">TBA</span>}
                    </dd></div>

                    {(meetings.length > 0 || !online) && <div><dt>Meetings</dt><dd>
                        {meetings.length > 0 ? <ul className="detail-list meeting-list">
                            {meetings.map((time, index) => <li key={`${time.crn}-${index}`}>
                                <div>{isUnscheduled(time) ? "TBA" : <>
                                    {(time.days ?? "").split("").map(day => dayNames[day] ?? day).join(", ")}
                                    {" · "}{moment(time.startTime, "HH:mm").format("h:mm A")}–{moment(time.endTime, "HH:mm").format("h:mm A")}
                                </>}</div>
                                {(time.location || !online) && <div className="detail-muted">
                                    {time.location || "Location TBA"}
                                </div>}
                            </li>)}
                        </ul> : <span className="detail-muted">TBA</span>}
                    </dd></div>}
                </dl>

                <RequirementsDropdown key={section.crn} prerequisites={section.prerequisites} corequisites={section.corequisites} />
            </Modal.Body>
            <Modal.Footer>
                <Button variant="secondary" onClick={props.onHide}>Close</Button>
                <Button variant={selected ? "outline-danger" : "primary"} onClick={toggleCourse}>
                    {selected ? "Remove course" : "Add to schedule"}
                </Button>
            </Modal.Footer>
        </Modal>
    );
}

import { Button, Modal, ModalBody, ModalFooter, ModalHeader } from "react-bootstrap";
import { Course, Professor, Seating, Time } from "../api/types";
import { useAtom } from "jotai";
import { profsAtom, seatingAtom, selectedCoursesAtom, timesAtom } from "../api/atoms";
import moment from "moment";

export default function SectionModal(props: {isOpen: boolean; onHide: () => void; section: Course}) {
    const [times] = useAtom(timesAtom);
    const [profs] = useAtom(profsAtom);
    const [selectedCourses, setSelectedCourses] = useAtom(selectedCoursesAtom);
    const [seatings] = useAtom(seatingAtom);

    function formatDateString(input: string){
        input = input.replace("M", "Monday, ").replace("T", "Tuesday, ").replace("W", "Wednesday, ").replace("R", "Thursday, ").replace("F", "Friday, ").replace("S", "Saturday, ").replace("U", "Sunday, ");
        if (input.endsWith(", ")) input = input.slice(0, -2);
        return input;
    }

    function addCourse() {
        props.onHide();
        setSelectedCourses([...selectedCourses, props.section]);
        const params = new URLSearchParams(window.location.search);
        let crns = "";
        selectedCourses.forEach((course: Course) => {
            crns += course.crn + ",";
        });
        crns += props.section.crn;
        params.set("crns", crns);
        window.history.replaceState(null, "", `?${params}`);
    }

    function removeCourse () {
        props.onHide();
        setSelectedCourses(selectedCourses.filter((course: Course) => course !== props.section));
        const params = new URLSearchParams(window.location.search);
        let crns = "";
        selectedCourses.forEach((course: Course) => {
            if (course.crn !== props.section.crn) crns += course.crn + ",";
        });
        params.set("crns", crns);
        window.history.replaceState(null, "", `?${params}`);
    }

    return(
        <Modal show={props.isOpen} onHide={() => props.onHide()} centered>
            <ModalHeader>{props.section.id} - {props.section.name} (Section: {props.section.section})</ModalHeader>
            <ModalBody>
                {/*
                Make:    it
                Like:    this
                */}
                {props.section.comment !== null &&
                <p><strong>Comment:</strong> {props.section.comment}</p>}
                <p><strong>CRN:</strong> {props.section.crn}</p>
                <p><strong>Credit Hours:</strong> {props.section.credits}</p>
                <p><strong>Section:</strong> {props.section.section}</p>
                <p><strong>Level:</strong> {props.section.level}</p>
                <p><strong>Campus:</strong> {props.section.campus}</p>
                <p><strong>Type:</strong> {props.section.type !== null ? props.section.type : "Unknown"}</p>
                <p><strong>Date Range:</strong> {props.section.dateRange !== null ? props.section.dateRange : "Unknown"}</p>
                {props.section.instructor != null && props.section.instructor != "" && (
                    <>
                        <p><strong>Instructors:</strong></p>
                        <ul className="m-0">
                            {props.section.instructor.split(", ").map((instructor: string) => {
                                if (profs !== undefined && profs.filter((prof: Professor) => prof.name == instructor).length > 0) {
                                    const prof = profs.filter((prof: Professor) => prof.name == instructor)[0];
                                    return <li key={instructor}>{instructor} {instructor !== "TBA" && <a href={`https://www.ratemyprofessors.com/professor/${prof.id}`} rel="noreferrer" target="_blank">(RateMyProfessors Rating: {prof.rating}/5)</a>}</li>;
                                }
                                else
                                    return <li key={instructor}>{instructor} {instructor !== "TBA" && <a href={`https://www.ratemyprofessors.com/search/professors/1441?q=${instructor.replace(/\s[A-Za-z]\.\s/g, ' ')}`} rel="noreferrer" target="_blank">(Search on RateMyProfessors)</a>}</li>;
                            })}
                        </ul>
                    </>
                )}
                {times.some(time => time.days && time.crn === props.section.crn) && (
                        <>
                            <p><strong>Times:</strong></p>
                            <ul className="m-0">
                                {times.filter((time: Time) => time.days != null && time.crn === props.section.crn).map((time: Time) => (
                                    <li key={time.id}>{formatDateString(time.days)} - {moment(time.startTime, "HH:mm").format("hh:mm A").replace("Invalid date", "TBA")}-{moment(time.endTime, "HH:mm").format("hh:mm A").replace("Invalid date", "TBA")} - {time.location} {props.section.type.includes(", ") ? `(${time.type})` : ""}</li>
                                ))}
                            </ul>
                        </>
                )}
                {seatings.filter((seating: Seating) => seating.crn == props.section.crn).map((seating: Seating) => {
                    return (
                        <div key={seating.crn}>
                            <p><strong>Seats Available:</strong> <span className={Number(seating.seats.remaining) <= 0 ? "text-danger" : ""}>{seating.seats.remaining}/{seating.seats.capacity}</span></p>
                            <p><strong>Waitlist Available:</strong> <span className={Number(seating.waitlist.remaining) <= 0 ? "text-danger" : ""}>{seating.waitlist.remaining}/{seating.waitlist.remaining}</span></p>
                        </div>
                    );
                })}
            </ModalBody>
            <ModalFooter>
                <Button variant="secondary" onClick={() => props.onHide()}>Close</Button>
                {selectedCourses.includes(props.section) &&
                <Button variant="danger" onClick={removeCourse}>Remove Course</Button>}
                {!selectedCourses.includes(props.section) &&
                <Button variant="primary" onClick={addCourse}>Add Course</Button>}
            </ModalFooter>
        </Modal>
    );
}
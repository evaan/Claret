import { Button, Modal, ModalBody, ModalFooter, ModalHeader } from "react-bootstrap";
import { Course, Professor, Seating, Time } from "../api/types";
import { useAtom } from "jotai";
import { profsAtom, seatingAtom, selectedCoursesAtom, timesAtom } from "../api/atoms";
import moment from "moment";

export default function SectionModal(props: {isOpen: boolean; onHide: () => void; section: Course}) {
    const [times] = useAtom(timesAtom);
    const [profs] = useAtom(profsAtom);
    const [selectedCourses, setSelectedCourses] = useAtom(selectedCoursesAtom);
    const [seatings, setSeatings] = useAtom(seatingAtom);

    async function updateSeatings(crn: string, semester: number) {
        fetch(`${process.env.NODE_ENV === "production" ? "https://api.claretformun.com" : "http://127.0.0.1:8080"}/seating?crn=${crn}&semester=${semester.toString()}`).then(response => response.json()).then((data: Seating[]) => {setSeatings(seatings.map((seating: Seating) => seating.identifier == props.section.identifier ? data[0] : seating));});
    }

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
                <p><strong>Instructors:</strong></p>
                <ul className="m-0">
                    {props.section.instructor != null && props.section.instructor.split(", ").map((instructor: string) => {
                        if (profs !== undefined && profs.filter((prof: Professor) => prof.name == instructor).length > 0) {
                            const prof = profs.filter((prof: Professor) => prof.name == instructor)[0];
                            return <li key={instructor}>{instructor} {instructor !== "TBA" && <a href={`https://www.ratemyprofessors.com/professor/${prof.id}`} rel="noreferrer" target="_blank">(RateMyProfessors Rating: {prof.rating}/5)</a>}</li>;
                        }
                        else
                            return <li key={instructor}>{instructor} {instructor !== "TBA" && <a href={`https://www.ratemyprofessors.com/search/professors/1441?q=${instructor}`} rel="noreferrer" target="_blank">(Search on RateMyProfessors)</a>}</li>;
                    })}
                </ul>
                <p><strong>Times:</strong></p>
                <ul className="m-0">
                    {times.filter((time: Time) => time.crn === props.section.crn).map((time: Time) => (
                        <li key={time.id}>{formatDateString(time.days)} - {moment(time.startTime, "HH:mm").format("hh:mm A").replace("Invalid date", "TBA")}-{moment(time.endTime, "HH:mm").format("hh:mm A").replace("Invalid date", "TBA")} - {time.location} {props.section.type.includes(", ") ? `(${time.courseType})` : ""}</li>
                    ))}
                </ul>
                {seatings.filter((seating: Seating) => seating.identifier == props.section.identifier).map((seating: Seating) => {
                    if (props.isOpen && (moment(seating.checked).isBefore(moment().subtract(1, "hours")) || seating.checked == "Never")) {
                        setTimeout(() => {updateSeatings(props.section.crn, props.section.semester);}, 200);
                    }
                    return (
                        <div key={seating.crn}>
                            <p><strong>Seats Available:</strong> <span className={seating.available == "0" ? "text-danger" : ""}>{seating.available}/{seating.max}</span></p>
                            <p><strong>Waitlist:</strong> {seating.waitlist}</p>
                            <p><strong>Last Checked:</strong> {moment(seating.checked).fromNow().replace("Invalid date", "Never")} <Button variant="link" style={{padding: "0"}} disabled={moment(seating.checked).isAfter(moment().subtract(5, "minutes"))} onClick={async () => await updateSeatings(props.section.crn, props.section.semester)}>(Update)</Button></p>
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
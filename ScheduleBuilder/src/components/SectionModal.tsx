import React from "react";
import { Button, Modal, ModalBody, ModalFooter, ModalHeader } from "react-bootstrap";
import { Course, Seating, Time } from "../api/types";
import { useAtom } from "jotai";
import { seatingAtom, selectedCoursesAtom, timesAtom } from "../api/atoms";
import moment from "moment";

export default function SectionModal(props: {isOpen: boolean; onHide: () => void; section: Course}) {
    const [times] = useAtom(timesAtom);
    const [selectedCourses, setSelectedCourses] = useAtom(selectedCoursesAtom);
    const [seatings, setSeatings] = useAtom(seatingAtom);

    async function updateSeatings(crn: string, semester: number) {
        fetch(`http://127.0.0.1:8080/seating/${crn}/${semester.toString()}`).then(response => response.json()).then((data: Seating[]) => {setSeatings(seatings.map((seating: Seating) => seating.crn == props.section.crn ? data[0] : seating));});
    }

    function formatDateString(input: string){
        input = input.replace("M", "Monday, ").replace("T", "Tuesday, ").replace("W", "Wednesday, ").replace("R", "Thursday, ").replace("F", "Friday, ");
        if (input.endsWith(", ")) input = input.slice(0, -2);
        return input;
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
                <p><strong>Campus:</strong> {props.section.campus}</p>
                <p><strong>Type:</strong> {props.section.type !== null ? props.section.type : "Unknown"}</p>
                <p><strong>Date Range:</strong> {props.section.dateRange !== null ? props.section.dateRange : "Unknown"}</p>
                <p><strong>Instructors:</strong></p>
                <ul className="m-0">
                    {/** Do keep in mind that the RMP searching doesnt work very well, potentially find a way to make it better? **/}
                    {props.section.instructor != null && props.section.instructor.split(", ").map((instructor: string) => (
                        <li key={instructor}>{instructor} {instructor !== "TBA" && <a href={`https://www.ratemyprofessors.com/search/professors/1441?q=${instructor}`} rel="noreferrer" target="_blank">(Search on RateMyProfessors)</a>}</li>
                    ))}
                </ul>
                <p><strong>Times:</strong></p>
                <ul className="m-0">
                    {times.filter((time: Time) => time.crn === props.section.crn).map((time: Time) => (
                        <li key={time.id}>{formatDateString(time.days)} - {moment(time.startTime, "HH:mm").format("hh:mm A").replace("Invalid date", "TBA")}-{moment(time.endTime, "HH:mm").format("hh:mm A").replace("Invalid date", "TBA")} - {time.location}</li>
                    ))}
                </ul>
                    {seatings.filter((seating: Seating) => seating.crn == props.section.crn).map((seating: Seating) => (
                        <div key={seating.crn}>
                            <p><strong>Seats Available:</strong> {seating.available}/{seating.max}</p>
                            <p><strong>Waitlist:</strong> {seating.waitlist}</p>
                            <p><strong>Last Checked:</strong> {moment(seating.checked).fromNow().replace("Invalid date", "Never")} <Button variant="link" style={{padding: "0"}} onClick={async () => await updateSeatings(props.section.crn, props.section.semester)}>(Update)</Button></p>
                        </div>
                    ))}
            </ModalBody>
            <ModalFooter>
                <Button variant="secondary" onClick={() => props.onHide()}>Close</Button>
                {selectedCourses.includes(props.section) &&
                <Button variant="danger" onClick={() => {props.onHide(); setSelectedCourses(selectedCourses.filter((course: Course) => course !== props.section));}}>Remove Course</Button>}
                {!selectedCourses.includes(props.section) &&
                <Button variant="primary" onClick={() => {props.onHide(); setSelectedCourses([...selectedCourses, props.section]);}}>Add Course</Button>}
            </ModalFooter>
        </Modal>
    );
}
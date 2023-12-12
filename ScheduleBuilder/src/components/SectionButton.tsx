import React from "react";
import { Button, Modal, ModalBody, ModalFooter, ModalHeader } from "react-bootstrap";
import { Course, Seating, Time } from "../api/types";
import { filterAtom, seatingAtom, selectedCoursesAtom, timesAtom } from "../api/atoms";
import { useAtom } from "jotai";
import moment from "moment";

export function SectionButton(props: {section: Course}) {
    const [times] = useAtom(timesAtom);
    const [selectedCourses, setSelectedCourses] = useAtom(selectedCoursesAtom);
    const [filter] = useAtom(filterAtom);
    const [seatings, setSeatings] = useAtom(seatingAtom);
    const [modalOpen, setModalOpen] = React.useState<boolean>(false);
    const tmp = times.filter((time: Time) => time.crn === props.section.crn).map((time: Time) => `${time.days}: ${time.startTime}-${time.endTime}`).join(", ")

    const shouldShow = (props.section.campus == "St. John's" && filter[0]) || (props.section.campus == "Grenfell" && filter[1]) || (props.section.campus == "Marine Institute" && filter[2]) || (props.section.campus == "Online" && filter[3])
        || (props.section.campus != "St. John's" && props.section.campus != "Grenfell" && props.section.campus != "Marine Institute" && props.section.campus != "Online" && filter[4])

    async function updateSeatings(crn: string) {
        fetch("http://127.0.0.1:8080/seating/" + crn).then(response => response.json()).then((data: Seating[]) => {setSeatings(seatings.map((seating: Seating) => seating.crn == props.section.crn ? data[0] : seating))})
    }

    function formatDateString(input: string){
        input = input.replace("M", "Monday, ").replace("T", "Tuesday, ").replace("W", "Wednesday, ").replace("R", "Thursday, ").replace("F", "Friday, ");
        if (input.endsWith(", ")) input = input.slice(0, -2);
        return input;
    }

    return (
        <div>
            <div>
                <Button variant={props.section.type == "Laboratory" ? "outline-primary" : "primary"} style={{width: "100%", marginBottom: "4px"}} onClick={() => setModalOpen(true)} disabled={!shouldShow}>
                    {props.section.dateRange !== null ? `${props.section.type} - ${tmp} ${props.section.type != "Laboratory" ? "- " + props.section.instructor : ""}` : `No Information, Section: ${props.section.section}`}
                </Button>
                <Modal show={modalOpen} onHide={() => setModalOpen(false)} centered>
                    <ModalHeader>{props.section.id} - {props.section.name} (Section: {props.section.section})</ModalHeader>
                    <ModalBody>
                        {/*
                        Make:    it
                        Like:    this
                        */}
                        {props.section.comment !== null &&
                        <div><strong>Comment:</strong> {props.section.comment}</div>}
                        <strong>CRN:</strong> {props.section.crn}<br/>
                        <strong>Credit Hours:</strong> {props.section.credits}<br/>
                        <strong>Section:</strong> {props.section.section}<br/>
                        <strong>Campus:</strong> {props.section.campus}<br/>
                        <strong>Type:</strong> {props.section.type !== null ? props.section.type : "Unknown"}<br/>
                        <strong>Date Range:</strong> {props.section.dateRange !== null ? props.section.dateRange : "Unknown"}<br/>
                        <strong>Instructors:</strong>
                        <ul style={{margin: "0"}}>
                            {/** Do keep in mind that the RMP searching doesnt work very well, potentially find a way to make it better? **/}
                            {props.section.instructor != null && props.section.instructor.split(", ").map((instructor: string) => (
                                <li key={instructor}>{instructor} {instructor !== "TBA" && <a href={`https://www.ratemyprofessors.com/search/professors/1441?q=${instructor}`} rel="noreferrer" target="_blank">(Search on RateMyProfessors)</a>}</li>
                            ))}
                        </ul>
                        <strong>Times:</strong>
                        <ul style={{margin: "0"}}>
                            {times.filter((time: Time) => time.crn === props.section.crn).map((time: Time) => (
                                <li key={time.id}>{formatDateString(time.days)} - {moment(time.startTime, "HH:mm").format("hh:mm A").replace("Invalid date", "TBA")}-{moment(time.endTime, "HH:mm").format("hh:mm A").replace("Invalid date", "TBA")} - {time.location}</li>
                            ))}
                        </ul>
                            {seatings.filter((seating: Seating) => seating.crn == props.section.crn).map((seating: Seating) => (
                                <div key={seating.crn}>
                                    <strong>Seats Available:</strong> {seating.available}/{seating.max}<br/>
                                    <strong>Waitlist:</strong> {seating.waitlist}<br/>
                                    <p style={{margin: "0"}}><strong>Last Checked:</strong> {moment(seating.checked).fromNow().replace("Invalid date", "Never")} <Button variant="link" style={{padding: "0"}} onClick={async () => await updateSeatings(props.section.crn)}>(Update)</Button></p>
                                </div>
                            ))}
                    </ModalBody>
                    <ModalFooter>
                        <Button variant="secondary" onClick={() => setModalOpen(false)}>Close</Button>
                        {selectedCourses.includes(props.section) &&
                        <Button variant="danger" onClick={() => {setModalOpen(false); setSelectedCourses(selectedCourses.filter((course: Course) => course !== props.section))}}>Remove Course</Button>}
                        {!selectedCourses.includes(props.section) &&
                        <Button variant="primary" onClick={() => {setModalOpen(false); setSelectedCourses([...selectedCourses, props.section])}}>Add Course</Button>}

                    </ModalFooter>
                </Modal>
            </div>
        </div>
    )
}
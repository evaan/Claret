import React from "react";
import { Button, Modal, ModalBody, ModalFooter, ModalHeader } from "react-bootstrap";
import { Course, Time } from "../api/types";
import { filterAtom, selectedCoursesAtom, timesAtom } from "../api/atoms";
import { useAtom } from "jotai";

export function SectionButton(props: {section: Course}) {
    const [times] = useAtom(timesAtom);
    const [selectedCourses, setSelectedCourses] = useAtom(selectedCoursesAtom);
    const [filter] = useAtom(filterAtom);
    const [modalOpen, setModalOpen] = React.useState<boolean>(false);
    const tmp = times.filter((time: Time) => time.crn === props.section.crn).map((time: Time) => `${time.days}: ${time.startTime}-${time.endTime}`).join(", ")

    const shouldShow = (props.section.campus == "St. John's" && filter[0]) || (props.section.campus == "Grenfell" && filter[1]) || (props.section.campus == "Marine Institute" && filter[2]) || (props.section.campus == "Online" && filter[3]) || props.section.campus == "Other"

    return (
        <div>
            {shouldShow &&
                <div>
                    <Button variant={props.section.type == "Laboratory" ? "outline-primary" : "primary"} style={{width: "100%", marginBottom: "4px"}} onClick={() => setModalOpen(true)}>
                        {props.section.dateRange !== null ? `${props.section.type} - ${tmp} ${props.section.type != "Laboratory" ? "- " + props.section.instructor : ""}` : `No Information, Section: ${props.section.section}`}
                    </Button>
                    <Modal show={modalOpen} onHide={() => setModalOpen(false)} centered>
                        <ModalHeader>{props.section.id} - {props.section.name} (Section: {props.section.section})</ModalHeader>
                        <ModalBody>
                            CRN: {props.section.crn}<br/>
                            Section: {props.section.section}<br/>
                            Type: {props.section.type}<br/>
                            Campus: {props.section.campus}<br/>
                            Instructor: {props.section.instructor} (TODO: RateMyProf Support)<br/>
                            Date Range: {props.section.dateRange}<br/>
                            {times.filter((time: Time) => time.crn === props.section.crn).map((time: Time) => (
                                <p key={time.id}>{JSON.stringify(time)}</p>
                            ))}
                        </ModalBody>
                        <ModalFooter>
                            <Button variant="secondary" onClick={() => setModalOpen(false)}>Close</Button>
                            <Button variant="primary" onClick={() => {setModalOpen(false); setSelectedCourses([...selectedCourses, props.section])}}>Add Course</Button>
                        </ModalFooter>
                    </Modal>
                </div>
            }
        </div>
    
    )
}
import React from "react";
import { Course, Time } from "../api/types";
import { seatingAtom, selectedCoursesAtom, timesAtom } from "../api/atoms";
import { useAtom } from "jotai";
import SectionModal from "./SectionModal";
import { Button } from "react-bootstrap";
import { availableSeats, formatInstructorName, formatMeeting, isOnline, isUnscheduled } from "../api/functions";

export function SectionButton(props: {section: Course}) {
    const [times] = useAtom(timesAtom);
    const [seatings] = useAtom(seatingAtom);
    const [selectedCourses] = useAtom(selectedCoursesAtom);
    const [modalOpen, setModalOpen] = React.useState<boolean>(false);
    const online = isOnline(props.section);
    const meetings = [...new Set(times.filter((time: Time) =>
        time.crn === props.section.crn && (!online || !isUnscheduled(time))
    ).map(formatMeeting))].join(" · ");
    const seating = seatings.find(seating => seating.crn === props.section.crn);
    const selected = selectedCourses.some(course => course.crn === props.section.crn);

    async function openModal() {
        setModalOpen(true);
    }

    const closeModal = () => setModalOpen(false);

    return (
        <div>
            <Button variant="outline-primary" className={`section-button ${selected ? "is-selected" : ""}`} onClick={openModal}>
                <span className="section-button-heading">
                    <strong>{props.section.type || "Section"} · {props.section.section}</strong>
                    <span className={!selected && seating && availableSeats(seating) <= 0 ? "seats-unavailable" : undefined}>
                        {selected ? "Added" : seating ? `${Math.max(0, availableSeats(seating))} seats left` : "Seats unknown"}
                    </span>
                </span>
                {(meetings || !online) && <span className="section-button-detail">{meetings || "TBA"}</span>}
                {(props.section.instructors ?? []).length > 0 && <span className="section-button-detail">
                    {[...new Set(props.section.instructors)].map(formatInstructorName).join(", ")}
                </span>}
            </Button>
            <SectionModal isOpen={modalOpen} onHide={closeModal} section={props.section}/>
        </div>
    );
}

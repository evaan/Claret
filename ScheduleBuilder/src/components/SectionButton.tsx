import React from "react";
import { Course, Seating, Time } from "../api/types";
import { seatingAtom, selectedSemesterAtom, timesAtom } from "../api/atoms";
import { useAtom } from "jotai";
import SectionModal from "./SectionModal";
import { Button } from "react-bootstrap";

export function SectionButton(props: {section: Course}) {
    const [times] = useAtom(timesAtom);
    const [modalOpen, setModalOpen] = React.useState<boolean>(false);
    const [seatings, setSeatings] = useAtom(seatingAtom);
    const tmp = times.filter((time: Time) => time.crn === props.section.crn && time.days != null).map((time: Time) => ` - ${time.days}: ${time.startTime}-${time.endTime}`).join(", ");
    const [semester] = useAtom(selectedSemesterAtom);

    async function openModal() {
        console.log(seatings);
        const existing = seatings.some(s => s.crn === props.section.crn);
        console.log(existing);
        if (existing) {
            setModalOpen(true);
        } else {
            fetch(`${process.env.NODE_ENV === "production" ? "https://api.claretformun.com" : "http://127.0.0.1:8080"}/seats?crn=${props.section.crn}&semester=${semester?.id.toString()}`)
                .then(response => response.json())
                .then((data: Seating) => {
                    setSeatings(prev => [...prev, data]);
                })
                .finally(() => {
                setModalOpen(true);
            });
        }
    }

    const closeModal = () => setModalOpen(false);

    return (
        <div>
            <Button variant={props.section.type == "Laboratory" ? "outline-primary" : "primary"} style={{width: "100%", marginBottom: "4px", color: "#fff"}} onClick={openModal}>
                {props.section.dateRange !== null ? `${props.section.type} ${tmp} ${!props.section.type.includes("Laboratory") && props.section.instructor != "" ? "- " + props.section.instructor : ""}` : `No Information, Section: ${props.section.section}`}
            </Button>
            <SectionModal isOpen={modalOpen} onHide={closeModal} section={props.section}/>
        </div>
    );
}
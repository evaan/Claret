import React from "react";
import { Course, Time } from "../api/types";
import { timesAtom } from "../api/atoms";
import { useAtom } from "jotai";
import SectionModal from "./SectionModal";
import { Button } from "react-bootstrap";

export function SectionButton(props: {section: Course}) {
    const [times] = useAtom(timesAtom);
    const [modalOpen, setModalOpen] = React.useState<boolean>(false);
    const tmp = times.filter((time: Time) => time.crn === props.section.crn).map((time: Time) => `${time.days}: ${time.startTime}-${time.endTime}`).join(", ");

    const closeModal = () => setModalOpen(false);

    return (
        <div>
            <Button variant={props.section.type == "Laboratory" ? "outline-primary" : "primary"} style={{width: "100%", marginBottom: "4px"}} onClick={async () => {setModalOpen(true);}}>
                {props.section.dateRange !== null ? `${props.section.type} - ${tmp} ${!props.section.type.includes("Laboratory") ? "- " + props.section.instructor : ""}` : `No Information, Section: ${props.section.section}`}
            </Button>
            <SectionModal isOpen={modalOpen} onHide={closeModal} section={props.section}/>
        </div>
    );
}
import React from "react";
import { Course, Time } from "../api/types";
import { timesAtom } from "../api/atoms";
import { useAtom } from "jotai";
import SectionModal from "./SectionModal";
import { Button } from "react-bootstrap";

export function SectionButton(props: {section: Course}) {
    const [times] = useAtom(timesAtom);
    const [modalOpen, setModalOpen] = React.useState<boolean>(false);
    const tmp = times.filter((time: Time) => time.crn === props.section.crn && time.days != null).map((time: Time) => ` - ${time.days}: ${time.startTime}-${time.endTime}`).join(", ");

    async function openModal() {
        setModalOpen(true);
    }

    const closeModal = () => setModalOpen(false);

    return (
        <div>
            <Button variant={props.section.type == "Laboratory" ? "outline-primary" : "primary"} style={{width: "100%", marginBottom: "4px", color: "#fff"}} onClick={openModal}>
                {`${props.section.type} ${tmp} ${!props.section.type.includes("Laboratory") && props.section.instructors.length <= 0 ? "- " + props.section.instructors.join(" - ") : ""}`}
            </Button>
            <SectionModal isOpen={modalOpen} onHide={closeModal} section={props.section}/>
        </div>
    );
}
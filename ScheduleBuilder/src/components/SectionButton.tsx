import React from "react";
import { Course, Time } from "../api/types";
import { filterAtom, timesAtom } from "../api/atoms";
import { useAtom } from "jotai";
import SectionModal from "./SectionModal";
import { Button } from "react-bootstrap";

export function SectionButton(props: {section: Course}) {
    const [times] = useAtom(timesAtom);
    const [filter] = useAtom(filterAtom);
    const [modalOpen, setModalOpen] = React.useState<boolean>(false);
    const tmp = times.filter((time: Time) => time.crn === props.section.crn).map((time: Time) => `${time.days}: ${time.startTime}-${time.endTime}`).join(", ");

    const closeModal = () => setModalOpen(false);

    const shouldShow = (props.section.campus == "St. John's" && filter[0]) || (props.section.campus == "Grenfell" && filter[1]) || (props.section.campus == "Marine Institute" && filter[2]) || (props.section.campus == "Online" && filter[3])
        || (props.section.campus != "St. John's" && props.section.campus != "Grenfell" && props.section.campus != "Marine Institute" && props.section.campus != "Online" && filter[4]);

    return (
        <div>
            <Button variant={props.section.type == "Laboratory" ? "outline-primary" : "primary"} style={{width: "100%", marginBottom: "4px"}} onClick={() => setModalOpen(true)} disabled={!shouldShow}>
                {props.section.dateRange !== null ? `${props.section.type} - ${tmp} ${props.section.type != "Laboratory" ? "- " + props.section.instructor : ""}` : `No Information, Section: ${props.section.section}`}
            </Button>
            <SectionModal isOpen={modalOpen} onHide={closeModal} section={props.section}/>
        </div>
    );
}
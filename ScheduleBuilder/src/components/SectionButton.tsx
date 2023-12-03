import React from "react";
import { Button } from "react-bootstrap";
import { Course, Time } from "../api/types";
import { timesAtom } from "../api/atoms";
import { useAtom } from "jotai";

export function SectionButton(props: {section: Course}) {
    const [times] = useAtom(timesAtom);
    const tmp = times.filter((time: Time) => time.crn === props.section.crn).map((time: Time) => `${time.days}: ${time.startTime}-${time.endTime}`).join(", ")

    if (props.section.type !== null) return (
        <Button variant={props.section.type == "Laboratory" ? "outline-primary" : "primary"} style={{width: "100%", marginBottom: "4px"}}>{props.section.type} - {tmp} {props.section.type != "Laboratory" ? "- " + props.section.instructor : ""}</Button>
    )
    else return (
        <Button variant={props.section.type == "Laboratory" ? "outline-primary" : "primary"} style={{width: "100%", marginBottom: "4px"}}>No Course Info - Section {props.section.section}</Button>
    )
}
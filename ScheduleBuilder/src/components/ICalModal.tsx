import React from "react";
import { Button, Modal, ModalBody, ModalHeader, InputGroup, Form } from "react-bootstrap";
import { selectedCoursesAtom } from "../api/atoms";
import { useAtom } from "jotai";

export default function ICalModal(props: { isOpen: boolean; onHide: () => void }) {
    const selectedCourses = useAtom(selectedCoursesAtom)[0];
    
    const generateiCalURL = () => {
        const crnString: string = selectedCourses.map(obj => obj.crn).join(",");
        return "https://ics.claretformun.com/feed.ics?crns=" + crnString;
    };

    return (
        <Modal show={props.isOpen} onHide={() => props.onHide()} centered>
            <ModalHeader>Add to External Calendar Application</ModalHeader>
            <ModalBody>
                <p className="mb-3">You can add your schedule to an external calendar application (such as Google Calendar, Apple Calendar, Outlook, and Thunderbird)</p>
                <p className="mb-3">Subscribe to the following iCalendar in your calendar application:</p>

                <InputGroup className="mb-3">
                    <Form.Control value={generateiCalURL()}/>
                    <Button variant="outline-secondary">
                        Copy
                    </Button>
                </InputGroup>
            </ModalBody>
        </Modal>
    );
}
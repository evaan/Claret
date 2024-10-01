import { Button, Modal, ModalBody, ModalHeader, InputGroup, Form, CloseButton } from "react-bootstrap";
import { selectedCoursesAtom, selectedSemesterAtom } from "../api/atoms";
import { useAtom } from "jotai";
import { useState } from "react";

export default function ICalModal(props: { isOpen: boolean; onHide: () => void }) {
    const selectedCourses = useAtom(selectedCoursesAtom)[0];
    const [selectedSemester] = useAtom(selectedSemesterAtom);
    const generateiCalURL = () => {
        const crnString: string = selectedCourses.map(obj => obj.crn).join(",");
        return `https://ics.claretformun.com/feed.ics?semester=${selectedSemester?.id}&crn=${crnString}`;
    };
    
    const copyURL = () => {
        const url: string = generateiCalURL();
        navigator.clipboard.writeText(url);
        setIsCopied(true);
        setTimeout(() => {setIsCopied(false);}, 1000);
    };

    const [isCopied, setIsCopied] = useState(false);

    return (
        <Modal show={props.isOpen} onHide={() => props.onHide()} centered>
            <ModalHeader>
                Add to External Calendar Application
                <CloseButton onClick={() => props.onHide()}/>
            </ModalHeader>
            <ModalBody>
                <p className="mb-3">You can add your schedule to an external calendar application (such as Google Calendar, Apple Calendar, Outlook, and Thunderbird)</p>
                <p className="mb-3">Subscribe to the following iCalendar in your calendar application:</p>

                <InputGroup className="mb-3">
                    <Form.Control value={generateiCalURL()}/>
                    <Button variant="outline-secondary" onClick={copyURL}>
                        {isCopied ? "Copied!" : "Copy"}
                    </Button>
                </InputGroup>
            </ModalBody>
        </Modal>
    );
}

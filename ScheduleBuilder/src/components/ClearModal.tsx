import { useAtom } from "jotai";
import React from "react";
import { Button, Modal, ModalHeader, ModalFooter } from "react-bootstrap";
import { selectedCoursesAtom } from "../api/atoms";

export default function ICalModal(props: { isOpen: boolean; onHide: () => void }) {
    const [, setSelectedCourses] = useAtom(selectedCoursesAtom);

    return (
        <Modal show={props.isOpen} onHide={() => props.onHide()} centered>
            <ModalHeader>
                Are you sure you want to clear courses?
            </ModalHeader>
            <ModalFooter>
                <Button variant="secondary" onClick={() => props.onHide()}>No</Button>
                <Button variant="danger" onClick={() => {setSelectedCourses([]); props.onHide();}}>Yes</Button>
            </ModalFooter>
        </Modal>
    );
}

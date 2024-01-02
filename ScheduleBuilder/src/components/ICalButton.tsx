import React from "react";
import ICalModal from "./ICalModal";
import { Button } from "react-bootstrap";

export function ICalButton() {

    const [modalOpen, setModalOpen] = React.useState<boolean>(false);
    const closeModal = () => setModalOpen(false);

    return (
        <div>
            <Button className="mt-3" onClick={() => setModalOpen(true)}>
                Subscribe to Calendar
            </Button>
            <ICalModal isOpen={modalOpen} onHide={closeModal}/>
        </div>
    );
}
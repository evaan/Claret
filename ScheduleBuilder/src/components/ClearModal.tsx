import { useAtom } from "jotai";
import { Button, Modal, ModalHeader, ModalFooter } from "react-bootstrap";
import { selectedCoursesAtom, selectedSemesterAtom } from "../api/atoms";

export default function ICalModal(props: { isOpen: boolean; onHide: () => void }) {
    const [, setSelectedCourses] = useAtom(selectedCoursesAtom);
    const [selectedSemester] = useAtom(selectedSemesterAtom);

    return (
        <Modal show={props.isOpen} onHide={() => props.onHide()} centered>
            <ModalHeader>
                Are you sure you want to clear courses?
            </ModalHeader>
            <ModalFooter>
                <Button variant="secondary" onClick={() => props.onHide()}>No</Button>
                <Button variant="danger" onClick={() => {setSelectedCourses([]); window.history.replaceState(null, "", `?semester=${selectedSemester?.id}`); props.onHide();}}>Yes</Button>
            </ModalFooter>
        </Modal>
    );
}

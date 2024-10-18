import { useAtom } from "jotai";
import { Button, Modal, ModalHeader, ModalFooter, ModalBody } from "react-bootstrap";
import { selectedCoursesAtom } from "../api/atoms";

export default function ExamModal(props: { isOpen: boolean; onHide: () => void }) {
    const [selectedCourses] = useAtom(selectedCoursesAtom);

    return (
        <Modal show={props.isOpen} onHide={() => props.onHide()} centered>
            <ModalHeader>
                Exam Times  
            </ModalHeader>
            <ModalBody>
                <table style={{width: "100%"}}>
                    <thead>
                        <tr>
                            <th>Course</th>
                            <th>Time</th>
                            <th>Location</th>
                        </tr>
                    </thead>
                    <tbody>
                        {selectedCourses.map((course: Course) => (
                            <tr>
                                <th>{course.id}-{course.section}</th>
                                <th>asd</th>
                                <th>asd</th>
                            </tr>
                        ))}
                        <tr>
                            <th>TEST-1000</th>
                            <th>0:00</th>
                            <th>idk</th>
                        </tr>
                    </tbody>
                </table>
            </ModalBody>
            <ModalFooter>
                <Button variant="secondary" onClick={() => props.onHide()}>Close</Button>
            </ModalFooter>
        </Modal>
    );
}

import { useAtom } from "jotai";
import { Button, Modal, ModalHeader, ModalFooter, ModalBody } from "react-bootstrap";
import { examsAtom, selectedCoursesAtom } from "../api/atoms";
import { Course, ExamTime } from "../api/types";

export default function ExamModal(props: { isOpen: boolean; onHide: () => void }) {
    const [selectedCourses] = useAtom(selectedCoursesAtom);
    const [exams] = useAtom(examsAtom);

    if (!selectedCourses.some((course: Course) => exams.some((exam: ExamTime) => course.crn === exam.crn))) {
        return <Modal show={props.isOpen} onHide={() => props.onHide()} centered>
            <ModalHeader>
                Exam Times
            </ModalHeader>
            <ModalBody>
                <p>No exam times could be found with the selected courses.</p>
            </ModalBody>
            <ModalFooter>
                <Button variant="secondary" onClick={() => props.onHide()}>Close</Button>
            </ModalFooter>
        </Modal>
    }

    return (
        <Modal show={props.isOpen} onHide={() => props.onHide()} centered dialogClassName="examModal">
            <ModalHeader>
                Exam Times  
            </ModalHeader>
            <ModalBody>
                <table style={{width: "100%"}}>
                    <thead>
                        <tr>
                            <th style={{textAlign: "center"}}>Course</th>
                            <th style={{textAlign: "center"}}>Time</th>
                            <th style={{textAlign: "center"}}>Location</th>
                        </tr>
                    </thead>
                    <tbody>
                        {selectedCourses.map((course: Course) => {
                            const exam: ExamTime | undefined = exams.filter((exam: ExamTime) => exam.crn == course.crn && exam.section == course.section)[0];
                            if (exam === undefined) return null;
                            return (
                                <tr>
                                    <th style={{textAlign: "center"}}>{course.id}-{course.section}</th>
                                    <th style={{textAlign: "center"}}>{exam.location}</th>
                                    <th style={{textAlign: "center"}}>{exam.time}</th>
                                </tr>
                            );
                        })}
                    </tbody>
                </table>
            </ModalBody>
            <ModalFooter>
                <Button variant="secondary" onClick={() => props.onHide()}>Close</Button>
            </ModalFooter>
        </Modal>
    );
}

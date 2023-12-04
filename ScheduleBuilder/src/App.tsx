import { Accordion, Button, Col, Row } from "react-bootstrap";
import React from "react";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faGithub } from "@fortawesome/free-brands-svg-icons";
import SubjectAccordion from "./components/SubjectAccordion";
import { coursesAtom, selectedCoursesAtom, subjectsAtom, timesAtom } from "./api/atoms";
import { useAtom } from "jotai";
import { Course, Subject, Time } from "./api/types";

export default function App() {
    const [subjects, setSubjects] = useAtom(subjectsAtom);
    const [, setCourses] = useAtom(coursesAtom);
    const [, setTimes] = useAtom(timesAtom);

    React.useEffect(() => {
        fetch("http://localhost:8080/all").then(response => response.json()).then((data: {subjects: Subject[], courses: Course[], times: Time[]}) => {
            setSubjects(data.subjects);
            setCourses(data.courses);
            setTimes(data.times);
        });
    }, [])

    const [selectedCourses] = useAtom(selectedCoursesAtom);

    return (
        <div>
            <Row style={{marginLeft: "2.5%", marginRight: "2.5%"}}>
                <Col xs={12} md={4}>
                    <h1 className="text-center">Courses</h1>
                    <Accordion style={{overflowY: "auto"}} className="h-md-75">
                    {subjects.map((subject, index) => (
                        <SubjectAccordion subject={subject} index={index} key={index} /> 
                    ))}
                    </Accordion>
                </Col>
                <Col xs={12} md={8}>
                    <h1 className="text-center">Schedule</h1>
                    {selectedCourses.map((course: Course) => (
                        <h1 key={course.crn}>{JSON.stringify(course)}</h1>
                    ))}
                </Col>
            </Row>
            <div style={{position: "absolute", top: "4px", right: "4px"}}>
                <Button variant="outline-info" onClick={() => window.open("https://github.com/evaan/Claret", '_blank')!.focus()}><FontAwesomeIcon icon={faGithub} /></Button>
            </div>
        </div>
    )
}
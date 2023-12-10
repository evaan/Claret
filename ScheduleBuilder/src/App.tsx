import { Accordion, Button, Col, Form, Row } from "react-bootstrap";
import React from "react";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faGithub } from "@fortawesome/free-brands-svg-icons";
import SubjectAccordion from "./components/SubjectAccordion";
import { coursesAtom, filterAtom, selectedTabAtom, subjectsAtom, timesAtom } from "./api/atoms";
import { useAtom } from "jotai";
import { Course, Subject, Time } from "./api/types";
import Schedule from "./components/Schedule";

export default function App() {
    const [subjects, setSubjects] = useAtom(subjectsAtom);
    const [, setCourses] = useAtom(coursesAtom);
    const [, setTimes] = useAtom(timesAtom);
    const [filters, setFilters] = useAtom(filterAtom);
    const [selectedTab, setSelectedTab] = useAtom(selectedTabAtom);

    React.useEffect(() => {
        fetch("http://localhost:8080/all").then(response => response.json()).then((data: {subjects: Subject[], courses: Course[], times: Time[]}) => {
            setSubjects(data.subjects);
            setCourses(data.courses);
            setTimes(data.times);
        })
    }, [])

    return (
        <div>
            {/**<LoadingScreen />**/}
            <Row style={{marginLeft: "2.5%", marginRight: "2.5%"}}>
                <Col xs={12} md={4}>
                    <h1 className="text-center">Courses</h1>
                    <Form style={{display: "flex", justifyContent: "center"}}>
                        <Form.Check inline type="switch" label="St. John's" defaultChecked={true} onChange={(event) => setFilters([event.target.checked, filters[1], filters[2], filters[3], filters[4]])} />
                        <Form.Check inline type="switch" label="Grenfell" defaultChecked={false} onChange={(event) => setFilters([filters[0], event.target.checked, filters[2], filters[3], filters[4]])} />
                        <Form.Check inline type="switch" label="Marine Institute" defaultChecked={false} onChange={(event) => setFilters([filters[0], filters[1], event.target.checked, filters[3], filters[4]])} />
                        <Form.Check inline type="switch" label="Online" defaultChecked={true} onChange={(event) => setFilters([filters[0], filters[1], filters[2], event.target.checked, filters[4]])} />
                        <Form.Check inline type="switch" label="Others" defaultChecked={false} onChange={(event) => setFilters([filters[0], filters[1], filters[2], filters[3], event.target.checked])} />
                    </Form>
                    <Accordion style={{overflowY: "auto"}} className="h-md-75" onSelect={(event) => {
                        //jank solution but it reduces the amount of lag the site has SIGNIFICANTLY
                        setSelectedTab([event, selectedTab[0]]);
                        setTimeout(() => {
                            setSelectedTab([event, "-1"]);
                        }, 500);
                    }}>
                        {subjects.map((subject, index) => (
                            <SubjectAccordion subject={subject} index={index} key={index} /> 
                        ))}
                    </Accordion>
                </Col>
                <Col xs={12} md={8}>
                    <h1 className="text-center">Schedule</h1>
                    <Schedule />
                </Col>
            </Row>
            <div style={{position: "absolute", top: "4px", right: "4px"}}>
                <Button variant="outline-info" onClick={() => window.open("https://github.com/evaan/Claret", '_blank')!.focus()}><FontAwesomeIcon icon={faGithub} /></Button>
            </div>
        </div>
    )
}
import { Accordion, Button, Col, Form, Row } from "react-bootstrap";
import React from "react";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faGithub } from "@fortawesome/free-brands-svg-icons";
import SubjectAccordion from "./components/SubjectAccordion";
import { coursesAtom, filterAtom, seatingAtom, selectedCoursesAtom, selectedTabAtom, subjectsAtom, timesAtom } from "./api/atoms";
import { useAtom } from "jotai";
import { Course, Seating, Subject, Time } from "./api/types";
import Schedule from "./components/Schedule";
import { shouldShow } from "./api/functions";

export default function App() {
    const [subjects, setSubjects] = useAtom(subjectsAtom);
    const [courses, setCourses] = useAtom(coursesAtom);
    const [, setTimes] = useAtom(timesAtom);
    const [, setSeating] = useAtom(seatingAtom);
    const [filters, setFilters] = useAtom(filterAtom);
    const [selectedTab, setSelectedTab] = useAtom(selectedTabAtom);
    const [selectedCourses, setSelectedCourses] = useAtom(selectedCoursesAtom);

    React.useEffect(() => {
        fetch((process.env.NODE_ENV === "production" ? "https://api.claretformun.com" : "http://127.0.0.1:8080")+"/all").then(response => response.json()).then((data: {subjects: Subject[], courses: Course[], times: Time[], seatings: Seating[]}) => {
            setSubjects(data.subjects);
            setCourses(data.courses);
            setTimes(data.times);
            setSeating(data.seatings);
            const crnsParam = new URLSearchParams(window.location.search).get("crns")?.split(",");
            const newCourses: Course[] = [];
            if (crnsParam !== undefined) {
                data.courses.forEach((course: Course) => {
                    if (crnsParam.includes(course.crn) && !selectedCourses.includes(course)) newCourses.push(course);
                });
            }
            setSelectedCourses(selectedCourses.concat(newCourses));
        });
    }, []);

    return (
        <div>
            <Row style={{marginLeft: "2.5%", marginRight: "2.5%"}}>
                <Col xs={12} md={8} className="order-md-2">
                    <h1 className="text-center">Schedule</h1>
                    <div style={{display: "flex", justifyContent: "center"}}>
                        {/* TODO: import/export buttons */}
                    </div>
                    <Schedule />
                </Col>
                <Col xs={12} md={4} className="order-md-1">
                    <h1 className="text-center">Courses</h1>
                    <div className="d-flex flex-wrap justify-content-center">
                        <Form.Check inline type="switch" label="St. John's" defaultChecked={true} onChange={(event) => setFilters([event.target.checked, filters[1], filters[2], filters[3], filters[4]])} />
                        <Form.Check inline type="switch" label="Grenfell" defaultChecked={false} onChange={(event) => setFilters([filters[0], event.target.checked, filters[2], filters[3], filters[4]])} />
                        <Form.Check inline type="switch" label="Marine Institute" defaultChecked={false} onChange={(event) => setFilters([filters[0], filters[1], event.target.checked, filters[3], filters[4]])} />
                        <div style={{flexBasis: "100%", height: "0"}} />
                        <Form.Check inline type="switch" label="Online" defaultChecked={true} onChange={(event) => setFilters([filters[0], filters[1], filters[2], event.target.checked, filters[4]])} />
                        <Form.Check inline type="switch" label="Others" defaultChecked={false} onChange={(event) => setFilters([filters[0], filters[1], filters[2], filters[3], event.target.checked])} />
                    </div>
                    <Accordion style={{overflowY: "auto"}} onSelect={(event) => {
                        //jank solution but it reduces the amount of lag the site has SIGNIFICANTLY
                        setSelectedTab([event, selectedTab[0]]);
                        setTimeout(() => {
                            setSelectedTab([event, "-1"]);
                        }, 500);
                    }}>
                        {subjects.map((subject, index) => {
                            if (courses.filter((course: Course) => course.subject == subject.name && shouldShow(course, filters)).length > 0) return (<SubjectAccordion subject={subject} index={index} key={index} />);
                        })}
                    </Accordion>
                </Col>
            </Row>
            <div style={{position: "absolute", top: "4px", right: "4px"}}>
                <Button style={{marginRight: "4px"}} variant="outline-info" onClick={() => window.open("https://github.com/evaan/Claret/issues", "_blank")!.focus()}>Issues? Leave them here!</Button>
                <Button variant="outline-info" onClick={() => window.open("https://github.com/evaan/Claret", "_blank")!.focus()}><FontAwesomeIcon icon={faGithub} aria-label="GitHub"/></Button>
            </div>
        </div>
    );
}
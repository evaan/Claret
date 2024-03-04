import { Accordion, Col, Form, Row } from "react-bootstrap";
import React from "react";
import SubjectAccordion from "./components/SubjectAccordion";
import { coursesAtom, filterAtom, searchQueryAtom, seatingAtom, selectedCoursesAtom, selectedSemesterAtom, selectedTabAtom, subjectsAtom, timesAtom } from "./api/atoms";
import { useAtom } from "jotai";
import { Course, Seating, Semester, Subject, Time } from "./api/types";
import Schedule from "./components/Schedule";
import { shouldShow } from "./api/functions";
import SearchBar from "./components/SearchBar";

export default function App() {
    const [subjects, setSubjects] = useAtom(subjectsAtom);
    const [courses, setCourses] = useAtom(coursesAtom);
    const [, setTimes] = useAtom(timesAtom);
    const [, setSeating] = useAtom(seatingAtom);
    const [filters, setFilters] = useAtom(filterAtom);
    const [selectedTab, setSelectedTab] = useAtom(selectedTabAtom);
    const [selectedCourses, setSelectedCourses] = useAtom(selectedCoursesAtom);
    const [searchQuery] = useAtom(searchQueryAtom);

    const [selectedSemester, setSelectedSemester] = useAtom(selectedSemesterAtom);
    const [semesters, setSemesters] = React.useState<Semester[]>([]);

    React.useEffect(() => {
        fetch((process.env.NODE_ENV === "production" ? "https://api.claretformun.com" : "http://127.0.0.1:8080")+"/semesters").then(response => response.json()).then((data: Semester[]) => {
            setSemesters(data);
            const params = new URLSearchParams(window.location.search);
            let semester = "";
            semester = data.filter((semester: Semester) => (params.get("semester") || "") == semester.id.toString()).length >= 1 ? params.get("semester") || data.filter((semester: Semester) => semester.latest)[0].id.toString() : data.filter((semester: Semester) => semester.latest)[0].id.toString(); 
            params.set("semester", semester);
            window.history.replaceState(null, "", `?${params}`);
            setSelectedSemester(data.filter((semester1: Semester) => semester == semester1.id.toString())[0]);
        });
    }, []);

    React.useEffect(() => {
        if (selectedSemester == null) return;
        fetch((process.env.NODE_ENV === "production" ? "https://api.claretformun.com" : "http://127.0.0.1:8080")+"/all?semester=" + selectedSemester.id).then(response => response.json()).then((data: {subjects: Subject[], courses: Course[], times: Time[], seatings: Seating[]}) => {
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
            setSelectedCourses(newCourses);
        });
    }, [selectedSemester]);

    return (
        <div className="mb-4">
            <Row style={{marginLeft: "2.5%", marginRight: "2.5%"}}>
                <Col xs={12} md={8} className="order-md-2">
                    <h1 className="text-center">Schedule</h1>
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
                    <Form.Select aria-label="Selected Semester" value={selectedSemester?.id} onChange={(e: React.ChangeEvent<HTMLSelectElement>) => {
                            const newSemester: Semester = semesters.filter((semester: Semester) => semester.id.toString() == e.target.value)[0];
                            setSelectedSemester(newSemester);
                            const params = new URLSearchParams(window.location.search);
                            params.set("semester", newSemester?.id.toString() || params.get("semester") || "");
                            window.history.replaceState(null, "", `?${params}`);
                        }}>
                        {semesters.sort((function(a: Semester, b: Semester) {return b.id - a.id;})).map((semester: Semester) => (
                            <option value={semester.id} key={semester.id}>{semester.name}</option>
                        ))}
                    </Form.Select>
                    <SearchBar />
                    <Accordion style={{overflowY: "auto"}} onSelect={(event) => {
                        //jank solution but it reduces the amount of lag the site has SIGNIFICANTLY
                        setSelectedTab([event, selectedTab[0]]);
                        setTimeout(() => {
                            if (event !== null) setSelectedTab([event, "-1"]);
                        }, 500);
                    }}>
                        {subjects.sort(function(a, b) {if (a.friendlyName < b.friendlyName) return -1; else return 1;}).map((subject, index) => {
                            if (courses.filter((course: Course) => course.subject == subject.name && shouldShow(course, filters) && (searchQuery == "" || course.id.toLowerCase().includes(searchQuery.toLowerCase()) || course.subjectFull.toLowerCase().includes(searchQuery.toLowerCase()))).length > 0) {
                                return (<SubjectAccordion subject={subject} index={index} key={index} />);
                            }
                        })}
                    </Accordion>
                </Col>
            </Row>
        </div>
    );
}
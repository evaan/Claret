import { Accordion } from "react-bootstrap";
import { coursesAtom, filteredCoursesAtom, selectedTabAtom } from "../api/atoms";
import { useAtom } from "jotai";
import { Course, Subject } from "../api/types";
import { SectionButton } from "./SectionButton";

export default function SubjectAccordion(props: {subject: Subject, index: string,}) {
    const [filteredCourses] = useAtom(filteredCoursesAtom);
    const [courses] = useAtom(coursesAtom);
    const subjectCourses = courses.filter((course: Course) => course.subject === props.subject.code);
    const uniqueCourses: [string, string, string][] = [];
    const [selectedTab] = useAtom(selectedTabAtom);
    const sortingOrder: {[name: string]: number} = {"Lecture": 1, "Laboratory": 2, "World Wide Web": 3};
    
    subjectCourses.forEach((course: Course) => {
        if(!uniqueCourses.some(([id]) => id === course.id) && (subjectCourses.filter((course1: Course) => course1.id === course.id && course1.name !== "Laboratory").length === 0 || course.name !== "Laboratory")) uniqueCourses.push([course.id, course.name, course.subject]);
    });


    return (
        <Accordion.Item eventKey={props.index}>
            <Accordion.Header>
                {props.subject.description}
            </Accordion.Header>
            <Accordion.Body>
                {selectedTab.includes(props.index) &&
                    <Accordion>
                    {uniqueCourses.sort(function(x, y) {return x>y ? 1: -1;}).map((course: [id: string, name: string, subject: string]) => {
                        if (filteredCourses.some(course1 => course1.id === course[0])) return (
                            <Accordion.Item eventKey={course[0]} key={course[0]}>
                                <Accordion.Header>
                                    {course[0]} - {course[1]}
                                </Accordion.Header>
                                <Accordion.Body>
                                    {filteredCourses.filter((section: Course) => course[0] === section.id).sort(function(a, b) {return (sortingOrder[a.type] ?? 0) - (sortingOrder[b.type] ?? 0);}).map((section: Course) => {                                        return (
                                            <SectionButton section={section} key={section.crn} />
                                        );
                                    })}
                                </Accordion.Body>
                            </Accordion.Item>
                        );
                    })}
                </Accordion>
                }
            </Accordion.Body>
        </Accordion.Item>
    );
}

import React from "react";
import { Accordion } from "react-bootstrap";
import { coursesAtom } from "../api/atoms";
import { useAtom } from "jotai";
import { Course, Subject } from "../api/types";
import { SectionButton } from "./SectionButton";

export default function SubjectAccordion(props: {subject: Subject, index: number}) {
    const [courses] = useAtom(coursesAtom);
    const subjectCourses = courses.filter((course: Course) => course.subject === props.subject.name);
    const [uniqueCourses, setUniqueCourses] = React.useState<[string, string][]>([]);

    React.useEffect(() => {
        subjectCourses.forEach((course: Course) => {
            if(!JSON.stringify(uniqueCourses).includes(course.id)) uniqueCourses.push([course.id, course.name]);
        });
        setUniqueCourses(uniqueCourses.sort(function(x, y) {return x>y ? 1: -1}));
    }, [])

    return (
        <Accordion.Item eventKey={props.index.toString()}>
            <Accordion.Header>
                {props.subject.friendlyName}
            </Accordion.Header>
            <Accordion.Body>
                <Accordion>
                    {uniqueCourses.map((course: [id: string, name: string]) => (
                        <Accordion.Item eventKey={course[0]} key={course[0]}>
                            <Accordion.Header>
                                {course[0]} - {course[1]}
                            </Accordion.Header>
                            <Accordion.Body>
                                {courses.filter((section: Course) => course[0] === section.id).map((section: Course) => (
                                    <SectionButton section={section} key={section.crn} />
                                ))}
                            </Accordion.Body>
                        </Accordion.Item>
                    ))}
                </Accordion>
            </Accordion.Body>
        </Accordion.Item>
    )
}
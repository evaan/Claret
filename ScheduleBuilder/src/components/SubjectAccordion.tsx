import React from "react";
import { Accordion } from "react-bootstrap";
import { coursesAtom, filterAtom, searchQueryAtom, selectedTabAtom } from "../api/atoms";
import { useAtom } from "jotai";
import { Course, Subject } from "../api/types";
import { SectionButton } from "./SectionButton";
import { shouldShow } from "../api/functions";

export default function SubjectAccordion(props: {subject: Subject, index: number,}) {
    const [filters] = useAtom(filterAtom);
    const [courses] = useAtom(coursesAtom);
    const subjectCourses = courses.filter((course: Course) => course.subject === props.subject.name);
    const uniqueCourses: [string, string, string][] = [];
    const [selectedTab] = useAtom(selectedTabAtom);
    const sortingOrder: {[name: string]: number} = {"Lecture": 1, "Laboratory": 2, "World Wide Web": 3};
    const [searchQuery] = useAtom(searchQueryAtom);
    
    subjectCourses.forEach((course: Course) => {
        if(!JSON.stringify(uniqueCourses).includes(course.id) && (subjectCourses.filter((course1: Course) => course.id == course.id && course1.name !== "Laboratory").length == 0 || course.name !== "Laboratory")) uniqueCourses.push([course.id, course.name, course.subjectFull]);
    });

    return (
        <Accordion.Item eventKey={props.index.toString()}>
            <Accordion.Header>
                {props.subject.friendlyName}
            </Accordion.Header>
            <Accordion.Body>
                {selectedTab.includes(props.index.toString()) &&
                    <Accordion>
                    {uniqueCourses.sort(function(x, y) {return x>y ? 1: -1;}).map((course: [id: string, name: string, subject: string]) => {
                        console.log(course[1]);
                        if (courses.filter((course1: Course) => course1.id == course[0] && shouldShow(course1, filters) && (searchQuery == "" || course[0].toLowerCase().includes(searchQuery.toLowerCase()) || course[2].toLowerCase().includes(searchQuery.toLowerCase()) || course[1].toLowerCase().includes(searchQuery.toLowerCase()))).length > 0) return (
                            <Accordion.Item eventKey={course[0]} key={course[0]}>
                                <Accordion.Header>
                                    {course[0]} - {course[1]}
                                </Accordion.Header>
                                <Accordion.Body>
                                    {courses.filter((section: Course) => course[0] === section.id && shouldShow(section, filters)).sort(function(a, b) {return (sortingOrder[a.type] ?? 0) - (sortingOrder[b.type] ?? 0);}).map((section: Course) => {
                                        return (
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
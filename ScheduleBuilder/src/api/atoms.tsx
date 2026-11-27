import { atom } from "jotai";
import { Course, CourseFilters, ExamTime, Professor, Seating, Semester, Subject, Time } from "./types";
import { AccordionEventKey } from "react-bootstrap/esm/AccordionContext";
import { defaultFilters, matchesFilters, matchesSearch } from "./functions";

export const subjectsAtom = atom<Subject[]>([]);
export const coursesAtom = atom<Course[]>([]);
export const timesAtom = atom<Time[]>([]);
export const seatingAtom = atom<Seating[]>([]);
export const profsAtom = atom<Professor[]>([]);
export const examsAtom = atom<ExamTime[]>([]);
export const selectedCoursesAtom = atom<Course[]>([]);
export const selectedTabAtom = atom<[AccordionEventKey, AccordionEventKey]>(["-1", "-1"]);
export const filterAtom = atom<CourseFilters>(defaultFilters());
export const searchQueryAtom = atom<string>("");
export const selectedSemesterAtom = atom<Semester | null>(null);

export const filteredCoursesAtom = atom(get => {
    const meetings = new Map<string, Time[]>();
    for (const time of get(timesAtom)) {
        const sectionTimes = meetings.get(time.crn) ?? [];
        sectionTimes.push(time);
        meetings.set(time.crn, sectionTimes);
    }
    const seats = new Map(get(seatingAtom).map(seating => [seating.crn, seating]));
    const subjects = new Map(get(subjectsAtom).map(subject => [subject.code, subject]));
    const filters = get(filterAtom);
    const query = get(searchQueryAtom);
    return get(coursesAtom).filter(course => {
        const times = meetings.get(course.crn) ?? [];
        return matchesFilters(course, filters, times, seats.get(course.crn)) &&
            matchesSearch(course, query, times, subjects.get(course.subject));
    });
});

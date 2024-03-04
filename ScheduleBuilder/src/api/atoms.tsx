import { atom } from "jotai";
import { Course, Seating, Semester, Subject, Time } from "./types";
import { AccordionEventKey } from "react-bootstrap/esm/AccordionContext";

export const subjectsAtom = atom<Subject[]>([]);
export const coursesAtom = atom<Course[]>([]);
export const timesAtom = atom<Time[]>([]);
export const seatingAtom = atom<Seating[]>([]);
export const selectedCoursesAtom = atom<Course[]>([]);
export const selectedTabAtom = atom<[AccordionEventKey, AccordionEventKey]>(["-1", "-1"]);
export const filterAtom = atom<[boolean, boolean, boolean, boolean, boolean]>([true, false, false, true, false]);
export const searchQueryAtom = atom<string>("");
export const selectedSemesterAtom = atom<Semester | null>(null);
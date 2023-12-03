import { atom } from "jotai";
import { Course, Subject, Time } from "./types";

export const subjectsAtom = atom<Subject[]>([]);
export const coursesAtom = atom<Course[]>([]);
export const timesAtom = atom<Time[]>([]);
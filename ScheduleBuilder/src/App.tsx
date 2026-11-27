import { Accordion, Alert, Button, Form } from "react-bootstrap";
import React from "react";
import SubjectAccordion from "./components/SubjectAccordion";
import { coursesAtom, examsAtom, filteredCoursesAtom, profsAtom, seatingAtom, selectedCoursesAtom, selectedSemesterAtom, selectedTabAtom, subjectsAtom, timesAtom } from "./api/atoms";
import { useAtom } from "jotai";
import { Course, ExamTime, Professor, Seating, Semester, Subject, Time } from "./api/types";
import Schedule from "./components/Schedule";
import SearchBar from "./components/SearchBar";
import { useQuery } from "@tanstack/react-query";
import MobileViews from "./components/MobileViews";

const apiURL = process.env.NODE_ENV === "production" ? "https://api.claretformun.com" : "http://127.0.0.1:8080";
const seatRefreshInterval = 10 * 60 * 1000;

async function fetchJSON<T>(path: string, signal: AbortSignal): Promise<T> {
    const response = await fetch(apiURL + path, { signal });
    if (!response.ok) throw new Error("Course data could not be loaded. Please try again.");
    return response.json();
}

export default function App() {
    const [subjects, setSubjects] = useAtom(subjectsAtom);
    const [, setCourses] = useAtom(coursesAtom);
    const [, setTimes] = useAtom(timesAtom);
    const [, setSeating] = useAtom(seatingAtom);
    const [, setProfs] = useAtom(profsAtom);
    const [, setExams] = useAtom(examsAtom);
    const [filteredCourses] = useAtom(filteredCoursesAtom);
    const [selectedTab, setSelectedTab] = useAtom(selectedTabAtom);
    const [, setSelectedCourses] = useAtom(selectedCoursesAtom);
    const [selectedSemester, setSelectedSemester] = useAtom(selectedSemesterAtom);
    const semesterID = selectedSemester?.id;
    const params = new URLSearchParams(window.location.search);
    const scheduleOnly = params.get("view") === "schedule";
    const scheduleCRNs = [...new Set((params.get("crns") ?? "").split(",").map(crn => crn.trim()).filter(Boolean))].sort().join(",");
    const semestersQuery = useQuery({
        queryKey: ["semesters"],
        queryFn: ({ signal }) => fetchJSON<Semester[]>("/semesters", signal),
        staleTime: seatRefreshInterval,
    });
    const frontendQuery = useQuery({
        queryKey: ["frontend", semesterID],
        queryFn: ({ signal }) => fetchJSON<{subjects: Subject[]; courses: Course[]; times: Time[]; seatings: Seating[]; profs: Professor[]; exams?: ExamTime[]}>(
            "/frontend?semester=" + semesterID, signal,
        ),
        enabled: semesterID != null && !scheduleOnly,
        staleTime: seatRefreshInterval,
        gcTime: 30 * 60 * 1000,
        refetchInterval: seatRefreshInterval,
    });
    const scheduleQuery = useQuery({
        queryKey: ["schedule", semesterID, scheduleCRNs],
        queryFn: async ({ signal }) => {
            const data = await fetchJSON<{courses: Omit<Course, "instructors">[]; times: Time[]}>(
                `/schedule?semester=${semesterID}&crns=${encodeURIComponent(scheduleCRNs)}`, signal,
            );
            return {
                courses: data.courses.map(course => ({ ...course, instructors: [] })),
                times: data.times,
                subjects: [], seatings: [], profs: [], exams: [],
            };
        },
        enabled: semesterID != null && scheduleOnly,
        staleTime: seatRefreshInterval,
        gcTime: 30 * 60 * 1000,
        refetchInterval: seatRefreshInterval,
    });
    const dataQuery = scheduleOnly ? scheduleQuery : frontendQuery;
    const semesters = React.useMemo(
        () => [...(semestersQuery.data ?? [])].sort((a, b) => b.id - a.id),
        [semestersQuery.data],
    );
    const error = (!semestersQuery.data && semestersQuery.error?.message)
        || (semestersQuery.data?.length === 0 && "No semesters are currently available.")
        || (!dataQuery.data && dataQuery.error?.message)
        || "";
    const loading = !error && (semesterID == null || dataQuery.isPending);
    const appliedSemester = React.useRef<number | null>(null);

    React.useEffect(() => {
        if (selectedSemester || !semestersQuery.data?.length) return;
        const params = new URLSearchParams(window.location.search);
        const semester = semestersQuery.data.find(s => s.id.toString() === params.get("semester"))
            ?? semestersQuery.data.find(s => s.latest) ?? semestersQuery.data[0];
        setSelectedSemester(semester);
        params.set("semester", semester.id.toString());
        window.history.replaceState(null, "", `?${params}`);
    }, [semestersQuery.data, selectedSemester, setSelectedSemester]);

    React.useLayoutEffect(() => {
        if (semesterID == null) return;
        const data = dataQuery.data;
        setSubjects(data?.subjects ?? []);
        setCourses(data?.courses ?? []);
        setTimes(data?.times ?? []);
        setSeating(data?.seatings ?? []);
        setProfs(data?.profs ?? []);
        setExams(data?.exams ?? []);
        if (!data) {
            setSelectedCourses([]);
            appliedSemester.current = null;
        } else if (appliedSemester.current !== semesterID) {
            const crns = new URLSearchParams(window.location.search).get("crns")?.split(",") ?? [];
            setSelectedCourses(data.courses.filter(course => crns.includes(course.crn)));
            appliedSemester.current = semesterID;
        } else {
            const coursesByCRN = new Map(data.courses.map(course => [course.crn, course]));
            setSelectedCourses(current => {
                const updated = current.map(course => coursesByCRN.get(course.crn) ?? course);
                return updated.some((course, index) => course !== current[index]) ? updated : current;
            });
        }
    }, [semesterID, dataQuery.data, setSubjects, setCourses, setTimes, setSeating, setProfs, setExams, setSelectedCourses]);

    const visibleSubjects = new Set(filteredCourses.map(course => course.subject));
    const errorAlert = error && <Alert variant="danger">{error} <Button variant="link" onClick={() => {
        void (semesterID == null ? semestersQuery.refetch() : dataQuery.refetch());
    }}>Retry</Button></Alert>;

    if (scheduleOnly) {
        return <main className="app-shell schedule-page">
            {errorAlert}
            {loading && <p className="empty-state" role="status">Loading schedule…</p>}
            {!loading && !error && <Schedule standalone />}
        </main>;
    }

    return (
        <main className="app-shell has-mobile-switcher">
            {errorAlert}
            <MobileViews
                schedule={
                    <section aria-label="Schedule">
                        <Schedule />
                    </section>
                }
                courses={
                    <section className="course-browser" aria-busy={loading} aria-label="Courses">
                        <Form.Label htmlFor="semester" className="visually-hidden">Semester</Form.Label>
                        <Form.Select id="semester" value={semesterID ?? ""} disabled={semesters.length === 0} onChange={event => {
                            const semester = semesters.find(semester => semester.id.toString() === event.target.value);
                            if (!semester) return;
                            const params = new URLSearchParams(window.location.search);
                            params.set("semester", semester.id.toString());
                            params.delete("crns");
                            window.history.replaceState(null, "", `?${params}`);
                            setSelectedSemester(semester);
                        }}>
                            {semesters.length === 0 && <option value="" />}
                            {semesters.map(semester => <option value={semester.id} key={semester.id}>{semester.name}</option>)}
                        </Form.Select>
                        {selectedSemester?.viewOnly && <p className="detail-muted mt-2">Archived semester. Availability may be out of date.</p>}
                        <SearchBar />
                        {loading && <div className="course-list-placeholder placeholder-glow" role="status">
                            <span className="visually-hidden">Loading courses…</span>
                            {Array.from({ length: 8 }, (_, index) => <div className="course-placeholder-row" key={index} aria-hidden="true">
                                <span className="placeholder" style={{ width: `${50 + index % 3 * 10}%` }} />
                            </div>)}
                        </div>}
                        {!loading && !error && visibleSubjects.size === 0 && <p className="empty-state">No matching courses.</p>}
                        {!loading && <Accordion onSelect={event => setSelectedTab([event, selectedTab[0]])}>
                            {[...subjects].sort((a, b) => a.description.localeCompare(b.description)).filter(subject => visibleSubjects.has(subject.code)).map(subject => (
                                <SubjectAccordion subject={subject} index={subject.description} key={subject.code} />
                            ))}
                        </Accordion>}
                    </section>
                }
            />
        </main>
    );
}

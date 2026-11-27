import { Course } from "../api/types";
import { linkedSectionStatus } from "../api/linkedSections";

export function LinkedSectionWarnings({ courses }: { courses: Course[] }) {
    const missing = courses.filter(course => linkedSectionStatus(course, courses) === "missing");
    if (!missing.length) return null;
    return <div className="bg-warning text-dark my-2 rounded p-3" role="status">
        Warning: Missing linked sections for {missing.map(course => `${course.id} · ${course.section}`).join(", ")}.
        <details className="mt-1">
            <summary>View options</summary>
            {missing.map(course => <div key={course.crn} className="mt-2">
                {course.id} · {course.section} — add one option:
                <ul className="mb-0">
                    {course.linkedSections!.options.map((option, index) => <li key={index}>
                        {option.sections.map(section => `${section.type} ${section.section} (CRN ${section.crn})${courses.some(course => course.crn === section.crn) ? " · In schedule" : ""}`).join(" + ")}
                    </li>)}
                </ul>
            </div>)}
        </details>
    </div>;
}

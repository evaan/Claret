import { Course } from "./types";

// Complete one whole option, not one arbitrary row or every option.
export function linkedSectionStatus(course: Course, selectedCourses: Course[]): "unknown" | "none" | "missing" | "complete" {
    const options = course.linkedSections?.options;
    if (!options) return "unknown";
    if (!options.length) return "none";
    if (options.some(option => !option.sections.length || option.sections.some(section => !section.crn.trim()))) return "unknown";
    const selected = new Set(selectedCourses.map(section => section.crn));
    return options.some(option => option.sections.every(section => selected.has(section.crn))) ? "complete" : "missing";
}

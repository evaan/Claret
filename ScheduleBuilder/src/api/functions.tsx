import { Course, CourseFilters, Seating, Subject, Time } from "./types";

export const campuses = ["St. John's", "Grenfell", "Marine Institute", "Online", "Other"];

const defaultCampuses = ["St. John's", "Online"];

export const defaultFilters = (): CourseFilters => ({
    campuses: [...defaultCampuses], types: [], delivery: [], days: [],
    startsAfter: "", endsBefore: "", minCredits: "", maxCredits: "", availableOnly: false,
});

export const activeFilterCount = (filters: CourseFilters) =>
    Number(filters.campuses.length !== defaultCampuses.length || !defaultCampuses.every(campus => filters.campuses.includes(campus))) +
    Number(filters.types.length > 0) + Number(filters.delivery.length > 0) +
    Number(filters.days.length > 0) + Number(!!(filters.startsAfter || filters.endsBefore)) +
    Number(!!(filters.minCredits || filters.maxCredits)) + Number(filters.availableOnly);

export const matchesFilters = (course: Course, filters: CourseFilters, times: Time[], seating?: Seating) => {
    const campus = campuses.includes(course.campus) ? course.campus : "Other";
    if (!filters.campuses.includes(campus)) return false;
    if (filters.types.length && !filters.types.includes(course.type)) return false;
    if (filters.delivery.length && !filters.delivery.includes(course.instructionalMethod ?? "")) return false;
    if (filters.minCredits !== "" && course.credits < Number(filters.minCredits)) return false;
    if (filters.maxCredits !== "" && course.credits > Number(filters.maxCredits)) return false;
    if (filters.availableOnly && (!seating || availableSeats(seating) <= 0 || seating.openSection === false ||
        (seating.crossListAvailable != null && seating.crossListAvailable <= 0))) return false;
    const scheduled = times.filter(time => !isUnscheduled(time));
    if (filters.days.length && !scheduled.some(time => filters.days.some(day => time.days?.includes(day)))) return false;
    if (filters.startsAfter || filters.endsBefore) {
        if (!scheduled.length) return false;
        if (scheduled.some(time => (filters.startsAfter && time.startTime < filters.startsAfter) ||
            (filters.endsBefore && time.endTime > filters.endsBefore))) return false;
    }
    return true;
};

// Self-Service's low/high bounds describe the course, not each section.
// A lecture and its zero-credit lab can both have the same 0–3 bounds.
export const formatCredits = (course: Course) => String(course.credits);

export const totalCredits = (courses: Course[]) => courses.reduce((sum, course) => sum + course.credits, 0);

export const uniqueCourseCount = (courses: Course[]) =>
    new Set(courses.filter(course => !/^(lab|laboratory)$/i.test(course.type.trim())).map(course => course.id)).size;

export const formatInstructorName = (name: string) => {
    const parts = name.trim().split(",").map(part => part.trim());
    return parts.length === 2 && parts.every(Boolean) ? `${parts[1]} ${parts[0]}` : name.trim();
};

export const availableSeats = (seating: Seating) => seating.seatsAvailable ?? seating.maxSeats - seating.seats;

export const isUnscheduled = (time: Time) =>
    !time.days || !/^\d{2}:\d{2}$/.test(time.startTime) || !/^\d{2}:\d{2}$/.test(time.endTime) ||
    (time.startTime === "00:00" && time.endTime === "00:01");

export const isOnline = (course: Course) =>
    course.campus.trim().toLowerCase() === "online" ||
    /\bonline\b/i.test(course.instructionalMethod ?? "") ||
    course.type === "World Wide Web";

export const formatTime = (value: string) => {
    const [hour, minutes] = value.split(":").map(Number);
    return `${hour % 12 || 12}:${String(minutes).padStart(2, "0")}${hour < 12 ? "am" : "pm"}`;
};

export const formatMeeting = (time: Time) => {
    if (isUnscheduled(time)) return "TBA";
    return `${time.days} ${formatTime(time.startTime)} - ${formatTime(time.endTime)}`;
};

export const matchesSearch = (course: Course, query: string, times: Time[] = [], subject?: Subject) => {
    const text = [
        course.id, course.name, course.crn, course.section, course.subject, subject?.description,
        course.campus, course.type, course.instructionalMethod, ...(course.instructors ?? []),
        ...times.flatMap(time => [time.location, time.buildingName]),
    ].filter(Boolean).join(" ").normalize("NFD").replace(/[\u0300-\u036f]/g, "").toLowerCase();
    return query.normalize("NFD").replace(/[\u0300-\u036f]/g, "").toLowerCase().trim().split(/\s+/)
        .every(word => text.includes(word));
};

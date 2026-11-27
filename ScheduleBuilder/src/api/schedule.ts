import moment from "moment";
import { isOnline, isUnscheduled } from "./functions";
import { Course, Time } from "./types";

export interface ScheduleEvent {
    title: string;
    start: string;
    end: string;
    classNames?: string[];
    extendedProps?: { unscheduled: boolean };
}

export function buildSchedule(courses: Course[], times: Time[], weekStart = moment().startOf("week")) {
    const events: ScheduleEvent[] = [];
    const startTimes: number[] = [];
    const endTimes: number[] = [];
    let weekend = false;

    courses.forEach(course => {
        times.filter(time => time.crn === course.crn && !isUnscheduled(time)).forEach(time => {
            startTimes.push(moment(time.startTime, "HH:mm").hour());
            const end = moment(time.endTime, "HH:mm");
            endTimes.push(end.hour() + (end.minute() > 0 ? 1 : 0));
            ["M", "T", "W", "R", "F", "S", "U"].forEach((day, index) => {
                if (!time.days?.includes(day)) return;
                if (index >= 5) weekend = true;
                const date = weekStart.clone().add(index + 1, "days").format("YYYY-MM-DD");
                events.push({
                    title: `${course.id}${time.location ? ` - ${time.location}` : ""}`,
                    start: `${date}T${time.startTime}`,
                    end: `${date}T${time.endTime}`,
                });
            });
        });
    });

    const min = Math.min(9, ...startTimes);
    const max = Math.max(17, ...endTimes);
    const unscheduledMeetings = courses.flatMap<{ course: Course; time?: Time }>(course => {
        const meetings = times.filter(time => time.crn === course.crn);
        const unscheduled = meetings.filter(isUnscheduled);
        if (isOnline(course)) {
            // Ignore placeholder records when an online course has an actual meeting.
            if (meetings.some(time => !isUnscheduled(time))) return [];
            return [{ course, time: unscheduled[0] }];
        }
        return meetings.length > 0
            ? unscheduled.map(time => ({ course, time }))
            : [{ course, time: undefined }];
    });
    const slotMinutes = Math.min(60, (max - min) * 60 / Math.max(1, unscheduledMeetings.length));
    const unscheduledStart = weekStart.clone().add(weekend ? 8 : 6, "days").hour(min);
    unscheduledMeetings.forEach(({ course, time }, index) => {
        events.push({
            title: `${course.id}${time?.location ? ` - ${time.location}` : ""}`,
            extendedProps: { unscheduled: true },
            start: unscheduledStart.clone().add(index * slotMinutes, "minutes").format("YYYY-MM-DDTHH:mm"),
            end: unscheduledStart.clone().add((index + 1) * slotMinutes, "minutes").format("YYYY-MM-DDTHH:mm"),
        });
    });

    const conflictingEvents = new Set<ScheduleEvent>();
    for (let i = 0; i < events.length; i++) {
        const event = events[i];
        if (event.extendedProps?.unscheduled) continue;
        for (let j = i + 1; j < events.length; j++) {
            const other = events[j];
            if (other.extendedProps?.unscheduled) continue;
            if (event.start < other.end && other.start < event.end) {
                conflictingEvents.add(event);
                conflictingEvents.add(other);
            }
        }
    }
    conflictingEvents.forEach(event => { event.classNames = ["has-conflict"]; });
    const scheduledEvents = events.filter(event => !event.extendedProps?.unscheduled);
    const maxConcurrent = Math.max(1, ...scheduledEvents.map(event =>
        scheduledEvents.filter(other => other.start <= event.start && event.start < other.end).length
    ));
    const days = ["Mon", "Tue", "Wed", "Thu", "Fri", ...(weekend ? ["Sat", "Sun"] : []), "Unscheduled"];
    return {
        events, min, max, weekend, days, conflictingEvents,
        calendarMinWidth: Math.max(480, 60 + days.length * Math.max(70, maxConcurrent * 60)),
        unscheduledDate: unscheduledStart.format("YYYY-MM-DD"),
        range: {
            start: weekStart.clone().add(1, "days").format("YYYY-MM-DD"),
            end: weekStart.clone().add(weekend ? 9 : 7, "days").format("YYYY-MM-DD"),
        },
    };
}

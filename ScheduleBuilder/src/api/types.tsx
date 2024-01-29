export interface Subject {
    name: string;
    friendlyName: string;
}

export interface Course {
    crn: string;
    id: string;
    name: string;
    section: string;
    dateRange: string;
    type: string;
    instructor: string;
    subjectFull: string;
    subject: string;
    campus: string;
    comment: string;
    credits: number;
    semester: number;
    level: string;
}

export interface Time {
    crn: string;
    days: string;
    startTime: string;
    endTime: string;
    location: string;
    courseType: string;
    id: number;
}

export interface Seating {
    crn: string;
    available: string;
    max: string;
    waitlist: string;
    checked: string;
}
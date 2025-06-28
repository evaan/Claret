export interface Subject {
    id: string;
    name: string;
}

export interface Semester {
    id: number
    name: string
    latest: boolean
    viewOnly: boolean
    medical: boolean
    mi: boolean
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
    identifier: string;
}

export interface Time {
    crn: string;
    days: string;
    startTime: string;
    endTime: string;
    location: string;
    type: string;
    id: number;
    identifier: string;
}

export interface Seating {
  semester: string;
  crn: string;
  seats: {
    capacity: number;
    actual: number;
    remaining: number;
  };
  waitlist: {
    capacity: number;
    actual: number;
    remaining: number;
  };
}

export interface Professor {
    name: string;
    rating: number;
    id: number;
    difficulty: number;
    ratings: number;
    wouldRetake: number;
}

export interface ExamTime {
   crn: string;
   time: string;
   location: string; 
}
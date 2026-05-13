export interface Subject {
    code: string;
    description: string;
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
    id: string;
    name: string;
    crn: string;
    section: string;
    credits: number;
    campus: string;
    subject: string;
    instructors: string[];
    type: string;
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
  seats: number;
  maxSeats: number;
  waitlist: number;
  maxWaitlist: number;
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
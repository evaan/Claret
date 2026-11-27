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
    instructionalMethod?: string;
    partOfTerm?: string;
    isSectionLinked?: boolean;
    linkIdentifier?: string;
    crossList?: string;
    creditsMin?: number | null;
    creditsMax?: number | null;
    prerequisites?: CourseRequirements | null;
    corequisites?: CourseRequirements | null;
    linkedSections?: { options: LinkedSectionOption[] } | null;
}

export interface LinkedSectionOption {
    sections: { crn: string; section: string; type: string; title: string }[];
}

export interface CourseRequirements {
    text: string;
    tables: { headers: string[]; rows: string[][] }[];
}

export interface Time {
    crn: string;
    days: string | null;
    startTime: string;
    endTime: string;
    location: string;
    type: string;
    id: number;
    identifier: string;
    dateRange?: string;
    startDate?: string;
    endDate?: string;
    buildingName?: string;
}

export interface Seating {
  semester: number;
  crn: string;
  seats: number;
  maxSeats: number;
  waitlist: number;
  maxWaitlist: number;
  seatsAvailable?: number | null;
  waitCount?: number | null;
  openSection?: boolean | null;
  crossListCapacity?: number | null;
  crossListCount?: number | null;
  crossListAvailable?: number | null;
  updatedAt?: string;
}

export interface CourseFilters {
    campuses: string[];
    types: string[];
    delivery: string[];
    days: string[];
    startsAfter: string;
    endsBefore: string;
    minCredits: string;
    maxCredits: string;
    availableOnly: boolean;
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

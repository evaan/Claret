import { Course } from "./types";

export const shouldShow = (section: Course, filter: [boolean, boolean, boolean, boolean, boolean]) => (section.campus == "St. John's" && filter[0]) || (section.campus == "Grenfell" && filter[1]) || (section.campus == "Marine Institute" && filter[2]) || (section.campus == "Online" && filter[3])
|| (section.campus != "St. John's" && section.campus != "Grenfell" && section.campus != "Marine Institute" && section.campus != "Online" && filter[4]);

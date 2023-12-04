import requests
from bs4 import BeautifulSoup
import sys
import re
from tqdm import tqdm
from dateutil import parser
from sqlalchemy import create_engine
from sqlalchemy.orm import Session
from dbClasses import Base, Course, CourseTime, Subject
import os
from dotenv import load_dotenv

BUILDING_CODES = {
    # Mostly sourced from https://www.mun.ca/regoff/registration-and-final-exams/course-offerings/building-abbreviations/

    # St. John's Campus
    # NOTE: Missing UC
    # NOTE: According to the link above, the new Faculty of Medicine building (which is
    # attached to HSC) has the building code "M". But it does not appear in banner or the
    # campus map. 
    "Arts and Administration Bldg": "A",
    "Henrietta Harvey Bldg": "HH",
    "Business Administration Bldg": "BN",
    "INCO Innovation Centre": "IIC",
    "Biotechnology Bldg": "BT",
    "St. John's College": "J",
    "Chemistry - Physics Bldg": "C",
    "Core Science Facility": "CSF",
    "M. O. Morgan Bldg": "MU",
    "Computing Services": "CS",
    "Physical Education Bldg": "PE",
    "G. A. Hickman Bldg": "ED",
    "Queen's College": "QC",
    "Queen Elizabeth II Library": "L",
    "S. J. Carew Bldg.": "EN",
    "Science Bldg": "S",
    "Alexander Murray Bldg": "ER",
    "Health Sciences Centre": "H",
    "Coughlan College": "CL",

    # Marine Institute
    "Marine Institute": "MI",

    # Center for Nursing Studies
    "Center for Nursing Studies": "N",

    # Grenfell Campus
    # NOTE: Western Memorial Hosp Library is a seperate building in banner,
    # but I have no evidence it has room numbers or a building code. So it is excluded
    "Arts and Science (SWGC)": "AS",
    "Fine Arts (SWGC)": "FA",
    "Forest Centre": "FC",
    "Library/Computing (SWGC)": "LC",
    "Western Memorial Hospital": "WMH",

    # Whitespace character
    "\u00A0": "N/A",
}

#TODO: seats remaining for class (maybe?)
#TODO: rate my prof support
#TODO: campus

parseTime = lambda time: parser.parse(time).strftime("%H:%M") if time != "TBA" else "TBA" 
regex = re.compile(r'(?<!\w)(' + '|'.join(re.escape(key) for key in BUILDING_CODES.keys()) + r')(?!\w)')

def processCourse(option):
    title = option.text.split(" - ")
    details = option.parent.findNext("td").find_all("td", attrs={"class", "dddefault"})
    for i in range(len(details)):
        details[i] = regex.sub(lambda x: BUILDING_CODES[x.group()], details[i].text)
        
    for line in option.parent.findNext("td").text.split("\n"):
        if "Campus" in line:
            campus = line[0:-7]


    if len(details) >= 7:
        session.merge(Course(
            name = " - ".join(title[0:-3]), #some courses have a hyphen in the name, this includes the remainder of the class
            id = title[-2], 
            crn = title[-3],
            section = title[-1],
            dateRange = details[4], #i dont really know if date range is even neccesary, may be something to remove eventually
            type = details[5],
            instructor = details[6][3:] if details[6].startswith("(P)") else details[6], #TODO: list of profs
            subject = title[-2].split()[0],
            campus = campus
        ))
    else:
        session.merge(Course(
            name = " - ".join(title[0:-3]), #some courses have a hyphen in the name, this includes the remainder of the class
            id = title[-2], 
            crn = title[-3],
            section = title[-1],
            subject = title[-2].split()[0],
            campus = campus
        ))

    #maybe do something with 12:00am to 12:01am?
    #remove times if there are more than how many there should be
    if session.query(CourseTime).filter(CourseTime.crn.like(title[-3] + "%")).count() != (len(details)//7):
        session.query(CourseTime).filter(CourseTime.crn.like(title[-3] + "%")).delete()
    for i in range(len(details)//7):
        time = details[1+(i*7)].split(" - ")
        session.merge(CourseTime(
            crn = title[-3],
            startTime = parseTime(time[0]),
            endTime = parseTime(time[1]) if len(time) > 1 else "TBA",
            days = details[2+(i*7)],
            location = details[3+(i*7)],
        ))


def recursivelyProcessSubjects(name, subject, semester):
    for i in range(1, 10):
        processSubject(name, subject, semester, i)

def processSubject(name, subject, semester, course=""):
    if verbose:
        if course == "":
            print(f"Processing Subject: {subject} ({name})")
        else:
            print(f"Processing Subject: {subject} ({name}) Iteration: {course}")
    postParams = {
        "term_in": semester,
        "sel_subj": ["dummy",subject],
        "sel_day": "dummy",
        "sel_schd": ["dummy","%"],
        "sel_insm": ["dummy","%"],
        "sel_camp": ["dummy","%"],
        "sel_levl": ["dummy","%"],
        "sel_sess": ["dummy","%"],
        "sel_instr": ["dummy","%"],
        "sel_ptrm": ["dummy","%"],
        "sel_attr": ["dummy","%"],
        "sel_crse": course,
        "sel_title": "",
        "sel_from_cred": "",
        "sel_to_cred": "",
        "begin_hh": "0",
        "begin_mi": "0",
        "begin_ap": "a",
        "end_hh": "0",
        "end_mi": "0",
        "end_ap": "a"
    }
    page = requests.post(
        "https://selfservice.mun.ca/direct/bwckschd.p_get_crse_unsec",
        data = postParams
    )
    soup = BeautifulSoup(page.content, "html.parser")
    courses = soup.find_all("th", attrs={"class": "ddtitle"})
    if len(courses) == 101 and course == "": #this does assume that each level has less than 101 courses as well
        recursivelyProcessSubjects(name, subject, semester)
        return
    for option in courses:
        course = processCourse(option)

def processSemester(semester, name):
    #TODO
    #courseDict["semester"] = name
    #courseDict["semesterId"] = semester
    page = requests.get(
        "https://selfservice.mun.ca/direct/bwckgens.p_proc_term_date",
        params={
            "p_calling_proc": "bwckschd.p_disp_dyn_sched",
            "p_term": semester
        },
        timeout=10
    )
    soup = BeautifulSoup(page.content, "html.parser")
    subjects = soup.find("select", attrs={"name": "sel_subj"})
    subjectList = []
    for option in subjects.find_all("option"):
        subjectList.append((option.text, option.get("value")))
    return subjectList

def getLatestSemester():
    page = requests.get("https://selfservice.mun.ca/direct/bwckschd.p_disp_dyn_sched")
    soup = BeautifulSoup(page.content, "html.parser")

    semesters = soup.find("select", attrs={"name": "p_term"})
    for option in semesters.find_all("option"):
        if not option.text.endswith("Medicine") and option.text != "None":
            if verbose:
                print(f"Latest Semester: {option.text}\nSemester ID: {option.get('value')}")
            return option.get("value"), option.text

if __name__ == "__main__":
    #intro and args
    print("\033[95m ####   #      ###  #####  ###### #####")
    print("#    #  #     #   # #    # #        #")
    print("#       #     ##### #####  ####     #")
    print("#       #     #   # #    # #        #")
    print("#    #  #     #   # #    # #        #")
    print(" ####   ##### #   # #    # ######   # Scraper v0.1")
    print("https://github.com/evaan/Claret\033[0m")
    print()
    verbose = "--verbose" in sys.argv
    
    #sql
    load_dotenv()
    engine = create_engine(os.getenv("DB_URL"), echo=verbose)
    Base.metadata.create_all(engine)
    session = Session(engine)

    #scraping
    semester, semesterId = getLatestSemester()
    subjects = processSemester(semester, semesterId)
    for subject in tqdm(subjects):
        if subject[1] != "%":
            session.merge(Subject(name=subject[1], friendlyName=subject[0]))
            processSubject(subject[0], subject[1], semester)
    print("\033[92mScrape complete!\033[0m")
    session.commit()
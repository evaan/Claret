import requests
from bs4 import BeautifulSoup
import re
from dateutil import parser
from sqlalchemy import create_engine
from sqlalchemy.orm import Session
from dbClasses import Base, Course, CourseTime, Semester, Subject, Seating
from os import getenv
from dotenv import load_dotenv

BUILDING_CODES = {
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
    "Marine Institute": "MI",
    "Center for Nursing Studies": "N",
    "Arts and Science (SWGC)": "AS",
    "Fine Arts (SWGC)": "FA",
    "Forest Centre": "FC",
    "Library/Computing (SWGC)": "LC",
    "Western Memorial Hospital": "WMH",
    "\u00A0": "N/A",
}

def getLatestSemester(medical = False):
    page = requests.get("https://selfservice.mun.ca/direct/bwckschd.p_disp_dyn_sched")
    soup = BeautifulSoup(page.content, "html.parser")

    semesters = soup.find("select", attrs={"name": "p_term"})
    for option in semesters.find_all("option"):
        if (not option.text.endswith("Medicine") or medical) and option.text != "None":
            print(f"Latest {'Medical ' if medical else ''}Semester: {option.text}\nSemester ID: {option.get('value')}")
            return option.get("value")
        
def processSemester(semester):
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
        subjectList.append((option.text, option.get("value"), semester))
    return subjectList

parseTime = lambda time: parser.parse(time).strftime("%H:%M") if time != "TBA" else "TBA" 
regex = re.compile(r'(?<!\w)(' + '|'.join(re.escape(key) for key in BUILDING_CODES.keys()) + r')(?!\w)')

timeId = 0

def processCourse(option, semester, medical = False):
    global timeId
    title = option.text.split(" - ")
    details = option.parent.findNext("td").find_all("td", attrs={"class", "dddefault"})
    for i in range(len(details)):
        details[i] = regex.sub(lambda x: BUILDING_CODES[x.group()], details[i].text)
        
    for line in option.parent.findNext("td").text.split("\n"):
        if re.match(".*Campus$", line):
            campus = line[0:-7]

    description = list(filter(lambda x: (not x == ""), option.parent.findNext("td").parent.text.split("\n\n")))

    if len(details) >= 7:
        session.merge(Course(
            name = " - ".join(title[0:-3]), #some courses have a hyphen in the name, this includes the remainder of the class
            id = title[-2], 
            crn = title[-3],
            section = title[-1],
            dateRange = details[4], #i dont really know if date range is even neccesary, may be something to remove eventually
            type = details[5],
            instructor = details[6][3:] if details[6].startswith("(P)") else details[6], #TODO: list of profs
            subject = title[-2].split()[0] + ("1" if medical else ""),
            campus = campus,
            comment = None if description[0].startswith("Associated Term") else description[0],
            credits = int(float(list(filter(lambda x: ("Credits" in x), description))[0].lstrip().split(" ")[0])),
            semester = semester
        ))
    else:
        session.merge(Course(
            name = " - ".join(title[0:-3]), #some courses have a hyphen in the name, this includes the remainder of the class
            id = title[-2], 
            crn = title[-3],
            section = title[-1],
            subject = title[-2].split()[0] + ("1" if medical else ""),
            campus = campus,
            comment = None if description[0].startswith("Associated Term") else description[0],
            credits = int(float(list(filter(lambda x: ("Credits" in x), description))[0].lstrip().split(" ")[0])),
            semester = semester
        ))
        
    if session.query(Seating.crn).filter_by(crn = title[-3]).first() is None:
        session.add(Seating(crn = title[-3], available = 0, max = 0, waitlist = 0, checked = "Never"))

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
            ignore = timeId
        ))
        timeId+=1

def processSubject(name, subject, semester, course="", medical=False):
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
        for i in range(1, 10):
            processSubject(name, subject, semester, i)
        return
    for option in courses:
        course = processCourse(option, semester, medical)

if __name__ == "__main__":
    #sql
    load_dotenv()
    engine = create_engine(getenv("DB_URL"))
    Base.metadata.create_all(engine)
    session = Session(engine)

    #scraping
    latestSemester = getLatestSemester()
    latestMedSemester = getLatestSemester(True)
    for semester in session.query(Semester).filter(Semester.semester != latestSemester, Semester.semester != latestMedSemester).all():
        session.delete(semester)
    session.merge(Semester(semester = latestSemester))
    for subject in processSemester(latestSemester):
        if subject[1] != "%":
            session.merge(Subject(name=subject[1], friendlyName=subject[0]))
            processSubject(subject[0], subject[1], subject[2])
    session.merge(Semester(semester = latestMedSemester))
    for subject in processSemester(latestMedSemester):
        if subject[1] != "%":
            session.merge(Subject(name=subject[1] + "1", friendlyName=subject[0] + " (Medical)"))
            processSubject(subject[0], subject[1], subject[2], medical=True)
    session.commit()
    print("Scrape Complete!")
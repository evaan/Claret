import requests
from bs4 import BeautifulSoup
import sys
import json

courseDict = {}

def processSubject(name, subject, semester):
    courseDict[name] = []
    if not quiet:
        print(f"Processing Subject: {subject} ({name})")
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
        "sel_crse": "",
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
    for option in courses:
        courseInfo = []
        title = option.text.split(" - ")
        courseInfo.append(title[0])
        courseInfo.append(title[1])
        courseInfo.append(title[2]) # with course identifier
        #courseInfo.append(title[2][5::]) # without course identifier
        courseInfo.append(title[3])
        details = option.findNext("abbr").parent.parent.find_all("td")[1::] #ignore type, seems to be always "Class"
        for detail in details:
            courseInfo.append(detail.text[3:] if detail.text.startswith("(P)") else detail.text)
        courseInfo = list(map(lambda x: x.replace("\u00A0", "N/A"), courseInfo))
        if not quiet:
            print(courseInfo)
        courseDict[name].append(courseInfo)

def processSemester(semester):
    page = requests.get(
        "https://selfservice.mun.ca/direct/bwckgens.p_proc_term_date",
        params={
            "p_calling_proc": "bwckschd.p_disp_dyn_sched",
            "p_term": semester
        }
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
            if not quiet:
                print(f"Latest Semester: {option.text}\nSemester ID: {option.get('value')}")
            return option.get("value")

if __name__ == "__main__":
    quiet = "--quiet" in sys.argv
    semester = getLatestSemester()
    subjects = processSemester(semester)
    for subject in subjects:
        if subject[1] != "%":
            processSubject(subject[0], subject[1], semester)
    if "--save" in sys.argv:
        output = open("courses.json", "w")
        output.write(json.dumps(courseDict))
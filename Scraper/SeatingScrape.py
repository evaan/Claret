import os
from sys import argv, exit
from bs4 import BeautifulSoup
from dotenv import load_dotenv
import requests
from sqlalchemy import create_engine
from sqlalchemy.orm import Session
from dbClasses import Base, Seating
from datetime import datetime

if __name__ == "__main__":
    if len(argv) <= 2:
        exit() #no crn or semester to scrape from
    load_dotenv()
    engine = create_engine(os.getenv("DB_URL"))
    Base.metadata.create_all(engine)
    session = Session(engine)
    page = requests.get(f"https://selfservice.mun.ca/direct/bwckschd.p_disp_detail_sched?term_in={argv[2]}&crn_in=" + argv[1])
    soup = BeautifulSoup(page.content, "html.parser")
    for caption in soup.find_all("caption"):
        if caption.text == "Registration Availability":
            cells = caption.parent.find_all("td", attrs={"class", "dddefault"})
            session.merge(Seating(
                crn = argv[1],
                available = cells[2].text,
                max = cells[0].text,
                waitlist = cells[4].text if len(cells) >= 6 else None,
                checked = datetime.now().strftime("%Y-%m-%dT%H:%M")
            ))
    session.commit()
FROM node:latest
COPY API /app/API
COPY Scraper /app/Scraper
WORKDIR /app/API
RUN apt-get update
RUN apt-get install python3-full tzdata -y
RUN python3 -m venv /app/venv
ENV PATH="/app/venv/bin:$PATH"
RUN pip install -r /app/Scraper/requirements.txt
CMD node server.js
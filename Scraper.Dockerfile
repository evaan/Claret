FROM alpine:latest
COPY Scraper /app/Scraper
RUN apk add --no-cache supercronic python3 py3-pip py3-virtualenv tzdata
RUN python3 -m venv /app/venv
ENV PATH="/app/venv/bin:$PATH"
RUN pip install -r /app/Scraper/requirements.txt
RUN echo "30 3 * * 1 python3 /app/Scraper/Scraper.py" > ./crontab
CMD python3 /app/Scraper/Scraper.py; supercronic -debug ./crontab
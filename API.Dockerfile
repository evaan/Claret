FROM golang:latest AS build
WORKDIR /app
COPY ./API /app
RUN CGO_ENABLED=0 GOOS=linux go build -o /api
FROM python:alpine
RUN mkdir -p /app/API
COPY --from=build /api /app/API/api
COPY ./Scraper /app/Scraper
WORKDIR /app/Scraper
RUN pip install -r requirements.txt
WORKDIR /app/API
CMD ["./api"]
FROM golang:alpine AS build-stage
WORKDIR /app
COPY ./Scraper /app
RUN CGO_ENABLED=0 GOOS=linux go build -o /scraper
FROM gcr.io/distroless/base-debian11:latest
COPY --from=build-stage /scraper /scraper
USER nonroot:nonroot
ENTRYPOINT [ "/scraper" ]
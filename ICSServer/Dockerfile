# Build Stage Container
FROM golang:1.21.5 AS build-stage
WORKDIR /app
COPY . /app
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /ics_server

# Production Container
FROM gcr.io/distroless/base-debian11:latest
COPY --from=build-stage /ics_server /ics_server
USER nonroot:nonroot
ENTRYPOINT [ "/ics_server" ]

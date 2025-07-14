FROM golang:alpine AS build-stage
WORKDIR /app
COPY . /app
RUN CGO_ENABLED=0 GOOS=linux go build -o /server
FROM gcr.io/distroless/base-debian11:latest
COPY --from=build-stage /server /server
USER nonroot:nonroot
ENTRYPOINT [ "/server" ]
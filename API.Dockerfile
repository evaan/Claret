FROM golang:latest AS build-stage
WORKDIR /app
COPY ./API /app
RUN CGO_ENABLED=0 GOOS=linux go build -o /api
FROM gcr.io/distroless/base-debian11:latest
COPY --from=build-stage /api /api
USER nonroot:nonroot
ENTRYPOINT [ "/api" ]
FROM golang:latest AS build
COPY API /app
WORKDIR /app/API
RUN CGO_ENABLED=0 GOOS=linux go build -o /api
FROM gcr.io/distroless/base-debian11:latest
COPY --from=build-stage /api /api
USER nonroot:nonroot
CMD ["/api"]
FROM golang:latest AS build
WORKDIR /app
COPY ./API /app
RUN CGO_ENABLED=0 GOOS=linux go build -o /api
FROM gcr.io/distroless/base-debian11:latest
COPY --from=build /api /api
USER nonroot:nonroot
CMD ["/api"]
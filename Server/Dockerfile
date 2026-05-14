FROM golang:alpine AS build-stage
WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY . /app
RUN CGO_ENABLED=0 GOOS=linux go build -o /server
FROM gcr.io/distroless/base-debian11:latest
COPY --from=build-stage /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build-stage /server /server
USER nonroot:nonroot
ENTRYPOINT [ "/server" ]
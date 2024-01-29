FROM golang:alpine AS build-stage
WORKDIR /app
COPY ./DiscordBot /app
RUN CGO_ENABLED=0 GOOS=linux go build -o /bot
FROM gcr.io/distroless/base-debian11:latest
COPY --from=build-stage /bot /bot
USER nonroot:nonroot
ENTRYPOINT [ "/bot" ]
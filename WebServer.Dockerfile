FROM node:alpine as build-stage
COPY ScheduleBuilder /app
WORKDIR /app/ScheduleBuilder
RUN npm install
RUN npm run build
FROM nginx:alpine
COPY --from=build-stage /app/build /usr/share/nginx/html
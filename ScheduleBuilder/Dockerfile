FROM node:alpine as build-stage
COPY . /app
WORKDIR /app/ScheduleBuilder
RUN npm install
RUN npm run build
FROM nginx:alpine
COPY nginx.conf /etc/nginx/nginx.conf
COPY --from=build-stage /app/dist /usr/share/nginx/html
FROM node:latest as build
COPY ScheduleBuilder /app
WORKDIR /app/ScheduleBuilder
RUN npm install
RUN npm run build
FROM nginx:latest
COPY --from=build /app/build /usr/share/nginx/html
CMD ["nginx", "-g", "daemon off;"]
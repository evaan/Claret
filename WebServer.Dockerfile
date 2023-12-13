FROM nginx:latest
COPY ScheduleBuilder /app/ScheduleBuilder
WORKDIR /app/ScheduleBuilder
RUN apt-get install -y nodejs npm
RUN npm install
RUN npm run build
RUN mv /app/ScheduleBuilder/build /usr/share/nginx/html
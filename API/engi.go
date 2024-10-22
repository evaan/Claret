package main

import (
	"encoding/json"
	"net/http"
)

func engi(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("semester") == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Semester was not provided, please add ?semester={semester} in your URL."))
		return
	}

	var output []EngSeats

	seats, err := db.Query("SELECT * FROM eng_seats WHERE semester = $1", r.URL.Query().Get("semester"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	defer seats.Close()

	for seats.Next() {
		var engSeat EngSeats

		err := seats.Scan(&engSeat.Id, &engSeat.Subject, &engSeat.Name, &engSeat.Course, &engSeat.Section, &engSeat.Registered, &engSeat.Date, &engSeat.Semester)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		output = append(output, engSeat)
	}

	course, err := json.Marshal(output)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(course))
}

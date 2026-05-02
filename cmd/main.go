package main

import (
	"cinemabooking/internal/adapters"
	"cinemabooking/internal/booking"
	"cinemabooking/internal/utils"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /movies", listMovies)
	mux.Handle("GET /", http.FileServer(http.Dir("static")))
	store := booking.NewRedisStore(adapters.NewClient("localhost:6379"))
	svc := booking.NewService(store)
	bookingHandler := booking.NewHandler(svc)
	mux.HandleFunc("GET /movies/{movieID}/seats", bookingHandler.ListSeats)
	mux.HandleFunc("POST /movies/{movieID}/seats/{seatID}/hold", bookingHandler.HoldSeat)

	mux.HandleFunc("PUT /sessions/{sessionID}/confirm", bookingHandler.ConfirmSession)
	mux.HandleFunc("DELETE /sessions/{sessionID}", bookingHandler.ReleaseSession)
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}

}

var movies = []movieResponse{
	{ID: "limitless", Title: "Области тьмы (Limitless)", Rows: 6, SeatsPerRow: 8},
	{ID: "escape-plan", Title: "План побега", Rows: 3, SeatsPerRow: 7},
}

func listMovies(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSON(w, http.StatusOK, movies)
}

type movieResponse struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Rows        int    `json:"rows"`
	SeatsPerRow int    `json:"seats_per_row"`
}

package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {

	jobsMux := http.NewServeMux()

	jobsMux.HandleFunc("POST /", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Created Jobs")
	})
	jobsMux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello Jobs")
	})

	mainMux := http.NewServeMux()
	mainMux.Handle("/api/v1/jobs/", http.StripPrefix("/api/v1/jobs", jobsMux))
	fmt.Println("Listening on port 8000")
	log.Fatal(http.ListenAndServe(":8000", mainMux))
}

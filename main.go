package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {

	jobsMux := http.NewServeMux()

	jobsMux.HandleFunc("POST /", createJob)
	jobsMux.HandleFunc("GET /", getJobs)
	jobsMux.HandleFunc("GET /:id", getJobById)

	mainMux := http.NewServeMux()
	mainMux.Handle("/jobs/", http.StripPrefix("/jobs", jobsMux))

	connectToDb()
	fmt.Println("Connected to Database")

	seedTables()
	fmt.Println("Listening on port 8000")
	log.Fatal(http.ListenAndServe(":8000", mainMux))
}

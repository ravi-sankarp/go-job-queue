package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ravi-sankarp/go-job-queue/db"
	"github.com/ravi-sankarp/go-job-queue/scheduler"
)

func main() {

	jobsMux := http.NewServeMux()

	jobsMux.HandleFunc("POST /", scheduler.CreateJob)
	jobsMux.HandleFunc("GET /", scheduler.GetJobs)

	mainMux := http.NewServeMux()
	mainMux.Handle("/jobs/", http.StripPrefix("/jobs", jobsMux))

	db.ConnectToDb()
	fmt.Println("Connected to Database")

	db.SeedTables()

	fmt.Println("Listening on port 8000")
	log.Fatal(http.ListenAndServe(":8000", mainMux))
}

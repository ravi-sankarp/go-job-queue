package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ravi-sankarp/go-job-queue/db"
	"github.com/ravi-sankarp/go-job-queue/scheduler"
	"github.com/ravi-sankarp/go-job-queue/workers"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer done()
	jobsMux := http.NewServeMux()

	jobsMux.HandleFunc("POST /", scheduler.CreateJob)
	jobsMux.HandleFunc("GET /", scheduler.GetJobs)

	mainMux := http.NewServeMux()
	mainMux.Handle("/jobs/", http.StripPrefix("/jobs", jobsMux))

	db.ConnectToDb()
	fmt.Println("Connected to Database")

	db.SeedTables(ctx)

	fmt.Println("Starting workers")
	go workers.Start(ctx)
	fmt.Println("Listening on port 8000")
	log.Fatal(http.ListenAndServe(":8000", mainMux))
}

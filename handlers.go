package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type JobCreate struct {
	Title        string    `json:"title"`
	Endpoint     string    `json:"endpoint"`
	Method       string    `json:"method"`
	Payload      string    `json:"payload"`
	Scheduled_at time.Time `json:"scheduled_at"`
}

type Job struct {
	Id           int       `json:"id"`
	Title        string    `json:"title"`
	Endpoint     string    `json:"endpoint"`
	Method       string    `json:"method"`
	Payload      string    `json:"payload"`
	Scheduled_at string    `json:"scheduled_at"`
	Created_on   time.Time `json:"created_on"`
	Status       string    `json:"status"`
	Retries      int       `json:"retries"`
	Updated_on   time.Time `json:"updated_on"`
}

type Response struct {
	Error   string `json:"error"`
	Data    any    `json:"data"`
	Success bool   `json:"success"`
}

func createJob(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var jobCreate JobCreate
	if err := json.NewDecoder(r.Body).Decode(&jobCreate); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Error: err.Error(), Success: false})
		return
	}
	jobCreate.Scheduled_at = jobCreate.Scheduled_at.UTC()
	if _, err := getDb().Exec("INSERT INTO jobs (title, endpoint, method, payload, scheduled_at) values (?, ?, ?, ?, ?)",
		jobCreate.Title, jobCreate.Endpoint, jobCreate.Method, jobCreate.Payload, jobCreate.Scheduled_at); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Error: err.Error(), Success: false})
		return
	}
	json.NewEncoder(w).Encode(Response{Success: true})
}

func getJobs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	rows, err := getDb().Query(`SELECT id, title, endpoint, method, payload, scheduled_at,
		created_on, status, updated_on, retries FROM jobs`)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Error: err.Error(), Success: false})
		return
	}
	defer rows.Close()
	jobs := make([]Job, 0, 10)
	for rows.Next() {
		var job Job
		if err := rows.Scan(&job.Id, &job.Title, &job.Endpoint, &job.Method, &job.Payload,
			&job.Scheduled_at, &job.Created_on, &job.Status, &job.Updated_on, &job.Retries); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(Response{Error: err.Error(), Success: false})
			return
		}
		jobs = append(jobs, job)

	}
	json.NewEncoder(w).Encode(Response{Success: true, Data: jobs})
}

func getJobById(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "ID")
}

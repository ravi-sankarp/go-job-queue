package scheduler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/ravi-sankarp/go-job-queue/db"
)

type JobCreate struct {
	Title        string    `json:"title"`
	Endpoint     string    `json:"endpoint"`
	Method       string    `json:"method"`
	Payload      string    `json:"payload"`
	Scheduled_at time.Time `json:"scheduled_at"`
}

type DbJob struct {
	Id           int
	Title        string
	Endpoint     string
	Method       string
	Payload      string
	Scheduled_at string
	Created_on   string
	Status       string
	Retries      sql.NullInt32
	Updated_on   sql.NullString
}

type Job struct {
	Id           int    `json:"id"`
	Title        string `json:"title"`
	Endpoint     string `json:"endpoint"`
	Method       string `json:"method"`
	Payload      string `json:"payload"`
	Scheduled_at string `json:"scheduled_at"`
	Created_on   string `json:"created_on"`
	Status       string `json:"status"`
	Retries      int    `json:"retries"`
	Updated_on   string `json:"updated_on"`
}
type Response struct {
	Error   string `json:"error"`
	Data    any    `json:"data"`
	Success bool   `json:"success"`
}

type DbRow interface {
	Scan(...any) error
}

func ParseJobRow(row DbRow) (Job, error) {
	var job DbJob
	if err := row.Scan(&job.Id, &job.Title, &job.Endpoint, &job.Method, &job.Payload,
		&job.Scheduled_at, &job.Created_on, &job.Status, &job.Retries, &job.Updated_on); err != nil {
		return Job{}, err
	}
	finalJob := Job{
		Id:           job.Id,
		Title:        job.Title,
		Endpoint:     job.Endpoint,
		Method:       job.Method,
		Payload:      job.Payload,
		Scheduled_at: job.Scheduled_at,
		Created_on:   job.Created_on,
		Status:       job.Status,
		Retries:      int(job.Retries.Int32),
		Updated_on:   job.Updated_on.String,
	}
	return finalJob, nil
}

func CreateJob(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var jobCreate JobCreate
	if err := json.NewDecoder(r.Body).Decode(&jobCreate); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Error: err.Error(), Success: false})
		return
	}
	jobCreate.Scheduled_at = jobCreate.Scheduled_at.UTC()
	if _, err := db.GetDb().Exec("INSERT INTO jobs (title, endpoint, method, payload, scheduled_at) values (?, ?, ?, ?, ?)",
		jobCreate.Title, jobCreate.Endpoint, jobCreate.Method, jobCreate.Payload, jobCreate.Scheduled_at); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Error: err.Error(), Success: false})
		return
	}
	json.NewEncoder(w).Encode(Response{Success: true})
}

func GetJobs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	jobId := r.URL.Query().Get("id")

	if jobId != "" {
		row := db.GetDb().QueryRow(`SELECT id, title, endpoint, method, payload, scheduled_at,
		created_on, status, retries, updated_on  FROM jobs WHERE id = ?`, jobId)
		if row.Err() != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(Response{Error: row.Err().Error(), Success: false})
			return
		}
		var job Job
		if row != nil {
			result, err := ParseJobRow(row)
			if err != nil {
				if err == sql.ErrNoRows {
					json.NewEncoder(w).Encode(Response{Success: true, Data: nil})
					return
				}
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(Response{Error: err.Error(), Success: false})
				return
			}
			job = result
		}
		json.NewEncoder(w).Encode(Response{Success: true, Data: job})

	} else {
		rows, err := db.GetDb().Query(`SELECT id, title, endpoint, method, payload, scheduled_at,
		created_on, status, retries, updated_on  FROM jobs`)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(Response{Error: err.Error(), Success: false})
			return
		}
		defer rows.Close()
		jobs := make([]Job, 0, 10)
		for rows.Next() {
			job, err := ParseJobRow(rows)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(Response{Error: err.Error(), Success: false})
				return
			}
			jobs = append(jobs, job)

		}
		json.NewEncoder(w).Encode(Response{Success: true, Data: jobs})
	}
}

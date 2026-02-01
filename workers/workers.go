package workers

import (
	"net/http"
	"sync"
	"time"

	"github.com/ravi-sankarp/go-job-queue/db"
	"github.com/ravi-sankarp/go-job-queue/scheduler"
)

var MAX_WORKERS int = 4

const (
	IDLE    string = "IDLE"
	RUNNING string = "RUNNING"
	SUCCESS string = "SUCCESS"
	FAILED  string = "FAILED"
)

type queue struct {
	jobs  []scheduler.Job
	mutex sync.Mutex
}

func (q *queue) dequeue() scheduler.Job {
	if q.len() == 0{
		return nil
	}
	q.mutex.Lock()
	defer q.mutex.Unlock()
	job:= q.jobs[0]
	q.jobs = q.jobs[1:]
	return job
}
func (q *queue) len() int {
	return len(q.jobs)
}

func pollJobs(q *queue) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		rows, err := db.GetDb().Query(`SELECT id, title, endpoint, method, payload, scheduled_at,
									   created_on, status, retries, updated_on  FROM jobs
									   WHERE scheduled_at <= datetime('now') AND status <> ?`, SUCCESS)
		if err != nil {
			panic(err)
		}
		q.mutex.Lock()
		for rows.Next() {
			job, err := scheduler.ParseJobRow(rows)
			if err != nil {
				panic(err)
			}
			q.jobs = append(q.jobs, job)
		}
		q.mutex.Unlock()
		rows.Close()
	}
}

func updateJob()error{
	return  nil
}

func worker(q *queue) {
	for {
			job := q.dequeue()
			if job ==nil{
				continue
			}
			req, err:=http.NewRequest(job.Method, job.Endpoint, job.Payload)

			if err!=nil{
				updateJob()
				continue
			}
			resp, err:=http.DefaultClient.Do(req)
			if err!=nil {
				updateJob()
				continue
			}
			defer resp.Body.Close()
		}
	}
}

func startWorkers(q *queue) {
	for i := 0; i < MAX_WORKERS; i++ {
		go worker(q)
	}
}

func Start() {
	q := &queue{
		jobs: make([]scheduler.Job, 0, MAX_WORKERS),
	}
	go pollJobs(q)

}

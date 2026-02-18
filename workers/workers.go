package workers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ravi-sankarp/go-job-queue/db"
	"github.com/ravi-sankarp/go-job-queue/scheduler"
)

var MAX_WORKERS int = 4

type status string

const (
	IDLE         status = "IDLE"
	RUNNING      status = "RUNNING"
	SUCCESS      status = "SUCCESS"
	FAILED       status = "FAILED"
	LOCK_TIMEOUT int    = 60
)

type HttpResponse struct {
	message string
}

type queue struct {
	jobs  []scheduler.Job
	mutex sync.Mutex
}

func (q *queue) dequeue() *scheduler.Job {
	if q.len() == 0 {
		return nil
	}
	q.mutex.Lock()
	defer q.mutex.Unlock()
	if q.len() == 0 {
		return nil
	}
	job := q.jobs[0]
	q.jobs = q.jobs[1:]
	return &job
}
func (q *queue) len() int {
	return len(q.jobs)
}

func pollJobs(q *queue) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		rows, err := db.GetDb().Query(`UPDATE jobs
									  SET locked_at = strftime('%s', 'now')
									  WHERE id IN
									  (
									   SELECT id FROM jobs
									   WHERE status <> ?
									   AND scheduled_at <= datetime('now')
									   AND (locked_at IS NULL OR locked_at < strftime('%s', 'now') - ?)
									   )
									   RETURNING id, title, endpoint, method, payload, scheduled_at,
									   created_on, status, retries, error_info, updated_on`, SUCCESS, LOCK_TIMEOUT)
		if err != nil {
			panic(err)
		}
		parsedRows := make([]scheduler.Job, 0, 20)
		for rows.Next() {
			job, err := scheduler.ParseJobRow(rows)
			if err != nil {
				panic(err)
			}
			parsedRows = append(parsedRows, job)
		}

		if err := rows.Err(); err != nil {
			panic(err)
		}

		q.mutex.Lock()
		q.jobs = append(q.jobs, parsedRows...)
		q.mutex.Unlock()
		rows.Close()

	}

}

func updateJob(id int, status status, err string) error {
	fmt.Println("Updating job with id = " + strconv.Itoa(id) + " status = " + string(status))
	_, error := db.GetDb().Exec(`UPDATE jobs SET status = ?, error_info = ?, updated_on = datetime('now'), locked_at = NULL
							   WHERE id = ?`, status, err, id)
	return error
}

func worker(q *queue) {
	for {
		job := q.dequeue()
		if job == nil {
			continue
		}
		fmt.Println("Executing job with title " + job.Title + " with id " + strconv.Itoa(job.Id))
		req, err := http.NewRequest(job.Method, job.Endpoint, bytes.NewReader([]byte(job.Payload)))

		if err != nil {
			updateJob(job.Id, FAILED, err.Error())
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			updateJob(job.Id, FAILED, err.Error())
			continue
		}
		if strings.HasPrefix(resp.Status, "2") == false {
			var msg string
			var result HttpResponse
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				msg = err.Error()
			}
			if err := json.Unmarshal(body, &result); err != nil {
				msg = err.Error()
			}

			updateJob(job.Id, FAILED, msg)
			continue
		}
		updateJob(job.Id, SUCCESS, "")
		resp.Body.Close()
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
	startWorkers(q)
}

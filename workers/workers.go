package workers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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

type status string

const (
	IDLE         status = "IDLE"
	RUNNING      status = "RUNNING"
	SUCCESS      status = "SUCCESS"
	FAILED       status = "FAILED"
	ABORTED      status = "ABORTED"
	LOCK_TIMEOUT int    = 60
	MAX_WORKERS  int    = 4
	MAX_RETIRES  int    = 5
	JOB_TIMEOUT  int    = 10
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
									   WHERE status NOT IN ( ?, ? )
									   AND scheduled_at <= datetime('now')
									   AND (locked_at IS NULL OR locked_at < strftime('%s', 'now') - ?)
									   )
									   RETURNING id, title, endpoint, method, payload, scheduled_at,
									   created_on, status, retries, error_info, updated_on`, SUCCESS, ABORTED, LOCK_TIMEOUT)
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

func updateJob(id int, status status, retries *int, err string) error {
	fmt.Println("Updating job with id = " + strconv.Itoa(id) + " status = " + string(status) + " error = " + err)
	_, error := db.GetDb().Exec(`UPDATE jobs SET status = ?, retries = ?, error_info = ?, updated_on = datetime('now'), locked_at = NULL
							   WHERE id = ?`, status, retries, err, id)
	return error
}

func updateFailedJob(job *scheduler.Job, err error) error {
	msg := err.Error()
	if errors.Is(err, context.DeadlineExceeded) {
		msg = "Error timeout"
	}
	status := FAILED
	if job.Retries == MAX_RETIRES {
		status = ABORTED
	}
	job.Retries++

	return updateJob(job.Id, status, &job.Retries, msg)
}

func executeJobRequest(job *scheduler.Job, ctx context.Context) error {
	fmt.Println("Executing job with title " + job.Title + " with id " + strconv.Itoa(job.Id))
	req, err := http.NewRequestWithContext(ctx, job.Method, job.Endpoint, bytes.NewReader([]byte(job.Payload)))

	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
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

		return errors.New(msg)
	}
	return nil
}

func worker(q *queue) {
	for {
		job := q.dequeue()
		if job == nil {
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), (time.Duration(JOB_TIMEOUT))*time.Second)
		if err := executeJobRequest(job, ctx); err != nil {
			updateFailedJob(job, err)
			cancel()
			continue
		}
		cancel()
		updateJob(job.Id, SUCCESS, nil, "")

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

package workers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
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
	BASE_BACKOFF int    = 1
	JOB_TIMEOUT  int    = 10
)

type HttpResponse struct {
	message string
}

func pollJobs(ch chan<- *scheduler.Job) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		rows, err := db.GetDb().Query(`UPDATE jobs
									  SET locked_at = strftime('%s', 'now')
									  WHERE id IN
									  (
									   SELECT id FROM jobs
									   WHERE status NOT IN ( ?, ? )
									   AND system_scheduled_at <= strftime('%s', 'now')
									   AND (locked_at IS NULL OR locked_at < strftime('%s', 'now') - ?)
									   )
									   RETURNING id, title, endpoint, method, payload, scheduled_at, system_scheduled_at,
									   created_on, status, retries, error_info, updated_on`, SUCCESS, ABORTED, LOCK_TIMEOUT)
		if err != nil {
			panic(err)
		}
		for rows.Next() {
			job, err := scheduler.ParseJobRow(rows)
			if err != nil {
				panic(err)
			}
			ch <- &job
		}

		if err := rows.Err(); err != nil {
			panic(err)
		}

	}

}

func updateJob(id int, status status, systemScheduled int64, retries *int, err string) error {
	fmt.Println("Updating job with id = " + strconv.Itoa(id) + " status = " + string(status) + " error = " + err)
	_, error := db.GetDb().Exec(`UPDATE jobs SET status = ?, system_scheduled_at =?, retries = ?, error_info = ?, updated_on = datetime('now'), locked_at = NULL
							   WHERE id = ?`, status, systemScheduled, retries, err, id)
	return error
}

func updateFailedJob(job *scheduler.Job, err error) error {
	msg := err.Error()
	if errors.Is(err, context.DeadlineExceeded) {
		msg = "Error timeout"
	}
	job.Retries++
	status := FAILED
	if job.Retries == MAX_RETIRES {
		status = ABORTED
	} else {
		retryAfter := time.Duration(BASE_BACKOFF * (1 << job.Retries))
		job.System_scheduled_at = time.Now().Unix() + int64(retryAfter)
		fmt.Println("current Time " + time.Now().UTC().String() + " Scheduled at " + time.Unix(job.System_scheduled_at, 0).UTC().String())
	}

	return updateJob(job.Id, status, job.System_scheduled_at,
		&job.Retries, msg)
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

func worker(ch <-chan *scheduler.Job) {
	defer func() {
		if r := recover(); r != nil {
			log.Fatal("Worker Panic")
			log.Fatal(r)
		}
	}()
	for {
		select {
		case job, ok := <-ch:
			if !ok {
				return
			}
			func() {
				defer func() {
					if r := recover(); r != nil {
						fmt.Println("Worker Panic")
						fmt.Println(r)
					}
				}()
				ctx, cancel := context.WithTimeout(context.Background(), (time.Duration(JOB_TIMEOUT))*time.Second)
				defer cancel()
				if err := executeJobRequest(job, ctx); err != nil {
					updateFailedJob(job, err)
					return
				}
			}()
		default:
			continue
		}
	}

}
func startWorkers(ch <-chan *scheduler.Job) {
	for i := 0; i < MAX_WORKERS; i++ {
		go worker(ch)
	}
}

func Start() {
	ch := make(chan *scheduler.Job, 10)
	go pollJobs(ch)
	startWorkers(ch)
}

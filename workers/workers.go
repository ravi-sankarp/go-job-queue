package workers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
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
	BASE_BACKOFF int    = 1
	JOB_TIMEOUT  int    = 10
)

type HttpResponse struct {
	message string
}

func pollJobs(ctx context.Context, ch chan<- *scheduler.Job) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		rows, err := db.GetDb().QueryContext(ctx, `UPDATE jobs
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

func updateJob(ctx context.Context, id int, status status, systemScheduled int64, retries *int, err error) error {
	var errorInfo sql.NullString
	if err != nil {
		errorInfo.String = err.Error()
		errorInfo.Valid = true
	}
	fmt.Println("Updating job with id = " + strconv.Itoa(id) + " status = " + string(status) + " error = " + errorInfo.String)
	_, execErr := db.GetDb().ExecContext(ctx, `UPDATE jobs SET status = ?, system_scheduled_at =?, retries = ?, error_info = ?, updated_on = datetime('now'), locked_at = NULL
							   WHERE id = ?`, status, systemScheduled, retries, errorInfo, id)
	return execErr
}

func updateFailedJob(ctx context.Context, job *scheduler.Job, err error) error {
	if errors.Is(err, context.DeadlineExceeded) {
		err = errors.New("Error timeout")
	}
	job.Retries++
	status := FAILED
	if job.Retries == MAX_RETIRES {
		status = ABORTED
	} else {
		retryAfter := time.Duration(BASE_BACKOFF * (1 << job.Retries))
		job.System_scheduled_at = time.Now().Unix() + int64(retryAfter)
	}

	return updateJob(ctx, job.Id, status, job.System_scheduled_at,
		&job.Retries, err)
}

func executeJobRequest(ctx context.Context, job *scheduler.Job) error {
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
	return updateJob(ctx, job.Id, SUCCESS, job.System_scheduled_at,
		nil, nil)
}

func worker(ctx context.Context, ch <-chan *scheduler.Job, wg *sync.WaitGroup) {
	defer func() {
		if r := recover(); r != nil {
			log.Fatal("Worker Panic")
			log.Fatal(r)
		}
	}()
	defer wg.Done()
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
				timeoutCtx, cancel := context.WithTimeout(ctx, (time.Duration(JOB_TIMEOUT))*time.Second)
				defer cancel()
				if err := executeJobRequest(timeoutCtx, job); err != nil {
					updateFailedJob(ctx, job, err)
					return
				}
			}()
		case <-ctx.Done():
			fmt.Println("Shutting down woker due to ctx cancel")
			return
		}
	}

}
func startWorkers(ctx context.Context, ch <-chan *scheduler.Job, wg *sync.WaitGroup) {
	for i := range MAX_WORKERS {
		wg.Add(i)
		go worker(ctx, ch, wg)
	}
}

func Start(ctx context.Context) {
	ch := make(chan *scheduler.Job, 10)
	wg := sync.WaitGroup{}
	go pollJobs(ctx, ch)
	startWorkers(ctx, ch, &wg)
	wg.Wait()
}

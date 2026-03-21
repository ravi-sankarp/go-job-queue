# go-job-queue

A lightweight multi-threaded job queue written in Go that schedules HTTP jobs, persists them in SQLite, and executes them concurrently with retries and exponential backoff.

## Overview

`go-job-queue` is a simple background job processing system built in Go. It allows clients to create scheduled jobs through an HTTP API, stores those jobs in SQLite, and runs worker goroutines that poll, lock, and execute due jobs.


## Features

- **HTTP API for job creation and retrieval**
- **Persistent job storage** in SQLite
- **Scheduled execution** 
- **Concurrent worker pool**
- **Lease-based job locking** to avoid duplicate pickup
- **Retry support** for failed jobs
- **Exponential backoff** between retries
- **Abort after max retries**
- **Per-job timeout handling**
- **Mock target server** for testing failure and latency scenarios

## How it works

1. A client submits a job to the API.
2. The job is stored in the `jobs` table with status `IDLE`.
3. Worker polling runs every 2 seconds.
4. Due jobs are atomically locked in SQLite and pushed into a worker channel.
5. Workers execute the HTTP request defined by the job.
6. If the request succeeds, the job is marked `SUCCESS`.
7. If it fails or times out, the job is marked `FAILED` and rescheduled using exponential backoff.
8. After the retry limit is reached, the job is marked `ABORTED`.

## API

### Create a job

**POST** `/jobs/`

#### Request body

```json
{
  "title": "Send webhook",
  "endpoint": "http://localhost:3000/",
  "method": "GET",
  "payload": "{}",
  "scheduled_at": "2026-03-21T12:00:00.000Z"
}
```

#### Example

```bash
curl --request POST \
  --url http://localhost:8000/jobs/ \
  --header 'content-type: application/json' \
  --data '{
    "title":"Test Job",
    "endpoint":"http://localhost:3000/",
    "payload":"{}",
    "method":"GET",
    "scheduled_at":"2026-03-21T12:00:00.000Z"
  }'
```

#### Success response

```json
{
  "error": "",
  "data": null,
  "success": true
}
```

---

### Get all jobs

**GET** `/jobs/`

#### Example

```bash
curl http://localhost:8000/jobs/
```

---

### Get a single job by id

**GET** `/jobs/?id=1`

#### Example

```bash
curl "http://localhost:8000/jobs/?id=1"
```

## Running locally

### Prerequisites

- Go installed
- CGO support available for `go-sqlite3`

### Start the scheduler service

```bash
go run main.go
```

Or with make:

```bash
make
```

The API server will start on:

```text
http://localhost:8000
```

### Start the mock server

In a separate terminal:

```bash
go run mock-server/main.go
```

Or with make:

```bash
make mock-server
```

The mock server runs on:

```text
http://localhost:3000
```

### Development mode

If you use `air` for hot reload:

```bash
make dev
```

## Load testing

The repository includes a script to enqueue a large number of jobs:

```bash
bash test.sh
```

This script creates 10,000 scheduled HTTP jobs targeting the local mock server.

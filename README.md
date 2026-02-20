# safe-transaction
Remake attempt of containerized electricity token submission using Go lang.

# Go Transaction-Safe API Service

A containerized backend service written in Go that demonstrates reliable API behavior under real-world failure conditions (retries, duplicate submissions, and service restarts).
This project focuses on **backend reliability and operational safety**, not just CRUD functionality.

---

## Purpose

In distributed systems, clients often retry requests due to:

* network timeout
* mobile connection drop
* double click / duplicate submission
* payment gateway retry

If the backend is not designed carefully, this creates **duplicate transactions** and data inconsistencies.

This service demonstrates a **retry-safe (idempotent) API** design where the same request can be safely repeated without producing multiple records.

---

## Key Features

### 1. REST API Service (Golang + Gin)

* HTTP server implemented as a compiled Go binary
* Clean request handling
* JSON request/response
* Structured routing

Endpoints:

| Method | Endpoint    | Description                           |
| ------ | ----------- | ------------------------------------- |
| GET    | `/users`    | Retrieve all users                    |
| POST   | `/users`    | Create a user                         |
| POST   | `/payments` | Create retry-safe payment transaction |

---

### 2. Idempotent Payment Handling (Important Part)

The `/payments` endpoint is designed to be **retry-safe**.

Each request contains:

```
external_id
```

If a client retries the same request:

* the server does **not create a second transaction**
* the server returns the original result

Implementation:

* Unique constraint on `external_id`
* Insert-first strategy
* Duplicate detection via database
* Safe recovery of original record

This prevents double-charging and maintains ledger consistency.

---

### 3. Database Integration (MySQL)

* Relational database storage
* Transactional integrity
* Unique constraint enforcement

Tables:

**users**

```
id
name
email
```

**payments**

```
id
external_id (UNIQUE)
amount
status
created_at
```

---

### 4. Containerized Environment (Docker + Docker Compose)

The service runs entirely in containers.

Services:

* Go API container
* MySQL database container

Benefits:

* reproducible environment
* easy local setup
* production-like behavior

Run locally:

```bash
docker compose up --build
```

---

### 5. Graceful Shutdown

The service handles termination signals correctly:

* stops accepting new requests
* finishes in-flight requests
* safely closes database connections

This prevents partial writes and inconsistent data during deployments.

---

### 6. Request Logging Middleware

Custom middleware logs:

* client IP
* HTTP method
* endpoint
* status code
* request latency

Example log:

```
[172.19.0.1] POST /payments | 200 | 3.2ms
```

This enables operational debugging and performance visibility.

---

### 7. Concurrency-Safe Design

The payment endpoint is safe even when two identical requests arrive simultaneously.

Protection mechanism:

* database uniqueness constraint
* retry-aware handler logic

This simulates real-world payment gateway behavior.

---

## Architecture Overview

Request flow:

```
Client Request
    ↓
Gin Router
    ↓
Logging Middleware
    ↓
Handler
    ↓
Database Query
    ↓
Response
```

The service acts as a standalone HTTP server (no external web server required).

---

## Why This Project Exists

Many API tutorials demonstrate only CRUD operations.
Real backend systems fail due to:

* retries
* timeouts
* duplicate submissions
* service restarts

This project demonstrates how a backend should behave under those conditions.

It focuses on:

* correctness
* safety
* predictable behavior

rather than UI or frontend.

---

## Planned Improvements

* Prometheus metrics endpoint
* Grafana dashboard
* authentication (JWT)
* integration tests
* request ID tracing

---

## Tech Stack

* Go (Golang)
* Gin Web Framework
* MySQL
* Docker / Docker Compose

---

## How to Test Retry Safety

Run the same request twice:

```bash
curl -X POST http://localhost:8080/payments \
-H "Content-Type: application/json" \
-d '{"external_id":"INV-777","amount":50000}'
```

The response ID will remain the same.

This confirms idempotent behavior.

---

## What This Demonstrates

This project demonstrates backend engineering practices:

* safe transaction handling
* failure-aware API design
* operational logging
* containerized deployment
* lifecycle management

The goal is to model a service that can be trusted in production-like scenarios.

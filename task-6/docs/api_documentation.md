# Task Management API — MongoDB Edition

Base URL: `http://localhost:8080`

## MongoDB Setup

- Install MongoDB locally or use Atlas (cloud).
- Environment variables (defaults shown):
  - `MONGO_URI=mongodb://localhost:27017`
  - `MONGO_DB=task_manager`
  - `MONGO_COLLECTION=tasks`

Run locally:
```bash
export MONGO_URI="mongodb://localhost:27017"
export MONGO_DB="task_manager"
export MONGO_COLLECTION="tasks"
go mod tidy
go run ./...

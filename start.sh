#!/usr/bin/env bash

export PORT=8080
export DATABASE_URL="postgres://user:password@localhost:5443/postgres?sslmode=disable"

go run ./cmd/api/main.go

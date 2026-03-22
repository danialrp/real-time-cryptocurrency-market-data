#!/bin/sh
docker run --rm -v $(pwd):/app -w /app golang:1.24-alpine sh -c "go mod tidy"

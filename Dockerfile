##########################################
# Stage 1: Development (Air + Debug)
##########################################
FROM golang:1.24-alpine AS dev

RUN apk add --no-cache git curl make tzdata bash

# Install Dev Tools > "Air" + "Delve"
RUN go install github.com/air-verse/air@latest \
    && go install github.com/go-delve/delve/cmd/dlv@latest

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# CMD ["air"]

##########################################
# Stage 2: Build All Binaries + Migrate
##########################################
FROM golang:1.24-alpine AS build

RUN apk add --no-cache git tzdata bash curl

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Build All Binaries
RUN go build -o runtime ./cmd/runtime \
 && go build -o scheduler ./cmd/scheduler \
 && go build -o worker ./cmd/worker

# Install "Migrate CLI" Based on architecture (For All Binaries)
RUN ARCH=$(uname -m) && \
    if [ "$ARCH" = "x86_64" ]; then \
        curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz; \
    elif [ "$ARCH" = "aarch64" ]; then \
        curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-arm64.tar.gz | tar xvz; \
    else \
        echo "Unsupported architecture: $ARCH" && exit 1; \
    fi

##########################################
# Stage 3: Runtime for Main Service
##########################################
FROM alpine:3.19 AS runtime-main

RUN apk add --no-cache tzdata bash

ENV TZ=Asia/Tehran
WORKDIR /app

COPY --from=build /app/runtime .
COPY --from=build /app/database/migrations ./database/migrations
COPY --from=build /app/migrate /usr/local/bin/migrate

EXPOSE 9090
CMD ["./runtime"]

##########################################
# Stage 4: Runtime for Worker
##########################################
FROM alpine:3.19 AS runtime-worker

RUN apk add --no-cache tzdata bash
ENV TZ=Asia/Tehran

WORKDIR /app
COPY --from=build /app/worker .
COPY --from=build /app/migrate /usr/local/bin/migrate
COPY wait-for-it.sh /wait-for-it.sh
RUN chmod +x /wait-for-it.sh

CMD ["./worker"]

##########################################
# Stage 5: Runtime for Scheduler
##########################################
FROM alpine:3.19 AS runtime-scheduler

RUN apk add --no-cache tzdata bash
ENV TZ=Asia/Tehran

WORKDIR /app
COPY --from=build /app/scheduler .
COPY --from=build /app/migrate /usr/local/bin/migrate
COPY wait-for-it.sh /wait-for-it.sh
RUN chmod +x /wait-for-it.sh

CMD ["./scheduler"]

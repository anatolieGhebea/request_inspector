# Stage 1: Build the Go app
FROM golang:1.21 AS builder

WORKDIR /app
COPY . .
RUN GOOS=linux GOARCH=amd64 go build -o request_inspector ./...

# Stage 2: Copy the binary into a scratch container
FROM scratch

COPY --from=builder /app/request_inspector /request_inspector
ENTRYPOINT ["/request_inspector"]
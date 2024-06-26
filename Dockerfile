# Stage 1: Build the Go app
FROM golang:1.21 AS builder

WORKDIR /app
COPY . .
RUN GOOS=linux GOARCH=amd64 go build -o request_inspector ./...

# Stage2: Use an image with bash and cp to copy the directory
FROM alpine AS builder_frontend
WORKDIR /app
COPY ./static ./static

# Stage 3: Copy the binary into a scratch container
FROM scratch

COPY --from=builder /app/request_inspector /request_inspector
COPY --from=builder_frontend /app/static /static
ENTRYPOINT ["/request_inspector"]
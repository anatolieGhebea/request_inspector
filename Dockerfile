# Stage 1: Build the Go app
FROM --platform=$BUILDPLATFORM golang:1.21 AS builder

ARG TARGETARCH
WORKDIR /app
COPY . .
RUN ls -la /app && \
    GOOS=linux GOARCH="$TARGETARCH" go build -o request_inspector . && \
    ls -la /app

# Stage 2: Use an image with bash and cp to copy the directory
FROM alpine:3.18 AS builder_frontend
WORKDIR /app
COPY ./static ./static
RUN ls -la /app

# Stage 3: Copy the binary into a scratch container
FROM scratch

COPY --from=builder /app/request_inspector /request_inspector
COPY --from=builder_frontend /app/static /static
ENTRYPOINT ["/request_inspector"]
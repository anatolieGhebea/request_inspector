# Stage 1: Build the Go app
FROM --platform=$BUILDPLATFORM golang:1.21 AS builder

ARG TARGETARCH
WORKDIR /app
COPY . .
# Ensure static linking
RUN ls -la /app && \
    CGO_ENABLED=0 GOOS=linux GOARCH="$TARGETARCH" go build -a -ldflags '-extldflags "-static"' -o request_inspector . && \
    ls -la /app

# Stage 2: Use an image with bash and cp to copy the directory
FROM alpine:3.18 AS builder_frontend
WORKDIR /app
COPY ./static ./static
RUN ls -la /app

# Stage 3: Final stage
FROM scratch

COPY --from=builder /app/request_inspector /request_inspector
COPY --from=builder_frontend /app/static /static
CMD ["/request_inspector"]
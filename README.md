[![Quality Gate Status](https://sonar.ga-dns.com/api/project_badges/measure?project=requestinspector&metric=alert_status&token=sqb_a26587bad346841e0e0cb411b6a289d8b078b54a)](https://sonar.ga-dns.com/dashboard?id=requestinspector)

# Introduction

Have you ever been in a situation where you needed to inspect a request from a third-party service that lacked clear documentation? This application provides a simple tool to inspect the request header and payload of any HTTP request. Generate a session, use the provided URL as the endpoint for the request you want to examine, and view the request header and payload on the session page.

# Usage
@TODO: 

An instance of this application is live at [request_inspector](http://request_inspector.ga-dns.com/), where you can see it in action. On the page, create a session. Then, using any HTTP client (e.g., curl, Postman), send a request to the URL provided on the session page. You will see the request header and payload displayed on the session page.

Each session is valid for one hour and can hold up to five requests. Throttling is in place to prevent abuse, allowing a maximum of ten requests per minute. 

None of the requests are persisted on disk or in a database; they are stored in memory and will be lost when the session expires or is terminated by the user.

**NOTE** 
the current version is missing the UI, so you will need to use an HTTP client to send requests to the provided URLs.
`/create` to create a session
`/session/<session_id>` to view request details made to the `/request/<session_id>` endpoint
`/request/<session_id>` the end point that logs the request
`/extend/<session_id>` to extend the session

# Installation 
When working with highly sensitive data, it is recommended to host your own instance of this application. The application is built using the Go programming language and can run on any platform that supports Go. For the best security, consider installing by compiling the source code. This ensures that the code you are running matches the code in the repository. A Linux-compiled binary is also provided for convenience.

## As Docker Container (Recommended)

To run the application as a Docker container, you will need Docker and Docker Compose installed on your system.
Then you can create a `docker-compose.yml` file with the following content:

```yaml
version: '3'

services:
  inspector:
    image: ghebby/request_inspector:0.0.1
    restart: unless-stopped  
    environment:
      - TZ=Europe/Rome
      - PORT=9001
    ports:
      # host:container
      - 9099:9001
```

> Check the [hub.docker.com/r/ghebby/request_inspector](https://hub.docker.com/r/ghebby/request_inspector) page for the latest available version of the image.

Run `docker-compose up -d` to start the application. The application will be available on port 9099 of the host machine.
Run `docker-compose down` to stop the application.

I you need https support, consider adding a reverse proxy like Nginx or Apache in front of the application, the best way is to add a container with the reverse proxy to the `docker-compose.yml` file, run them in the same docker network and configure the reverse proxy to forward the requests to the application container.

## Compiled Binary

Download a release from the release page. The compressed file contains a binary file that you can run on your system and the static folder with the UI files.
Make sure you have the `make` cli tool installed on your system.
1.  `make install_local f=<file_name>.tar.gz` to install the application
2.  `make start_local` to start the application
3.  `make show_running` to view the running instances 
4.  `make stop_local` to stop the application

## From Source 

You can clone this project and run `go run main.go` to start the application in development mode.
`GOOS=<platform> GOARCH=<arch> go build -o <output_file>` to compile the application for a specific platform and architecture.



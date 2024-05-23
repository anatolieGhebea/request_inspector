# Introduction

Have you ever been in a situation where you needed to inspect a request from a third-party service that lacked clear documentation? This application provides a simple tool to inspect the request header and payload of any HTTP request. Generate a session, use the provided URL as the endpoint for the request you want to examine, and view the request header and payload on the session page.

# Usage

An instance of this application is live at [request_inspector.app](http://request_inspector.ga-dns.com/), where you can see it in action. On the page, create a session. Then, using any HTTP client (e.g., curl, Postman), send a request to the URL provided on the session page. You will see the request header and payload displayed on the session page.

Each session is valid for one hour and can hold up to five requests. Throttling is in place to prevent abuse, allowing a maximum of five requests per minute. 

None of the requests are persisted on disk or in a database; they are stored in memory and will be lost when the session expires or is terminated by the user.

# Installation 

When working with highly sensitive data, it is recommended to host your own instance of this application. The application is built using the Go programming language and can run on any platform that supports Go. For the best security, consider installing by compiling the source code. This ensures that the code you are running matches the code in the repository. A Linux-compiled binary is also provided for convenience.

## Compiled Binary

Use `wget` to download the binary. Set the execution permission on the file with `chmod +x request_inspector` and run the binary with `./request_inspector`.

## From Source 

Ensure that Go is installed on your system and that you can compile Go projects. Clone this repository and build the project; a Makefile is provided for convenience.

Copy the binary to the server where you want to run the application and execute it.

## As Docker Container

Run `make install_local f=request_inspector<version>.tar.gz` to extract the binary file in the bin directory. 
Then create a copy of the `.env.example` file and name it `.env`. Update the values in the `.env` file as needed. 
Finally, run `docker-compose up -d` to start the application.

When running the application in a Docker container, the application won't be available on the host machine's port, you will need a proxy container like nginx to act as reverse proxy to the application.
If you want to run the application on the host machine's port, you can edit the docker compose file to expose the application on the host machine's port.

# Note 

If HTTPS support is needed, consider running this application behind a reverse proxy like Nginx or Apache. Additionally, add a firewall rule to block all incoming traffic except from the reverse proxy.

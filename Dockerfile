FROM scratch
# WORKDIR /
# RUN mkdir app
COPY ./bin/request_inspector /request_inspector
# RUN ls -la
# RUN chmod +x /app/request_inspector
ENTRYPOINT ["/request_inspector"]
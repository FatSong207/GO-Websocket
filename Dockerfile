FROM        golang:alpine
RUN         mkdir -p /app
WORKDIR     /app
COPY        . .
RUN         go mod tidy
RUN         go build -o app

EXPOSE 8082
ENTRYPOINT  ["./app"]
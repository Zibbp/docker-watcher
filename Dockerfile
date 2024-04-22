FROM golang:1 AS build-stage-01

RUN mkdir /app
ADD . /app
WORKDIR /app

RUN go build -o docker-watcher main.go

FROM debian:12-slim AS production

COPY --from=build-stage-01 /app/docker-watcher /usr/local/bin/

CMD ["/usr/local/bin/docker-watcher"]
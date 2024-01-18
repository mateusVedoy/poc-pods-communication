FROM golang:1.20.0-alpine3.17 as builder

WORKDIR /go/src/github.com/poc/poc-pods-communication.git/go-app

RUN apk update && apk add build-base

COPY . .

RUN go mod download

# RUN go test ./...

RUN go build -ldflags "-s -w" src/main.go

FROM alpine:3.17

WORKDIR /app

ENV KUBERNETES_SERVICE_HOST=10.96.0.1
ENV KUBERNETES_SERVICE_PORT=443
ENV USER=inCluster

COPY --from=builder /go/src/github.com/poc/poc-pods-communication.git/go-app/main .

ENV TZ=America/Sao_Paulo
RUN apk update && \
    apk add --no-cache tzdata curl unzip && \
    ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone && \
    curl -sS https://infraestructure-dependencies.s3.amazonaws.com/bootstrap/linux/tzupdate-alpine.sh|sh


ADD k8s k8s

CMD [ "./main" ]
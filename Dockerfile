FROM golang:1.12-alpine
RUN apk update
RUN apk add openssl ca-certificates git curl
ENV GO111MODULE on
RUN mkdir -p /root/.kube/certs
WORKDIR /go/src/github.com/akkeris/node-watcher-f5
COPY . .
RUN go build .
CMD "./start.sh"
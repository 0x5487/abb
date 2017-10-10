FROM golang:1.9 AS builder
COPY ./ /go/src/github.com/jasonsoft/abb/
WORKDIR /go/src/github.com/jasonsoft/abb/cmd/api/
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build

FROM alpine:3.6
RUN apk update && \
    apk upgrade && \
    apk add --no-cache curl && \
    rm -rf /var/cache/apk/* && \
    mkdir -p /jasonsoft/abb_api

COPY --from=builder /go/src/github.com/jasonsoft/abb/cmd/api/api /jasonsoft/abb_api/api
WORKDIR /jasonsoft/abb_api/
CMD ./api


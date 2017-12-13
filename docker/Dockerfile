FROM golang:1.9 AS builder
COPY ./ /go/src/github.com/jasonsoft/abb/
WORKDIR /go/src/github.com/jasonsoft/abb/cmd/api/
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build

FROM alpine:3.7
RUN apk update && \
    apk upgrade && \
    apk add --no-cache curl && \
    rm -rf /var/cache/apk/* && \
    mkdir -p /jasonsoft/abb_api


COPY --from=builder /go/src/github.com/jasonsoft/abb/cmd/api/api /jasonsoft/abb_api/api
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser && \
    chown -R appuser:appuser /jasonsoft/abb_api/
USER appuser

WORKDIR /jasonsoft/abb_api/
CMD ./api


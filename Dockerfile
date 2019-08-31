FROM golang:1.12.9-alpine3.10 AS builder
RUN apk --no-cache add build-base git gcc
ADD . /charlyedu
WORKDIR /charlyedu
RUN go build -o service cmd/main.go

FROM alpine:3.10
WORKDIR /app
COPY --from=builder /charlyedu/service /app/
ENTRYPOINT /app/service
FROM golang:1.16.4 as builder

WORKDIR /go/src

COPY go.mod go.sum ./
RUN go mod download

COPY ./main.go  ./

ARG CGO_ENABLED=0
ARG GOOS=linux
ARG GOARCH=amd64
RUN go build \
    -o /go/bin/main \
    -ldflags '-s -w'

FROM alpine as runner

RUN apk add --no-cache ca-certificates

COPY --from=builder /go/bin/main /app/main

ENTRYPOINT ["/app/main"]

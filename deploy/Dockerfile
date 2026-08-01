FROM golang:1.26-alpine AS builder

ENV CGO_ENABLED=0 GOOS=linux

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY --parents **/*.go internal ./

RUN go build -o /no-time-to-explain main.go

# Deployable image
FROM alpine:latest

RUN adduser -D appuser
USER appuser

COPY --from=builder /no-time-to-explain /no-time-to-explain

EXPOSE 3000
ENTRYPOINT ["/no-time-to-explain"]

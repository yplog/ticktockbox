FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o /ticktockbox ./cmd/ticktockbox

FROM debian:bullseye-slim

COPY --from=builder /ticktockbox /ticktockbox

CMD ["/ticktockbox"]
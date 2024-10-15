FROM --platform=linux/arm64 golang:1.23-alpine AS build_arm64

WORKDIR /app

RUN apk add --no-cache gcc musl-dev libc-dev

COPY go.mod go.sum ./

RUN go mod download

COPY . .

ENV CGO_ENABLED=1 GOOS=linux GOARCH=arm64 CC="gcc" CXX="g++"
RUN go build -o /ticktockbox ./cmd/ticktockbox

FROM --platform=linux/arm64 alpine:latest AS runtime_arm64

RUN apk add --no-cache libc6-compat

COPY --from=build_arm64 /ticktockbox /ticktockbox

CMD ["/ticktockbox"]

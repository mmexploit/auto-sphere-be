FROM golang:1.23.2 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy

COPY . .

RUN go build -o main cmd/api/main.go

FROM debian:bookworm-slim


WORKDIR /app

RUN apt-get update && apt-get install -y libpq-dev

COPY --from=builder /app/main .

EXPOSE 4321

CMD ["./main"]

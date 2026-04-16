# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/money-transfer ./cmd/main

FROM alpine:3.22

RUN adduser -D -u 10001 appuser

WORKDIR /app

COPY --from=builder /out/money-transfer /app/money-transfer

USER appuser

ENTRYPOINT ["/app/money-transfer"]
EXPOSE 8080 9090
CMD ["--app-port=8080", "--metrics-port=9090"]

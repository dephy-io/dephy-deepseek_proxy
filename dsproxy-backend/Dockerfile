FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o dsproxy-backend ./main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/dsproxy-backend .
COPY config.yaml ./config.yaml 
EXPOSE 8080
CMD ["./dsproxy-backend", "config.yaml"] 
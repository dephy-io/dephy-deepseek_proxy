FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o dsproxy-backend ./main.go
# 添加调试：列出生成的文件
RUN ls -lh /app/dsproxy-backend

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/dsproxy-backend .
# 添加调试：确认文件是否复制到最终镜像
RUN ls -lh /app/dsproxy-backend
COPY config.yaml ./config.yaml 
EXPOSE 8080
CMD ["./dsproxy-backend", "config.yaml"] 
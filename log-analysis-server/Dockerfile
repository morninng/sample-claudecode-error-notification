FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY ./log-analysis-server/go.mod ./log-analysis-server/go.sum ./
RUN go mod download

COPY ./log-analysis-server .
RUN CGO_ENABLED=0 GOOS=linux go build -o log-analysis-server .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/log-analysis-server .

EXPOSE 8080

CMD ["./log-analysis-server"]

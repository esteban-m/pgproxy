FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o pgproxy .

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/pgproxy .
EXPOSE 23306
EXPOSE 33306
ENV LISTEN_MYSQL_ADDR=0.0.0.0:23306 \
    NEW_LISTEN_MYSQL_ADDR=0.0.0.0:33306 \
    REMOTE_MYSQL_ADDR=127.0.0.1:3306
CMD ["./pgproxy", \
  "-local=0.0.0.0:23306", \
  "-remote=127.0.0.1:3306", \
  "-newlocal=0.0.0.0:33306" \
]

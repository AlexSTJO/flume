FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .


RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o flume .

FROM alpine:3.19

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/flume .
COPY --from=builder /app/.flume /app/.flume

EXPOSE 8080

ENV PORT=8080
ENV URL=http://localhost

ENTRYPOINT ["./flume"]

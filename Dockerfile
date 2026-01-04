FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .


RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o flume .

FROM alpine:3.19


RUN apk add --no-cache \
    ca-certificates \
    git \
    openssh \
    curl \
    unzip \
    nodejs \
    npm

RUN curl -fsSL https://releases.hashicorp.com/terraform/1.8.5/terraform_1.8.5_linux_amd64.zip -o terraform.zip \
    && unzip terraform.zip \
    && mv terraform /usr/local/bin/terraform \
    && chmod +x /usr/local/bin/terraform \
    && rm terraform.zip

WORKDIR /app

COPY --from=builder /app/flume .
RUN mkdir -p /app/.flume

EXPOSE 8080

ENV PORT=8080
ENV URL=http://localhost

ENTRYPOINT ["./flume"]

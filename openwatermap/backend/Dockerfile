FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY . .

RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -o openwatermap .

FROM alpine:3.19

RUN apk add --no-cache wget

WORKDIR /app

COPY --from=builder /app/openwatermap .

RUN mkdir -p /data

EXPOSE 8080

CMD ["./openwatermap"]

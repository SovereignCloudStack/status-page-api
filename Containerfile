FROM golang:1.20-alpine3.18 as builder

WORKDIR /build

COPY . .

RUN go build -o status-page-api main.go

FROM alpine:3.18

WORKDIR /app

COPY --from=builder /build/status-page-api .
COPY provisioning.yaml .
COPY LICENSE .

EXPOSE 3000/tcp

ENTRYPOINT /app/status-page-api

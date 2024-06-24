FROM docker.io/golang:1.22.4-alpine3.20 as builder

RUN apk upgrade --no-cache

WORKDIR /build

COPY . .

RUN go build -ldflags "-s -w" -o status-page-api cmd/status-page-api/main.go

FROM docker.io/alpine:3.20

RUN apk upgrade --no-cache

WORKDIR /app

COPY --from=builder /build/status-page-api .
COPY provisioning.yaml .
COPY LICENSE .
COPY entrypoint.sh .

EXPOSE 3000/tcp

ENTRYPOINT [ "/app/entrypoint.sh" ]
CMD [ "/app/status-page-api" ]

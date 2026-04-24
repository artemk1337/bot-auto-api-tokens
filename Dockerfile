FROM golang:1.25.6-alpine AS builder

WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=linux go build -o /out/bot ./cmd/bot

FROM alpine:3.22

WORKDIR /app
COPY --from=builder /out/bot /usr/local/bin/bot
COPY docs ./docs
COPY configs/gpt-oss-20b.json ./config.json

USER nobody

ENTRYPOINT ["/usr/local/bin/bot"]
CMD ["-config", "/app/config.json"]

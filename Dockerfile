FROM golang:1.27.2-alpine AS builder

WORKDIR /app

COPY go.sum go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/do-together ./cmd
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/email-worker ./cmd/email-worker



FROM alpine:3.22

RUN addgroup -S app && adduser -S -G app app \
    && mkdir -p /data/avatars \
    && chown -R app:app /data/avatars

WORKDIR /app

COPY --from=builder /app/do-together ./do-together
COPY --from=builder /app/email-worker ./email-worker

USER app

EXPOSE 8080

CMD ["./do-together"]

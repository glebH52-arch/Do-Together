FROM golang:1.25.0-alpine AS builder

WORKDIR /app

COPY go.sum go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/do-together ./cmd



FROM alpine:3.22

RUN addgroup -S app && adduser -S -G app app

WORKDIR /app

COPY --from=builder /app/do-together ./do-together

USER app

EXPOSE 8080

ENTRYPOINT ["./do-together"]
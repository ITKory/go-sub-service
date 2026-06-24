FROM golang:1.26-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /app/server ./cmd/app

FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata wget && adduser -D -g '' appuser

WORKDIR /app

COPY --from=builder /app/server /app/server
COPY --from=builder /app/migrations /app/migrations
COPY --from=builder /app/config.yaml /app/config.yaml

USER appuser

EXPOSE 8080

CMD ["/app/server"]
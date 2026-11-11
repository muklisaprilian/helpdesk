FROM golang:1.23-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git ca-certificates
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/helpdesk ./cmd/server

FROM alpine:3.20
WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY --from=builder /bin/helpdesk /app/helpdesk
COPY migrations /app/migrations
ENV MIGRATIONS_DIR=/app/migrations
ENV PORT=8080
EXPOSE 8080
CMD ["/app/helpdesk"]

FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o taskapi .

FROM alpine:3.19
RUN adduser -D appuser
USER appuser
COPY --from=builder /app/taskapi /taskapi
COPY --from=builder /app/migrations /migrations
EXPOSE 8080
ENTRYPOINT ["/taskapi"]

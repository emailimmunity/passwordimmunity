# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o passwordimmunity ./src

# Final stage
FROM alpine:3.18

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app

COPY --from=builder /app/passwordimmunity .
COPY --from=builder /app/src/static ./static

ENV SERVER_ADDR=:8000
EXPOSE 8000

CMD ["./passwordimmunity"]

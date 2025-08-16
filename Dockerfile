FROM alpine:latest AS builder

RUN apk add --no-cache \
    build-base \
    go \
    git

WORKDIR /app

COPY . .

RUN go mod tidy && go build -o /docker-manager


FROM alpine:latest

WORKDIR /app

COPY --from=builder /docker-manager /app/docker-manager

RUN chmod +x /app/docker-manager

EXPOSE 15000

CMD ["/app/docker-manager", "-ip", "0.0.0.0", "-port", "15000"]
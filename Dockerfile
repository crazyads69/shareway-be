FROM golang:1.23.2-alpine3.20 AS builder
WORKDIR /app
COPY . .
RUN go build -o main main.go


# Running stage
FROM alpine:3.20

WORKDIR /app
COPY --from=builder /app/main .
COPY ./app.env .
COPY ./serviceAccountKey.json .

# Install netcat for the wait-for-it functionality
RUN apk add --no-cache netcat-openbsd

COPY ./entrypoint.sh .

EXPOSE 8080

ENTRYPOINT [ "/app/entrypoint.sh" ]
CMD ["/app/main"]

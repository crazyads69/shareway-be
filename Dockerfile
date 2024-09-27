FROM golang:1.21.3-alpine3.18 AS builder
WORKDIR /app
COPY . .
RUN go build -o main main.go


# Running stage
FROM alpine:3.18

WORKDIR /app
COPY --from=builder /app/main .
COPY ./app.env .
COPY ./entrypoint.sh .

EXPOSE 8080

ENTRYPOINT [ "/app/entrypoint.sh" ]
CMD ["/app/main"]

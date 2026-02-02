FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server ./cmd/server

FROM alpine:3.19

WORKDIR /app

COPY --from=builder /app/server /app/server

EXPOSE 3000

ENV SERVER_HOST=0.0.0.0
ENV SERVER_PORT=3000

CMD ["/app/server"]


FROM golang:1.24-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o payaja-service ./wallet-service/main.go

EXPOSE 50051

CMD ["./payaja-service"]
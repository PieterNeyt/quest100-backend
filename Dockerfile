FROM golang:1.25 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /quest100 ./cmd/api/main.go

FROM alpine:3.23
COPY --from=builder /quest100 /quest100
EXPOSE 8080
CMD ["/quest100"]

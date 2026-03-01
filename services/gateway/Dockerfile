FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/gateway ./cmd/gateway

# ---

FROM alpine:3.21

RUN apk add --no-cache ca-certificates

COPY --from=builder /bin/gateway /bin/gateway
COPY --from=builder /app/gen/openapiv2/v1/gateway.swagger.json /etc/swagger.json

EXPOSE 8080

ENTRYPOINT ["/bin/gateway"]

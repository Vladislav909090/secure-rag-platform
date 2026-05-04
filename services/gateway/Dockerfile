FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY . .

WORKDIR /app/services/gateway
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/gateway ./cmd/gateway

# ---

FROM alpine:3.21

RUN apk add --no-cache ca-certificates

COPY --from=builder /bin/gateway /bin/gateway
COPY --from=builder /app/services/gateway/gen/openapiv2/v1/gateway.swagger.json /etc/swagger.json

EXPOSE 8080

ENTRYPOINT ["/bin/gateway"]

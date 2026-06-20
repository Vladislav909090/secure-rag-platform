FROM golang:1.25-alpine AS builder

WORKDIR /repo

COPY api/go.mod ./api/go.mod
COPY services/gateway/go.mod services/gateway/go.sum* ./services/gateway/
RUN cd services/gateway && go mod download

COPY api ./api
COPY services/gateway ./services/gateway

WORKDIR /repo/services/gateway
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/gateway ./cmd/gateway

FROM alpine:3.21

RUN apk add --no-cache ca-certificates

COPY --from=builder /bin/gateway /bin/gateway
COPY --from=builder /repo/api/gen/openapiv2/gateway/v1/gateway.swagger.json /etc/swagger.json

EXPOSE 8080

ENTRYPOINT ["/bin/gateway"]

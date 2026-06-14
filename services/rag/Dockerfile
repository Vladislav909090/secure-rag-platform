FROM golang:1.25-alpine AS builder

WORKDIR /repo

COPY api/go.mod ./api/go.mod
COPY services/rag/go.mod services/rag/go.sum* ./services/rag/
RUN cd services/rag && go mod download

COPY api ./api
COPY services/rag ./services/rag

WORKDIR /repo/services/rag
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/rag ./cmd/rag

# ---

FROM alpine:3.21

RUN apk add --no-cache ca-certificates

COPY --from=builder /bin/rag /bin/rag
COPY --from=builder /repo/api/gen/openapiv2/rag/v1/rag.swagger.json /etc/swagger.json

EXPOSE 8083

ENTRYPOINT ["/bin/rag"]

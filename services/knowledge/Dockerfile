FROM golang:1.25-alpine AS builder

WORKDIR /repo

COPY api/go.mod ./api/go.mod
COPY services/knowledge/go.mod services/knowledge/go.sum* ./services/knowledge/
RUN cd services/knowledge && go mod download

COPY api ./api
COPY services/knowledge ./services/knowledge

WORKDIR /repo/services/knowledge
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/knowledge ./cmd/knowledge

# ---

FROM alpine:3.21

RUN apk add --no-cache ca-certificates

COPY --from=builder /bin/knowledge /bin/knowledge
COPY --from=builder /repo/api/gen/openapiv2/knowledge/v1/knowledge.swagger.json /etc/swagger.json

EXPOSE 8082

ENTRYPOINT ["/bin/knowledge"]

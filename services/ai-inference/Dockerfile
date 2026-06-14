FROM golang:1.25-alpine AS builder

WORKDIR /repo

COPY api/go.mod ./api/go.mod
COPY services/ai-inference/go.mod services/ai-inference/go.sum* ./services/ai-inference/
RUN cd services/ai-inference && go mod download

COPY api ./api
COPY services/ai-inference ./services/ai-inference

WORKDIR /repo/services/ai-inference
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/ai-inference ./cmd/ai-inference

# ---

FROM alpine:3.21

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /bin/ai-inference /bin/ai-inference
RUN mkdir -p /app/config
COPY --from=builder /repo/services/ai-inference/config/models.example.json /app/config/models.example.json
COPY --from=builder /repo/api/gen/openapiv2/aiinference/v1/ai_inference.swagger.json /etc/swagger.json

EXPOSE 8084 9094

ENTRYPOINT ["/bin/ai-inference"]

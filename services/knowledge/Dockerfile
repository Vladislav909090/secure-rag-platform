FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/knowledge ./cmd/knowledge

# ---

FROM alpine:3.21

RUN apk add --no-cache ca-certificates

COPY --from=builder /bin/knowledge /bin/knowledge
COPY --from=builder /app/gen/openapiv2/v1/knowledge.swagger.json /etc/swagger.json

EXPOSE 8082

ENTRYPOINT ["/bin/knowledge"]

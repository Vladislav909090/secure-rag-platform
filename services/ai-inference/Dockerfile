FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/ai-inference ./cmd/ai-inference

# ---

FROM alpine:3.21

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /bin/ai-inference /bin/ai-inference
COPY --from=builder /app/config /app/config

EXPOSE 9094

ENTRYPOINT ["/bin/ai-inference"]

FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/iam ./cmd/iam

# ---

FROM alpine:3.21

RUN apk add --no-cache ca-certificates

COPY --from=builder /bin/iam /bin/iam
COPY --from=builder /app/gen/openapiv2/v1/iam.swagger.json /etc/swagger.json

EXPOSE 8081

ENTRYPOINT ["/bin/iam"]

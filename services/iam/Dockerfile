FROM golang:1.25-alpine AS builder

WORKDIR /repo

COPY api/go.mod ./api/go.mod
COPY services/iam/go.mod services/iam/go.sum* ./services/iam/
RUN cd services/iam && go mod download

COPY api ./api
COPY services/iam ./services/iam

WORKDIR /repo/services/iam
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/iam ./cmd/iam

# ---

FROM alpine:3.21

RUN apk add --no-cache ca-certificates

COPY --from=builder /bin/iam /bin/iam
COPY --from=builder /repo/api/gen/openapiv2/iam/v1/iam.swagger.json /etc/swagger.json

EXPOSE 8081

ENTRYPOINT ["/bin/iam"]

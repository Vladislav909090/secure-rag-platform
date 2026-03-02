.PHONY: lint lint\:gateway lint\:iam lint\:knowledge lint\:rag \
	test test\:gateway test\:iam test\:knowledge test\:rag \
	build build\:gateway build\:iam build\:knowledge build\:rag \
	proto\:tools proto\:deps proto\:gen \
	proto\:gen\:gateway proto\:gen\:iam proto\:gen\:knowledge proto\:gen\:rag \
	grpc\:stubs grpc\:stubs\:gateway grpc\:stubs\:iam grpc\:stubs\:knowledge grpc\:stubs\:rag \
	compose\:up compose\:down

GOOGLEAPIS_RAW = https://raw.githubusercontent.com/googleapis/googleapis/master
PROTO_INC      = -I third_party
PROTOC         = protoc

# ── Platform-specific helpers ─────────────────────────

ifeq ($(OS),Windows_NT)

MKDIRP = powershell -NoProfile -Command "New-Item -ItemType Directory -Force -Path '$1' | Out-Null"

# Авто-поиск protoc на Windows:
# 1) в PATH
# 2) winget-путь
# 3) локальная распаковка в %LOCALAPPDATA%\protoc\bin
PROTOC = $(shell powershell -NoProfile -Command "$$cmd=Get-Command protoc -ErrorAction SilentlyContinue; if($$cmd){$$cmd.Source}else{$$p1=Join-Path $$env:LOCALAPPDATA 'Microsoft\\WinGet\\Packages\\Google.Protobuf_Microsoft.Winget.Source_8wekyb3d8bbwe\\bin\\protoc.exe'; $$p2=Join-Path $$env:LOCALAPPDATA 'protoc\\bin\\protoc.exe'; if(Test-Path $$p1){$$p1}elseif(Test-Path $$p2){$$p2}else{'protoc'}}")

PROTO_DEPS_CMD = powershell -NoProfile -Command "New-Item -ItemType Directory -Force third_party/google/api | Out-Null; if(-not(Test-Path third_party/google/api/annotations.proto)){Invoke-WebRequest '$(GOOGLEAPIS_RAW)/google/api/annotations.proto' -OutFile third_party/google/api/annotations.proto -UseBasicParsing; Invoke-WebRequest '$(GOOGLEAPIS_RAW)/google/api/http.proto' -OutFile third_party/google/api/http.proto -UseBasicParsing; Write-Host '==> Downloaded google/api protos'}else{Write-Host '==> google/api protos already present'}"

else

MKDIRP = mkdir -p $1

PROTO_DEPS_CMD = mkdir -p third_party/google/api && \
  if [ ! -f third_party/google/api/annotations.proto ]; then \
    curl -sSL $(GOOGLEAPIS_RAW)/google/api/annotations.proto -o third_party/google/api/annotations.proto && \
    curl -sSL $(GOOGLEAPIS_RAW)/google/api/http.proto -o third_party/google/api/http.proto && \
    echo "==> Downloaded google/api protos"; \
  else echo "==> google/api protos already present"; fi

endif

# ── Lint ──────────────────────────────────────────────

lint\:gateway:
	cd services/gateway && golangci-lint run ./...

lint\:iam:
	cd services/iam && golangci-lint run ./...

lint\:knowledge:
	cd services/knowledge && golangci-lint run ./...

lint\:rag:
	cd services/rag && golangci-lint run ./...

lint: lint\:gateway lint\:iam lint\:knowledge lint\:rag

# ── Test ──────────────────────────────────────────────

test\:gateway:
	cd services/gateway && go test ./...

test\:iam:
	cd services/iam && go test ./...

test\:knowledge:
	cd services/knowledge && go test ./...

test\:rag:
	cd services/rag && go test ./...

test: test\:gateway test\:iam test\:knowledge test\:rag

# ── Build ─────────────────────────────────────────────

build\:gateway:
	cd services/gateway && go build ./cmd/gateway

build\:iam:
	cd services/iam && go build ./cmd/iam

build\:knowledge:
	cd services/knowledge && go build ./cmd/knowledge

build\:rag:
	cd services/rag && go build ./cmd/rag

build: build\:gateway build\:iam build\:knowledge build\:rag

# ── Proto: install tools ─────────────────────────────

proto\:tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.11
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.27.2
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.27.2

# ── Proto: fetch google/api deps ─────────────────────

proto\:deps:
	@$(PROTO_DEPS_CMD)

# ── Proto: per-service generation ─────────────────────

proto\:gen\:gateway: proto\:deps
	@$(call MKDIRP,services/gateway/gen/v1)
	@$(call MKDIRP,services/gateway/gen/openapiv2)
	"$(PROTOC)" $(PROTO_INC) -I services/gateway/api \
		--go_out=services/gateway/gen --go_opt=paths=source_relative \
		--go-grpc_out=services/gateway/gen --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=services/gateway/gen --grpc-gateway_opt=paths=source_relative,generate_unbound_methods=true \
		--openapiv2_out=services/gateway/gen/openapiv2 \
		services/gateway/api/v1/gateway.proto

proto\:gen\:iam: proto\:deps
	@$(call MKDIRP,services/iam/gen/v1)
	@$(call MKDIRP,services/iam/gen/openapiv2)
	"$(PROTOC)" $(PROTO_INC) -I services/iam/api \
		--go_out=services/iam/gen --go_opt=paths=source_relative \
		--go-grpc_out=services/iam/gen --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=services/iam/gen --grpc-gateway_opt=paths=source_relative,generate_unbound_methods=true \
		--openapiv2_out=services/iam/gen/openapiv2 \
		services/iam/api/v1/iam.proto

proto\:gen\:knowledge: proto\:deps
	@$(call MKDIRP,services/knowledge/gen/v1)
	@$(call MKDIRP,services/knowledge/gen/openapiv2)
	"$(PROTOC)" $(PROTO_INC) -I services/knowledge/api \
		--go_out=services/knowledge/gen --go_opt=paths=source_relative \
		--go-grpc_out=services/knowledge/gen --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=services/knowledge/gen --grpc-gateway_opt=paths=source_relative,generate_unbound_methods=true \
		--openapiv2_out=services/knowledge/gen/openapiv2 \
		services/knowledge/api/v1/knowledge.proto

proto\:gen\:rag: proto\:deps
	@$(call MKDIRP,services/rag/gen/v1)
	@$(call MKDIRP,services/rag/gen/openapiv2)
	"$(PROTOC)" $(PROTO_INC) -I services/rag/api \
		--go_out=services/rag/gen --go_opt=paths=source_relative \
		--go-grpc_out=services/rag/gen --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=services/rag/gen --grpc-gateway_opt=paths=source_relative,generate_unbound_methods=true \
		--openapiv2_out=services/rag/gen/openapiv2 \
		services/rag/api/v1/rag.proto

proto\:gen: proto\:gen\:gateway proto\:gen\:iam proto\:gen\:knowledge proto\:gen\:rag

# ── gRPC: generate server stubs (transport/grpc) ─────

grpc\:stubs\:gateway:
	go run ./tools/grpcstubgen --service services/gateway --out internal/transport/grpc

grpc\:stubs\:iam:
	go run ./tools/grpcstubgen --service services/iam --out internal/transport/grpc

grpc\:stubs\:knowledge:
	go run ./tools/grpcstubgen --service services/knowledge --out internal/transport/grpc

grpc\:stubs\:rag:
	go run ./tools/grpcstubgen --service services/rag --out internal/transport/grpc

grpc\:stubs: grpc\:stubs\:gateway grpc\:stubs\:iam grpc\:stubs\:knowledge grpc\:stubs\:rag


# ── Docker Compose ───────────────────────────────────

compose\:up:
	docker compose -f deploy/compose/docker-compose.yml up -d --build

compose\:down:
	docker compose -f deploy/compose/docker-compose.yml down

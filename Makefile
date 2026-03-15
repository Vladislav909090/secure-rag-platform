.PHONY: lint lint\:gateway lint\:iam lint\:knowledge lint\:rag \
	test test\:gateway test\:iam test\:knowledge test\:rag \
	build build\:gateway build\:iam build\:knowledge build\:rag \
	proto\:tools proto\:deps proto\:gen \
	proto\:gen\:gateway proto\:gen\:iam proto\:gen\:knowledge proto\:gen\:rag \
	grpc\:stubs grpc\:stubs\:gateway grpc\:stubs\:iam grpc\:stubs\:knowledge grpc\:stubs\:rag \
	migrate\:status migrate\:status\:gateway migrate\:status\:iam migrate\:status\:knowledge migrate\:status\:rag \
	migrate\:up migrate\:up\:gateway migrate\:up\:iam migrate\:up\:knowledge migrate\:up\:rag \
	migrate\:down migrate\:down\:gateway migrate\:down\:iam migrate\:down\:knowledge migrate\:down\:rag \
	migrate\:create\:gateway migrate\:create\:iam migrate\:create\:knowledge migrate\:create\:rag \
	compose\:up compose\:down

COMPOSE_DEV ?= 0
COMPOSE_FILES = -f deploy/compose/docker-compose.yml

ifeq ($(DEV),1)
COMPOSE_DEV = 1
endif

ifeq ($(COMPOSE_DEV),1)
COMPOSE_FILES += -f deploy/compose/compose.dev.yml
endif

GOOGLEAPIS_RAW = https://raw.githubusercontent.com/googleapis/googleapis/master
PROTO_INC      = -I third_party
PROTOC         = protoc
GOOSE          = go run github.com/pressly/goose/v3/cmd/goose@v3.24.3
MIGRATION_NAME ?= new_migration

GATEWAY_MIGRATIONS_DIR   = services/gateway/migrations
IAM_MIGRATIONS_DIR       = services/iam/migrations
KNOWLEDGE_MIGRATIONS_DIR = services/knowledge/migrations
RAG_MIGRATIONS_DIR       = services/rag/migrations

GATEWAY_DB_DSN   ?= postgres://gateway:gateway@localhost:5432/gateway?sslmode=disable
IAM_DB_DSN       ?= postgres://iam:iam@localhost:5433/iam?sslmode=disable
KNOWLEDGE_DB_DSN ?= postgres://knowledge:knowledge@localhost:5434/knowledge?sslmode=disable
RAG_DB_DSN       ?= postgres://rag:rag@localhost:5435/rag?sslmode=disable

# ── Platform-specific helpers ─────────────────────────

ifeq ($(OS),Windows_NT)

MKDIRP = powershell -NoProfile -Command "New-Item -ItemType Directory -Force -Path '$1' | Out-Null"

# Авто-поиск protoc на Windows:
# 1) в PATH
# 2) winget-путь
# 3) локальная распаковка в %LOCALAPPDATA%\protoc\bin
PROTOC = $(shell powershell -NoProfile -Command "$$cmd=Get-Command protoc -ErrorAction SilentlyContinue; if($$cmd){$$cmd.Source}else{$$p1=Join-Path $$env:LOCALAPPDATA 'Microsoft\\WinGet\\Packages\\Google.Protobuf_Microsoft.Winget.Source_8wekyb3d8bbwe\\bin\\protoc.exe'; $$p2=Join-Path $$env:LOCALAPPDATA 'protoc\\bin\\protoc.exe'; if(Test-Path $$p1){$$p1}elseif(Test-Path $$p2){$$p2}else{'protoc'}}")

PROTO_DEPS_CMD = powershell -NoProfile -Command "New-Item -ItemType Directory -Force third_party/google/api | Out-Null; $$updated=$$false; if(-not(Test-Path third_party/google/api/annotations.proto)){Invoke-WebRequest '$(GOOGLEAPIS_RAW)/google/api/annotations.proto' -OutFile third_party/google/api/annotations.proto -UseBasicParsing; $$updated=$$true}; if(-not(Test-Path third_party/google/api/http.proto)){Invoke-WebRequest '$(GOOGLEAPIS_RAW)/google/api/http.proto' -OutFile third_party/google/api/http.proto -UseBasicParsing; $$updated=$$true}; if(-not(Test-Path third_party/google/api/httpbody.proto)){Invoke-WebRequest '$(GOOGLEAPIS_RAW)/google/api/httpbody.proto' -OutFile third_party/google/api/httpbody.proto -UseBasicParsing; $$updated=$$true}; if($$updated){Write-Host '==> Downloaded missing google/api protos'}else{Write-Host '==> google/api protos already present'}"

else

MKDIRP = mkdir -p $1

PROTO_DEPS_CMD = mkdir -p third_party/google/api && \
	changed=0; \
	if [ ! -f third_party/google/api/annotations.proto ]; then \
		curl -sSL $(GOOGLEAPIS_RAW)/google/api/annotations.proto -o third_party/google/api/annotations.proto && changed=1; \
	fi; \
	if [ ! -f third_party/google/api/http.proto ]; then \
		curl -sSL $(GOOGLEAPIS_RAW)/google/api/http.proto -o third_party/google/api/http.proto && changed=1; \
	fi; \
	if [ ! -f third_party/google/api/httpbody.proto ]; then \
		curl -sSL $(GOOGLEAPIS_RAW)/google/api/httpbody.proto -o third_party/google/api/httpbody.proto && changed=1; \
	fi; \
	if [ $$changed -eq 1 ]; then echo "==> Downloaded missing google/api protos"; else echo "==> google/api protos already present"; fi

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


# ── Database migrations (goose) ─────────────────────

migrate\:status\:gateway:
	@$(GOOSE) -dir $(GATEWAY_MIGRATIONS_DIR) postgres "$(GATEWAY_DB_DSN)" status

migrate\:status\:iam:
	@$(GOOSE) -dir $(IAM_MIGRATIONS_DIR) postgres "$(IAM_DB_DSN)" status

migrate\:status\:knowledge:
	@$(GOOSE) -dir $(KNOWLEDGE_MIGRATIONS_DIR) postgres "$(KNOWLEDGE_DB_DSN)" status

migrate\:status\:rag:
	@$(GOOSE) -dir $(RAG_MIGRATIONS_DIR) postgres "$(RAG_DB_DSN)" status

migrate\:status: migrate\:status\:gateway migrate\:status\:iam migrate\:status\:knowledge migrate\:status\:rag

migrate\:up\:gateway:
	@$(GOOSE) -dir $(GATEWAY_MIGRATIONS_DIR) postgres "$(GATEWAY_DB_DSN)" up

migrate\:up\:iam:
	@$(GOOSE) -dir $(IAM_MIGRATIONS_DIR) postgres "$(IAM_DB_DSN)" up

migrate\:up\:knowledge:
	@$(GOOSE) -dir $(KNOWLEDGE_MIGRATIONS_DIR) postgres "$(KNOWLEDGE_DB_DSN)" up

migrate\:up\:rag:
	@$(GOOSE) -dir $(RAG_MIGRATIONS_DIR) postgres "$(RAG_DB_DSN)" up

migrate\:up: migrate\:up\:gateway migrate\:up\:iam migrate\:up\:knowledge migrate\:up\:rag

migrate\:down\:gateway:
	@$(GOOSE) -dir $(GATEWAY_MIGRATIONS_DIR) postgres "$(GATEWAY_DB_DSN)" down

migrate\:down\:iam:
	@$(GOOSE) -dir $(IAM_MIGRATIONS_DIR) postgres "$(IAM_DB_DSN)" down

migrate\:down\:knowledge:
	@$(GOOSE) -dir $(KNOWLEDGE_MIGRATIONS_DIR) postgres "$(KNOWLEDGE_DB_DSN)" down

migrate\:down\:rag:
	@$(GOOSE) -dir $(RAG_MIGRATIONS_DIR) postgres "$(RAG_DB_DSN)" down

migrate\:down: migrate\:down\:gateway migrate\:down\:iam migrate\:down\:knowledge migrate\:down\:rag

migrate\:create\:gateway:
	@$(GOOSE) -dir $(GATEWAY_MIGRATIONS_DIR) create $(MIGRATION_NAME) sql

migrate\:create\:iam:
	@$(GOOSE) -dir $(IAM_MIGRATIONS_DIR) create $(MIGRATION_NAME) sql

migrate\:create\:knowledge:
	@$(GOOSE) -dir $(KNOWLEDGE_MIGRATIONS_DIR) create $(MIGRATION_NAME) sql

migrate\:create\:rag:
	@$(GOOSE) -dir $(RAG_MIGRATIONS_DIR) create $(MIGRATION_NAME) sql


# ── Docker Compose ───────────────────────────────────

compose\:up:
	docker compose $(COMPOSE_FILES) up -d --build

compose\:down:
	docker compose -f deploy/compose/docker-compose.yml -f deploy/compose/compose.dev.yml down

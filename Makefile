.PHONY: lint lint\:gateway lint\:iam lint\:knowledge lint\:rag lint\:ai-inference \
	test test\:gateway test\:iam test\:knowledge test\:rag test\:ai-inference \
	build build\:gateway build\:iam build\:knowledge build\:rag build\:ai-inference \
	api\:sync api\:gen api\:clean \
	proto\:tools proto\:deps proto\:gen \
	proto\:gen\:gateway proto\:gen\:iam proto\:gen\:knowledge proto\:gen\:rag proto\:gen\:ai-inference \
	grpc\:stubs grpc\:stubs\:gateway grpc\:stubs\:iam grpc\:stubs\:knowledge grpc\:stubs\:rag grpc\:stubs\:ai-inference \
	migrate\:validate migrate\:validate\:iam migrate\:validate\:knowledge migrate\:validate\:rag \
	migrate\:status migrate\:status\:iam migrate\:status\:knowledge migrate\:status\:rag \
	migrate\:up migrate\:up\:iam migrate\:up\:knowledge migrate\:up\:rag \
	migrate\:down migrate\:down\:iam migrate\:down\:knowledge migrate\:down\:rag \
	migrate\:create\:iam migrate\:create\:knowledge migrate\:create\:rag \
	compose\:up compose\:recreate compose\:down

COMPOSE_DEV ?= 0
COMPOSE_FILES = -f deploy/compose/docker-compose.yml

ifeq ($(DEV),1)
COMPOSE_DEV = 1
endif

ifeq ($(COMPOSE_DEV),1)
COMPOSE_FILES += -f deploy/compose/compose.dev.yml
endif

GOOGLEAPIS_REF ?= 1526e545e9d26f23b9c5d0f04af17297def8d045
GOOGLEAPIS_RAW = https://raw.githubusercontent.com/googleapis/googleapis/$(GOOGLEAPIS_REF)
PROTO_INC      = -I third_party
PROTOC         = protoc
GOOSE          = go run github.com/pressly/goose/v3/cmd/goose@v3.24.3
MIGRATION_NAME ?= new_migration

API_PROTO_FILES = \
	gateway/v1/gateway.proto \
	iam/v1/iam.proto \
	knowledge/v1/knowledge.proto \
	rag/v1/rag.proto \
	aiinference/v1/ai_inference.proto

IAM_MIGRATIONS_DIR       = services/iam/migrations
KNOWLEDGE_MIGRATIONS_DIR = services/knowledge/migrations
RAG_MIGRATIONS_DIR       = services/rag/migrations

IAM_DB_DSN       ?= postgres://iam:iam@localhost:5433/iam?sslmode=disable
KNOWLEDGE_DB_DSN ?= postgres://knowledge:knowledge@localhost:5434/knowledge?sslmode=disable
RAG_DB_DSN       ?= postgres://rag:rag@localhost:5435/rag?sslmode=disable

# ── Платформенные вспомогательные команды ─────────────

ifeq ($(OS),Windows_NT)

MKDIRP = powershell -NoProfile -Command "New-Item -ItemType Directory -Force -Path '$1' | Out-Null"

# Авто-поиск protoc на Windows:
# 1) в PATH
# 2) winget-путь
# 3) локальная распаковка в %LOCALAPPDATA%\protoc\bin
PROTOC = $(shell powershell -NoProfile -Command "$$cmd=Get-Command protoc -ErrorAction SilentlyContinue; if($$cmd){$$cmd.Source}else{$$p1=Join-Path $$env:LOCALAPPDATA 'Microsoft\\WinGet\\Packages\\Google.Protobuf_Microsoft.Winget.Source_8wekyb3d8bbwe\\bin\\protoc.exe'; $$p2=Join-Path $$env:LOCALAPPDATA 'protoc\\bin\\protoc.exe'; if(Test-Path -LiteralPath $$p1 -ErrorAction SilentlyContinue){$$p1}elseif(Test-Path -LiteralPath $$p2 -ErrorAction SilentlyContinue){$$p2}else{'protoc'}}")

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

lint\:ai-inference:
	cd services/ai-inference && golangci-lint run ./...

lint: lint\:gateway lint\:iam lint\:knowledge lint\:rag lint\:ai-inference

# ── Test ──────────────────────────────────────────────

test\:gateway:
	cd services/gateway && go test ./...

test\:iam:
	cd services/iam && go test ./...

test\:knowledge:
	cd services/knowledge && go test ./...

test\:rag:
	cd services/rag && go test ./...

test\:ai-inference:
	cd services/ai-inference && go test ./...

test: test\:gateway test\:iam test\:knowledge test\:rag test\:ai-inference

# ── Build ─────────────────────────────────────────────

build\:gateway:
	cd services/gateway && go build ./cmd/gateway

build\:iam:
	cd services/iam && go build ./cmd/iam

build\:knowledge:
	cd services/knowledge && go build ./cmd/knowledge

build\:rag:
	cd services/rag && go build ./cmd/rag

build\:ai-inference:
	cd services/ai-inference && go build ./cmd/ai-inference

build: build\:gateway build\:iam build\:knowledge build\:rag build\:ai-inference

# ── Proto: установка инструментов ─────────────────────

proto\:tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.11
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.27.2
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.27.2

# ── Proto: загрузка зависимостей google/api ───────────

proto\:deps:
	@$(PROTO_DEPS_CMD)

# ── API: генерация общих контрактов ───────────────────

api\:sync:
ifeq ($(OS),Windows_NT)
	powershell -NoProfile -Command "New-Item -ItemType Directory -Force api/proto/gateway/v1,api/proto/iam/v1,api/proto/knowledge/v1,api/proto/rag/v1,api/proto/aiinference/v1 | Out-Null; Copy-Item -Force services/gateway/api/v1/gateway.proto api/proto/gateway/v1/gateway.proto; Copy-Item -Force services/iam/api/v1/iam.proto api/proto/iam/v1/iam.proto; Copy-Item -Force services/knowledge/api/v1/knowledge.proto api/proto/knowledge/v1/knowledge.proto; Copy-Item -Force services/rag/api/v1/rag.proto api/proto/rag/v1/rag.proto; Copy-Item -Force services/ai-inference/api/v1/ai_inference.proto api/proto/aiinference/v1/ai_inference.proto"
else
	mkdir -p api/proto/gateway/v1 api/proto/iam/v1 api/proto/knowledge/v1 api/proto/rag/v1 api/proto/aiinference/v1
	cp services/gateway/api/v1/gateway.proto api/proto/gateway/v1/gateway.proto
	cp services/iam/api/v1/iam.proto api/proto/iam/v1/iam.proto
	cp services/knowledge/api/v1/knowledge.proto api/proto/knowledge/v1/knowledge.proto
	cp services/rag/api/v1/rag.proto api/proto/rag/v1/rag.proto
	cp services/ai-inference/api/v1/ai_inference.proto api/proto/aiinference/v1/ai_inference.proto
endif

api\:gen: api\:sync proto\:deps
ifeq ($(OS),Windows_NT)
	powershell -NoProfile -Command "New-Item -ItemType Directory -Force api/gen/go,api/gen/openapiv2 | Out-Null; & '$(PROTOC)' -I third_party -I api/proto --go_out=api/gen/go --go_opt=paths=source_relative --go-grpc_out=api/gen/go --go-grpc_opt=paths=source_relative --grpc-gateway_out=api/gen/go --grpc-gateway_opt=paths=source_relative --openapiv2_out=api/gen/openapiv2 $(API_PROTO_FILES)"
else
	mkdir -p api/gen/go api/gen/openapiv2
	$(PROTOC) -I third_party -I api/proto \
		--go_out=api/gen/go --go_opt=paths=source_relative \
		--go-grpc_out=api/gen/go --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=api/gen/go --grpc-gateway_opt=paths=source_relative \
		--openapiv2_out=api/gen/openapiv2 \
		$(API_PROTO_FILES)
endif

api\:clean:
ifeq ($(OS),Windows_NT)
	powershell -NoProfile -Command "Remove-Item -Recurse -Force api/gen/go,api/gen/openapiv2 -ErrorAction SilentlyContinue"
else
	rm -rf api/gen/go api/gen/openapiv2
endif

proto\:gen: api\:gen

# Совместимые алиасы: контракты теперь генерируются единым api-модулем.
proto\:gen\:gateway proto\:gen\:iam proto\:gen\:knowledge proto\:gen\:rag proto\:gen\:ai-inference: api\:gen

# ── gRPC: генерация server stubs (transport/grpc) ─────

grpc\:stubs\:gateway:
ifeq ($(OS),Windows_NT)
	powershell -NoProfile -Command "$$env:GOWORK='off'; $$env:GOCACHE=Join-Path $$env:TEMP 'secure-rag-platform-go-build-cache'; go -C tools/grpcstubgen run . --service ../../services/gateway --gen ../../api/gen/go/gateway --pb-import secure-rag-platform/api/gen/go/gateway/v1 --out internal/transport/grpc"
else
	GOWORK=off go -C tools/grpcstubgen run . --service ../../services/gateway --gen ../../api/gen/go/gateway --pb-import secure-rag-platform/api/gen/go/gateway/v1 --out internal/transport/grpc
endif

grpc\:stubs\:iam:
ifeq ($(OS),Windows_NT)
	powershell -NoProfile -Command "$$env:GOWORK='off'; $$env:GOCACHE=Join-Path $$env:TEMP 'secure-rag-platform-go-build-cache'; go -C tools/grpcstubgen run . --service ../../services/iam --gen ../../api/gen/go/iam --pb-import secure-rag-platform/api/gen/go/iam/v1 --out internal/transport/grpc"
else
	GOWORK=off go -C tools/grpcstubgen run . --service ../../services/iam --gen ../../api/gen/go/iam --pb-import secure-rag-platform/api/gen/go/iam/v1 --out internal/transport/grpc
endif

grpc\:stubs\:knowledge:
ifeq ($(OS),Windows_NT)
	powershell -NoProfile -Command "$$env:GOWORK='off'; $$env:GOCACHE=Join-Path $$env:TEMP 'secure-rag-platform-go-build-cache'; go -C tools/grpcstubgen run . --service ../../services/knowledge --gen ../../api/gen/go/knowledge --pb-import secure-rag-platform/api/gen/go/knowledge/v1 --out internal/transport/grpc"
else
	GOWORK=off go -C tools/grpcstubgen run . --service ../../services/knowledge --gen ../../api/gen/go/knowledge --pb-import secure-rag-platform/api/gen/go/knowledge/v1 --out internal/transport/grpc
endif

grpc\:stubs\:rag:
ifeq ($(OS),Windows_NT)
	powershell -NoProfile -Command "$$env:GOWORK='off'; $$env:GOCACHE=Join-Path $$env:TEMP 'secure-rag-platform-go-build-cache'; go -C tools/grpcstubgen run . --service ../../services/rag --gen ../../api/gen/go/rag --pb-import secure-rag-platform/api/gen/go/rag/v1 --out internal/transport/grpc"
else
	GOWORK=off go -C tools/grpcstubgen run . --service ../../services/rag --gen ../../api/gen/go/rag --pb-import secure-rag-platform/api/gen/go/rag/v1 --out internal/transport/grpc
endif

grpc\:stubs\:ai-inference:
ifeq ($(OS),Windows_NT)
	powershell -NoProfile -Command "$$env:GOWORK='off'; $$env:GOCACHE=Join-Path $$env:TEMP 'secure-rag-platform-go-build-cache'; go -C tools/grpcstubgen run . --service ../../services/ai-inference --gen ../../api/gen/go/aiinference --pb-import secure-rag-platform/api/gen/go/aiinference/v1 --out internal/transport/grpc"
else
	GOWORK=off go -C tools/grpcstubgen run . --service ../../services/ai-inference --gen ../../api/gen/go/aiinference --pb-import secure-rag-platform/api/gen/go/aiinference/v1 --out internal/transport/grpc
endif

grpc\:stubs: grpc\:stubs\:gateway grpc\:stubs\:iam grpc\:stubs\:knowledge grpc\:stubs\:rag grpc\:stubs\:ai-inference


# ── Миграции баз данных (goose) ───────────────────────

migrate\:validate\:iam:
	@$(GOOSE) -dir $(IAM_MIGRATIONS_DIR) validate

migrate\:validate\:knowledge:
	@$(GOOSE) -dir $(KNOWLEDGE_MIGRATIONS_DIR) validate

migrate\:validate\:rag:
	@$(GOOSE) -dir $(RAG_MIGRATIONS_DIR) validate

migrate\:validate: migrate\:validate\:iam migrate\:validate\:knowledge migrate\:validate\:rag

migrate\:status\:iam:
	@$(GOOSE) -dir $(IAM_MIGRATIONS_DIR) postgres "$(IAM_DB_DSN)" status

migrate\:status\:knowledge:
	@$(GOOSE) -dir $(KNOWLEDGE_MIGRATIONS_DIR) postgres "$(KNOWLEDGE_DB_DSN)" status

migrate\:status\:rag:
	@$(GOOSE) -dir $(RAG_MIGRATIONS_DIR) postgres "$(RAG_DB_DSN)" status

migrate\:status: migrate\:status\:iam migrate\:status\:knowledge migrate\:status\:rag

migrate\:up\:iam:
	@$(GOOSE) -dir $(IAM_MIGRATIONS_DIR) postgres "$(IAM_DB_DSN)" up

migrate\:up\:knowledge:
	@$(GOOSE) -dir $(KNOWLEDGE_MIGRATIONS_DIR) postgres "$(KNOWLEDGE_DB_DSN)" up

migrate\:up\:rag:
	@$(GOOSE) -dir $(RAG_MIGRATIONS_DIR) postgres "$(RAG_DB_DSN)" up

migrate\:up: migrate\:up\:iam migrate\:up\:knowledge migrate\:up\:rag

migrate\:down\:iam:
	@$(GOOSE) -dir $(IAM_MIGRATIONS_DIR) postgres "$(IAM_DB_DSN)" down

migrate\:down\:knowledge:
	@$(GOOSE) -dir $(KNOWLEDGE_MIGRATIONS_DIR) postgres "$(KNOWLEDGE_DB_DSN)" down

migrate\:down\:rag:
	@$(GOOSE) -dir $(RAG_MIGRATIONS_DIR) postgres "$(RAG_DB_DSN)" down

migrate\:down: migrate\:down\:iam migrate\:down\:knowledge migrate\:down\:rag

# ── Docker Compose ───────────────────────────────────

compose\:up:
	docker compose $(COMPOSE_FILES) up -d --no-recreate

compose\:recreate:
	docker compose $(COMPOSE_FILES) up -d --build --force-recreate --remove-orphans

compose\:down:
	docker compose -f deploy/compose/docker-compose.yml -f deploy/compose/compose.dev.yml down

.PHONY: lint test build docker\:build \
	proto\:tools proto\:deps proto\:gen proto\:clean \
	api\:sync api\:gen api\:clean grpc\:stubs \
	migrate\:validate migrate\:status migrate\:up migrate\:down migrate\:create \
	compose\:up compose\:recreate compose\:down compose\:config

SERVICE        = rag
BINARY         = rag
IMAGE         ?= secure-rag-platform/$(SERVICE):local
COMPOSE_FILE  ?= compose.yaml -f compose.local.yaml
COMPOSE_PROJECT ?= secure-rag-$(SERVICE)

GOOGLEAPIS_REF ?= 1526e545e9d26f23b9c5d0f04af17297def8d045
GOOGLEAPIS_RAW = https://raw.githubusercontent.com/googleapis/googleapis/$(GOOGLEAPIS_REF)
PROTOC         = protoc
GOOSE          = go run github.com/pressly/goose/v3/cmd/goose@v3.24.3
DB_DSN        ?= postgres://rag:rag@localhost:5435/rag?sslmode=disable
MIGRATION_NAME ?= new_migration

LOCAL_PROTO_FILE = v1/rag.proto
ROOT_PROTO_DIR   = ../../api/proto/rag/v1
ROOT_PROTO_FILE  = $(ROOT_PROTO_DIR)/rag.proto
API_GEN_DIR      = rag
PB_IMPORT        = secure-rag-platform/api/gen/go/rag/v1

API_PROTO_FILES = \
	gateway/v1/gateway.proto \
	iam/v1/iam.proto \
	knowledge/v1/knowledge.proto \
	rag/v1/rag.proto \
	aiinference/v1/ai_inference.proto

ifeq ($(OS),Windows_NT)
PROTOC = $(shell powershell -NoProfile -Command "$$cmd=Get-Command protoc -ErrorAction SilentlyContinue; if($$cmd){$$cmd.Source}else{$$p1=Join-Path $$env:LOCALAPPDATA 'Microsoft\\WinGet\\Packages\\Google.Protobuf_Microsoft.Winget.Source_8wekyb3d8bbwe\\bin\\protoc.exe'; $$p2=Join-Path $$env:LOCALAPPDATA 'protoc\\bin\\protoc.exe'; if(Test-Path -LiteralPath $$p1 -ErrorAction SilentlyContinue){$$p1}elseif(Test-Path -LiteralPath $$p2 -ErrorAction SilentlyContinue){$$p2}else{'protoc'}}")
PROTO_DEPS_CMD = powershell -NoProfile -Command "New-Item -ItemType Directory -Force third_party/google/api | Out-Null; $$updated=$$false; if(-not(Test-Path third_party/google/api/annotations.proto)){Invoke-WebRequest '$(GOOGLEAPIS_RAW)/google/api/annotations.proto' -OutFile third_party/google/api/annotations.proto -UseBasicParsing; $$updated=$$true}; if(-not(Test-Path third_party/google/api/http.proto)){Invoke-WebRequest '$(GOOGLEAPIS_RAW)/google/api/http.proto' -OutFile third_party/google/api/http.proto -UseBasicParsing; $$updated=$$true}; if(-not(Test-Path third_party/google/api/httpbody.proto)){Invoke-WebRequest '$(GOOGLEAPIS_RAW)/google/api/httpbody.proto' -OutFile third_party/google/api/httpbody.proto -UseBasicParsing; $$updated=$$true}; if($$updated){Write-Host '==> Downloaded missing google/api protos'}else{Write-Host '==> google/api protos already present'}"
else
PROTO_DEPS_CMD = mkdir -p third_party/google/api && \
	changed=0; \
	if [ ! -f third_party/google/api/annotations.proto ]; then curl -sSL $(GOOGLEAPIS_RAW)/google/api/annotations.proto -o third_party/google/api/annotations.proto && changed=1; fi; \
	if [ ! -f third_party/google/api/http.proto ]; then curl -sSL $(GOOGLEAPIS_RAW)/google/api/http.proto -o third_party/google/api/http.proto && changed=1; fi; \
	if [ ! -f third_party/google/api/httpbody.proto ]; then curl -sSL $(GOOGLEAPIS_RAW)/google/api/httpbody.proto -o third_party/google/api/httpbody.proto && changed=1; fi; \
	if [ $$changed -eq 1 ]; then echo "==> Downloaded missing google/api protos"; else echo "==> google/api protos already present"; fi
endif

lint:
	golangci-lint run ./...

test:
	go test ./...

build:
	go build ./cmd/$(BINARY)

docker\:build:
	docker build -f Dockerfile -t $(IMAGE) ../..

proto\:tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.11
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.27.2
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.27.2

proto\:deps:
	@$(PROTO_DEPS_CMD)

proto\:gen: proto\:deps
ifeq ($(OS),Windows_NT)
	powershell -NoProfile -Command "New-Item -ItemType Directory -Force gen,gen/openapiv2 | Out-Null; & '$(PROTOC)' -I third_party -I api --go_out=gen --go_opt=paths=source_relative --go-grpc_out=gen --go-grpc_opt=paths=source_relative --grpc-gateway_out=gen --grpc-gateway_opt=paths=source_relative --openapiv2_out=gen/openapiv2 $(LOCAL_PROTO_FILE)"
else
	mkdir -p gen gen/openapiv2
	$(PROTOC) -I third_party -I api \
		--go_out=gen --go_opt=paths=source_relative \
		--go-grpc_out=gen --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=gen --grpc-gateway_opt=paths=source_relative \
		--openapiv2_out=gen/openapiv2 \
		$(LOCAL_PROTO_FILE)
endif

proto\:clean:
ifeq ($(OS),Windows_NT)
	powershell -NoProfile -Command "Remove-Item -Recurse -Force gen/v1,gen/openapiv2 -ErrorAction SilentlyContinue"
else
	rm -rf gen/v1 gen/openapiv2
endif

api\:sync:
ifeq ($(OS),Windows_NT)
	powershell -NoProfile -Command "New-Item -ItemType Directory -Force '$(ROOT_PROTO_DIR)' | Out-Null; Copy-Item -Force api/$(LOCAL_PROTO_FILE) '$(ROOT_PROTO_FILE)'"
else
	mkdir -p $(ROOT_PROTO_DIR)
	cp api/$(LOCAL_PROTO_FILE) $(ROOT_PROTO_FILE)
endif

api\:gen: api\:sync proto\:deps
ifeq ($(OS),Windows_NT)
	powershell -NoProfile -Command "New-Item -ItemType Directory -Force ../../api/gen/go,../../api/gen/openapiv2 | Out-Null; & '$(PROTOC)' -I third_party -I ../../api/proto --go_out=../../api/gen/go --go_opt=paths=source_relative --go-grpc_out=../../api/gen/go --go-grpc_opt=paths=source_relative --grpc-gateway_out=../../api/gen/go --grpc-gateway_opt=paths=source_relative --openapiv2_out=../../api/gen/openapiv2 $(API_PROTO_FILES)"
else
	mkdir -p ../../api/gen/go ../../api/gen/openapiv2
	$(PROTOC) -I third_party -I ../../api/proto \
		--go_out=../../api/gen/go --go_opt=paths=source_relative \
		--go-grpc_out=../../api/gen/go --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=../../api/gen/go --grpc-gateway_opt=paths=source_relative \
		--openapiv2_out=../../api/gen/openapiv2 \
		$(API_PROTO_FILES)
endif

api\:clean:
ifeq ($(OS),Windows_NT)
	powershell -NoProfile -Command "Remove-Item -Recurse -Force ../../api/gen/go,../../api/gen/openapiv2 -ErrorAction SilentlyContinue"
else
	rm -rf ../../api/gen/go ../../api/gen/openapiv2
endif

grpc\:stubs: api\:gen
ifeq ($(OS),Windows_NT)
	powershell -NoProfile -Command "$$env:GOWORK='off'; $$env:GOCACHE=Join-Path $$env:TEMP 'secure-rag-platform-go-build-cache'; go -C tools/grpcstubgen run . --service ../.. --gen ../../../../api/gen/go/$(API_GEN_DIR) --pb-import $(PB_IMPORT) --out internal/transport/grpc"
else
	GOWORK=off go -C tools/grpcstubgen run . --service ../.. --gen ../../../../api/gen/go/$(API_GEN_DIR) --pb-import $(PB_IMPORT) --out internal/transport/grpc
endif

migrate\:validate:
	@$(GOOSE) -dir migrations validate

migrate\:status:
	@$(GOOSE) -dir migrations postgres "$(DB_DSN)" status

migrate\:up:
	@$(GOOSE) -dir migrations postgres "$(DB_DSN)" up

migrate\:down:
	@$(GOOSE) -dir migrations postgres "$(DB_DSN)" down

migrate\:create:
	@$(GOOSE) -dir migrations create $(MIGRATION_NAME) sql

compose\:up:
	docker compose -p $(COMPOSE_PROJECT) -f $(COMPOSE_FILE) up -d --no-recreate

compose\:recreate:
	docker compose -p $(COMPOSE_PROJECT) -f $(COMPOSE_FILE) up -d --build --force-recreate --remove-orphans

compose\:down:
	docker compose -p $(COMPOSE_PROJECT) -f $(COMPOSE_FILE) down

compose\:config:
	docker compose -p $(COMPOSE_PROJECT) -f $(COMPOSE_FILE) config --quiet

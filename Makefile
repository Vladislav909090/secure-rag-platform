.PHONY: lint lint\:gateway lint\:iam lint\:knowledge lint\:rag lint\:ai-inference \
	test test\:gateway test\:iam test\:knowledge test\:rag test\:ai-inference \
	test\:cover test\:cover\:gateway test\:cover\:iam test\:cover\:knowledge test\:cover\:rag test\:cover\:ai-inference coverage \
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
	compose\:up compose\:up\:dev compose\:up\:prod \
	compose\:recreate compose\:recreate\:dev compose\:recreate\:prod \
	compose\:down compose\:down\:dev compose\:down\:prod \
	compose\:config compose\:config\:dev compose\:config\:prod

COMPOSE_ENV ?= dev

ifeq ($(DEV),1)
COMPOSE_ENV = dev
endif

ifeq ($(PROD),1)
COMPOSE_ENV = prod
endif

COMPOSE_FILE_dev = deploy/compose/docker-compose.yml -f deploy/compose/docker-compose.prod.yml -f deploy/compose/docker-compose.dev.yml
COMPOSE_FILE_prod = deploy/compose/docker-compose.yml -f deploy/compose/docker-compose.prod.yml
COMPOSE_FILE = $(COMPOSE_FILE_$(COMPOSE_ENV))

PROTOC         = protoc
MIGRATION_NAME ?= new_migration
PROTO_THIRD_PARTY ?= services/gateway/third_party

IAM_DB_DSN       ?= postgres://iam:iam@localhost:5433/iam?sslmode=disable
KNOWLEDGE_DB_DSN ?= postgres://knowledge:knowledge@localhost:5434/knowledge?sslmode=disable
RAG_DB_DSN       ?= postgres://rag:rag@localhost:5435/rag?sslmode=disable

API_PROTO_FILES = \
	gateway/v1/gateway.proto \
	iam/v1/iam.proto \
	knowledge/v1/knowledge.proto \
	rag/v1/rag.proto \
	aiinference/v1/ai_inference.proto

ifeq ($(OS),Windows_NT)
PROTOC = $(shell powershell -NoProfile -Command "$$cmd=Get-Command protoc -ErrorAction SilentlyContinue; if($$cmd){$$cmd.Source}else{$$p1=Join-Path $$env:LOCALAPPDATA 'Microsoft\\WinGet\\Packages\\Google.Protobuf_Microsoft.Winget.Source_8wekyb3d8bbwe\\bin\\protoc.exe'; $$p2=Join-Path $$env:LOCALAPPDATA 'protoc\\bin\\protoc.exe'; if(Test-Path -LiteralPath $$p1 -ErrorAction SilentlyContinue){$$p1}elseif(Test-Path -LiteralPath $$p2 -ErrorAction SilentlyContinue){$$p2}else{'protoc'}}")
endif

# Проверка

lint\:gateway:
	$(MAKE) -C services/gateway lint

lint\:iam:
	$(MAKE) -C services/iam lint

lint\:knowledge:
	$(MAKE) -C services/knowledge lint

lint\:rag:
	$(MAKE) -C services/rag lint

lint\:ai-inference:
	$(MAKE) -C services/ai-inference lint

lint: lint\:gateway lint\:iam lint\:knowledge lint\:rag lint\:ai-inference

# Тесты

test\:gateway:
	$(MAKE) -C services/gateway test

test\:iam:
	$(MAKE) -C services/iam test

test\:knowledge:
	$(MAKE) -C services/knowledge test

test\:rag:
	$(MAKE) -C services/rag test

test\:ai-inference:
	$(MAKE) -C services/ai-inference test

test: test\:gateway test\:iam test\:knowledge test\:rag test\:ai-inference

test\:cover\:gateway:
	$(MAKE) -C services/gateway test:cover

test\:cover\:iam:
	$(MAKE) -C services/iam test:cover

test\:cover\:knowledge:
	$(MAKE) -C services/knowledge test:cover

test\:cover\:rag:
	$(MAKE) -C services/rag test:cover

test\:cover\:ai-inference:
	$(MAKE) -C services/ai-inference test:cover

test\:cover: test\:cover\:gateway test\:cover\:iam test\:cover\:knowledge test\:cover\:rag test\:cover\:ai-inference

coverage: test\:cover

# Сборка

build\:gateway:
	$(MAKE) -C services/gateway build

build\:iam:
	$(MAKE) -C services/iam build

build\:knowledge:
	$(MAKE) -C services/knowledge build

build\:rag:
	$(MAKE) -C services/rag build

build\:ai-inference:
	$(MAKE) -C services/ai-inference build

build: build\:gateway build\:iam build\:knowledge build\:rag build\:ai-inference

# Proto и публичный API

proto\:tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.11
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.27.2
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.27.2

proto\:deps:
	$(MAKE) -C services/gateway proto:deps
	$(MAKE) -C services/iam proto:deps
	$(MAKE) -C services/knowledge proto:deps
	$(MAKE) -C services/rag proto:deps
	$(MAKE) -C services/ai-inference proto:deps

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
	powershell -NoProfile -Command "New-Item -ItemType Directory -Force api/gen/go,api/gen/openapiv2 | Out-Null; & '$(PROTOC)' -I $(PROTO_THIRD_PARTY) -I api/proto --go_out=api/gen/go --go_opt=paths=source_relative --go-grpc_out=api/gen/go --go-grpc_opt=paths=source_relative --grpc-gateway_out=api/gen/go --grpc-gateway_opt=paths=source_relative --openapiv2_out=api/gen/openapiv2 $(API_PROTO_FILES)"
else
	mkdir -p api/gen/go api/gen/openapiv2
	$(PROTOC) -I $(PROTO_THIRD_PARTY) -I api/proto \
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
proto\:gen\:gateway: api\:gen
proto\:gen\:iam: api\:gen
proto\:gen\:knowledge: api\:gen
proto\:gen\:rag: api\:gen
proto\:gen\:ai-inference: api\:gen

# gRPC-заглушки транспорта

grpc\:stubs\:gateway:
	$(MAKE) -C services/gateway grpc:stubs

grpc\:stubs\:iam:
	$(MAKE) -C services/iam grpc:stubs

grpc\:stubs\:knowledge:
	$(MAKE) -C services/knowledge grpc:stubs

grpc\:stubs\:rag:
	$(MAKE) -C services/rag grpc:stubs

grpc\:stubs\:ai-inference:
	$(MAKE) -C services/ai-inference grpc:stubs

grpc\:stubs: grpc\:stubs\:gateway grpc\:stubs\:iam grpc\:stubs\:knowledge grpc\:stubs\:rag grpc\:stubs\:ai-inference

# Миграции баз данных

migrate\:validate\:iam:
	$(MAKE) -C services/iam migrate:validate

migrate\:validate\:knowledge:
	$(MAKE) -C services/knowledge migrate:validate

migrate\:validate\:rag:
	$(MAKE) -C services/rag migrate:validate

migrate\:validate: migrate\:validate\:iam migrate\:validate\:knowledge migrate\:validate\:rag

migrate\:status\:iam:
	$(MAKE) -C services/iam migrate:status DB_DSN="$(IAM_DB_DSN)"

migrate\:status\:knowledge:
	$(MAKE) -C services/knowledge migrate:status DB_DSN="$(KNOWLEDGE_DB_DSN)"

migrate\:status\:rag:
	$(MAKE) -C services/rag migrate:status DB_DSN="$(RAG_DB_DSN)"

migrate\:status: migrate\:status\:iam migrate\:status\:knowledge migrate\:status\:rag

migrate\:up\:iam:
	$(MAKE) -C services/iam migrate:up DB_DSN="$(IAM_DB_DSN)"

migrate\:up\:knowledge:
	$(MAKE) -C services/knowledge migrate:up DB_DSN="$(KNOWLEDGE_DB_DSN)"

migrate\:up\:rag:
	$(MAKE) -C services/rag migrate:up DB_DSN="$(RAG_DB_DSN)"

migrate\:up: migrate\:up\:iam migrate\:up\:knowledge migrate\:up\:rag

migrate\:down\:iam:
	$(MAKE) -C services/iam migrate:down DB_DSN="$(IAM_DB_DSN)"

migrate\:down\:knowledge:
	$(MAKE) -C services/knowledge migrate:down DB_DSN="$(KNOWLEDGE_DB_DSN)"

migrate\:down\:rag:
	$(MAKE) -C services/rag migrate:down DB_DSN="$(RAG_DB_DSN)"

migrate\:down: migrate\:down\:iam migrate\:down\:knowledge migrate\:down\:rag

migrate\:create\:iam:
	$(MAKE) -C services/iam migrate:create MIGRATION_NAME="$(MIGRATION_NAME)"

migrate\:create\:knowledge:
	$(MAKE) -C services/knowledge migrate:create MIGRATION_NAME="$(MIGRATION_NAME)"

migrate\:create\:rag:
	$(MAKE) -C services/rag migrate:create MIGRATION_NAME="$(MIGRATION_NAME)"

# Docker Compose

compose\:up:
	docker compose -f $(COMPOSE_FILE) up -d --no-recreate

compose\:recreate:
	docker compose -f $(COMPOSE_FILE) up -d --build --force-recreate --remove-orphans

compose\:down:
	docker compose -f $(COMPOSE_FILE) down

compose\:config:
	docker compose -f $(COMPOSE_FILE) config --quiet

compose\:up\:dev:
	$(MAKE) compose:up COMPOSE_ENV=dev

compose\:up\:prod:
	$(MAKE) compose:up COMPOSE_ENV=prod

compose\:recreate\:dev:
	$(MAKE) compose:recreate COMPOSE_ENV=dev

compose\:recreate\:prod:
	$(MAKE) compose:recreate COMPOSE_ENV=prod

compose\:down\:dev:
	$(MAKE) compose:down COMPOSE_ENV=dev

compose\:down\:prod:
	$(MAKE) compose:down COMPOSE_ENV=prod

compose\:config\:dev:
	$(MAKE) compose:config COMPOSE_ENV=dev

compose\:config\:prod:
	$(MAKE) compose:config COMPOSE_ENV=prod

# Makefile - 统一开发命令入口
#
# 常用任务:
#   make help            打印所有目标说明
#   make build           编译全部子包(default tag)
#   make build-all       编译 default + integration + miniredis_test 4 路径
#   make test            跑单元测试(无需 Redis,带 -race)
#   make test-mr         跑 miniredis 套(无需 Redis,带 -race)
#   make test-int        跑集成测试(需 Redis)
#   make test-all        以上三类全跑(需 Redis)
#   make cover           生成 cover 报告(需 Redis)
#   make cover-stats     仅打印覆盖率(总体 + 排除 cmd_gen)
#   make lint            golangci-lint
#   make fmt             gofmt -w
#   make builder-check   builder AST 不变式 + spec 漂移检测
#   make tidy            go mod tidy
#   make clean           清理产物
#   make ci              CI 等价完整流水(build/lint/test-mr/test-int/builder-check)
#
# 环境变量:
#   REDIS_ADDR=127.0.0.1:6379    集成测试用的 Redis 地址(被 helpers_test 默认使用)

GO        ?= go
TIMEOUT   ?= 5m
PKG       ?= ./...
COVER_OUT ?= coverage.out

.DEFAULT_GOAL := help

.PHONY: help
help:
	@awk 'BEGIN{FS=":.*?##"} /^[a-zA-Z_-]+:.*?##/{printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: build
build: ## 编译 default tag
	$(GO) build $(PKG)

.PHONY: build-all
build-all: ## 编译 4 个 tag 路径
	$(GO) build $(PKG)
	$(GO) build -tags integration $(PKG)
	$(GO) build -tags miniredis $(PKG)
	$(GO) build -tags 'miniredis_test redisson_miniredis' $(PKG)

.PHONY: vet
vet: ## go vet 3 个 tag 路径
	$(GO) vet $(PKG)
	$(GO) vet -tags integration $(PKG)
	$(GO) vet -tags 'miniredis_test redisson_miniredis' $(PKG)

.PHONY: test
test: ## 跑单元测试(无需 Redis,带 -race)
	$(GO) test -race -count=1 -timeout=$(TIMEOUT) $(PKG)

.PHONY: test-mr
test-mr: ## 跑 miniredis 套(无需 Redis,带 -race)
	$(GO) test -tags 'miniredis_test redisson_miniredis' -race -count=1 -timeout=$(TIMEOUT) $(PKG)

.PHONY: test-int
test-int: ## 跑集成测试(需 Redis,带 -race)
	$(GO) test -tags integration -race -count=1 -timeout=$(TIMEOUT) $(PKG)

.PHONY: test-all
test-all: test test-mr test-int ## 三类测试全跑

.PHONY: cover
cover: ## 生成集成测试 cover 报告
	$(GO) test -tags integration -coverprofile=$(COVER_OUT) -count=1 -timeout=$(TIMEOUT) .
	$(GO) tool cover -html=$(COVER_OUT) -o coverage.html
	@echo "report: coverage.html"

.PHONY: cover-stats
cover-stats: ## 打印覆盖率统计(总体 + 排除 cmd_gen)
	@$(GO) test -tags integration -coverprofile=$(COVER_OUT) -count=1 -timeout=$(TIMEOUT) . | tail -1
	@$(GO) tool cover -func=$(COVER_OUT) | tail -1
	@awk 'NR>1 && !/cmd_gen\.go/ {total+=$$2; if ($$3>0) covered+=$$2} END {printf "排除 cmd_gen.go: %d/%d = %.1f%%\n", covered, total, covered*100/total}' $(COVER_OUT)

.PHONY: cover-stats-mr
cover-stats-mr: ## 打印 miniredis 套覆盖率(主要是 builder_*.go 单测)
	@$(GO) test -tags 'miniredis_test redisson_miniredis' -coverprofile=coverage-mr.out -count=1 -timeout=$(TIMEOUT) . | tail -1
	@$(GO) tool cover -func=coverage-mr.out | tail -1
	@awk 'NR>1 && /builder_/ {total+=$$2; if ($$3>0) covered+=$$2} END {printf "miniredis 套 builder_*.go: %d/%d = %.1f%%\n", covered, total, covered*100/total}' coverage-mr.out

.PHONY: lint
lint: ## golangci-lint
	golangci-lint run --timeout=$(TIMEOUT) $(PKG)

.PHONY: fmt
fmt: ## gofmt -w
	gofmt -w .

.PHONY: fmt-check
fmt-check: ## gofmt 干净度检查(CI 用)
	@diff=$$(gofmt -l .); \
	if [ -n "$$diff" ]; then echo "gofmt issues:"; echo "$$diff"; exit 1; fi

.PHONY: builder-check
builder-check: ## builder AST 不变式 + spec 漂移检测
	$(GO) run ./cmd/genbuilder -check
	$(GO) run ./cmd/genbuilder -check-spec specs/builder.yaml

.PHONY: tidy
tidy: ## go mod tidy
	$(GO) mod tidy

.PHONY: ci
ci: build-all vet fmt-check lint builder-check test-mr test-int ## CI 等价完整流水(需 Redis)

.PHONY: clean
clean: ## 清理产物
	rm -f $(COVER_OUT) coverage.html

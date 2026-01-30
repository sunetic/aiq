.PHONY: build build-test clean test-version help

# 正常构建（使用 dev 版本，用于开发）
build:
	go build -o aiq cmd/aiq/main.go

# 测试构建（指定版本号）
# 用法: make build-test VERSION=v1.0.0 [COMMIT=abc1234]
build-test:
	@if [ -z "$(VERSION)" ]; then \
		echo "Usage: make build-test VERSION=v1.0.0 [COMMIT=abc1234]"; \
		echo "   or: make test-v<version> (e.g., make test-v0.0.4)"; \
		exit 1; \
	fi
	@COMMIT=$${COMMIT:-test1234}; \
	echo "Building with version: $(VERSION), commit: $$COMMIT"; \
	go build -ldflags "-X github.com/aiq/aiq/internal/version.Version=$(VERSION) -X github.com/aiq/aiq/internal/version.CommitID=$$COMMIT" -o aiq cmd/aiq/main.go
	@echo "Build complete. Version: $(VERSION)"

# 快速测试版本命令（需要先 build-test）
test-version: build-test
	@./aiq -v

# 通用版本测试规则：make test-v<version>
# 例如: make test-v0.0.4, make test-v1.0.0, make test-v1.2.3
# Commit ID 使用当前 git commit ID（如果 git 可用），否则使用版本号的哈希
test-v%:
	@VERSION="v$*"; \
	if git rev-parse HEAD >/dev/null 2>&1; then \
		COMMIT=$$(git rev-parse HEAD | cut -c1-7); \
		echo "Using real git commit ID: $$COMMIT"; \
	else \
		COMMIT=$$(echo -n "$$VERSION" | shasum -a 256 | cut -c1-7); \
		echo "Using hash-based commit ID (git not available): $$COMMIT"; \
	fi; \
	echo "Building and testing version: $$VERSION (commit: $$COMMIT)"; \
	go build -ldflags "-X github.com/aiq/aiq/internal/version.Version=$$VERSION -X github.com/aiq/aiq/internal/version.CommitID=$$COMMIT" -o aiq cmd/aiq/main.go && \
	./aiq -v

# 预定义的快捷命令（可选）
test-v1.0.0:
	@make build-test VERSION=v1.0.0 COMMIT=abc1234
	@./aiq -v

test-v1.1.0:
	@make build-test VERSION=v1.1.0 COMMIT=def5678
	@./aiq -v

clean:
	rm -f aiq

help:
	@echo "Available targets:"
	@echo "  make build              - Build with dev version (for development)"
	@echo "  make build-test VERSION=v1.0.0 [COMMIT=abc1234] - Build with specific version"
	@echo "  make test-v<version>   - Build and test version (e.g., make test-v0.0.4)"
	@echo "  make clean              - Remove built binary"

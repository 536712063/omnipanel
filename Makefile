.PHONY: all build build-plugins dev clean proto

# OmniPanel Makefile
# 企业级构建管线

APP_NAME := omnipanel
BUILD_DIR := build
PLUGINS := docker ssh frp sdt
GO := go
WAILS := wails
PROTOC := protoc

all: proto build-plugins build

# ---------- Protobuf 代码生成 ----------
proto:
	@echo "生成 protobuf 代码..."
	$(PROTOC) --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/common/v1/common.proto \
		proto/sdt/v1/sdt.proto \
		proto/agent/v1/agent.proto

# ---------- 构建插件二进制文件 ----------
build-plugins:
	@mkdir -p $(BUILD_DIR)/plugins
	@for plugin in $(PLUGINS); do \
		echo "构建插件: $$plugin"; \
		$(GO) build -o $(BUILD_DIR)/plugins/omnipanel-plugin-$$plugin \
			./plugins/$$plugin/; \
	done

# ---------- 构建前端 ----------
build-frontend:
	cd ui && npm install && npm run build

# ---------- 构建 Wails 应用 (Go + 前端) ----------
build: build-plugins build-frontend
	@echo "构建 Wails 应用..."
	$(WAILS) build -o $(BUILD_DIR)/$(APP_NAME)

# ---------- 开发模式 ----------
dev:
	$(WAILS) dev

# ---------- 清理 ----------
clean:
	rm -rf $(BUILD_DIR)
	rm -rf ui/dist

# ---------- 测试 ----------
test:
	$(GO) test ./...

# ---------- 交叉编译 (Windows) ----------
build-windows:
	GOOS=windows GOARCH=amd64 $(WAILS) build -platform windows/amd64 -o $(BUILD_DIR)/$(APP_NAME).exe

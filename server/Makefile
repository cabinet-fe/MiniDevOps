.PHONY: dev build run clean install-air

# 开发模式 - 使用Air热重载
dev:
	air

# 构建项目
build:
	go build -o bin/main ./cmd

# 运行项目
run:
	go run ./cmd

# 清理构建文件
clean:
	rm -rf tmp/
	rm -rf bin/
	rm -f build-errors.log

# 安装Air（如果没有安装）
install-air:
	go install github.com/air-verse/air@latest

# 初始化项目依赖
init:
	go mod tidy

# 格式化代码
fmt:
	go fmt ./...

# 运行测试
test:
	go test ./...

# 构建并运行
build-run: build
	./bin/main

# 帮助信息
help:
	@echo "可用的命令:"
	@echo "  dev         - 开发模式（热重载）"
	@echo "  build       - 构建项目"
	@echo "  run         - 运行项目"
	@echo "  clean       - 清理构建文件"
	@echo "  install-air - 安装Air工具"
	@echo "  init        - 初始化项目依赖"
	@echo "  fmt         - 格式化代码"
	@echo "  test        - 运行测试"
	@echo "  build-run   - 构建并运行"
	@echo "  help        - 显示帮助信息"
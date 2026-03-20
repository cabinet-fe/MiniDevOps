# 构建体系优化与项目文档

> 状态: 已执行

## 目标

优化开发/生产构建流程，精简 Makefile，添加 GitHub Actions 自动发布工作流，并补充项目 README 文档。

## 内容

### 步骤 1：分离开发与生产环境的前端嵌入

使用 Go build tags 区分开发和生产模式：

- 创建 `cmd/server/embed_prod.go`（`//go:build !dev`）：保留 `//go:embed all:dist` 和 `serveSPA` 逻辑。
- 创建 `cmd/server/embed_dev.go`（`//go:build dev`）：提供空的 `serveSPA` 实现（开发环境由 Vite dev server 代理处理前端）。
- 修改 `cmd/server/main.go`：移除 embed 指令和 serveSPA 函数，改为调用独立文件中定义的函数。同时在开发模式下使用 `gin.DebugMode` 替代 `gin.ReleaseMode`。
- 开发模式下无需 `mkdir -p cmd/server/dist`，消除空目录依赖。

**完成标准**：`go run -tags dev ./cmd/server` 启动时不需要 dist 目录；`go build ./cmd/server` 正常嵌入前端产物。

### 步骤 2：精简 Makefile

仅保留三个命令：

- `dev`：使用 `-tags dev` 启动后端 + Vite 前端开发服务器。
- `build-linux`：构建前端 → 拷贝到 dist → 交叉编译 Linux amd64 二进制。
- `build-win`：构建前端 → 拷贝到 dist → 交叉编译 Windows amd64 二进制。

额外保留 `clean` 作为辅助命令。

**完成标准**：`make dev` 可正常启动开发环境；`make build-linux` 和 `make build-win` 可正常生成对应平台二进制。

### 步骤 3：添加 GitHub Actions 自动发布工作流

创建 `.github/workflows/release.yml`：

- 触发条件：推送 `v*` 标签。
- 构建矩阵：
  - linux/amd64
  - linux/arm64
  - windows/amd64
- 步骤：
  1. Checkout 代码。
  2. 安装 Go、Bun。
  3. 构建前端（`cd web && bun install && bun run build`）。
  4. 拷贝前端产物到 `cmd/server/dist/`。
  5. 按矩阵交叉编译 Go 二进制（CGO_ENABLED=1，使用 zig 或对应交叉编译工具链）。
  6. 创建 GitHub Release 并上传所有二进制资源。

采用 `modernc.org/sqlite` 纯 Go 替代 `mattn/go-sqlite3`，彻底消除 CGO 依赖，简化交叉编译。需将 GORM driver 更换为 `gorm.io/driver/sqlite` 配合 modernc 后端（`github.com/glebarez/sqlite`）。所有 build 命令使用 `CGO_ENABLED=0`。

**完成标准**：推送 `v*.*.*` 标签后，GitHub Actions 自动构建三平台二进制并创建 Release。

### 步骤 4：添加 README.md

在项目根目录创建 `README.md`，包含：

- 项目简介（BuildFlow CI/CD 平台）
- 功能特性列表
- 快速开始（Docker / 二进制部署）
- 开发指南（环境要求、本地开发、构建命令）
- 配置说明（config.yaml 各字段含义）
- 技术栈概览
- 许可证

**完成标准**：README 内容完整、可读，覆盖用户从安装到使用的全流程。

## 影响范围

- `cmd/server/main.go` — 移除 embed/fs/strings 导入、webFS 变量、serveSPA 函数、gin.SetMode 调用；新增 version 变量
- `cmd/server/embed_prod.go` — 新增：生产模式 embed + serveSPA（build tag: !dev）
- `cmd/server/embed_dev.go` — 新增：开发模式空 serveSPA（build tag: dev）
- `internal/model/database.go` — SQLite 驱动替换为 github.com/glebarez/sqlite
- `internal/engine/pipeline_test.go` — SQLite 驱动替换
- `internal/service/build_service_test.go` — SQLite 驱动替换
- `go.mod` / `go.sum` — 依赖变更（移除 mattn/go-sqlite3，新增 glebarez/sqlite + modernc.org/sqlite）
- `Makefile` — 精简为 dev / build-linux / build-win / clean
- `.github/workflows/release.yml` — 新增：GitHub Actions 自动发布工作流
- `README.md` — 新增：项目文档

## 历史补丁

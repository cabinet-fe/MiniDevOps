# 修复 make dev 配置文件路径

## 补丁内容

Makefile 的 `dev` 和 `dev-backend` 目标执行 `cd cmd/server && go run .`，将工作目录切换到 `cmd/server/`，
而 `config.yaml` 位于项目根目录，导致后端启动时报 `open config.yaml: no such file or directory`。

修复方式：在 `go run .` 后添加 `--config ../../config.yaml` 参数，显式指定配置文件的正确相对路径。

## 影响范围

- 修改文件: `Makefile`

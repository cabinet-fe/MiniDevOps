# 文档：部署 Agent 使用说明

## 补丁内容

在根目录 `README.md` 与前端「项目手册」页补充 **HTTP Agent（buildflow-agent）** 的部署与对接说明：Bearer 鉴权、与主控「服务器管理」Token 一致、默认 `buildflow-agent.yaml` 与 `-config`、环境变量、配置优先级（YAML → 环境变量 → 命令行）、可选 TLS、以及 `/healthz` 自检。并清理 `project-manual.tsx` 中未使用的 import。

## 影响范围

- 修改文件: `/home/whj/codes/dev-ops/README.md`
- 修改文件: `/home/whj/codes/dev-ops/web/src/pages/project-manual.tsx`

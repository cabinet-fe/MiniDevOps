# MiniDevOps

迷你持续集成工具，本项目主要用于前端的项目构建，打包等等。

## 项目结构

```
mini-devops
├── server // 服务端代码
├── web // 前端代码
├── README.md // 项目说明
```

## 技术栈

- 服务端：Golang + go fiber + ent + sqlite3
- 前端：Vue3 + TypeScript + Pinia + UltraUI + Vite

## 需求

- 支持账户登录，每个账户可以关联多个项目。
- 支持多项目管理，每个项目除了创建人还可以选择其他参与人员。
- 项目包含仓库地址，源码存放目录，项目名称，gitee 用户名和密码，分支，构建脚本(需要注意安全性，防止脚本注入)。

## 主要模块

- 登录模块，默认提供一个初始的管理员账号。
- 项目管理模块，支持项目的创建，分页查询，删除，编辑，查看构建历史(可以保留最新 10 次的构建记录)，并且构建时的命令输出要通过 SSE 或者 WebSocket 输出到前端。

## 如何运行

### 服务端

1. 确保安装了 Go 环境 (推荐 Go 1.21+)
2. 进入 server 目录

```bash
cd server
```

3. 生成 Ent 模型代码

```bash
make init-ent
make generate
```

4. 构建并运行服务端

```bash
make run
```

服务端将在 http://localhost:8080 上运行。

### 前端

1. 确保安装了 Node.js 环境 (推荐 Node.js 16+) 和 Bun
2. 进入 web 目录

```bash
cd web
```

3. 安装依赖

```bash
bun install
```

4. 运行开发服务器

```bash
bun run dev
```

前端将在 http://localhost:3001 上运行。

## 账号信息

默认管理员账号:

- 用户名: admin
- 密码: admin123

## 功能说明

### 用户管理

- 管理员可以创建、编辑和删除用户
- 用户分为管理员和普通用户两种角色
- 管理员账号不能被删除

### 项目管理

- 可以创建、编辑和删除项目
- 可以为项目添加参与者
- 可以手动触发项目构建
- 可以查看项目的构建历史

### 构建历史

- 显示所有项目的构建记录
- 可以查看构建日志
- 可以跳转到相应的项目详情

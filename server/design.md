以下是迷你构建工具服务端的设计文档，基于技术栈（Golang + Go Fiber + Ent + SQLite）和核心功能要求：

---

### 功能名：迷你构建工具

#### 需求描述

开发轻量级构建工具，支持用户登录、管理项目代码仓库（从码云自动拉取代码）、执行自动化构建任务，并提供构建历史追溯功能，满足开发者针对小型项目的快速部署需求。

---

### 概述

工具定位为开发者的个人/小团队生产力助手，核心功能价值：

- **代码托管集成**：直接从码云(Gitee)拉取代码到服务器
- **任务隔离性**：每个构建任务在独立目录中执行，避免污染
- **轻量化部署**：用 SQLite 作为嵌入式数据库，无外部依赖
- **实时日志**：支持 Web 端实时查看构建输出日志

---

### 相关页面设计

#### 1. 用户登录页

- 用户名/密码输入框 + "注册"跳转链接
- 支持第三方码云账号快捷登录（可选）

#### 2. 项目列表页

- 展示用户所有项目的卡片列表（项目名、仓库地址、最后构建时间、构建状态）
- 操作按钮：新建项目、触发构建、查看构建历史

#### 3. 项目新增/编辑页

- 表单字段：项目名称、码云仓库 URL、分支名、构建命令（如`go build -o app`）
- 仓库鉴权：码云 Access Token 输入（加密存储）

#### 4. 构建任务详情页

- 实时日志展示区（类似终端风格，自动滚动）
- 任务状态：运行中/成功/失败 + 耗时统计
- 历史构建记录列表（按时间倒序）

---

### 用户旅程

**场景：用户从码云拉取代码并构建**

1. **登录系统** → 进入项目列表页 → 点击“新建项目”
2. **填写仓库信息**：
   - 输入码云仓库 HTTPS 地址（如`https://gitee.com/user/demo.git`）
   - 提供码云 Access Token（用于私有仓库拉取）
   - 设置构建命令 `go build -o app`
3. **触发构建**：
   - 系统自动拉取代码到`/data/projects/{project_id}/`目录
   - 执行用户预设的构建命令
4. **查看结果**：
   - 实时日志页面显示`git clone`进度和构建输出
   - 构建成功时生成可执行文件到`/data/output/{project_id}/`目录

---

### 用户故事

1. **作为开发者**：
   - 我希望通过码云账号登录系统，避免单独管理密码
   - 我需要在创建项目时自动下载代码仓库，并保存构建配置
2. **作为团队技术负责人**：
   - 我需要查看所有历史构建记录，分析失败原因
   - 我需确保每个构建任务的环境隔离，避免依赖冲突
3. **作为系统管理员**：
   - 我需要监控服务器资源占用（CPU/内存），防止构建任务耗尽资源

---

### 技术实现逻辑

#### 1. 后端架构分层设计

```text
├── main.go         # 入口文件（路由初始化）
├── internal
│   ├── auth       # 鉴权模块（JWT签发/验证）
│   ├── model      # Ent生成的数据库模型
│   ├── gitee      # 码云API封装（仓库校验/元数据获取）
│   ├── builder    # 构建任务引擎
│   └── task       # 构建任务调度管理
```

#### 2. 关键技术实现

- **码云集成**：

  ```go
  // 仓库校验示例
  func ValidateGiteeRepo(url, token string) error {
      client := gitee.NewClient(token)
      _, err := client.GetRepoFromURL(url) // 调用码云API验证仓库存在性
      return err
  }
  ```

- **构建任务执行器**：

  ```go
  // 使用Go的os/exec执行shell命令，并捕获实时输出
  func RunBuildCommand(projectID string, cmd string) {
      dir := fmt.Sprintf("/data/projects/%s", projectID)
      command := exec.Command("sh", "-c", cmd)
      command.Dir = dir

      // 实时输出日志到WebSocket/SSE
      stdoutPipe, _ := command.StdoutPipe()
      go streamLogs(stdoutPipe, projectID)

      command.Start()
      // 记录任务状态到数据库
  }
  ```

- **资源隔离方案**：
  - 每个项目的代码存放在独立目录：`/data/projects/{project_id}/`
  - 构建输出产物分离：`/data/output/{project_id}/build_{timestamp}/`
  - 任务进程树监控：通过`syscall`获取 PID 并限制资源

#### 3. 数据库表设计（Ent Schema）

```go
// 用户表
type User struct {
    ent.Schema
    fields.Field{
        Field("username").Unique(),
        Field("password").Sensitive(),
        Field("gitee_token").Optional().Sensitive(),
    }
}

// 项目表
type Project struct {
    ent.Schema
    fields.Field{
        Field("name"),
        Field("repo_url"),
        Field("branch").Default("master"),
        Field("build_cmd"),
        Edge("owner", User.Type).Unique(), // 项目归属用户
    }
}

// 构建任务表
type BuildTask struct {
    ent.Schema
    fields.Field{
        Field("status").Enum("pending", "running", "success", "failed"),
        Field("log_path"),
        Field("duration").Optional(),
        Edge("project", Project.Type), // 任务归属项目
    }
}
```

---

### 功能细节描述

#### 1. 用户认证安全设计

- **密码存储**：bcrypt 哈希 + 随机盐值
- **JWT 有效期**：Access Token(2 小时) + Refresh Token(7 天)
- **码云 Token 加密**：使用 AES-GCM 加密后存储到数据库

#### 2. 构建任务保障机制

| 机制         | 实现方案                              |
| ------------ | ------------------------------------- |
| **超时中断** | 任务启动时设置 15 分钟超时定时器      |
| **异常捕获** | 拦截 panic 并标记任务状态为失败       |
| **重试策略** | 允许用户手动重试失败任务（最多 3 次） |
| **空间清理** | 定时任务清理超过 30 天的构建产物目录  |

#### 3. RESTful API 设计示例

```text
POST /api/login          # 用户登录（返回JWT）
POST /api/projects       # 创建新项目
GET  /api/projects/{id}/tasks  # 获取项目的构建历史
POST /api/projects/{id}/build  # 触发新的构建
WS   /api/builds/{id}/logs     # WebSocket实时日志推送
```

---

### 扩展性考虑（可选）

- **多仓库支持**：未来可扩展 GitHub/GitLab 适配器
- **通知模块**：构建结果通过 Webhook 或邮件通知
- **分布式构建**：通过 Redis 队列实现多节点任务分发
- **构建缓存**：支持 go mod/vendor 缓存加速构建

如果需要进一步细化某个模块（如 Ent 的完整 Schema 定义或安全设计细节），可随时告知！

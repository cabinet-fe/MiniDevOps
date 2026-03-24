import {
  BookOpen,
  Rocket,
  ShieldCheck,
  Key,
  Layers,
  Terminal,
  Server,
  Workflow,
  Github,
  CheckCircle2,
  Info,
  Clock,
  Shield,
  Activity,
  Code2,
} from 'lucide-react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  BUILD_SCRIPT_TYPES,
  DEPLOY_METHODS,
  REPO_AUTH_TYPES,
  ARTIFACT_FORMATS,
  ROLES,
} from '@/lib/constants'

const toc = [
  { id: 'intro', label: '项目简介', icon: BookOpen },
  { id: 'rbac', label: '角色权限', icon: ShieldCheck },
  { id: 'credential', label: '凭证管理', icon: Key },
  { id: 'project', label: '项目配置', icon: Layers },
  { id: 'environment', label: '环境与变量', icon: SettingsIcon },
  { id: 'pipeline', label: '流水线原理', icon: Workflow },
  { id: 'deploy', label: '部署方式', icon: Server },
  { id: 'agent', label: '独立 Agent', icon: Rocket },
] as const

function SettingsIcon({ className }: { className?: string }) {
  return <Terminal className={className} />
}

function Section({
  id,
  title,
  icon: Icon,
  children,
}: {
  id: string
  title: string
  icon: any
  children: React.ReactNode
}) {
  return (
    <section id={id} className="scroll-mt-24 space-y-6">
      <div className="flex items-center gap-3">
        <div className="rounded-lg bg-primary/10 p-2 text-primary">
          <Icon className="h-6 w-6" />
        </div>
        <h2 className="text-2xl font-bold tracking-tight text-foreground">{title}</h2>
      </div>
      <div className="space-y-4 text-sm leading-relaxed text-muted-foreground [&_strong]:font-semibold [&_strong]:text-foreground [&_code]:rounded-md [&_code]:bg-muted/80 [&_code]:px-1.5 [&_code]:py-0.5 [&_code]:font-mono [&_code]:text-[13px] [&_code]:text-foreground [&_ul]:list-disc [&_ul]:space-y-2 [&_ul]:pl-5">
        {children}
      </div>
    </section>
  )
}

export function ProjectManualPage() {
  return (
    <div className="mx-auto max-w-5xl space-y-12 pb-20">
      {/* Header */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div className="space-y-2">
          <h1 className="text-4xl font-extrabold tracking-tight text-foreground lg:text-5xl">项目手册</h1>
          <p className="text-lg text-muted-foreground">
            全面了解 BuildFlow 的核心概念、操作流程与最佳实践。
          </p>
        </div>
        <Button variant="outline" className="w-fit gap-2" asChild>
          <a href="https://github.com/cabinet-fe/MiniDevOps" target="_blank" rel="noreferrer">
            <Github className="h-4 w-4" />
            源码仓库
          </a>
        </Button>
      </div>

      <Separator />

      {/* Quick Navigation */}
      <div className="grid grid-cols-2 gap-4 md:grid-cols-4">
        {toc.map((item) => (
          <a
            key={item.id}
            href={`#${item.id}`}
            className="flex items-center gap-2 rounded-xl border bg-card p-4 transition-all hover:border-primary/50 hover:bg-accent/50 group"
          >
            <item.icon className="h-5 w-5 text-muted-foreground group-hover:text-primary" />
            <span className="text-sm font-medium">{item.label}</span>
          </a>
        ))}
      </div>

      <Card className="border-none bg-muted/30 shadow-none">
        <CardContent className="space-y-20 p-8 sm:p-12">
          {/* Introduction */}
          <Section id="intro" title="项目简介" icon={BookOpen}>
            <p className="text-base">
              <strong>BuildFlow</strong> 是一款轻量级 CI/CD 构建部署平台。采用 Go 后端 + React 前端单体仓库架构，
              支持从代码克隆、自动化构建到多种模式部署的全生命周期管理。
            </p>
            <div className="grid gap-4 sm:grid-cols-2">
              <div className="flex items-start gap-3 rounded-lg border bg-background p-4">
                <CheckCircle2 className="mt-1 h-5 w-5 text-green-500" />
                <div>
                  <h4 className="font-medium text-foreground">单文件部署</h4>
                  <p className="text-xs">前端产物嵌入后端二进制，一个文件即可运行整个平台。</p>
                </div>
              </div>
              <div className="flex items-start gap-3 rounded-lg border bg-background p-4">
                <CheckCircle2 className="mt-1 h-5 w-5 text-green-500" />
                <div>
                  <h4 className="font-medium text-foreground">多模式部署</h4>
                  <p className="text-xs">支持 Rsync、SFTP、SCP 以及更安全的专用 Agent 模式。</p>
                </div>
              </div>
            </div>
          </Section>

          <Separator />

          {/* RBAC */}
          <Section id="rbac" title="角色与权限" icon={ShieldCheck}>
            <p>系统内置基于角色的访问控制（RBAC），支持三级角色体系：</p>
            <div className="grid gap-4 sm:grid-cols-3">
              {ROLES.map((role) => (
                <Card key={role.value} className="relative overflow-hidden">
                  <CardHeader className="pb-2">
                    <Badge variant="outline" className="w-fit">
                      {role.label}
                    </Badge>
                  </CardHeader>
                  <CardContent className="text-xs">
                    {role.value === 'admin' && '最高权限。管理系统配置、用户角色、审计日志及所有业务数据。'}
                    {role.value === 'ops' && '运维权限。管理服务器、凭证、环境配置及执行部署操作。'}
                    {role.value === 'dev' && '开发权限。查看项目详情、实时日志、手动触发构建。'}
                  </CardContent>
                </Card>
              ))}
            </div>
            <p className="flex items-center gap-2 text-xs italic">
              <Info className="h-4 w-4" />
              权限校验覆盖前端菜单显隐与后端 API 鉴权。
            </p>
          </Section>

          <Separator />

          {/* Credential */}
          <Section id="credential" title="凭证管理" icon={Key}>
            <p>
              BuildFlow 采用<strong>集中式凭证管理</strong>，确保敏感信息（如密码、Token、私钥）不被明文硬编码在项目配置中。
            </p>
            <ul>
              <li>
                <strong>用户名/密码</strong>：适用于传统的 Git 认证或服务器 SSH 密码登录。
              </li>
              <li>
                <strong>私钥</strong>：支持 RSA/ED25519 等格式的 SSH 私钥，建议用于服务器连接。
              </li>
              <li>
                <strong>Token</strong>：适用于 GitHub PAT、码云私人令牌或 Agent 认证密钥。
              </li>
            </ul>
            <p className="rounded-md bg-yellow-500/10 p-4 text-xs text-yellow-600 dark:text-yellow-400">
              <strong>安全提示</strong>：所有敏感凭据在数据库中均通过 <code>AES-GCM-256</code> 高强度加密存储。
            </p>
          </Section>

          <Separator />

          {/* Project */}
          <Section id="project" title="项目配置" icon={Layers}>
            <p>
              项目是构建任务的最小单元。创建项目时需指定<strong>仓库地址 (HTTPS/SSH)</strong>。
            </p>
            <ul>
              <li>
                <strong>仓库认证</strong>：
                {REPO_AUTH_TYPES.map((t) => (
                  <span key={t.value}>
                    {' '}
                    <code>{t.label}</code>
                    {t.value === 'credential' ? '（引用「凭证管理」中创建的凭据）' : ''}
                    {t.value !== REPO_AUTH_TYPES[REPO_AUTH_TYPES.length - 1].value ? '、' : '。'}
                  </span>
                ))}
              </li>
              <li>
                <strong>默认分支</strong>：系统将根据此分支拉取代码进行初次分析或默认构建。
              </li>
              <li>
                <strong>产物格式</strong>：支持{' '}
                {ARTIFACT_FORMATS.map((f) => (
                  <code key={f.value}>{f.label}</code>
                ))}
                。后端会将构建结果打包为此格式以便传输。
              </li>
            </ul>
          </Section>

          <Separator />

          {/* Environment */}
          <Section id="environment" title="环境与变量组" icon={SettingsIcon}>
            <p>
              环境（Environment）定义了具体的运行实例（如 Dev, Staging, Prod）。
            </p>
            <div className="space-y-4">
              <div className="rounded-lg border bg-background p-4">
                <h4 className="mb-2 flex items-center gap-2 font-medium text-foreground">
                  <Clock className="h-4 w-4 text-primary" /> 定时任务 (Cron)
                </h4>
                <p className="text-xs">支持标准 Cron 表达式（如 <code>0 0 * * *</code> 表示每日零点构建），由后端基于内存调度器自动触发。</p>
              </div>
              <div className="rounded-lg border bg-background p-4">
                <h4 className="mb-2 flex items-center gap-2 font-medium text-foreground">
                  <Shield className="h-4 w-4 text-primary" /> 变量组 (Variable Groups)
                </h4>
                <p className="text-xs">可创建共享的变量组并关联至多个环境。变量将以环境变量的形式注入构建脚本中。支持<strong>敏感变量</strong>（界面展示掩码，后端加密保存）。</p>
              </div>
            </div>
          </Section>

          <Separator />

          {/* Pipeline */}
          <Section id="pipeline" title="流水线原理" icon={Workflow}>
            <p>一次完整的流水线包含以下四个核心阶段：</p>
            <div className="relative space-y-8 before:absolute before:left-[17px] before:top-2 before:h-[calc(100%-16px)] before:w-0.5 before:bg-border">
              {[
                { label: '代码准备', desc: '根据环境配置的凭证与分支，执行 git fetch/checkout，支持浅克隆以加速。', icon: Github },
                { label: '脚本构建', desc: '在指定的工作区执行自定义脚本（Bash/Node/Python）。支持依赖缓存保留。', icon: Code2 },
                { label: '产物打包', desc: '扫描构建产物目录，按项目设定的格式（Zip/Gzip）生成压缩包。', icon: Layers },
                { label: '分发部署', desc: '将打包好的产物通过选定的协议推送至目标服务器并解压。', icon: Rocket },
              ].map((step, i) => (
                <div key={i} className="relative flex items-start gap-6">
                  <div className="z-10 flex h-9 w-9 items-center justify-center rounded-full border bg-background text-primary shadow-sm">
                    <step.icon className="h-4 w-4" />
                  </div>
                  <div className="space-y-1 pt-1">
                    <h4 className="font-semibold text-foreground">{step.label}</h4>
                    <p className="text-xs">{step.desc}</p>
                  </div>
                </div>
              ))}
            </div>
          </Section>

          <Separator />

          {/* Deploy Methods */}
          <Section id="deploy" title="部署方式" icon={Server}>
            <div className="grid gap-4 sm:grid-cols-2">
              {DEPLOY_METHODS.map((method) => (
                <div key={method.value} className="rounded-lg border bg-background p-4">
                  <div className="mb-2 flex items-center justify-between">
                    <Badge>{method.label}</Badge>
                    <Activity className="h-4 w-4 text-muted-foreground" />
                  </div>
                  <p className="text-xs">
                    {method.value === 'rsync' && '增量同步，仅传输变化的文件。适合 Linux 生产环境。'}
                    {method.value === 'sftp' && '基于 SSH 子协议的文件传输。无需服务器安装 rsync，兼容性好。'}
                    {method.value === 'scp' && '全量覆盖传输。简单直接，适合小规模产物发布。'}
                    {method.value === 'agent' && '最推荐模式。专用协议推送，支持跨内网、解压后脚本执行。'}
                  </p>
                </div>
              ))}
            </div>
          </Section>

          <Separator />

          {/* Agent */}
          <Section id="agent" title="独立 Agent 部署" icon={Rocket}>
            <p>
              Agent 是一个极小的二进制程序，部署在生产服务器上，作为 BuildFlow 主控的接收端。
            </p>
            <div className="space-y-4">
              <div className="space-y-2">
                <h4 className="text-sm font-medium text-foreground">启动参数示例：</h4>
                <div className="rounded-lg bg-black/90 p-4 font-mono text-xs text-white">
                  <div># 启动并设置监听端口与 Token</div>
                  <div className="mt-1">./buildflow-agent -addr :9091 -token YOUR_SECRET_TOKEN</div>
                </div>
              </div>
              <div className="rounded-lg border border-primary/20 bg-primary/5 p-4">
                <h4 className="mb-2 flex items-center gap-2 text-sm font-medium text-primary">
                  <ShieldCheck className="h-4 w-4" /> 为什么使用 Agent？
                </h4>
                <ul className="!list-none !pl-0 text-xs">
                  <li className="flex items-start gap-2">
                    <CheckCircle2 className="mt-0.5 h-3 w-3 shrink-0" />
                    <span><strong>安全性</strong>：无需开放 SSH 端口（22），可自定义高位端口。</span>
                  </li>
                  <li className="flex items-start gap-2 mt-1">
                    <CheckCircle2 className="mt-0.5 h-3 w-3 shrink-0" />
                    <span><strong>高性能</strong>：原生 HTTP 上传解压，比 SSH 协议封装更轻量。</span>
                  </li>
                  <li className="flex items-start gap-2 mt-1">
                    <CheckCircle2 className="mt-0.5 h-3 w-3 shrink-0" />
                    <span><strong>后置脚本</strong>：Agent 可在本地环境执行解压后的重启服务等操作。</span>
                  </li>
                </ul>
              </div>
            </div>
          </Section>
        </CardContent>
      </Card>

      {/* Footer */}
      <div className="flex flex-col items-center justify-center gap-4 text-center">
        <p className="text-sm text-muted-foreground">
          需要更多帮助？请查阅项目 <a href="https://github.com/cabinet-fe/MiniDevOps" className="font-medium text-primary underline underline-offset-4">README.md</a> 或联系管理员。
        </p>
        <div className="flex items-center gap-4">
          <Badge variant="secondary">Version 1.0.0</Badge>
          <Badge variant="outline">MIT Licensed</Badge>
        </div>
      </div>
    </div>
  )
}


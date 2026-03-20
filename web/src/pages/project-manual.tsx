import type { ReactNode } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import { BUILD_SCRIPT_TYPES, DEPLOY_METHODS, REPO_AUTH_TYPES, ARTIFACT_FORMATS } from '@/lib/constants'

const toc = [
  { id: 'new-project', label: '新建项目' },
  { id: 'repo-auth', label: '仓库认证（GitHub / 码云）' },
  { id: 'new-environment', label: '新建环境' },
  { id: 'deploy-methods', label: '部署方式与原理' },
  { id: 'standalone-agent', label: '独立部署 Agent' },
] as const

function Section({
  id,
  title,
  children,
}: {
  id: string
  title: string
  children: ReactNode
}) {
  return (
    <section id={id} className="scroll-mt-24 space-y-4">
      <h2 className="border-b border-border pb-2 text-lg font-semibold tracking-tight text-foreground">
        {title}
      </h2>
      <div className="space-y-3 text-sm leading-relaxed text-muted-foreground [&_strong]:font-medium [&_strong]:text-foreground [&_code]:rounded-md [&_code]:bg-muted/80 [&_code]:px-1.5 [&_code]:py-0.5 [&_code]:font-mono [&_code]:text-[13px] [&_code]:text-foreground">
        {children}
      </div>
    </section>
  )
}

export function ProjectManualPage() {
  return (
    <div className="space-y-8 pb-10">
      <div className="space-y-2">
        <h1 className="text-2xl font-semibold tracking-tight text-foreground">项目手册</h1>
        <p className="max-w-3xl text-sm text-muted-foreground">
          本文说明如何在 BuildFlow 中创建项目与环境、配置仓库访问、理解构建与部署流程，以及如何独立运行部署 Agent。
        </p>
      </div>

      <Card>
        <CardHeader className="pb-3">
          <CardTitle className="text-base">目录</CardTitle>
          <CardDescription>点击跳转到对应章节</CardDescription>
        </CardHeader>
        <CardContent>
          <nav className="flex flex-col gap-2 sm:flex-row sm:flex-wrap">
            {toc.map((item) => (
              <a
                key={item.id}
                href={`#${item.id}`}
                className="text-sm font-medium text-emerald-600/90 underline-offset-4 hover:underline dark:text-emerald-400/90"
              >
                {item.label}
              </a>
            ))}
          </nav>
        </CardContent>
      </Card>

      <Card className="gap-0">
        <CardContent className="space-y-10 pt-8">
          <Section id="new-project" title="新建项目">
            <p>
              在侧边栏进入<strong>项目</strong>，点击创建并填写基本信息。仓库地址请使用 <strong>HTTPS</strong> 形式（例如{' '}
              <code>https://github.com/org/repo.git</code>），并在项目中指定<strong>默认分支</strong>（构建时将检出该分支）。
            </p>
            <p>
              <strong>仓库认证</strong>对应界面中的选项：
            </p>
            <ul className="list-inside list-disc space-y-1 pl-1">
              {REPO_AUTH_TYPES.map((t) => (
                <li key={t.value}>
                  <strong>{t.label}</strong>
                  {t.value === 'none' && '：公开仓库或已在网络层放行时可选用。'}
                  {t.value === 'password' && '：在用户名、密码字段填写 HTTPS 基本认证所需信息。'}
                  {t.value === 'token' && '：在用户名、密码字段填写平台要求的凭据（常见为「用户名 + 令牌」组合，见下文）。'}
                </li>
              ))}
            </ul>
            <p>
              后端在拉取代码时会将凭据写入克隆 URL 的 <code>userinfo</code> 段，等价于 HTTPS 基本认证形式（形如{' '}
              <code>https://user:password@host/...</code>
              ）。实现见服务端 <code>buildAuthURL</code>（<code>internal/engine/git.go</code>）：无论选择「用户名/密码」还是「Token」，均通过{' '}
              <code>url.UserPassword(username, password)</code> 注入；因此请按各平台惯例把令牌放在<strong>密码</strong>字段（及需要的<strong>用户名</strong>字段）中。
            </p>
            <p>
              <strong>产物格式</strong>（项目级，在创建/编辑项目时选择）：
              {ARTIFACT_FORMATS.map((f) => (
                <span key={f.value}>
                  {' '}
                  <code>{f.label}</code>
                  {f.hint ? `（${f.hint}）` : ''}
                  {f.value !== ARTIFACT_FORMATS[ARTIFACT_FORMATS.length - 1].value ? '；' : '。'}
                </span>
              ))}
            </p>
          </Section>

          <Separator />

          <Section id="repo-auth" title="仓库认证（GitHub / 码云）">
            <div className="space-y-3">
              <h3 className="text-base font-medium text-foreground">GitHub</h3>
              <p>
                私有仓库通常使用 <strong>HTTPS + Personal Access Token</strong>。在 GitHub 设置中创建{' '}
                <strong>classic PAT</strong> 或 <strong>Fine-grained PAT</strong>：classic 需勾选仓库访问相关 scope（如{' '}
                <code>repo</code>
                ）；fine-grained 需为指定仓库授予 Contents 等只读/读写权限。
              </p>
              <p>
                在 BuildFlow 中选择<strong>Token</strong>时，常见填法为：用户名为你的 GitHub 用户名，密码为 PAT；亦有不少团队使用{' '}
                <code>x-access-token</code> 作为用户名、PAT 作为密码（以平台当前文档为准）。<strong>公开仓库</strong>可选用<strong>无需认证</strong>。
              </p>
            </div>
            <div className="space-y-3 pt-2">
              <h3 className="text-base font-medium text-foreground">码云（Gitee）</h3>
              <p>
                在码云「设置 → 安全设置 → 私人令牌」中生成令牌。BuildFlow 中选择<strong>Token</strong>后，将码云用户名填入<strong>用户名</strong>，将生成的私人令牌填入<strong>密码</strong>字段。企业版或私有库同样使用 HTTPS；请确保令牌具备对应仓库的拉取权限。
              </p>
            </div>
          </Section>

          <Separator />

          <Section id="new-environment" title="新建环境">
            <p>
              打开<strong>项目详情</strong>，在环境列表中新建或编辑环境。以下为与表单一致的主要字段（与「环境配置」对话框对应）：
            </p>
            <ul className="list-inside list-disc space-y-1 pl-1">
              <li>
                <strong>分支</strong>：构建时检出的 Git 分支（可从列表拉取远程分支）。
              </li>
              <li>
                <strong>构建脚本类型</strong>：
                {BUILD_SCRIPT_TYPES.map((t) => (
                  <span key={t.value}>
                    {' '}
                    <code>{t.label}</code>
                    {t.value !== BUILD_SCRIPT_TYPES[BUILD_SCRIPT_TYPES.length - 1].value ? '、' : ''}
                  </span>
                ))}
                。
              </li>
              <li>
                <strong>构建脚本</strong>：按所选类型编写构建命令或脚本内容。
              </li>
              <li>
                <strong>构建产物目录</strong>：构建完成后用于打包归档的目录（相对仓库路径）。
              </li>
              <li>
                <strong>部署服务器 / 部署路径 / 部署方式</strong>：选择已配置的服务器、目标路径，以及部署方式（见下一节）。
              </li>
              <li>
                <strong>部署后脚本</strong>：可选，在目标环境上于部署完成后执行。
              </li>
              <li>
                <strong>缓存路径</strong>：可选，用于在清理工作区时保留依赖缓存目录（多行路径）。
              </li>
              <li>
                <strong>定时构建</strong>：可开启 Cron，并填写 Cron 表达式（界面提供常用预设）。
              </li>
              <li>
                <strong>变量组</strong>：可勾选已在系统中维护的变量组，将其中变量注入构建环境。
              </li>
            </ul>
          </Section>

          <Separator />

          <Section id="deploy-methods" title="部署方式与原理">
            <p>
              一次构建的流水线顺序为：<strong>克隆/更新仓库 → 执行构建脚本 → 将构建产物目录打包归档 → 按所选方式部署</strong>。
            </p>
            <p>部署方式与界面中的名称一致：</p>
            <ul className="list-inside list-disc space-y-2 pl-1">
              <li>
                <strong>{DEPLOY_METHODS.find((m) => m.value === 'rsync')?.label}</strong>
                ：通过 SSH 在远端执行 rsync，适合 Linux 类主机、增量同步、带宽友好。
              </li>
              <li>
                <strong>{DEPLOY_METHODS.find((m) => m.value === 'sftp')?.label}</strong>
                ：基于 SFTP 上传文件，不依赖本机 rsync，适合受限环境。
              </li>
              <li>
                <strong>{DEPLOY_METHODS.find((m) => m.value === 'scp')?.label}</strong>
                ：一次性拷贝，简单直接，适合小体量或临时场景。
              </li>
              <li>
                <strong>{DEPLOY_METHODS.find((m) => m.value === 'agent')?.label}</strong>
                ：由运行中的 BuildFlow Agent 接收归档包并解压到指定目录；适合不便开放 SSH 或需统一由 Agent 落盘的场景。
              </li>
            </ul>
            <p>
              <strong>Agent 部署（主控 → Agent）</strong>：服务端将归档以 <code>POST</code> 请求上传到 Agent 的 <code>upload</code> 路径（由服务器上的「Agent
              URL」与路径拼接得到，见 <code>internal/deployer/agent.go</code>）。请求头包括：
            </p>
            <ul className="list-inside list-disc space-y-1 pl-1">
              <li>
                <code>Authorization: Bearer &lt;token&gt;</code>（与服务器配置中的 Agent Token 一致）
              </li>
              <li>
                <code>Content-Type</code>：与归档格式一致（如 gzip 为 <code>application/gzip</code>，zip 为 <code>application/zip</code>）
              </li>
              <li>
                <code>X-Archive-Format</code>：<code>gzip</code> 或 <code>zip</code>
              </li>
              <li>
                <code>X-Target-Path</code>：解压目标绝对路径（与环境中配置的部署路径一致）
              </li>
            </ul>
            <p>
              在<strong>服务器</strong>配置中，认证方式选择 Agent 时，需填写 <strong>Agent URL</strong>（指向 Agent 根地址即可，系统会拼接{' '}
              <code>/upload</code>
              ）与 <strong>Agent Token</strong>（与 Agent 启动时配置的 Bearer Token 一致）。
            </p>
          </Section>

          <Separator />

          <Section id="standalone-agent" title="独立部署 Agent">
            <p>
              BuildFlow 提供独立的 Agent 二进制（<code>cmd/agent</code>），用于在目标机器上接收构建产物。默认监听地址可通过环境变量{' '}
              <code>BUILDFLOW_AGENT_ADDR</code> 配置（默认 <code>:9091</code>），<strong>必须</strong>设置{' '}
              <code>BUILDFLOW_AGENT_TOKEN</code>（或通过命令行 <code>-token</code>）作为 Bearer 校验。
            </p>
            <p>
              若需 HTTPS，请设置 <code>BUILDFLOW_AGENT_TLS_CERT</code> 与 <code>BUILDFLOW_AGENT_TLS_KEY</code>（或命令行 <code>-tls-cert</code> /{' '}
              <code>-tls-key</code>），与 README 说明一致。Agent 暴露 <code>/healthz</code>、<code>/upload</code>、<code>/exec</code> 等路由，上传接口要求请求头{' '}
              <code>Authorization: Bearer &lt;token&gt;</code> 与 <code>X-Target-Path</code>。
            </p>
            <p>
              在 BuildFlow 的<strong>服务器</strong>中，将 <strong>Agent URL</strong> 填为可访问的 Agent 根 URL（例如{' '}
              <code>https://agent.example.com:9091</code>
              ），并保证与主控之间网络可达；生产环境建议使用 HTTPS 并配置防火墙仅放行主控出口 IP。
            </p>
          </Section>
        </CardContent>
      </Card>
    </div>
  )
}

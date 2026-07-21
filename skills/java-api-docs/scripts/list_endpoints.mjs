#!/usr/bin/env node

import path from 'path';
import { listEndpoints } from './lib/endpoints.mjs';

function usage() {
  console.error(`用法: node list_endpoints.mjs <srcRoot> [选项]

扫描 Spring @RestController / @Controller 的路径映射，并拼接网关前缀。
结果以紧凑 JSON 打印到标准输出。

选项:
  --files a.java,b.java   只扫这些文件
  --project <name>        项目名（参与网关服务匹配；默认不强制）
  --service <name>        显式 spring.application.name / 网关服务名
  --repo-root <dir>       git/仓根（用于读 application.yml；默认=srcRoot）
  --gateway-json <path>   网关路由 JSON（默认: 技能 references/ic-gateway-dev.json）
  --gateway <path>        同 --gateway-json
  -h, --help

输出要点:
  - endpoints[].servicePath  服务内 Controller 映射
  - endpoints[].path         对外完整 path = 网关前缀 + servicePath（未匹配时=servicePath）
  - endpoints[].docFile      建议文档名（kebab，如 sys-user.md）
  - controllers[]            按 Controller 分组，便于一文件一 md
  - gateway                  匹配结果；未匹配时含 warning，禁止手算前缀

文档布局: <out>/<project>/<kebab>.md（见 SKILL / changed_since / stamp_commit）`);
  process.exit(2);
}

function parseArgs(argv) {
  const args = argv.slice(2);
  if (!args.length || args[0] === '-h' || args[0] === '--help') usage();

  const positional = [];
  const opts = {
    files: null,
    project: null,
    service: null,
    repoRoot: null,
    gatewayJson: null,
  };
  for (let i = 0; i < args.length; i += 1) {
    const a = args[i];
    if (a === '--files') {
      const val = args[++i] || '';
      opts.files = val
        .split(',')
        .map((s) => s.trim())
        .filter(Boolean);
    } else if (a.startsWith('--files=')) {
      opts.files = a
        .slice('--files='.length)
        .split(',')
        .map((s) => s.trim())
        .filter(Boolean);
    } else if (a === '--project') {
      opts.project = args[++i];
    } else if (a.startsWith('--project=')) {
      opts.project = a.slice('--project='.length);
    } else if (a === '--service') {
      opts.service = args[++i];
    } else if (a.startsWith('--service=')) {
      opts.service = a.slice('--service='.length);
    } else if (a === '--repo-root' || a === '--repoRoot') {
      opts.repoRoot = args[++i];
    } else if (a.startsWith('--repo-root=')) {
      opts.repoRoot = a.slice('--repo-root='.length);
    } else if (a.startsWith('--repoRoot=')) {
      opts.repoRoot = a.slice('--repoRoot='.length);
    } else if (a === '--gateway-json' || a === '--gatewayJson' || a === '--gateway') {
      opts.gatewayJson = args[++i];
    } else if (a.startsWith('--gateway-json=')) {
      opts.gatewayJson = a.slice('--gateway-json='.length);
    } else if (a.startsWith('--gatewayJson=')) {
      opts.gatewayJson = a.slice('--gatewayJson='.length);
    } else if (a.startsWith('--gateway=')) {
      opts.gatewayJson = a.slice('--gateway='.length);
    } else if (a === '-h' || a === '--help') {
      usage();
    } else if (a.startsWith('-')) {
      console.error(`未知参数: ${a}`);
      usage();
    } else {
      positional.push(a);
    }
  }
  if (!positional.length) usage();
  return {
    srcRoot: positional[0],
    ...opts,
  };
}

function main() {
  const opts = parseArgs(process.argv);
  const result = listEndpoints(path.resolve(opts.srcRoot), {
    files: opts.files,
    project: opts.project,
    service: opts.service,
    repoRoot: opts.repoRoot ? path.resolve(opts.repoRoot) : null,
    gatewayJson: opts.gatewayJson,
  });
  process.stdout.write(`${JSON.stringify(result, null, 2)}\n`);
}

main();

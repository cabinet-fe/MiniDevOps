#!/usr/bin/env node

import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';
import { ensureDir } from './lib/fs-utils.mjs';
import {
  resolveOutRoot,
  displayPath,
  DEFAULT_OUT,
  CONVENTIONS_FILE,
} from './lib/out-paths.mjs';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const SKILL_ROOT = path.resolve(__dirname, '..');
const SOURCE = path.join(SKILL_ROOT, 'references', 'project-conventions.md');

function usage() {
  console.error(`用法: node ensure_conventions.mjs [--out <dir>] [--dry-run]

根据技能内 references/project-conventions.md 生成产出根下唯一的约定页：
  <out>/${CONVENTIONS_FILE}

选项:
  --out <dir>   输出根目录（默认: ${DEFAULT_OUT}，相对 process.cwd()；也可绝对路径）
  --dry-run     只打印将写入的路径与正文，不落盘
  -h, --help    显示帮助`);
  process.exit(2);
}

function parseArgs(argv) {
  const args = argv.slice(2);
  const opts = { out: null, dryRun: false };
  for (let i = 0; i < args.length; i += 1) {
    const a = args[i];
    if (a === '--dry-run') {
      opts.dryRun = true;
    } else if (a === '--out') {
      opts.out = args[++i];
    } else if (a.startsWith('--out=')) {
      opts.out = a.slice('--out='.length);
    } else if (a === '-h' || a === '--help') {
      usage();
    } else if (a.startsWith('-')) {
      console.error(`未知参数: ${a}`);
      usage();
    } else {
      console.error(`未知参数: ${a}`);
      usage();
    }
  }
  return opts;
}

/**
 * 将技能内规范投影为面向读者的约定页（去掉 Agent 指令段与脚本目录细节）。
 * @param {string} src
 */
function projectForReaders(src) {
  let body = src.replace(/\r\n/g, '\n').trim();
  // 去掉首行标题与紧随的 Agent 提示引用块
  body = body.replace(/^#[^\n]*\n+/, '');
  body = body.replace(/^>[^\n]*(?:\n>[^\n]*)*\n+/, '');
  // 去掉面向脚本/技能的「多仓库与文档目录」节，改写为读者向说明
  body = body.replace(/\n## 7\.[^\n]*[\s\S]*$/, '\n');

  const header = `# API 约定

> 由 java-api-docs 技能根据 \`references/project-conventions.md\` 生成。
> 单仓与多仓共用本文件；各项目接口文档通过相对链接引用（\`../${CONVENTIONS_FILE}\`）。

`;

  const footer = `
## 7. 文档目录

- 产出根：\`<out>/\`（默认工作区 \`${DEFAULT_OUT}/\`）
- 各项目接口文档：\`<out>/<project>/<kebab>.md\`（每个 Controller 一个文件）
- 本约定（唯一）：\`<out>/${CONVENTIONS_FILE}\`
`;

  return `${header}${body.trim()}\n${footer}`;
}

function main() {
  const opts = parseArgs(process.argv);
  if (!fs.existsSync(SOURCE)) {
    console.error('找不到规范源文件:', SOURCE);
    process.exit(1);
  }
  const outRoot = resolveOutRoot(opts.out);
  const dest = path.join(outRoot, CONVENTIONS_FILE);
  const text = projectForReaders(fs.readFileSync(SOURCE, 'utf8'));
  const shown = displayPath(process.cwd(), dest);

  if (opts.dryRun) {
    process.stdout.write(
      `${JSON.stringify(
        {
          dryRun: true,
          path: shown,
          outRoot: displayPath(process.cwd(), outRoot),
          bytes: Buffer.byteLength(text, 'utf8'),
          preview: text.slice(0, 400),
        },
        null,
        2,
      )}\n`,
    );
    return;
  }

  ensureDir(outRoot);
  fs.writeFileSync(dest, text, 'utf8');
  process.stdout.write(
    `${JSON.stringify(
      {
        written: true,
        path: shown,
        outRoot: displayPath(process.cwd(), outRoot),
        bytes: Buffer.byteLength(text, 'utf8'),
      },
      null,
      2,
    )}\n`,
  );
}

main();

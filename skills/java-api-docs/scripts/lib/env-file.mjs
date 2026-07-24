import fs from 'fs';
import path from 'path';

/**
 * 解析简单 dotenv 文本（KEY=value）。
 * 支持 # 注释、空行、可选引号；不覆盖调用方传入的已有键。
 * @param {string} text
 * @returns {Record<string, string>}
 */
function parseEnvText(text) {
  const out = {};
  const lines = String(text || '').replace(/\r\n/g, '\n').split('\n');
  for (const raw of lines) {
    const line = raw.trim();
    if (!line || line.startsWith('#')) continue;
    const eq = line.indexOf('=');
    if (eq <= 0) continue;
    const key = line.slice(0, eq).trim();
    if (!key || key.includes('\n')) continue;
    let value = line.slice(eq + 1);
    // 去掉行尾注释（未加引号时）
    if (!value.startsWith('"') && !value.startsWith("'")) {
      const hash = value.indexOf(' #');
      if (hash >= 0) value = value.slice(0, hash);
      value = value.trim();
    } else {
      value = value.trim();
    }
    value = unquoteEnvValue(value);
    out[key] = value;
  }
  return out;
}

/**
 * @param {string} value
 */
function unquoteEnvValue(value) {
  if (value.length >= 2) {
    const q = value[0];
    if ((q === '"' || q === "'") && value[value.length - 1] === q) {
      const inner = value.slice(1, -1);
      if (q === '"') {
        return inner
          .replace(/\\n/g, '\n')
          .replace(/\\r/g, '\r')
          .replace(/\\t/g, '\t')
          .replace(/\\\\/g, '\\')
          .replace(/\\"/g, '"');
      }
      return inner;
    }
  }
  return value;
}

/**
 * 按优先级解析 env 文件路径（第一个存在的）。
 * `--env-file` → `$BEDROCK_AGENT_ENV_FILE` → `$BEDROCK_AGENT_WORKDIR/.env` → `./.env`
 * @param {string|null|undefined} envFileOpt
 * @param {{ cwd?: string, env?: NodeJS.ProcessEnv }} [opts]
 * @returns {{ path: string|null, candidates: string[] }}
 */
function resolveEnvFilePath(envFileOpt, opts = {}) {
  const env = opts.env || process.env;
  const cwd = path.resolve(opts.cwd || process.cwd());
  const candidates = [];

  const explicit = envFileOpt != null && String(envFileOpt).trim() ? String(envFileOpt).trim() : null;
  if (explicit) {
    candidates.push(path.isAbsolute(explicit) ? path.resolve(explicit) : path.resolve(cwd, explicit));
  }
  if (env.BEDROCK_AGENT_ENV_FILE && String(env.BEDROCK_AGENT_ENV_FILE).trim()) {
    const p = String(env.BEDROCK_AGENT_ENV_FILE).trim();
    candidates.push(path.isAbsolute(p) ? path.resolve(p) : path.resolve(cwd, p));
  }
  if (env.BEDROCK_AGENT_WORKDIR && String(env.BEDROCK_AGENT_WORKDIR).trim()) {
    candidates.push(path.resolve(String(env.BEDROCK_AGENT_WORKDIR).trim(), '.env'));
  }
  candidates.push(path.join(cwd, '.env'));

  // 去重保序
  const seen = new Set();
  const unique = [];
  for (const c of candidates) {
    if (seen.has(c)) continue;
    seen.add(c);
    unique.push(c);
  }

  for (const p of unique) {
    try {
      if (fs.existsSync(p) && fs.statSync(p).isFile()) {
        return { path: p, candidates: unique };
      }
    } catch {
      // ignore
    }
  }
  return { path: null, candidates: unique };
}

/**
 * 读取 env 文件并 merge 进 target（默认 process.env）。
 * **不覆盖** target 中已存在的键（便于本机调试覆盖）。
 * @param {string|null|undefined} envFileOpt
 * @param {{ cwd?: string, target?: NodeJS.ProcessEnv, env?: NodeJS.ProcessEnv }} [opts]
 * @returns {{ loaded: boolean, path: string|null, keys: string[], candidates: string[] }}
 */
function loadEnvFile(envFileOpt, opts = {}) {
  const target = opts.target || process.env;
  const lookupEnv = opts.env || process.env;
  const resolved = resolveEnvFilePath(envFileOpt, { cwd: opts.cwd, env: lookupEnv });
  if (!resolved.path) {
    return { loaded: false, path: null, keys: [], candidates: resolved.candidates };
  }
  const text = fs.readFileSync(resolved.path, 'utf8');
  const parsed = parseEnvText(text);
  const keys = [];
  for (const [k, v] of Object.entries(parsed)) {
    if (Object.prototype.hasOwnProperty.call(target, k) && target[k] !== undefined) {
      continue;
    }
    target[k] = v;
    keys.push(k);
  }
  keys.sort();
  return { loaded: true, path: resolved.path, keys, candidates: resolved.candidates };
}

export { parseEnvText, resolveEnvFilePath, loadEnvFile, unquoteEnvValue };

import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const SKILL_ROOT = path.resolve(__dirname, '../..');
const DEFAULT_GATEWAY_JSON = path.join(SKILL_ROOT, 'references', 'ic-gateway-dev.json');

/**
 * 规范化服务名便于匹配：小写、_→-、去掉常见前缀变体比较时另算。
 * @param {string} name
 */
function normalizeServiceKey(name) {
  return String(name || '')
    .trim()
    .toLowerCase()
    .replace(/_/g, '-')
    .replace(/\s+/g, '');
}

/**
 * 候选名列表（含去 ic- 前缀 / -biz 后缀）。
 * @param {string} name
 */
function serviceNameCandidates(name) {
  const n = normalizeServiceKey(name);
  if (!n) return [];
  const out = [n];
  if (n.startsWith('ic-')) out.push(n.slice(3));
  if (n.endsWith('-biz')) out.push(n.slice(0, -4));
  if (n.startsWith('ic-') && n.endsWith('-biz')) out.push(n.slice(3, -4));
  return [...new Set(out.filter(Boolean))];
}

/**
 * 从仓库 / 源码树读取 spring.application.name。
 * @param {string} root
 * @returns {string|null}
 */
function readSpringApplicationName(root) {
  const abs = path.resolve(root);
  const candidates = [];
  const pushIf = (p) => {
    if (fs.existsSync(p)) candidates.push(p);
  };

  for (const rel of [
    'src/main/resources/application.yml',
    'src/main/resources/application.yaml',
    'src/main/resources/bootstrap.yml',
    'src/main/resources/bootstrap.yaml',
    'application.yml',
    'application.yaml',
  ]) {
    pushIf(path.join(abs, rel));
  }

  try {
    for (const ent of fs.readdirSync(abs, { withFileTypes: true })) {
      if (!ent.isDirectory()) continue;
      if (ent.name.startsWith('.') || ent.name === 'target' || ent.name === 'node_modules') {
        continue;
      }
      for (const f of ['application.yml', 'application.yaml', 'bootstrap.yml', 'bootstrap.yaml']) {
        pushIf(path.join(abs, ent.name, 'src', 'main', 'resources', f));
      }
    }
  } catch {
    /* ignore */
  }

  for (const file of candidates) {
    let text;
    try {
      text = fs.readFileSync(file, 'utf8');
    } catch {
      continue;
    }
    const flat = text.match(/^\s*spring\.application\.name\s*:\s*['"]?([^\s#'"]+)/m);
    if (flat) return flat[1].trim();
    const nested = text.match(
      /(?:^|\n)spring\s*:\s*\n(?:[ \t]+.+\n)*?[ \t]+application\s*:\s*\n(?:[ \t]+.+\n)*?[ \t]+name\s*:\s*['"]?([^\s#'"]+)/,
    );
    if (nested) return nested[1].trim();
  }
  return null;
}

/**
 * 解析网关 JSON 路径：显式参数 → 技能 references/ic-gateway-dev.json
 * @param {string|null|undefined} gatewayOpt
 */
function resolveGatewayJsonPath(gatewayOpt) {
  if (gatewayOpt != null && String(gatewayOpt).trim()) {
    const raw = String(gatewayOpt).trim();
    return path.isAbsolute(raw) ? path.resolve(raw) : path.resolve(process.cwd(), raw);
  }
  return DEFAULT_GATEWAY_JSON;
}

/**
 * 读取手维网关 JSON。
 * @param {string} jsonPath
 * @returns {{
 *   defaultStripPrefix: number|null,
 *   routes: Array<{
 *     id: string,
 *     label: string,
 *     service: string,
 *     prefix: string,
 *     gatewayPrefix: string,
 *     serviceName: string,
 *   }>,
 * }}
 */
function loadGatewayConfig(jsonPath) {
  const text = fs.readFileSync(jsonPath, 'utf8');
  const raw = JSON.parse(text);
  const defaultStripPrefix =
    raw.defaultStripPrefix != null ? Number(raw.defaultStripPrefix) : null;
  const routes = (raw.routes || []).map((r) => {
    const prefix = String(r.prefix || '').trim() || null;
    const service = String(r.service || '').trim();
    const id = String(r.id || '').trim();
    return {
      id,
      label: String(r.label || '').trim(),
      service,
      prefix,
      gatewayPrefix: prefix,
      serviceName: service || id,
    };
  });
  return { defaultStripPrefix, routes };
}

/**
 * 按服务名匹配网关前缀。
 *
 * 匹配顺序（任一命中即返回）：
 * 1. route.service
 * 2. route.id
 * 候选来自：显式 --service → spring.application.name → project 名（及去 ic- / -biz 变体）
 *
 * @param {{
 *   serviceNames?: string[],
 *   gatewayJson?: string|null,
 *   gateway?: string|null,
 * }} opts
 * @returns {{
 *   matched: boolean,
 *   gatewayPrefix: string|null,
 *   route: object|null,
 *   serviceName: string|null,
 *   triedNames: string[],
 *   gatewayJson: string,
 *   warning: string|null,
 *   defaultStripPrefix: number|null,
 * }}
 */
function resolveGatewayPrefix(opts = {}) {
  const gatewayJson = resolveGatewayJsonPath(opts.gatewayJson ?? opts.gateway);
  if (!fs.existsSync(gatewayJson)) {
    return {
      matched: false,
      gatewayPrefix: null,
      route: null,
      serviceName: null,
      triedNames: [],
      gatewayJson,
      warning: `网关配置不存在: ${gatewayJson}；文档 path 仅写服务内映射，并标注「网关前缀未配置」`,
      defaultStripPrefix: null,
    };
  }

  let parsed;
  try {
    parsed = loadGatewayConfig(gatewayJson);
  } catch (err) {
    return {
      matched: false,
      gatewayPrefix: null,
      route: null,
      serviceName: null,
      triedNames: [],
      gatewayJson,
      warning: `网关 JSON 解析失败: ${err.message}`,
      defaultStripPrefix: null,
    };
  }

  const tried = [];
  for (const n of opts.serviceNames || []) {
    for (const c of serviceNameCandidates(n)) {
      if (!tried.includes(c)) tried.push(c);
    }
  }

  for (const key of tried) {
    for (const route of parsed.routes) {
      const serviceKey = normalizeServiceKey(route.service || '');
      const idKey = normalizeServiceKey(route.id || '');
      const serviceCands = serviceNameCandidates(serviceKey);
      const idCands = serviceNameCandidates(idKey);
      if (
        serviceCands.includes(key) ||
        idCands.includes(key) ||
        serviceKey === key ||
        idKey === key
      ) {
        if (!route.gatewayPrefix) {
          return {
            matched: false,
            gatewayPrefix: null,
            route,
            serviceName: route.serviceName,
            triedNames: tried,
            gatewayJson,
            warning: `路由 ${route.id || route.service} 无可用 prefix；勿手算拼接`,
            defaultStripPrefix: parsed.defaultStripPrefix,
          };
        }
        return {
          matched: true,
          gatewayPrefix: route.gatewayPrefix,
          route,
          serviceName: route.serviceName,
          triedNames: tried,
          gatewayJson,
          warning: null,
          defaultStripPrefix: parsed.defaultStripPrefix,
        };
      }
    }
  }

  return {
    matched: false,
    gatewayPrefix: null,
    route: null,
    serviceName: tried[0] || null,
    triedNames: tried,
    gatewayJson,
    warning: `服务 [${tried.join(', ') || '(空)'}] 在网关 JSON 中未找到路由（查 service / id）。文档 path 使用服务内映射，并醒目标记「网关前缀未匹配」；禁止臆造前缀。`,
    defaultStripPrefix: parsed.defaultStripPrefix,
  };
}

/**
 * 拼接对外完整 path = 网关前缀 + 服务内 path。
 * @param {string|null} gatewayPrefix
 * @param {string} servicePath
 */
function joinGatewayPath(gatewayPrefix, servicePath) {
  const svc = String(servicePath || '').trim() || '/';
  if (!gatewayPrefix) return svc.startsWith('/') ? svc : `/${svc}`;
  const a = gatewayPrefix.replace(/\/+$/, '');
  const b = svc.replace(/^\/+/, '');
  if (!b || b === '') return a || '/';
  return `${a}/${b}`.replace(/\/+/g, '/');
}

/**
 * 收集用于匹配的服务名候选。
 * @param {{ repoRoot?: string, srcRoot?: string, service?: string|null, project?: string|null }} opts
 */
function collectServiceNames(opts = {}) {
  const names = [];
  if (opts.service) names.push(String(opts.service).trim());
  const roots = [opts.srcRoot, opts.repoRoot].filter(Boolean);
  for (const r of roots) {
    const app = readSpringApplicationName(r);
    if (app) names.push(app);
  }
  if (opts.project) names.push(String(opts.project).trim());
  return [...new Set(names.filter(Boolean))];
}

export {
  DEFAULT_GATEWAY_JSON,
  normalizeServiceKey,
  serviceNameCandidates,
  readSpringApplicationName,
  resolveGatewayJsonPath,
  loadGatewayConfig,
  resolveGatewayPrefix,
  joinGatewayPath,
  collectServiceNames,
};

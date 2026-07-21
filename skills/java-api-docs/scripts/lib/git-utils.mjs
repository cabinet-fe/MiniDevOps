import fs from 'fs';
import path from 'path';
import { execFileSync } from 'child_process';

function runGit(repoRoot, args, { allowFail = false } = {}) {
  try {
    const out = execFileSync('git', args, {
      cwd: repoRoot,
      encoding: 'utf8',
      stdio: ['ignore', 'pipe', 'pipe'],
    });
    return String(out).trimEnd();
  } catch (err) {
    if (allowFail) return null;
    const stderr = err.stderr ? String(err.stderr).trim() : err.message;
    throw new Error(`git ${args.join(' ')} 失败: ${stderr}`);
  }
}

function isGitRepo(repoRoot) {
  const out = runGit(repoRoot, ['rev-parse', '--is-inside-work-tree'], { allowFail: true });
  return out === 'true';
}

function gitHead(repoRoot) {
  return runGit(repoRoot, ['rev-parse', 'HEAD']);
}

function commitExists(repoRoot, sha) {
  if (!sha) return false;
  // rev-parse 比 cat-file 更轻；^{commit} 确保是 commit 对象
  const out = runGit(repoRoot, ['rev-parse', '--verify', `${sha}^{commit}`], {
    allowFail: true,
  });
  return Boolean(out);
}

function diffNameOnly(repoRoot, args) {
  const out = runGit(repoRoot, ['diff', '--name-only', ...args], { allowFail: true });
  if (!out) return [];
  return out.split('\n').map((l) => l.trim()).filter(Boolean);
}

function changedFilesSince(repoRoot, baseCommit) {
  const committed = diffNameOnly(repoRoot, [`${baseCommit}..HEAD`]);
  const unstaged = diffNameOnly(repoRoot, []);
  const staged = diffNameOnly(repoRoot, ['--cached']);
  const set = new Set([...committed, ...unstaged, ...staged]);
  return [...set].sort();
}

function todayYmd(d = new Date()) {
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, '0');
  const day = String(d.getDate()).padStart(2, '0');
  return `${y}-${m}-${day}`;
}

/**
 * 把 repoRoot 尽量显示成相对 workspace 的路径（写入 .sync.json 的 repoRel）。
 * @param {string} workspaceRoot
 * @param {string} repoRoot
 */
function repoRelFromWorkspace(workspaceRoot, repoRoot) {
  const ws = path.resolve(workspaceRoot || process.cwd());
  const repo = path.resolve(repoRoot);
  const rel = path.relative(ws, repo);
  if (!rel) return '.';
  if (rel.startsWith('..') || path.isAbsolute(rel)) return repo;
  return rel;
}

/**
 * 解析 sync.repoRel / 传入路径为绝对 repo 根。
 * @param {string} workspaceRoot
 * @param {string|null|undefined} repoRel
 */
function resolveRepoRel(workspaceRoot, repoRel) {
  if (repoRel == null || !String(repoRel).trim()) return null;
  const raw = String(repoRel).trim();
  if (path.isAbsolute(raw)) return path.resolve(raw);
  return path.resolve(workspaceRoot || process.cwd(), raw);
}

/**
 * 列出工作区下一层疑似 git 仓库目录（含自身）。
 * 优先匹配 `repo-*`、常见 checkout 名；避免深扫。
 * @param {string} workspaceRoot
 * @returns {string[]} 绝对路径
 */
function listNearbyGitRepos(workspaceRoot) {
  const root = path.resolve(workspaceRoot || process.cwd());
  const found = [];
  if (isGitRepo(root)) found.push(root);

  let entries;
  try {
    entries = fs.readdirSync(root);
  } catch {
    return found;
  }

  const preferred = entries
    .filter((n) => n === 'repo' || /^repo[-_]?\d+/i.test(n) || n.startsWith('repo-'))
    .sort();
  const others = entries
    .filter((n) => !preferred.includes(n) && !n.startsWith('.'))
    .sort();

  for (const name of [...preferred, ...others]) {
    const abs = path.join(root, name);
    let st;
    try {
      st = fs.statSync(abs);
    } catch {
      continue;
    }
    if (!st.isDirectory()) continue;
    if (isGitRepo(abs)) found.push(abs);
  }
  return [...new Set(found)];
}

/**
 * 在候选仓库中查找包含该 commit 的第一个仓库。
 * @param {string[]} repoRoots
 * @param {string} sha
 * @returns {string|null}
 */
function findRepoContainingCommit(repoRoots, sha) {
  if (!sha) return null;
  for (const root of repoRoots) {
    if (commitExists(root, sha)) return root;
  }
  return null;
}

export {
  runGit,
  isGitRepo,
  gitHead,
  commitExists,
  changedFilesSince,
  todayYmd,
  repoRelFromWorkspace,
  resolveRepoRel,
  listNearbyGitRepos,
  findRepoContainingCommit,
};

import fs from 'fs';
import path from 'path';

function isJavaFile(filePath) {
  return filePath.endsWith('.java');
}

function walkJavaFiles(rootDir, { filter } = {}) {
  const absRoot = path.resolve(rootDir);
  if (!fs.existsSync(absRoot)) {
    throw new Error(`路径不存在: ${absRoot}`);
  }
  const out = [];
  const stack = [absRoot];
  while (stack.length) {
    const cur = stack.pop();
    let st;
    try {
      st = fs.statSync(cur);
    } catch {
      continue;
    }
    if (st.isDirectory()) {
      const base = path.basename(cur);
      if (base === 'target' || base === 'node_modules' || base === '.git' || base === 'build') {
        continue;
      }
      let entries;
      try {
        entries = fs.readdirSync(cur);
      } catch {
        continue;
      }
      for (const name of entries) {
        stack.push(path.join(cur, name));
      }
    } else if (st.isFile() && isJavaFile(cur)) {
      if (!filter || filter(cur)) out.push(cur);
    }
  }
  out.sort();
  return out;
}

function readText(filePath) {
  return fs.readFileSync(filePath, 'utf8');
}

function ensureDir(dirPath) {
  fs.mkdirSync(dirPath, { recursive: true });
}

function writeJson(filePath, obj) {
  ensureDir(path.dirname(filePath));
  fs.writeFileSync(filePath, `${JSON.stringify(obj, null, 2)}\n`, 'utf8');
}

function readJsonIfExists(filePath) {
  if (!fs.existsSync(filePath)) return null;
  try {
    return JSON.parse(fs.readFileSync(filePath, 'utf8'));
  } catch (err) {
    throw new Error(`无效 JSON: ${filePath}: ${err.message}`);
  }
}

function resolveUnder(root, maybeRel) {
  if (path.isAbsolute(maybeRel)) return path.resolve(maybeRel);
  return path.resolve(root, maybeRel);
}

export {
  walkJavaFiles,
  readText,
  writeJson,
  readJsonIfExists,
  ensureDir,
  resolveUnder,
  isJavaFile,
};

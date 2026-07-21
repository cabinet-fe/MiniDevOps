import path from 'path';
import { walkJavaFiles, readText } from './fs-utils.mjs';
import {
  stripComments,
  mapJavaType,
  simpleTypeName,
  findAnnotationsBefore,
} from './java-parse.mjs';

function buildTypeIndex(srcRoot) {
  const root = path.resolve(srcRoot);
  const files = walkJavaFiles(root);
  const index = new Map(); // 简单类名 -> [{ file, packageName, fqcn, raw }]
  for (const f of files) {
    const raw = readText(f);
    const pkgMatch = raw.match(/^\s*package\s+([\w.]+)\s*;/m);
    const packageName = pkgMatch ? pkgMatch[1] : '';
    const classRe =
      /\b((?:public|protected|private)\s+)?((?:static\s+)?(?:abstract\s+|final\s+)?)(class|interface|enum|record)\s+([A-Za-z_][A-Za-z0-9_]*)/g;
    let m;
    const src = stripComments(raw);
    while ((m = classRe.exec(src))) {
      const name = m[4];
      const fqcn = packageName ? `${packageName}.${name}` : name;
      const entry = { file: f, relative: path.relative(root, f), packageName, fqcn, kind: m[3], raw, src };
      if (!index.has(name)) index.set(name, []);
      index.get(name).push(entry);
    }
  }
  return { root, index };
}

function pickType(index, typeName) {
  const simple = simpleTypeName(typeName) || typeName;
  const list = index.get(simple);
  if (!list || !list.length) return null;
  if (typeName.includes('.')) {
    const hit = list.find((e) => e.fqcn === typeName || e.fqcn.endsWith(`.${simple}`));
    return hit || list[0];
  }
  return list[0];
}

function parseFieldsFromClass(entry) {
  const { src } = entry;
  const classRe = new RegExp(
    `\\b(?:public|protected|private)?\\s*(?:static\\s+)?(?:abstract\\s+|final\\s+)?(?:class|record)\\s+${entry.fqcn.split('.').pop()}\\b`,
  );
  const cm = classRe.exec(src);
  if (!cm) return { fields: [], extendsType: null, lombok: false };

  const annos = findAnnotationsBefore(src, cm.index);
  const lombok = annos.some((a) =>
    ['Data', 'Value', 'Getter', 'Setter', 'Builder', 'SuperBuilder'].includes(a.name),
  );

  let extendsType = null;
  const after = src.slice(cm.index, cm.index + 400);
  const ext = after.match(/\bextends\s+([A-Za-z_][\w.<>,\s]*)/);
  if (ext) {
    extendsType = ext[1].trim().split(/\s+implements/)[0].trim();
  }

  const braceStart = src.indexOf('{', cm.index);
  if (braceStart < 0) return { fields: [], extendsType, lombok };

  // 类体：在深度 1 收集字段，跳过嵌套类型
  let depth = 0;
  let i = braceStart;
  const fields = [];
  while (i < src.length) {
    const c = src[i];
    if (c === '{') {
      depth += 1;
      i += 1;
      continue;
    }
    if (c === '}') {
      depth -= 1;
      if (depth === 0) break;
      i += 1;
      continue;
    }
    if (depth !== 1) {
      i += 1;
      continue;
    }
    // 尝试匹配字段：注解? 修饰符 类型 名称 [=...] ;
    const slice = src.slice(i);
    if (/^\/\//.test(slice) || /^\/\*/.test(slice)) {
      i += 1;
      continue;
    }
    const fieldMatch = slice.match(
      /^((?:@\w+(?:\((?:[^()]|\([^()]*\))*\))?\s*)*)((?:(?:public|protected|private|static|final|transient|volatile)\s+)*)([\w.<>,\s\[\]]+?)\s+([A-Za-z_][A-Za-z0-9_]*)\s*(?:=|;)/,
    );
    if (fieldMatch) {
      const mods = fieldMatch[2] || '';
      if (!/\bstatic\b/.test(mods)) {
        const typeExpr = fieldMatch[3].trim();
        const name = fieldMatch[4];
        if (!typeExpr.includes('(') && name !== 'serialVersionUID') {
          const schemaDesc = extractSchemaDescription(fieldMatch[1] || '');
          const mapped = mapJavaType(typeExpr);
          fields.push({
            name,
            ...mapped,
            required: true,
            description: schemaDesc || '',
          });
        }
      }
      i += fieldMatch[0].length;
      continue;
    }
    i += 1;
  }
  return { fields, extendsType, lombok };
}

function extractSchemaDescription(annoBlock) {
  const m = annoBlock.match(/@Schema\s*\(([^)]*)\)/);
  if (!m) return '';
  const dm = m[1].match(/description\s*=\s*"((?:\\.|[^"\\])*)"/);
  return dm ? dm[1] : '';
}

function resolveType(indexBundle, typeName, { depth = 0, seen = new Set() } = {}) {
  const simple = simpleTypeName(typeName) || typeName;
  const mapped = mapJavaType(typeName);

  // 基本类型 / 无本地源码的集合
  if (['string', 'number', 'boolean', 'null'].includes(mapped.type) && !mapped.needs_source) {
    return { name: simple, ...mapped, fields: [], resolved: true };
  }
  if (mapped.type.endsWith('[]') && mapped.item) {
    const itemResolved = resolveType(indexBundle, mapped.item.javaType || mapped.item.simple, {
      depth: depth + 1,
      seen,
    });
    return {
      name: simple,
      type: mapped.type,
      javaType: mapped.javaType,
      item: itemResolved,
      fields: [],
      resolved: true,
    };
  }
  if (mapped.type.startsWith('Record<') && mapped.value) {
    const valResolved = resolveType(indexBundle, mapped.value.javaType || mapped.value.simple, {
      depth: depth + 1,
      seen,
    });
    return {
      name: simple,
      type: mapped.type,
      javaType: mapped.javaType,
      value: valResolved,
      fields: [],
      resolved: true,
    };
  }

  if (depth > 6) {
    return { name: simple, ...mapped, fields: [], resolved: false, needs_source: true, reason: 'max_depth' };
  }
  if (seen.has(simple)) {
    return { name: simple, type: 'object', javaType: typeName, fields: [], resolved: false, circular: true };
  }

  const entry = pickType(indexBundle.index, typeName);
  if (!entry) {
    return {
      name: simple,
      type: mapped.type || 'object',
      javaType: typeName,
      fields: [],
      resolved: false,
      needs_source: true,
      reason: 'not_found_in_srcRoot',
    };
  }

  seen.add(simple);
  const parsed = parseFieldsFromClass(entry);
  let fields = [...parsed.fields];
  let extendsUnresolved = null;

  if (parsed.extendsType) {
    const parentSimple = simpleTypeName(parsed.extendsType);
    const parent = resolveType(indexBundle, parsed.extendsType, { depth: depth + 1, seen: new Set(seen) });
    if (parent.resolved && parent.fields) {
      fields = [...parent.fields, ...fields];
    } else {
      // 不要注入 _extends_* / _(继承)_ 这类假字段名；留给文档说明或项目规范
      extendsUnresolved = {
        simple: parentSimple,
        javaType: parsed.extendsType,
        needs_source: true,
        description: `继承自 ${parsed.extendsType}（本仓库无源码；字段见项目规范）`,
      };
    }
  }

  // 源码在本地时，充实嵌套 / 集合元素类型
  fields = fields.map((f) => {
    if (f.needs_source && f.simple && indexBundle.index.has(f.simple) && depth < 4) {
      const nested = resolveType(indexBundle, f.simple, { depth: depth + 1, seen: new Set(seen) });
      if (nested.resolved) {
        return { ...f, needs_source: false, fields: nested.fields, type: f.type };
      }
    }
    if (f.item && f.item.simple && indexBundle.index.has(f.item.simple) && depth < 4) {
      const nested = resolveType(indexBundle, f.item.simple, { depth: depth + 1, seen: new Set(seen) });
      if (nested.resolved) {
        return {
          ...f,
          item: { ...f.item, needs_source: false, fields: nested.fields, resolved: true },
        };
      }
    }
    return f;
  });

  return {
    name: simple,
    fqcn: entry.fqcn,
    file: entry.relative,
    type: 'object',
    javaType: entry.fqcn,
    lombok: parsed.lombok,
    extendsType: parsed.extendsType,
    extendsUnresolved,
    fields,
    resolved: true,
  };
}

function resolveTypes(srcRoot, typeNames) {
  const bundle = buildTypeIndex(srcRoot);
  const types = {};
  for (const name of typeNames) {
    const trimmed = name.trim();
    if (!trimmed) continue;
    types[trimmed] = resolveType(bundle, trimmed);
  }
  return { srcRoot: bundle.root, types };
}

export {
  buildTypeIndex,
  resolveTypes,
  resolveType,
};

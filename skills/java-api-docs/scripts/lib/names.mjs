/**
 * Java Controller 类名 → 文档文件名（kebab-case，去掉 Controller 后缀）。
 *
 * 例:
 *   FileController → file
 *   SysUserController → sys-user
 *   OAuth2ClientController → oauth2-client
 *   DynModuleListConfigController → dyn-module-list-config
 *   XMLParserController → xml-parser
 *
 * @param {string} className
 * @returns {string}
 */
function controllerToKebab(className) {
  let name = String(className || '').trim();
  if (!name) return 'unknown';
  if (/Controller$/i.test(name)) {
    name = name.replace(/Controller$/i, '');
  }
  if (!name) return 'controller';

  // 连续大写缩写（≥2）后接词首大写：XMLParser → XML-Parser
  // 要求前段至少 2 个大写，避免 OAuth → O-Auth
  name = name.replace(/([A-Z]{2,})([A-Z][a-z])/g, '$1-$2');
  // 小写或数字后接大写：SysUser → Sys-User；OAuth2Client → OAuth2-Client
  name = name.replace(/([a-z0-9])([A-Z])/g, '$1-$2');

  return name
    .replace(/_+/g, '-')
    .replace(/-+/g, '-')
    .replace(/^-|-$/g, '')
    .toLowerCase();
}

/**
 * 文档相对项目目录的文件名：`<kebab>.md`
 * @param {string} className
 */
function controllerDocFileName(className) {
  return `${controllerToKebab(className)}.md`;
}

export { controllerToKebab, controllerDocFileName };

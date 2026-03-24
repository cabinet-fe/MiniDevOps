/// <reference types="vite/client" />

interface ImportMetaEnv {
  /** 与后端 `encryption.key` 相同的 64 位 hex；dev / 非 Go 托管时用于登录加密 */
  readonly VITE_BUILDFLOW_ENCRYPTION_KEY?: string;
  readonly VITE_APP_VERSION?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}

declare global {
  interface Window {
    /** 嵌入二进制由服务端在 index.html 中注入，优先于 VITE_*，与运行时 encryption.key 一致 */
    __BUILDFLOW_ENCRYPTION_KEY__?: string;
  }
}

export {};

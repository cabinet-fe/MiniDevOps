import { AES, CBC, Hex, Pkcs7, Utf8, WordArray } from "crypto-es";

function isValidHexKey64(s: string): boolean {
  return /^[0-9a-fA-F]{64}$/.test(s);
}

function hexToBytes(hex: string): Uint8Array {
  const out = new Uint8Array(hex.length / 2);
  for (let i = 0; i < out.length; i++) {
    out[i] = parseInt(hex.slice(i * 2, i * 2 + 2), 16);
  }
  return out;
}

function bytesToHex(bytes: Uint8Array): string {
  let s = "";
  for (let i = 0; i < bytes.length; i++) {
    s += bytes[i]!.toString(16).padStart(2, "0");
  }
  return s;
}

function randomIV(): Uint8Array {
  const iv = new Uint8Array(16);
  crypto.getRandomValues(iv);
  return iv;
}

/** 安全上下文（HTTPS、localhost 等）下可用 Web Crypto；纯 HTTP 内网等非安全上下文用 crypto-es。 */
function useSubtleForLoginEncryption(): boolean {
  return typeof crypto !== "undefined" && typeof crypto.subtle !== "undefined";
}

function getEncryptionKeyHex(): string {
  // 1) 嵌入二进制：Go 在返回的 index.html 中注入，与运行时 config encryption.key 一致
  if (typeof window !== "undefined") {
    const w = window as Window & { __BUILDFLOW_ENCRYPTION_KEY__?: string };
    const injected = typeof w.__BUILDFLOW_ENCRYPTION_KEY__ === "string" ? w.__BUILDFLOW_ENCRYPTION_KEY__.trim() : "";
    if (isValidHexKey64(injected)) {
      return injected;
    }
  }
  // 2) Vite 编译时注入（dev、vite preview、非 Go 托管的静态资源等）
  const vite = import.meta.env.VITE_BUILDFLOW_ENCRYPTION_KEY;
  const fromEnv = typeof vite === "string" ? vite.trim() : "";
  if (isValidHexKey64(fromEnv)) {
    return fromEnv;
  }
  throw new Error(
    "登录需要有效的加密密钥（64 位 hex，与后端 encryption.key 一致）：嵌入部署由服务端注入 window.__BUILDFLOW_ENCRYPTION_KEY__，本地开发可设 VITE_BUILDFLOW_ENCRYPTION_KEY",
  );
}

function getEncryptionKeyBytes(): Uint8Array {
  const trimmed = getEncryptionKeyHex();
  const keyBytes = hexToBytes(trimmed);
  if (keyBytes.length !== 32) {
    throw new Error("加密密钥长度应为 32 字节（64 hex 字符）");
  }
  return keyBytes;
}

async function encryptSubtle(plain: string, keyBytes: Uint8Array): Promise<string> {
  const iv = randomIV();
  const key = await crypto.subtle.importKey("raw", keyBytes, "AES-CBC", false, ["encrypt"]);
  const ciphertext = await crypto.subtle.encrypt(
    { name: "AES-CBC", iv },
    key,
    new TextEncoder().encode(plain),
  );
  const ct = new Uint8Array(ciphertext);
  const combined = new Uint8Array(iv.length + ct.length);
  combined.set(iv, 0);
  combined.set(ct, iv.length);
  return bytesToHex(combined);
}

function encryptCryptoEs(plain: string, keyBytes: Uint8Array): string {
  const iv = randomIV();
  const keyWA = new WordArray(keyBytes);
  const ivWA = new WordArray(iv);
  const encrypted = AES.encrypt(Utf8.parse(plain), keyWA, {
    iv: ivWA,
    mode: CBC,
    padding: Pkcs7,
  });
  const ctHex = encrypted.ciphertext.toString(Hex);
  return bytesToHex(iv) + ctHex;
}

/**
 * AES-256-CBC，输出 `hex(IV(16 字节) || PKCS#7 密文)`，与后端 `DecryptLoginPasswordCipher` 一致。
 * 无有效密钥或加密失败时抛错，不再回退为明文 `password`。
 */
export async function encryptLoginPassword(plain: string): Promise<string> {
  const keyBytes = getEncryptionKeyBytes();
  if (useSubtleForLoginEncryption()) {
    return encryptSubtle(plain, keyBytes);
  }
  return encryptCryptoEs(plain, keyBytes);
}

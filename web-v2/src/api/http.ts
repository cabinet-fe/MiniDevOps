import { HTTPClient, TokenPlugin } from "@cat-kit/http";

import type { ApiEnvelope, TokenPair } from "./types";

const ACCESS_KEY = "access_token";
const REFRESH_KEY = "refresh_token";

export function getAccessToken(): string | null {
  return localStorage.getItem(ACCESS_KEY);
}

export function getRefreshToken(): string | null {
  return localStorage.getItem(REFRESH_KEY);
}

export function setTokens(access: string, refresh: string): void {
  localStorage.setItem(ACCESS_KEY, access);
  localStorage.setItem(REFRESH_KEY, refresh);
}

export function clearTokens(): void {
  localStorage.removeItem(ACCESS_KEY);
  localStorage.removeItem(REFRESH_KEY);
}

/** Bare client for login/refresh (no Bearer injection). */
export const bareHttp = new HTTPClient("/api/v1", {
  timeout: 30_000,
  credentials: false,
});

let onAuthExpired: (() => void) | null = null;

export function setOnAuthExpired(cb: (() => void) | null): void {
  onAuthExpired = cb;
}

async function refreshAccessToken(): Promise<void> {
  const refresh = getRefreshToken();
  if (!refresh) {
    throw new Error("missing refresh token");
  }
  const res = await bareHttp.post<ApiEnvelope<TokenPair>>("/auth/refresh", {
    refresh_token: refresh,
  });
  const body = res.body;
  if (res.code !== 200 || !body || body.code !== 0 || !body.data) {
    throw new Error(body?.message || "refresh failed");
  }
  setTokens(body.data.access_token, body.data.refresh_token);
}

export const http = new HTTPClient("/api/v1", {
  timeout: 30_000,
  credentials: false,
  plugins: [
    TokenPlugin({
      getter: () => getAccessToken(),
      authType: "Bearer",
      maxRetries: 1,
      onRefresh: refreshAccessToken,
      shouldRefresh: (response) => response.code === 401,
      onRefreshExpired: () => {
        clearTokens();
        onAuthExpired?.();
      },
      isRefreshExpired: () => !getRefreshToken(),
    }),
  ],
});

/** Unwrap Bedrock envelope; throw on business or HTTP failure. */
export async function apiData<T>(
  promise: Promise<{ code: number; body: ApiEnvelope<T> }>,
): Promise<T> {
  const res = await promise;
  const body = res.body;
  if (res.code < 200 || res.code >= 300) {
    throw new Error(body?.message || `HTTP ${res.code}`);
  }
  if (!body || body.code !== 0) {
    throw new Error(body?.message || "request failed");
  }
  return body.data as T;
}

import { apiData, bareHttp, clearTokens, http, setTokens } from "./http";
import type { MeResponse, TokenPair, User } from "./types";

export async function loginApi(username: string, passwordCipher: string): Promise<TokenPair> {
  return apiData(
    bareHttp.post("/auth/login", {
      username,
      password_cipher: passwordCipher,
    }),
  );
}

export async function logoutApi(): Promise<void> {
  await apiData(http.post("/auth/logout"));
}

export async function meApi(): Promise<MeResponse> {
  return apiData(http.get("/auth/me"));
}

export { clearTokens, setTokens };
export type { MeResponse, TokenPair, User };

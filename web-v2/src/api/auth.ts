import { bareHttp, clearTokens, http, setTokens } from "./http";
import type { MeResponse, TokenPair, User } from "./types";

export async function loginApi(username: string, passwordCipher: string): Promise<TokenPair> {
  const { body } = await bareHttp.post<TokenPair>("/auth/login", {
    username,
    password_cipher: passwordCipher,
  });
  return body;
}

export async function logoutApi(): Promise<void> {
  await http.post("/auth/logout");
}

export async function meApi(): Promise<MeResponse> {
  const { body } = await http.get<MeResponse>("/auth/me");
  return body;
}

export { clearTokens, setTokens };
export type { MeResponse, TokenPair, User };

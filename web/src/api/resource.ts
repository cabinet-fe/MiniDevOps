import { http } from "./http";
import type { Credential, PageResult, Repository, Server } from "./types";

export type ListQuery = Record<string, string | number | boolean | undefined | null>;

function toQuery(params?: ListQuery): Record<string, string | number | boolean> {
  const out: Record<string, string | number | boolean> = {};
  if (!params) return out;
  for (const [k, v] of Object.entries(params)) {
    if (v === undefined || v === null || v === "") continue;
    out[k] = v;
  }
  return out;
}

// —— Credentials ——
export async function listCredentials(params?: ListQuery): Promise<PageResult<Credential>> {
  const { body } = await http.get<PageResult<Credential>>("/resource/credentials", {
    query: toQuery(params),
  });
  return body;
}

export async function createCredential(body: Record<string, unknown>): Promise<Credential> {
  const { body: data } = await http.post<Credential>("/resource/credentials", body);
  return data;
}

export async function updateCredential(
  id: number,
  body: Record<string, unknown>,
): Promise<Credential> {
  const { body: data } = await http.put<Credential>(`/resource/credentials/${id}`, body);
  return data;
}

export async function deleteCredential(id: number): Promise<void> {
  await http.delete(`/resource/credentials/${id}`);
}

// —— Repositories ——
export async function listRepositories(params?: ListQuery): Promise<PageResult<Repository>> {
  const { body } = await http.get<PageResult<Repository>>("/resource/repositories", {
    query: toQuery(params),
  });
  return body;
}

export async function listRepositoryBranches(id: number): Promise<string[]> {
  const { body } = await http.get<{ items: string[] }>(`/resource/repositories/${id}/branches`);
  return body.items ?? [];
}

export async function createRepository(body: Record<string, unknown>): Promise<Repository> {
  const { body: data } = await http.post<Repository>("/resource/repositories", body);
  return data;
}

export async function updateRepository(
  id: number,
  body: Record<string, unknown>,
): Promise<Repository> {
  const { body: data } = await http.put<Repository>(`/resource/repositories/${id}`, body);
  return data;
}

export async function deleteRepository(id: number): Promise<void> {
  await http.delete(`/resource/repositories/${id}`);
}

export async function testRepository(id: number): Promise<{ ok: boolean; branches?: string[] }> {
  const { body } = await http.post<{ ok: boolean; branches?: string[] }>(
    `/resource/repositories/${id}/test`,
    {},
  );
  return body;
}

// —— Servers ——
export async function listServers(params?: ListQuery): Promise<PageResult<Server>> {
  const { body } = await http.get<PageResult<Server>>("/resource/servers", {
    query: toQuery(params),
  });
  return body;
}

export async function createServer(body: Record<string, unknown>): Promise<Server> {
  const { body: data } = await http.post<Server>("/resource/servers", body);
  return data;
}

export async function updateServer(id: number, body: Record<string, unknown>): Promise<Server> {
  const { body: data } = await http.put<Server>(`/resource/servers/${id}`, body);
  return data;
}

export async function deleteServer(id: number): Promise<void> {
  await http.delete(`/resource/servers/${id}`);
}

export async function testServer(id: number): Promise<{ ok: boolean; output?: string }> {
  const { body } = await http.post<{ ok: boolean; output?: string }>(
    `/resource/servers/${id}/test`,
    {},
  );
  return body;
}

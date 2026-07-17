import { getAccessToken, http } from "./http";
import type { BuildJob, BuildRun, Credential, PageResult, Repository, Server } from "./types";

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
  const { body } = await http.get<PageResult<Credential>>("/credentials", {
    query: toQuery(params),
  });
  return body;
}

export async function createCredential(body: Record<string, unknown>): Promise<Credential> {
  const { body: data } = await http.post<Credential>("/credentials", body);
  return data;
}

export async function updateCredential(
  id: number,
  body: Record<string, unknown>,
): Promise<Credential> {
  const { body: data } = await http.put<Credential>(`/credentials/${id}`, body);
  return data;
}

export async function deleteCredential(id: number): Promise<void> {
  await http.delete(`/credentials/${id}`);
}

// —— Repositories ——
export async function listRepositories(params?: ListQuery): Promise<PageResult<Repository>> {
  const { body } = await http.get<PageResult<Repository>>("/repositories", {
    query: toQuery(params),
  });
  return body;
}

export async function listRepositoryBranches(id: number): Promise<string[]> {
  const { body } = await http.get<{ items: string[] }>(`/repositories/${id}/branches`);
  return body.items ?? [];
}

export async function createRepository(body: Record<string, unknown>): Promise<Repository> {
  const { body: data } = await http.post<Repository>("/repositories", body);
  return data;
}

export async function updateRepository(
  id: number,
  body: Record<string, unknown>,
): Promise<Repository> {
  const { body: data } = await http.put<Repository>(`/repositories/${id}`, body);
  return data;
}

export async function deleteRepository(id: number): Promise<void> {
  await http.delete(`/repositories/${id}`);
}

export async function testRepository(id: number): Promise<{ ok: boolean; branches?: string[] }> {
  const { body } = await http.post<{ ok: boolean; branches?: string[] }>(
    `/repositories/${id}/test`,
    {},
  );
  return body;
}

// —— Servers ——
export async function listServers(params?: ListQuery): Promise<PageResult<Server>> {
  const { body } = await http.get<PageResult<Server>>("/servers", { query: toQuery(params) });
  return body;
}

export async function createServer(body: Record<string, unknown>): Promise<Server> {
  const { body: data } = await http.post<Server>("/servers", body);
  return data;
}

export async function updateServer(id: number, body: Record<string, unknown>): Promise<Server> {
  const { body: data } = await http.put<Server>(`/servers/${id}`, body);
  return data;
}

export async function deleteServer(id: number): Promise<void> {
  await http.delete(`/servers/${id}`);
}

export async function testServer(id: number): Promise<{ ok: boolean; output?: string }> {
  const { body } = await http.post<{ ok: boolean; output?: string }>(`/servers/${id}/test`, {});
  return body;
}

// —— Build jobs ——
export async function listBuildJobs(params?: ListQuery): Promise<PageResult<BuildJob>> {
  const { body } = await http.get<PageResult<BuildJob>>("/build-jobs", { query: toQuery(params) });
  return body;
}

export async function getBuildJob(id: number): Promise<BuildJob> {
  const { body } = await http.get<BuildJob>(`/build-jobs/${id}`);
  return body;
}

export async function createBuildJob(body: Record<string, unknown>): Promise<BuildJob> {
  const { body: data } = await http.post<BuildJob>("/build-jobs", body);
  return data;
}

export async function updateBuildJob(id: number, body: Record<string, unknown>): Promise<BuildJob> {
  const { body: data } = await http.put<BuildJob>(`/build-jobs/${id}`, body);
  return data;
}

export async function deleteBuildJob(id: number): Promise<void> {
  await http.delete(`/build-jobs/${id}`);
}

export async function getBuildJobWebhookSecret(
  id: number,
): Promise<{ webhook_secret: string; webhook_url: string }> {
  const { body } = await http.get<{ webhook_secret: string; webhook_url: string }>(
    `/build-jobs/${id}/webhook-secret`,
  );
  return body;
}

export async function rotateBuildJobWebhookSecret(
  id: number,
): Promise<{ webhook_secret: string; webhook_url: string }> {
  const { body } = await http.post<{ webhook_secret: string; webhook_url: string }>(
    `/build-jobs/${id}/webhook-secret/rotate`,
    {},
  );
  return body;
}

export async function enqueueBuildRun(
  jobId: number,
  body?: Record<string, unknown>,
): Promise<BuildRun> {
  const { body: data } = await http.post<BuildRun>(`/build-jobs/${jobId}/runs`, body ?? {});
  return data;
}

// —— Build runs ——
export async function getBuildRun(id: number): Promise<BuildRun> {
  const { body } = await http.get<BuildRun>(`/build-runs/${id}`);
  return body;
}

export async function cancelBuildRun(id: number): Promise<BuildRun> {
  const { body } = await http.post<BuildRun>(`/build-runs/${id}/cancel`, {});
  return body;
}

export async function retryBuildRun(id: number): Promise<BuildRun> {
  const { body } = await http.post<BuildRun>(`/build-runs/${id}/retry`, {});
  return body;
}

export async function redeployBuildRun(
  id: number,
  body?: { target_ids?: number[] },
): Promise<BuildRun> {
  const { body: data } = await http.post<BuildRun>(`/build-runs/${id}/redeploy`, body ?? {});
  return data;
}

/** Artifact download URL (Bearer via browser navigation with token query is not used; open with fetch blob). */
export function buildRunArtifactURL(id: number): string {
  return `/api/v1/build-runs/${id}/artifact`;
}

export async function getBuildRunLog(id: number): Promise<string> {
  const token = getAccessToken();
  const res = await fetch(`/api/v1/build-runs/${id}/log`, {
    headers: token ? { Authorization: `Bearer ${token}` } : {},
  });
  if (!res.ok) {
    throw new Error(`HTTP ${res.status}`);
  }
  return res.text();
}

export function buildRunLogsWSURL(id: number, token: string): string {
  const proto = location.protocol === "https:" ? "wss:" : "ws:";
  return `${proto}//${location.host}/ws/build-runs/${id}/logs?token=${encodeURIComponent(token)}`;
}

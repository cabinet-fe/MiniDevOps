import { apiData, getAccessToken, http } from "./http";
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
  return apiData(http.get("/credentials", { query: toQuery(params) }));
}

export async function createCredential(body: Record<string, unknown>): Promise<Credential> {
  return apiData(http.post("/credentials", body));
}

export async function updateCredential(
  id: number,
  body: Record<string, unknown>,
): Promise<Credential> {
  return apiData(http.put(`/credentials/${id}`, body));
}

export async function deleteCredential(id: number): Promise<void> {
  await apiData(http.delete(`/credentials/${id}`));
}

// —— Repositories ——
export async function listRepositories(params?: ListQuery): Promise<PageResult<Repository>> {
  return apiData(http.get("/repositories", { query: toQuery(params) }));
}

export async function getRepository(id: number): Promise<Repository> {
  return apiData(http.get(`/repositories/${id}`));
}

export async function createRepository(body: Record<string, unknown>): Promise<Repository> {
  return apiData(http.post("/repositories", body));
}

export async function updateRepository(
  id: number,
  body: Record<string, unknown>,
): Promise<Repository> {
  return apiData(http.put(`/repositories/${id}`, body));
}

export async function deleteRepository(id: number): Promise<void> {
  await apiData(http.delete(`/repositories/${id}`));
}

export async function testRepository(id: number): Promise<{ ok: boolean; branches?: string[] }> {
  return apiData(http.post(`/repositories/${id}/test`, {}));
}

export async function getWebhookSecret(
  id: number,
): Promise<{ webhook_secret: string; webhook_url: string }> {
  return apiData(http.get(`/repositories/${id}/webhook-secret`));
}

export async function rotateWebhookSecret(
  id: number,
): Promise<{ webhook_secret: string; webhook_url: string }> {
  return apiData(http.post(`/repositories/${id}/webhook-secret/rotate`, {}));
}

// —— Servers ——
export async function listServers(params?: ListQuery): Promise<PageResult<Server>> {
  return apiData(http.get("/servers", { query: toQuery(params) }));
}

export async function createServer(body: Record<string, unknown>): Promise<Server> {
  return apiData(http.post("/servers", body));
}

export async function updateServer(id: number, body: Record<string, unknown>): Promise<Server> {
  return apiData(http.put(`/servers/${id}`, body));
}

export async function deleteServer(id: number): Promise<void> {
  await apiData(http.delete(`/servers/${id}`));
}

export async function testServer(id: number): Promise<{ ok: boolean; output?: string }> {
  return apiData(http.post(`/servers/${id}/test`, {}));
}

// —— Build jobs ——
export async function listBuildJobs(params?: ListQuery): Promise<PageResult<BuildJob>> {
  return apiData(http.get("/build-jobs", { query: toQuery(params) }));
}

export async function getBuildJob(id: number): Promise<BuildJob> {
  return apiData(http.get(`/build-jobs/${id}`));
}

export async function createBuildJob(body: Record<string, unknown>): Promise<BuildJob> {
  return apiData(http.post("/build-jobs", body));
}

export async function updateBuildJob(id: number, body: Record<string, unknown>): Promise<BuildJob> {
  return apiData(http.put(`/build-jobs/${id}`, body));
}

export async function deleteBuildJob(id: number): Promise<void> {
  await apiData(http.delete(`/build-jobs/${id}`));
}

export async function enqueueBuildRun(
  jobId: number,
  body?: Record<string, unknown>,
): Promise<BuildRun> {
  return apiData(http.post(`/build-jobs/${jobId}/runs`, body ?? {}));
}

// —— Build runs ——
export async function listBuildRuns(params?: ListQuery): Promise<PageResult<BuildRun>> {
  return apiData(http.get("/build-runs", { query: toQuery(params) }));
}

export async function getBuildRun(id: number): Promise<BuildRun> {
  return apiData(http.get(`/build-runs/${id}`));
}

export async function cancelBuildRun(id: number): Promise<BuildRun> {
  return apiData(http.post(`/build-runs/${id}/cancel`, {}));
}

export async function retryBuildRun(id: number): Promise<BuildRun> {
  return apiData(http.post(`/build-runs/${id}/retry`, {}));
}

export async function redeployBuildRun(
  id: number,
  body?: { target_ids?: number[] },
): Promise<BuildRun> {
  return apiData(http.post(`/build-runs/${id}/redeploy`, body ?? {}));
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

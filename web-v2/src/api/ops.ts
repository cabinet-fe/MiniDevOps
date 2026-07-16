import { http } from "./http";
import type {
  InstallSource,
  PageResult,
  ProcessInfo,
  ToolchainDefinition,
  ToolchainInstallJob,
} from "./types";

type Query = Record<string, string | number | boolean | undefined>;

function compactQuery(query?: Query): Record<string, string | number | boolean> {
  return Object.fromEntries(
    Object.entries(query ?? {}).filter(([, value]) => value !== undefined && value !== ""),
  ) as Record<string, string | number | boolean>;
}

export async function listProcesses(query?: Query): Promise<ProcessInfo[]> {
  const { body } = await http.get<{ items: ProcessInfo[] }>("/ops/processes", {
    query: compactQuery(query),
  });
  return body.items;
}

export async function killProcess(pid: number): Promise<void> {
  await http.post(`/ops/processes/${pid}/kill`, {});
}

export async function listToolchains(): Promise<ToolchainDefinition[]> {
  const { body } = await http.get<{ items: ToolchainDefinition[] }>("/ops/toolchains");
  return body.items;
}

export async function createToolchain(
  input: Record<string, unknown>,
): Promise<ToolchainDefinition> {
  const { body } = await http.post<ToolchainDefinition>("/ops/toolchains", input);
  return body;
}

export async function updateToolchain(
  id: number,
  input: Record<string, unknown>,
): Promise<ToolchainDefinition> {
  const { body } = await http.put<ToolchainDefinition>(`/ops/toolchains/${id}`, input);
  return body;
}

export async function deleteToolchain(id: number): Promise<void> {
  await http.delete(`/ops/toolchains/${id}`);
}

export async function detectToolchain(id: number): Promise<{ detected: boolean; output: string }> {
  const { body } = await http.post<{ detected: boolean; output: string }>(
    `/ops/toolchains/${id}/detect`,
    {},
  );
  return body;
}

export async function enqueueToolchainOperation(
  id: number,
  operation: "install" | "upgrade" | "uninstall" | "switch",
  version = "",
): Promise<ToolchainInstallJob> {
  const { body } = await http.post<ToolchainInstallJob>(`/ops/toolchains/${id}/${operation}`, {
    version,
  });
  return body;
}

export async function listInstallSources(): Promise<InstallSource[]> {
  const { body } = await http.get<{ items: InstallSource[] }>("/ops/install-sources");
  return body.items;
}

export async function createInstallSource(input: Record<string, unknown>): Promise<InstallSource> {
  const { body } = await http.post<InstallSource>("/ops/install-sources", input);
  return body;
}

export async function updateInstallSource(
  id: number,
  input: Record<string, unknown>,
): Promise<InstallSource> {
  const { body } = await http.put<InstallSource>(`/ops/install-sources/${id}`, input);
  return body;
}

export async function deleteInstallSource(id: number): Promise<void> {
  await http.delete(`/ops/install-sources/${id}`);
}

export async function pingInstallSource(id: number): Promise<{ ok: boolean; detail: string }> {
  const { body } = await http.post<{ ok: boolean; detail: string }>(
    `/ops/install-sources/${id}/ping`,
    {},
  );
  return body;
}

export async function listInstallJobs(query?: Query): Promise<PageResult<ToolchainInstallJob>> {
  const { body } = await http.get<PageResult<ToolchainInstallJob>>("/ops/install-jobs", {
    query: compactQuery(query),
  });
  return body;
}

export async function getInstallJob(id: number): Promise<ToolchainInstallJob> {
  const { body } = await http.get<ToolchainInstallJob>(`/ops/install-jobs/${id}`);
  return body;
}

export async function getInstallJobLogs(id: number): Promise<string> {
  const { body } = await http.get<string>(`/ops/install-jobs/${id}/logs`);
  return body;
}

export async function retryInstallJob(id: number): Promise<ToolchainInstallJob> {
  const { body } = await http.post<ToolchainInstallJob>(`/ops/install-jobs/${id}/retry`, {});
  return body;
}

import { http } from "./http";
import type {
  AgentRun,
  AgentTrigger,
  AiAgent,
  CliExecuteResult,
  CliInstallSource,
  CliRuntimeDefinition,
  PageResult,
  PersonalAccessToken,
  SkillPackage,
} from "./types";

type Query = Record<string, string | number | boolean | undefined>;

function compactQuery(query?: Query): Record<string, string | number | boolean> {
  return Object.fromEntries(
    Object.entries(query ?? {}).filter(([, value]) => value !== undefined && value !== ""),
  ) as Record<string, string | number | boolean>;
}

export async function listCLIs(): Promise<{ items: CliRuntimeDefinition[]; risk_notice: string }> {
  const { body } = await http.get<{ items: CliRuntimeDefinition[]; risk_notice: string }>(
    "/ai/clis",
  );
  return body;
}

export async function detectCLI(key: string) {
  const { body } = await http.post<{
    detected: boolean;
    output: string;
    path: string;
    version: string;
    healthy: boolean;
    risk_notice: string;
  }>(`/ai/clis/${key}/detect`, {});
  return body;
}

export async function executeCLI(
  key: string,
  operation: "install" | "upgrade" | "uninstall",
  version = "",
): Promise<CliExecuteResult> {
  const { body } = await http.post<CliExecuteResult>(
    `/ai/clis/${key}/${operation}`,
    { version },
    { timeout: 300_000 },
  );
  return body;
}

export async function listCLISources(cliKey?: string): Promise<CliInstallSource[]> {
  const { body } = await http.get<{ items: CliInstallSource[] }>("/ai/cli-sources", {
    query: compactQuery({ cli_key: cliKey }),
  });
  return body.items;
}

export async function createCLISource(input: {
  cli_key: string;
  name: string;
  base_url: string;
  priority?: number;
  enabled?: boolean;
}): Promise<CliInstallSource> {
  const { body } = await http.post<CliInstallSource>("/ai/cli-sources", input);
  return body;
}

export async function updateCLISource(
  id: number,
  input: {
    cli_key?: string;
    name?: string;
    base_url?: string;
    priority?: number;
    enabled?: boolean;
  },
): Promise<CliInstallSource> {
  const { body } = await http.put<CliInstallSource>(`/ai/cli-sources/${id}`, input);
  return body;
}

export async function deleteCLISource(id: number): Promise<void> {
  await http.delete(`/ai/cli-sources/${id}`);
}

export async function listAgents(query?: Query): Promise<PageResult<AiAgent>> {
  const { body } = await http.get<PageResult<AiAgent>>("/ai/agents", {
    query: compactQuery(query),
  });
  return body;
}

export async function createAgent(input: Record<string, unknown>): Promise<AiAgent> {
  const { body } = await http.post<AiAgent>("/ai/agents", input);
  return body;
}

export async function updateAgent(id: number, input: Record<string, unknown>): Promise<AiAgent> {
  const { body } = await http.put<AiAgent>(`/ai/agents/${id}`, input);
  return body;
}

export async function deleteAgent(id: number): Promise<void> {
  await http.delete(`/ai/agents/${id}`);
}

export async function listTriggers(agentID: number): Promise<AgentTrigger[]> {
  const { body } = await http.get<{ items: AgentTrigger[] }>(`/ai/agents/${agentID}/triggers`);
  return body.items;
}

export async function createTrigger(
  agentID: number,
  input: Record<string, unknown>,
): Promise<AgentTrigger> {
  const { body } = await http.post<AgentTrigger>(`/ai/agents/${agentID}/triggers`, input);
  return body;
}

export async function manualRunAgent(agentID: number): Promise<AgentRun> {
  const { body } = await http.post<AgentRun>(`/ai/agents/${agentID}/runs`, {});
  return body;
}

export async function getRun(id: number): Promise<AgentRun> {
  const { body } = await http.get<AgentRun>(`/ai/runs/${id}`);
  return body;
}

export async function uploadSkill(form: FormData): Promise<SkillPackage> {
  const { body } = await http.post<SkillPackage>("/skills", form);
  return body;
}

export async function overwriteSkill(id: number, form: FormData): Promise<SkillPackage> {
  const { body } = await http.put<SkillPackage>(`/skills/${id}`, form);
  return body;
}

export async function deleteSkill(id: number): Promise<void> {
  await http.delete(`/skills/${id}`);
}

export async function downloadSkill(id: number): Promise<Blob> {
  const { body } = await http.get<Blob>(`/skills/${id}/package`, { responseType: "blob" });
  return body;
}

export async function listTokens(): Promise<PersonalAccessToken[]> {
  const { body } = await http.get<{ items: PersonalAccessToken[] }>("/tokens");
  return body.items;
}

export async function createToken(input: {
  name: string;
  scopes: string[];
  expires_at?: string;
}): Promise<{ token: string; metadata: PersonalAccessToken }> {
  const { body } = await http.post<{ token: string; metadata: PersonalAccessToken }>(
    "/tokens",
    input,
  );
  return body;
}

export async function deleteToken(id: number): Promise<void> {
  await http.delete(`/tokens/${id}`);
}

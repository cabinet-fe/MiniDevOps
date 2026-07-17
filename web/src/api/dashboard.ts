import { http } from "./http";
import type {
  AgentRunSummary,
  BuildSummary,
  DashboardLayout,
  SystemInfo,
  SystemStatus,
} from "./types";

export async function getDashboardLayout(): Promise<DashboardLayout> {
  const { body } = await http.get<DashboardLayout>("/dashboard/layout");
  return body;
}

export async function saveDashboardLayout(layout: DashboardLayout): Promise<DashboardLayout> {
  const { body } = await http.put<DashboardLayout>("/dashboard/layout", layout);
  return body;
}

export async function getBuildSummary(): Promise<BuildSummary> {
  const { body } = await http.get<BuildSummary>("/dashboard/build-summary");
  return body;
}

export async function getAgentRunSummary(): Promise<AgentRunSummary> {
  const { body } = await http.get<AgentRunSummary>("/dashboard/agent-run-summary");
  return body;
}

export async function getSystemInfo(): Promise<SystemInfo> {
  const { body } = await http.get<SystemInfo>("/dashboard/system-info");
  return body;
}

export async function getSystemStatus(): Promise<SystemStatus> {
  const { body } = await http.get<SystemStatus>("/dashboard/system-status");
  return body;
}

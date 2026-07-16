import type { ISearchOptions } from "@xterm/addon-search";
import type { ITerminalOptions, ITheme } from "@xterm/xterm";

export type BuildLogStatus =
  | "pending"
  | "queued"
  | "running"
  | "cloning"
  | "building"
  | "deploying"
  | "distributing"
  | "success"
  | "failed"
  | "cancelled"
  | "interrupted";

export const BUILD_LOG_STATUS_LABEL: Record<BuildLogStatus, string> = {
  pending: "等待中",
  queued: "排队中",
  running: "运行中",
  cloning: "拉取代码",
  building: "构建中",
  deploying: "部署中",
  distributing: "分发中",
  success: "成功",
  failed: "失败",
  cancelled: "已取消",
  interrupted: "已中断",
};

export const BUILD_LOG_STATUS_TAG: Record<
  BuildLogStatus,
  "info" | "primary" | "success" | "danger" | "warning" | undefined
> = {
  pending: "info",
  queued: "info",
  running: "primary",
  cloning: "primary",
  building: "primary",
  deploying: "primary",
  distributing: "primary",
  success: "success",
  failed: "danger",
  cancelled: "warning",
  interrupted: "warning",
};

export const TERMINAL_THEME: ITheme = {
  background: "#09090b",
  foreground: "#d4d4d8",
  cursor: "#d4d4d8",
  selectionBackground: "#3b82f680",
  selectionForeground: "#ffffff",
  black: "#27272a",
  red: "#ef4444",
  green: "#22c55e",
  yellow: "#eab308",
  blue: "#3b82f6",
  magenta: "#a855f7",
  cyan: "#06b6d4",
  white: "#d4d4d8",
  brightBlack: "#52525b",
  brightRed: "#f87171",
  brightGreen: "#4ade80",
  brightYellow: "#facc15",
  brightBlue: "#60a5fa",
  brightMagenta: "#c084fc",
  brightCyan: "#22d3ee",
  brightWhite: "#fafafa",
};

export const TERMINAL_OPTIONS: ITerminalOptions = {
  allowProposedApi: true,
  disableStdin: true,
  cursorBlink: false,
  cursorStyle: "bar",
  cursorInactiveStyle: "none",
  fontSize: 13,
  fontFamily:
    "'JetBrains Mono', 'Fira Code', 'Cascadia Code', Menlo, Monaco, 'Courier New', monospace",
  lineHeight: 1.4,
  scrollback: 100000,
  convertEol: true,
  theme: TERMINAL_THEME,
};

export const SEARCH_OPTIONS: ISearchOptions = {
  caseSensitive: false,
  regex: false,
  wholeWord: false,
  decorations: {
    matchBackground: "#854d0e",
    matchBorder: "#a16207",
    matchOverviewRuler: "#eab308",
    activeMatchBackground: "#1d4ed8",
    activeMatchBorder: "#3b82f6",
    activeMatchColorOverviewRuler: "#3b82f6",
  },
};

export function normalizeLogLines(text: string): string[] {
  const lines = text.split("\n");
  if (lines.at(-1) === "") {
    lines.pop();
  }
  return lines;
}

export function writeLinesToTerminal(term: import("@xterm/xterm").Terminal, lines: string[]) {
  term.clear();
  if (lines.length > 0) {
    term.write(lines.join("\r\n"));
  }
}

export function resolveBuildLogStatus(
  status: string | undefined,
  distributionSummary?: string,
): BuildLogStatus {
  if (
    status === "success" &&
    (distributionSummary === "running" || distributionSummary === "pending")
  ) {
    return "distributing";
  }
  if (status && status in BUILD_LOG_STATUS_LABEL) {
    return status as BuildLogStatus;
  }
  return "pending";
}

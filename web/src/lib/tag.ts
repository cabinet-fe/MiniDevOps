import type { ColorType } from "@veltra/utils";

/** 省略 type 时为默认灰底标签 */
export type TagType = ColorType | undefined;

export function tagType(value: string | undefined | null, map: Record<string, TagType>): TagType {
  if (!value) return undefined;
  return map[value];
}

export function boolTagType(value: boolean, falseType?: TagType): TagType {
  return value ? "success" : falseType;
}

/** 构建 / Agent / 安装任务等异步状态 */
export const JOB_STATUS_TAG: Record<string, TagType> = {
  queued: "info",
  pending: "info",
  running: "primary",
  success: "success",
  failed: "danger",
  cancelled: "warning",
  interrupted: "warning",
};

export const TRIGGER_TYPE_TAG: Record<string, TagType> = {
  manual: undefined,
  api: "info",
  webhook: "info",
  cron: "primary",
  build_event: "warning",
  docs_generate: "info",
};

export const INSTALL_OP_TAG: Record<string, TagType> = {
  install: "primary",
  upgrade: "info",
  uninstall: "danger",
  switch: "warning",
};

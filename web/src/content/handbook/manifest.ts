import openApi from "./open-api.md?raw";
import ops from "./ops.md?raw";
import usage from "./usage.md?raw";

export type HandbookSection = {
  key: string;
  title: string;
  content: string;
};

export const handbookSections: HandbookSection[] = [
  { key: "usage", title: "使用说明", content: usage },
  { key: "ops", title: "运维手册", content: ops },
  { key: "open-api", title: "开放接口", content: openApi },
];

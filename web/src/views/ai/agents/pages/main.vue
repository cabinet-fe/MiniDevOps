<script setup lang="ts">
defineOptions({ name: "AiAgents" });

import { computed, onMounted, reactive, ref, useTemplateRef } from "vue";
import { useRouter } from "vue-router";
import { o } from "@cat-kit/core";
import { message } from "@veltra/desktop";

import {
  createAgent,
  createTrigger,
  deleteAgent,
  deleteTrigger,
  listSkills,
  listTriggers,
  manualRunAgent,
  updateAgent,
} from "@/api/ai";
import { listBuildJobs } from "@/api/cicd";
import type { AiAgent, BuildJob, SkillPackage } from "@/api/types";
import FormDialog from "@/components/form-dialog";
import ProTable, { defineProTableColumns } from "@/components/pro-table";
import { usePermission } from "@/composables/use-permission";
import { tagType, type TagType } from "@/lib/tag";

const CLI_KEY_TAG: Record<string, TagType> = {
  claude_code: "primary",
  opencode: "info",
  reasonix: "success",
  codex: "warning",
};

const TRIGGER_TYPE_LABEL: Record<string, string> = {
  manual: "手动",
  api: "API",
  cron: "Cron",
  build_event: "构建事件",
};

type TriggerDraft = {
  /** Existing server id; undefined = newly added locally */
  id?: number;
  type: string;
  cron_expression: string;
  cron_timezone: string;
  build_job_id?: number;
  build_event: string;
};

const { hasPermission } = usePermission();
const router = useRouter();
const table = useTemplateRef("table");
const dialogOpen = ref(false);
const editing = ref<AiAgent | null>(null);
const skills = ref<SkillPackage[]>([]);
const buildJobs = ref<BuildJob[]>([]);
/** Triggers shown in the agent form (existing + newly added drafts). */
const formTriggers = ref<TriggerDraft[]>([]);
/** Snapshot of server trigger ids when the edit dialog opened. */
const initialTriggerIDs = ref<number[]>([]);

const form = reactive({
  name: "",
  description: "",
  enabled: true,
  cli_key: "claude_code",
  system_prompt: "",
  skill_ids: [] as number[],
  build_job_ids: [] as number[],
  output_dir: "output",
  artifact_format: "gzip" as "zip" | "gzip",
  max_artifacts: 10,
  timeout_sec: 600,
});

const triggerDraft = reactive({
  type: "manual",
  cron_expression: "0 * * * *",
  cron_timezone: "Asia/Shanghai",
  build_job_id: undefined as number | undefined,
  build_event: "artifact_ready",
});

const skillOptions = computed(() =>
  skills.value.map((s) => ({
    label: `${s.name}${s.visibility === "private" ? " (私有)" : ""}`,
    value: s.id,
  })),
);

const buildJobOptions = computed(() =>
  buildJobs.value.map((j) => ({
    label: `${j.name} (job-${j.id})`,
    value: j.id,
  })),
);

const columns = defineProTableColumns([
  { key: "id", name: "ID", width: 70 },
  { key: "name", name: "名称" },
  { key: "cli_key", name: "CLI", width: 120 },
  { key: "enabled", name: "启用", width: 80 },
  { key: "action", name: "操作", width: 260, align: "center", fixed: "right" },
]);

function openRunHistory(row: AiAgent) {
  void router.push({ path: "/ai/runs", query: { agent_id: String(row.id) } });
}

onMounted(async () => {
  const tasks: Promise<void>[] = [];
  if (hasPermission("ai.skills:view")) {
    tasks.push(
      listSkills({ page: 1, page_size: 200 })
        .then((res) => {
          skills.value = res.items ?? [];
        })
        .catch(() => {
          skills.value = [];
        }),
    );
  }
  if (hasPermission("cicd.build_jobs:view")) {
    tasks.push(
      listBuildJobs({ page: 1, page_size: 200 })
        .then((res) => {
          buildJobs.value = res.items ?? [];
        })
        .catch(() => {
          buildJobs.value = [];
        }),
    );
  }
  await Promise.all(tasks);
});

function resetTriggerDraft() {
  triggerDraft.type = "manual";
  triggerDraft.cron_expression = "0 * * * *";
  triggerDraft.cron_timezone = "Asia/Shanghai";
  triggerDraft.build_job_id = undefined;
  triggerDraft.build_event = "artifact_ready";
}

function resetFormFields() {
  form.name = "";
  form.description = "";
  form.enabled = true;
  form.cli_key = "claude_code";
  form.system_prompt = "";
  form.skill_ids = [];
  form.build_job_ids = [];
  form.output_dir = "output";
  form.artifact_format = "gzip";
  form.max_artifacts = 10;
  form.timeout_sec = 600;
}

function openCreate() {
  editing.value = null;
  resetFormFields();
  formTriggers.value = [];
  initialTriggerIDs.value = [];
  resetTriggerDraft();
  dialogOpen.value = true;
}

async function openEdit(row: AiAgent) {
  editing.value = row;
  o(form).extend(row);
  form.skill_ids = [...(row.skill_ids ?? [])];
  form.build_job_ids = [...(row.build_job_ids ?? [])];
  form.output_dir = row.output_dir || "output";
  form.artifact_format = row.artifact_format === "zip" ? "zip" : "gzip";
  form.max_artifacts = row.max_artifacts || 10;
  resetTriggerDraft();
  try {
    const items = await listTriggers(row.id);
    formTriggers.value = items.map((t) => ({
      id: t.id,
      type: t.type,
      cron_expression: t.cron_expression ?? "",
      cron_timezone: t.cron_timezone ?? "UTC",
      build_job_id: t.build_job_id ?? undefined,
      build_event: t.build_event ?? "artifact_ready",
    }));
    initialTriggerIDs.value = items.map((t) => t.id);
  } catch {
    formTriggers.value = [];
    initialTriggerIDs.value = [];
    message.error("加载触发器失败");
  }
  dialogOpen.value = true;
}

function buildJobLabel(jobID?: number) {
  if (!jobID) return "";
  const job = buildJobs.value.find((j) => j.id === jobID);
  return job ? `${job.name} (job-${job.id})` : `job-${jobID}`;
}

function triggerSummary(t: TriggerDraft): string {
  const typeLabel = TRIGGER_TYPE_LABEL[t.type] ?? t.type;
  if (t.type === "cron") {
    return `${typeLabel} · ${t.cron_expression} (${t.cron_timezone})`;
  }
  if (t.type === "build_event") {
    return `${typeLabel} · ${buildJobLabel(t.build_job_id)} · ${t.build_event}`;
  }
  return typeLabel;
}

function addTriggerDraft() {
  if (triggerDraft.type === "cron") {
    if (!triggerDraft.cron_expression.trim() || !triggerDraft.cron_timezone.trim()) {
      message.error("请填写 Cron 表达式与时区");
      return;
    }
  }
  if (triggerDraft.type === "build_event") {
    if (!triggerDraft.build_job_id) {
      message.error("请选择构建任务");
      return;
    }
  }
  formTriggers.value.push({
    type: triggerDraft.type,
    cron_expression: triggerDraft.cron_expression,
    cron_timezone: triggerDraft.cron_timezone,
    build_job_id: triggerDraft.build_job_id,
    build_event: triggerDraft.build_event,
  });
  resetTriggerDraft();
}

function removeFormTrigger(index: number) {
  formTriggers.value.splice(index, 1);
}

function triggerPayload(t: TriggerDraft) {
  return {
    type: t.type,
    enabled: true,
    cron_expression: t.type === "cron" ? t.cron_expression : "",
    cron_timezone: t.type === "cron" ? t.cron_timezone : "UTC",
    build_job_id: t.type === "build_event" ? t.build_job_id : undefined,
    build_event: t.type === "build_event" ? t.build_event : "",
  };
}

async function syncTriggers(agentID: number) {
  const keptIDs = new Set(
    formTriggers.value.map((t) => t.id).filter((id): id is number => id != null),
  );
  const toDelete = initialTriggerIDs.value.filter((id) => !keptIDs.has(id));
  for (const tid of toDelete) {
    await deleteTrigger(agentID, tid);
  }
  const toCreate = formTriggers.value.filter((t) => t.id == null);
  for (const draft of toCreate) {
    await createTrigger(agentID, triggerPayload(draft));
  }
}

async function save() {
  const body = {
    name: form.name,
    description: form.description,
    enabled: form.enabled,
    cli_key: form.cli_key,
    system_prompt: form.system_prompt,
    skill_ids: form.skill_ids,
    build_job_ids: form.build_job_ids,
    output_dir: form.output_dir || "output",
    artifact_format: form.artifact_format,
    max_artifacts: form.max_artifacts,
    timeout_sec: form.timeout_sec,
  };
  try {
    let agentID: number;
    if (editing.value) {
      await updateAgent(editing.value.id, body);
      agentID = editing.value.id;
    } else {
      const created = await createAgent(body);
      agentID = created.id;
    }
    await syncTriggers(agentID);
    dialogOpen.value = false;
    table.value?.reload();
  } catch (error) {
    message.error(error instanceof Error ? error.message : "保存失败");
  }
}

async function run(row: AiAgent) {
  try {
    const run = await manualRunAgent(row.id);
    message.success(`已创建运行 #${run.id}`);
    await router.push(`/ai/runs/${run.id}`);
  } catch (error) {
    message.error(error instanceof Error ? error.message : "触发失败");
  }
}

async function remove(row: AiAgent) {
  try {
    await deleteAgent(row.id);
    table.value?.reload();
  } catch (error) {
    message.error(error instanceof Error ? error.message : "删除失败");
  }
}
</script>

<template>
  <div>
    <ProTable ref="table" url="/ai/agents" pagination :columns="columns">
      <template #filters>
        <u-button
          v-if="hasPermission('ai.agents:create')"
          type="primary"
          style="margin-left: auto"
          @click.prevent="openCreate"
        >
          新建
        </u-button>
      </template>
      <template #column:cli_key="{ rowData }">
        <u-tag size="small" :type="tagType((rowData as AiAgent).cli_key, CLI_KEY_TAG)">
          {{ (rowData as AiAgent).cli_key }}
        </u-tag>
      </template>
      <template #column:enabled="{ rowData }">
        <u-tag size="small" :type="(rowData as AiAgent).enabled ? 'success' : undefined">
          {{ (rowData as AiAgent).enabled ? "启用" : "停用" }}
        </u-tag>
      </template>
      <template #column:action="{ rowData }">
        <u-action-group :max="3">
          <u-action v-if="hasPermission('ai.agents:update')" @run="openEdit(rowData as AiAgent)">
            编辑
          </u-action>
          <u-action v-if="hasPermission('ai.agents:execute')" @run="run(rowData as AiAgent)">
            运行
          </u-action>
          <u-action
            v-if="hasPermission('ai.runs:view')"
            @run="openRunHistory(rowData as AiAgent)"
          >
            运行历史
          </u-action>
          <u-action
            v-if="hasPermission('ai.agents:delete')"
            type="danger"
            @run="remove(rowData as AiAgent)"
          >
            删除
          </u-action>
        </u-action-group>
      </template>
    </ProTable>

    <FormDialog
      v-model="dialogOpen"
      :title="editing ? '编辑智能体' : '新建智能体'"
      :model="form"
      label-width="110px"
      style="width: 960px"
      @submit="save"
    >
      <u-input label="名称" field="name" :rules="{ required: '必填' }" />
      <u-input label="描述" field="description" />
      <u-select
        label="CLI"
        field="cli_key"
        :options="[
          { label: 'Claude Code', value: 'claude_code' },
          { label: 'OpenCode', value: 'opencode' },
          { label: 'Reasonix', value: 'reasonix' },
          { label: 'Codex', value: 'codex' },
        ]"
        :rules="{ required: '必填' }"
      />
      <u-textarea
        label="系统提示词"
        field="system_prompt"
        span="full"
        :rows="6"
        placeholder="描述任务目标；若需访问绑定的构建任务，请写相对路径，如 ./job-12（与选项「名称 (job-12)」一致）"
      />
      <u-multi-select
        label="技能"
        field="skill_ids"
        :options="skillOptions"
        placeholder="选择可访问的技能"
        filterable
        clearable
      />
      <u-multi-select
        label="构建任务"
        field="build_job_ids"
        :options="buildJobOptions"
        placeholder="软链到构建任务工作区"
        filterable
        clearable
      />
      <u-select
        label="制品格式"
        field="artifact_format"
        :options="[
          { label: 'gzip (tar.gz)', value: 'gzip' },
          { label: 'zip', value: 'zip' },
        ]"
        :rules="{ required: '必填' }"
      />
      <u-input label="产出目录名" field="output_dir" placeholder="默认 output" />
      <u-number-input label="保留制品数" field="max_artifacts" :min="1" />
      <u-number-input label="超时(秒)" field="timeout_sec" :min="30" />
      <u-switch label="启用" field="enabled" />

      <u-form-item label="触发器" span="full">
        <div class="trigger-section">
          <ul v-if="formTriggers.length" class="trigger-list">
            <li v-for="(t, index) in formTriggers" :key="t.id ?? `new-${index}`" class="trigger-row">
              <span class="trigger-summary">{{ triggerSummary(t) }}</span>
              <u-button text type="danger" size="small" @click="removeFormTrigger(index)">
                移除
              </u-button>
            </li>
          </ul>
          <p v-else class="trigger-empty">暂无触发器，可在下方添加</p>

          <div class="trigger-draft">
            <div class="trigger-draft-row">
              <span class="trigger-draft-label">类型</span>
              <u-select
                v-model="triggerDraft.type"
                :options="[
                  { label: '手动', value: 'manual' },
                  { label: 'API', value: 'api' },
                  { label: 'Cron', value: 'cron' },
                  { label: '构建事件', value: 'build_event' },
                ]"
              />
            </div>
            <template v-if="triggerDraft.type === 'cron'">
              <div class="trigger-draft-row">
                <span class="trigger-draft-label">表达式</span>
                <u-input v-model="triggerDraft.cron_expression" placeholder="如 0 * * * *" />
              </div>
              <div class="trigger-draft-row">
                <span class="trigger-draft-label">时区</span>
                <u-input
                  v-model="triggerDraft.cron_timezone"
                  placeholder="IANA，如 Asia/Shanghai"
                />
              </div>
            </template>
            <template v-if="triggerDraft.type === 'build_event'">
              <div class="trigger-draft-row">
                <span class="trigger-draft-label">构建任务</span>
                <u-select
                  v-model="triggerDraft.build_job_id"
                  :options="buildJobOptions"
                  filterable
                  clearable
                  placeholder="选择构建任务"
                />
              </div>
              <div class="trigger-draft-row">
                <span class="trigger-draft-label">事件</span>
                <u-select
                  v-model="triggerDraft.build_event"
                  :options="[
                    { label: 'artifact_ready（默认）', value: 'artifact_ready' },
                    { label: 'distribution_finished', value: 'distribution_finished' },
                  ]"
                />
              </div>
            </template>
            <u-button size="small" @click="addTriggerDraft">添加到列表</u-button>
          </div>
        </div>
      </u-form-item>
    </FormDialog>
  </div>
</template>

<style scoped lang="scss">
.trigger-section {
  width: 100%;
}

.trigger-list {
  margin: 0 0 12px;
  padding: 0;
  list-style: none;
}

.trigger-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 4px 0;
  font-size: 13px;
}

.trigger-summary {
  flex: 1;
  min-width: 0;
}

.trigger-empty {
  margin: 0 0 12px;
  font-size: 13px;
  opacity: 0.65;
}

.trigger-draft {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding-top: 4px;
}

.trigger-draft-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.trigger-draft-label {
  flex: 0 0 64px;
  font-size: 13px;
  opacity: 0.8;
}

.trigger-draft-row > :deep(.u-select),
.trigger-draft-row > :deep(.u-input) {
  flex: 1;
  min-width: 0;
}
</style>

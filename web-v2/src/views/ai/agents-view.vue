<script setup lang="ts">
import { reactive, ref, useTemplateRef } from "vue";
import { useRouter } from "vue-router";
import { defineTableColumns, message } from "@veltra/desktop";

import {
  createAgent,
  createTrigger,
  deleteAgent,
  listTriggers,
  manualRunAgent,
  updateAgent,
} from "@/api/ai";
import type { AgentTrigger, AiAgent } from "@/api/types";
import FormDialog from "@/components/form-dialog.vue";
import ProTable from "@/components/pro-table.vue";
import { usePermission } from "@/composables/use-permission";

const { hasPermission } = usePermission();
const router = useRouter();
const table = useTemplateRef("table");
const dialogOpen = ref(false);
const editing = ref<AiAgent | null>(null);
const triggers = ref<AgentTrigger[]>([]);
const triggerOpen = ref(false);
const triggerAgentID = ref(0);

const form = reactive({
  name: "",
  description: "",
  enabled: true,
  cli_key: "claude_code",
  system_prompt: "",
  skill_ids: "",
  repository_id: undefined as number | undefined,
  timeout_sec: 600,
});

const triggerForm = reactive({
  type: "manual",
  cron_expression: "0 * * * *",
  cron_timezone: "Asia/Shanghai",
  build_job_id: undefined as number | undefined,
  build_event: "artifact_ready",
});

const columns = defineTableColumns([
  { key: "id", name: "ID", width: 70 },
  { key: "name", name: "名称", minWidth: 140 },
  { key: "cli_key", name: "CLI", width: 120 },
  { key: "enabled", name: "启用", width: 80 },
  { key: "action", name: "操作", width: 280, minWidth: 240 },
]);

function openCreate() {
  editing.value = null;
  Object.assign(form, {
    name: "",
    description: "",
    enabled: true,
    cli_key: "claude_code",
    system_prompt: "",
    skill_ids: "",
    repository_id: undefined,
    timeout_sec: 600,
  });
  dialogOpen.value = true;
}

function openEdit(row: AiAgent) {
  editing.value = row;
  Object.assign(form, {
    name: row.name,
    description: row.description,
    enabled: row.enabled,
    cli_key: row.cli_key,
    system_prompt: row.system_prompt,
    skill_ids: (row.skill_ids ?? []).join(","),
    repository_id: row.repository_id ?? undefined,
    timeout_sec: row.timeout_sec,
  });
  dialogOpen.value = true;
}

function parseSkillIDs() {
  return form.skill_ids
    .split(/[,\s]+/)
    .map((s) => Number(s.trim()))
    .filter((n) => Number.isFinite(n) && n > 0);
}

async function save() {
  const body = {
    name: form.name,
    description: form.description,
    enabled: form.enabled,
    cli_key: form.cli_key,
    system_prompt: form.system_prompt,
    skill_ids: parseSkillIDs(),
    repository_id: form.repository_id || null,
    timeout_sec: form.timeout_sec,
  };
  try {
    if (editing.value) {
      await updateAgent(editing.value.id, body);
    } else {
      await createAgent(body);
    }
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

async function openTriggers(row: AiAgent) {
  triggerAgentID.value = row.id;
  triggers.value = await listTriggers(row.id);
  triggerOpen.value = true;
}

async function addTrigger() {
  try {
    await createTrigger(triggerAgentID.value, { ...triggerForm });
    triggers.value = await listTriggers(triggerAgentID.value);
    message.success("触发器已创建");
  } catch (error) {
    message.error(error instanceof Error ? error.message : "创建失败");
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
  <div class="page">
    <header class="page-head">
      <h2>智能体</h2>
      <p>上下文仅为系统提示词 + 选定代码仓库；构建事件异步触发，不影响 BuildRun 状态。</p>
      <u-button v-if="hasPermission('ai.agents:create')" type="primary" @click="openCreate">
        新建
      </u-button>
    </header>

    <ProTable ref="table" url="/ai/agents" mode="pagination" :columns="columns">
      <template #enabled="{ rowData }">
        {{ (rowData as AiAgent).enabled ? "是" : "否" }}
      </template>
      <template #action="{ rowData }">
        <u-action-group :max="4">
          <u-action v-if="hasPermission('ai.agents:update')" @run="openEdit(rowData as AiAgent)">
            编辑
          </u-action>
          <u-action
            v-if="hasPermission('ai.agents:update')"
            @run="openTriggers(rowData as AiAgent)"
          >
            触发器
          </u-action>
          <u-action v-if="hasPermission('ai.agents:execute')" @run="run(rowData as AiAgent)">
            运行
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
      style="width: 560px"
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
      <u-textarea label="系统提示词" field="system_prompt" :rows="6" />
      <u-input label="Skill IDs" field="skill_ids" placeholder="逗号分隔" />
      <u-number-input label="仓库 ID" field="repository_id" :min="1" placeholder="可选" />
      <u-number-input label="超时(秒)" field="timeout_sec" :min="30" />
      <u-switch label="启用" field="enabled" />
    </FormDialog>

    <FormDialog
      v-model="triggerOpen"
      title="触发器"
      :model="triggerForm"
      label-width="120px"
      style="width: 520px"
      confirm-text="添加"
      @submit="addTrigger"
    >
      <ul class="trigger-list">
        <li v-for="t in triggers" :key="t.id">
          #{{ t.id }} {{ t.type }}
          <span v-if="t.type === 'cron'">{{ t.cron_expression }} {{ t.cron_timezone }}</span>
          <span v-if="t.type === 'build_event'">
            job={{ t.build_job_id }} event={{ t.build_event }}
          </span>
        </li>
      </ul>
      <u-select
        label="类型"
        field="type"
        :options="[
          { label: '手动', value: 'manual' },
          { label: 'API', value: 'api' },
          { label: 'Cron', value: 'cron' },
          { label: '构建事件', value: 'build_event' },
        ]"
        :rules="{ required: '必填' }"
      />
      <template v-if="triggerForm.type === 'cron'">
        <u-input
          label="表达式"
          field="cron_expression"
          placeholder="如 0 * * * *"
          :rules="{ required: '必填' }"
        />
        <u-input
          label="时区"
          field="cron_timezone"
          placeholder="IANA，如 Asia/Shanghai"
          :rules="{ required: '必填' }"
        />
      </template>
      <template v-if="triggerForm.type === 'build_event'">
        <u-number-input
          label="BuildJob ID"
          field="build_job_id"
          :min="1"
          :rules="{ required: '必填' }"
        />
        <u-select
          label="事件"
          field="build_event"
          :options="[
            { label: 'artifact_ready（默认）', value: 'artifact_ready' },
            { label: 'distribution_finished', value: 'distribution_finished' },
          ]"
        />
      </template>
    </FormDialog>
  </div>
</template>

<style scoped lang="scss">
.page {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.page-head p {
  margin: 4px 0 12px;
  color: var(--u-color-text-secondary, #666);
  font-size: 13px;
}
.trigger-list {
  margin: 0 0 12px;
  padding-left: 18px;
  font-size: 13px;
}
</style>

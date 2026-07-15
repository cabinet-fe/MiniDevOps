<script setup lang="ts">
import { onMounted, reactive, ref, useTemplateRef } from "vue";
import { useRouter } from "vue-router";
import { defineTableColumns, message } from "@veltra/desktop";

import {
  createBuildJob,
  deleteBuildJob,
  enqueueBuildRun,
  getBuildJob,
  listBuildJobs,
  listRepositories,
  listServers,
  updateBuildJob,
} from "@/api/cicd";
import type { BuildJob, DeployTarget, Repository, Server } from "@/api/types";
import FormDialog from "@/components/form-dialog.vue";
import ResourceList from "@/components/resource-list.vue";
import { usePermission } from "@/composables/use-permission";

const BRANCH_POLICY_OPTIONS = [
  { label: "固定分支", value: "fixed" },
  { label: "Webhook 任意分支", value: "param" },
];

const METHOD_OPTIONS = [
  { label: "rsync", value: "rsync" },
  { label: "sftp", value: "sftp" },
  { label: "scp", value: "scp" },
  { label: "agent", value: "agent" },
  { label: "local", value: "local" },
];

const ARTIFACT_OPTIONS = [
  { label: "gzip", value: "gzip" },
  { label: "zip", value: "zip" },
];

const { hasPermission } = usePermission();
const router = useRouter();
const listRef = useTemplateRef("list");
const filters = reactive({ keyword: "", repository_id: undefined as number | undefined });
const dialogOpen = ref(false);
const editing = ref<BuildJob | null>(null);
const repoOptions = ref<{ label: string; value: number }[]>([]);
const serverOptions = ref<{ label: string; value: number }[]>([]);
const form = reactive({
  repository_id: undefined as number | undefined,
  name: "",
  description: "",
  enabled: true,
  branch_policy: "fixed",
  branch: "main",
  shallow_clone: true,
  build_script_type: "bash",
  build_script: "",
  work_dir: "",
  output_dir: "",
  env_var_names: "",
  trigger_manual: true,
  trigger_webhook: false,
  trigger_cron: false,
  cron_expression: "",
  cron_timezone: "Asia/Shanghai",
  max_artifacts: 5,
  artifact_format: "gzip",
  deploy_targets: [] as DeployTarget[],
});

const columns = defineTableColumns([
  { key: "id", name: "ID", width: 80, minWidth: 60 },
  { key: "name", name: "名称", minWidth: 140 },
  { key: "repository_id", name: "仓库", width: 90, minWidth: 70 },
  { key: "branch", name: "分支", width: 110, minWidth: 80 },
  { key: "enabled", name: "启用", width: 80, minWidth: 60 },
  { key: "triggers", name: "触发", minWidth: 140 },
  { key: "action", name: "操作", width: 240, minWidth: 180 },
]);

async function fetcher(params: { page: number; page_size: number }) {
  return listBuildJobs({
    ...params,
    keyword: filters.keyword,
    repository_id: filters.repository_id,
  });
}

onMounted(async () => {
  try {
    const [repos, servers] = await Promise.all([
      listRepositories({ page: 1, page_size: 100 }),
      listServers({ page: 1, page_size: 100 }),
    ]);
    repoOptions.value = (repos.items ?? []).map((r: Repository) => ({
      label: r.name,
      value: r.id,
    }));
    serverOptions.value = (servers.items ?? []).map((s: Server) => ({
      label: `${s.name} (${s.host})`,
      value: s.id,
    }));
  } catch {
    /* ignore */
  }
});

function triggerLabel(job: BuildJob) {
  const parts: string[] = [];
  if (job.trigger_manual) parts.push("手动");
  if (job.trigger_webhook) parts.push("Webhook");
  if (job.trigger_cron) parts.push("Cron");
  return parts.length ? parts.join(" / ") : "—";
}

function resetForm() {
  Object.assign(form, {
    repository_id: undefined,
    name: "",
    description: "",
    enabled: true,
    branch_policy: "fixed",
    branch: "main",
    shallow_clone: true,
    build_script_type: "bash",
    build_script: "",
    work_dir: "",
    output_dir: "",
    env_var_names: "",
    trigger_manual: true,
    trigger_webhook: false,
    trigger_cron: false,
    cron_expression: "",
    cron_timezone: "Asia/Shanghai",
    max_artifacts: 5,
    artifact_format: "gzip",
    deploy_targets: [],
  });
}

function openCreate() {
  editing.value = null;
  resetForm();
  dialogOpen.value = true;
}

async function openEdit(row: BuildJob) {
  try {
    const full = await getBuildJob(row.id);
    editing.value = full;
    Object.assign(form, {
      repository_id: full.repository_id,
      name: full.name,
      description: full.description || "",
      enabled: full.enabled,
      branch_policy: full.branch_policy || "fixed",
      branch: full.branch || "main",
      shallow_clone: full.shallow_clone !== false,
      build_script_type: full.build_script_type || "bash",
      build_script: full.build_script || "",
      work_dir: full.work_dir || "",
      output_dir: full.output_dir || "",
      env_var_names: (full.env_var_names ?? []).join(","),
      trigger_manual: full.trigger_manual,
      trigger_webhook: full.trigger_webhook,
      trigger_cron: full.trigger_cron,
      cron_expression: full.cron_expression || "",
      cron_timezone: full.cron_timezone || "Asia/Shanghai",
      max_artifacts: full.max_artifacts || 5,
      artifact_format: full.artifact_format || "gzip",
      deploy_targets: (full.deploy_targets ?? []).map((t) => ({ ...t })),
    });
    dialogOpen.value = true;
  } catch (err) {
    message.error(err instanceof Error ? err.message : "加载失败");
  }
}

function addTarget() {
  form.deploy_targets.push({
    server_id: undefined,
    remote_path: "",
    method: "rsync",
    post_deploy_script: "",
    sort_order: form.deploy_targets.length,
  });
}

function removeTarget(idx: number) {
  form.deploy_targets.splice(idx, 1);
}

function buildBody(): Record<string, unknown> {
  return {
    repository_id: form.repository_id,
    name: form.name,
    description: form.description,
    enabled: form.enabled,
    branch_policy: form.branch_policy,
    branch: form.branch,
    shallow_clone: form.shallow_clone,
    build_script_type: form.build_script_type,
    build_script: form.build_script,
    work_dir: form.work_dir,
    output_dir: form.output_dir,
    env_var_names: form.env_var_names
      .split(/[,;\s]+/)
      .map((s) => s.trim())
      .filter(Boolean),
    trigger_manual: form.trigger_manual,
    trigger_webhook: form.trigger_webhook,
    trigger_cron: form.trigger_cron,
    cron_expression: form.cron_expression,
    cron_timezone: form.cron_timezone,
    max_artifacts: form.max_artifacts,
    artifact_format: form.artifact_format,
    deploy_targets: form.deploy_targets.map((t, i) => ({
      server_id: t.method === "local" ? null : t.server_id,
      remote_path: t.remote_path,
      method: t.method,
      post_deploy_script: t.post_deploy_script || "",
      sort_order: t.sort_order ?? i,
    })),
  };
}

async function save() {
  try {
    const body = buildBody();
    if (editing.value) {
      await updateBuildJob(editing.value.id, body);
      message.success("已更新");
    } else {
      await createBuildJob(body);
      message.success("已创建");
    }
    dialogOpen.value = false;
    await listRef.value?.refresh();
  } catch (err) {
    message.error(err instanceof Error ? err.message : "保存失败");
  }
}

async function remove(row: BuildJob) {
  try {
    await deleteBuildJob(row.id);
    message.success("已删除");
    await listRef.value?.refresh();
  } catch (err) {
    message.error(err instanceof Error ? err.message : "删除失败");
  }
}

async function trigger(row: BuildJob) {
  try {
    const run = await enqueueBuildRun(row.id, { trigger_type: "manual" });
    message.success(`已入队 #${run.build_number}`);
    await router.push({ name: "cicd-build-run-detail", params: { id: run.id } });
  } catch (err) {
    message.error(err instanceof Error ? err.message : "触发失败");
  }
}
</script>

<template>
  <div class="page">
    <div class="page-head">
      <h2>构建任务</h2>
      <u-button v-if="hasPermission('cicd.build_jobs:create')" type="primary" @click="openCreate">
        新建任务
      </u-button>
    </div>

    <ResourceList ref="list" :fetcher="fetcher" :columns="columns" :filters="filters">
      <template #filters="{ reload }">
        <u-select
          v-model="filters.repository_id"
          :options="repoOptions"
          placeholder="全部仓库"
          clearable
          style="width: 180px"
        />
        <u-input v-model="filters.keyword" placeholder="名称" style="width: 160px" />
        <u-button @click="reload">刷新</u-button>
      </template>
      <template #column:enabled="{ rowData }">
        {{ (rowData as BuildJob).enabled ? "是" : "否" }}
      </template>
      <template #column:triggers="{ rowData }">
        {{ triggerLabel(rowData as BuildJob) }}
      </template>
      <template #column:action="{ rowData }">
        <u-action-group :max="4">
          <u-action
            v-if="hasPermission('cicd.build_jobs:update')"
            @run="openEdit(rowData as BuildJob)"
          >
            编辑
          </u-action>
          <u-action
            v-if="hasPermission('cicd.build_jobs:execute')"
            @run="trigger(rowData as BuildJob)"
          >
            触发
          </u-action>
          <u-action
            v-if="hasPermission('cicd.build_jobs:delete')"
            need-confirm
            type="danger"
            @run="remove(rowData as BuildJob)"
          >
            删除
          </u-action>
        </u-action-group>
      </template>
    </ResourceList>

    <FormDialog
      v-model="dialogOpen"
      :title="editing ? '编辑任务' : '新建任务'"
      :model="form"
      label-width="110px"
      style="width: 780px"
      @submit="save"
    >
      <u-select
        label="仓库"
        field="repository_id"
        :options="repoOptions"
        :disabled="!!editing"
        :rules="{ required: '必填' }"
      />
      <u-input label="名称" field="name" :rules="{ required: '必填' }" />
      <u-input label="描述" field="description" />
      <u-switch label="启用" field="enabled" />
      <u-select label="分支策略" field="branch_policy" :options="BRANCH_POLICY_OPTIONS" />
      <u-input label="分支" field="branch" placeholder="分支名或匹配模式" />
      <u-switch label="浅克隆" field="shallow_clone" />
      <u-textarea label="构建脚本" field="build_script" :rows="5" />
      <u-input label="工作目录" field="work_dir" placeholder="相对仓库根" />
      <u-input label="输出目录" field="output_dir" />
      <u-input label="环境变量名" field="env_var_names" placeholder="逗号分隔，仅名称" />

      <u-form-item label="触发方式">
        <div class="trigger-row">
          <u-checkbox v-model="form.trigger_manual">手动</u-checkbox>
          <u-checkbox v-model="form.trigger_webhook">Webhook</u-checkbox>
          <u-checkbox v-model="form.trigger_cron">Cron</u-checkbox>
        </div>
      </u-form-item>
      <template v-if="form.trigger_cron">
        <u-input label="Cron 表达式" field="cron_expression" placeholder="如 0 */6 * * *" />
        <u-input label="时区" field="cron_timezone" placeholder="IANA，如 Asia/Shanghai" />
      </template>

      <u-number-input label="制品保留" field="max_artifacts" />
      <u-select label="制品格式" field="artifact_format" :options="ARTIFACT_OPTIONS" />

      <div class="targets-head">
        <strong>部署目标（Job 私有）</strong>
        <u-button size="small" @click="addTarget">添加</u-button>
      </div>
      <div v-for="(t, idx) in form.deploy_targets" :key="idx" class="target-block">
        <div class="target-row">
          <u-select v-model="t.method" :options="METHOD_OPTIONS" style="width: 110px" />
          <u-select
            v-if="t.method !== 'local'"
            v-model="t.server_id"
            :options="serverOptions"
            placeholder="服务器"
            style="width: 200px"
          />
          <u-input v-model="t.remote_path" placeholder="远程路径" style="flex: 1" />
          <u-button size="small" @click="removeTarget(idx)">删</u-button>
        </div>
        <u-textarea
          v-model="t.post_deploy_script"
          :rows="2"
          placeholder="部署后脚本（可选）"
          class="post-script"
        />
      </div>
    </FormDialog>
  </div>
</template>

<style scoped>
.page {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.page-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.page-head h2 {
  margin: 0;
  font-size: 18px;
}
.trigger-row {
  display: flex;
  gap: 16px;
  align-items: center;
  flex-wrap: wrap;
}
.targets-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin: 12px 0 8px;
}
.target-block {
  margin-bottom: 12px;
  padding: 8px;
  border: 1px solid rgba(0, 0, 0, 0.08);
  border-radius: 6px;
}
.target-row {
  display: flex;
  gap: 8px;
  align-items: center;
}
.post-script {
  margin-top: 8px;
  width: 100%;
}
</style>

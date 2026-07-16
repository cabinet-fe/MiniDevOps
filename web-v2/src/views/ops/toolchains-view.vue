<script setup lang="ts">
import { onUnmounted, reactive, ref, useTemplateRef } from "vue";
import { defineTableColumns, message } from "@veltra/desktop";

import {
  createInstallSource,
  createToolchain,
  deleteInstallSource,
  deleteToolchain,
  detectToolchain,
  enqueueToolchainOperation,
  getInstallJob,
  getInstallJobLogs,
  listInstallSources,
  listInstallJobs,
  pingInstallSource,
  retryInstallJob,
  updateInstallSource,
  updateToolchain,
} from "@/api/ops";
import type { InstallSource, ToolchainDefinition, ToolchainInstallJob } from "@/api/types";
import FormDialog from "@/components/form-dialog.vue";
import ProTable from "@/components/pro-table.vue";
import { usePermission } from "@/composables/use-permission";

const { hasPermission } = usePermission();
const toolchainList = useTemplateRef("toolchainList");
const jobList = useTemplateRef("jobList");
const sourceList = ref<InstallSource[]>([]);
const sourceLoading = ref(false);
const sourceDialogOpen = ref(false);
const toolchainDialogOpen = ref(false);
const editingToolchain = ref<ToolchainDefinition | null>(null);
const editingSource = ref<InstallSource | null>(null);
const logViewerOpen = ref(false);
const jobLog = ref("");
const jobLogTitle = ref("");
const jobQuery = reactive({ status: "" });
const viewedJobID = ref<number | null>(null);
const pendingJobIDs = new Set<number>();
let jobPollTimer: ReturnType<typeof setInterval> | undefined;

const toolchainForm = reactive({
  name: "",
  executable: "",
  description: "",
  detect_command: "",
  install_template: "",
  upgrade_template: "",
  uninstall_template: "",
  versions_command: "",
  switch_template: "",
  default_version: "",
});

const sourceForm = reactive({
  name: "",
  base_url: "",
  priority: 10,
  enabled: true,
});

const toolchainColumns = defineTableColumns([
  { key: "name", name: "工具链", minWidth: 130 },
  { key: "kind", name: "类型", width: 90, minWidth: 70 },
  { key: "executable", name: "可执行文件", minWidth: 120 },
  { key: "default_version", name: "默认版本", minWidth: 100 },
  { key: "action", name: "操作", width: 270, minWidth: 220 },
]);

const sourceColumns = defineTableColumns([
  { key: "name", name: "名称", minWidth: 120 },
  { key: "base_url", name: "地址", minWidth: 240 },
  { key: "priority", name: "优先级", width: 100, minWidth: 80 },
  { key: "enabled", name: "启用", width: 80, minWidth: 60 },
  { key: "action", name: "操作", width: 160, minWidth: 120 },
]);

const jobColumns = defineTableColumns([
  { key: "id", name: "ID", width: 70, minWidth: 60 },
  { key: "toolchain_id", name: "工具链", width: 90, minWidth: 70 },
  { key: "operation", name: "操作", width: 100, minWidth: 80 },
  { key: "requested_version", name: "版本", minWidth: 100 },
  { key: "status", name: "状态", width: 110, minWidth: 80 },
  { key: "created_at", name: "创建时间", minWidth: 160 },
  { key: "action", name: "操作", width: 120, minWidth: 100 },
]);

async function loadSources() {
  sourceLoading.value = true;
  try {
    sourceList.value = await listInstallSources();
  } catch (error) {
    showError(error);
  } finally {
    sourceLoading.value = false;
  }
}

function resetToolchainForm(item?: ToolchainDefinition) {
  Object.assign(toolchainForm, {
    name: item?.name ?? "",
    executable: item?.executable ?? "",
    description: item?.description ?? "",
    detect_command: item?.detect_command ?? "",
    install_template: item?.install_template ?? "",
    upgrade_template: item?.upgrade_template ?? "",
    uninstall_template: item?.uninstall_template ?? "",
    versions_command: item?.versions_command ?? "",
    switch_template: item?.switch_template ?? "",
    default_version: item?.default_version ?? "",
  });
}

function openCreateToolchain() {
  editingToolchain.value = null;
  resetToolchainForm();
  toolchainDialogOpen.value = true;
}

function openEditToolchain(item: ToolchainDefinition) {
  editingToolchain.value = item;
  resetToolchainForm(item);
  toolchainDialogOpen.value = true;
}

async function saveToolchain() {
  try {
    if (editingToolchain.value) {
      await updateToolchain(editingToolchain.value.id, toolchainForm);
      message.success("自定义工具链已更新");
    } else {
      await createToolchain(toolchainForm);
      message.success("自定义工具链已创建");
    }
    toolchainDialogOpen.value = false;
    await toolchainList.value?.reload();
  } catch (error) {
    showError(error);
  }
}

async function removeToolchain(item: ToolchainDefinition) {
  if (!window.confirm(`删除自定义工具链 ${item.name}？`)) return;
  try {
    await deleteToolchain(item.id);
    message.success("已删除");
    await toolchainList.value?.reload();
  } catch (error) {
    showError(error);
  }
}

async function runOperation(
  item: ToolchainDefinition,
  operation: "install" | "upgrade" | "uninstall" | "switch",
) {
  let version = "";
  if (operation !== "uninstall") {
    const requested = window.prompt(`输入 ${item.name} 的目标版本（可留空）`, item.default_version);
    if (requested === null) return;
    version = requested;
  }
  if (operation === "uninstall" && !window.confirm(`确认卸载 ${item.name}？`)) return;
  try {
    const job = await enqueueToolchainOperation(item.id, operation, version);
    message.success("安装任务已排队");
    trackJob(job);
    await jobList.value?.reload();
  } catch (error) {
    showError(error);
  }
}

async function runDetect(item: ToolchainDefinition) {
  try {
    const result = await detectToolchain(item.id);
    message[result.detected ? "success" : "warning"](
      result.detected
        ? `已检测到 ${item.name}`
        : `${item.name} 未检测到：${result.output || "无输出"}`,
    );
  } catch (error) {
    showError(error);
  }
}

function openCreateSource() {
  editingSource.value = null;
  Object.assign(sourceForm, { name: "", base_url: "", priority: 10, enabled: true });
  sourceDialogOpen.value = true;
}

function openEditSource(item: InstallSource) {
  editingSource.value = item;
  Object.assign(sourceForm, item);
  sourceDialogOpen.value = true;
}

async function saveSource() {
  try {
    if (editingSource.value) {
      await updateInstallSource(editingSource.value.id, sourceForm);
      message.success("安装源已更新");
    } else {
      await createInstallSource(sourceForm);
      message.success("安装源已创建");
    }
    sourceDialogOpen.value = false;
    await loadSources();
  } catch (error) {
    showError(error);
  }
}

async function removeSource(item: InstallSource) {
  if (!window.confirm(`删除安装源 ${item.name}？`)) return;
  try {
    await deleteInstallSource(item.id);
    await loadSources();
  } catch (error) {
    showError(error);
  }
}

async function pingSource(item: InstallSource) {
  try {
    const result = await pingInstallSource(item.id);
    message[result.ok ? "success" : "warning"](result.ok ? "连通性正常" : result.detail);
  } catch (error) {
    showError(error);
  }
}

async function showJobLog(item: ToolchainInstallJob) {
  try {
    jobLog.value = await getInstallJobLogs(item.id);
    jobLogTitle.value = `安装任务 #${item.id} 日志`;
    viewedJobID.value = item.id;
    logViewerOpen.value = true;
    trackJob(item);
  } catch (error) {
    showError(error);
  }
}

async function retryJob(item: ToolchainInstallJob) {
  try {
    const job = await retryInstallJob(item.id);
    message.success("已创建重试任务");
    trackJob(job);
    await jobList.value?.reload();
  } catch (error) {
    showError(error);
  }
}

function trackJob(job: ToolchainInstallJob) {
  if (["queued", "running"].includes(job.status)) {
    pendingJobIDs.add(job.id);
    ensureJobPolling();
  }
}

function ensureJobPolling() {
  if (jobPollTimer || pendingJobIDs.size === 0) return;
  jobPollTimer = setInterval(() => {
    void pollPendingJobs();
  }, 2000);
  void pollPendingJobs();
}

async function pollPendingJobs() {
  if (pendingJobIDs.size === 0) {
    stopJobPolling();
    return;
  }
  try {
    const jobs = await Promise.all([...pendingJobIDs].map((id) => getInstallJob(id)));
    const viewedJob = jobs.find((job) => job.id === viewedJobID.value);
    for (const job of jobs) {
      if (!["queued", "running"].includes(job.status)) {
        pendingJobIDs.delete(job.id);
      }
    }
    if (viewedJob) {
      jobLog.value = await getInstallJobLogs(viewedJob.id);
    }
    await jobList.value?.reload();
  } catch {
    // A transient poll failure should not close the task monitor.
  }
  if (pendingJobIDs.size === 0) stopJobPolling();
}

function stopJobPolling() {
  if (jobPollTimer) {
    clearInterval(jobPollTimer);
    jobPollTimer = undefined;
  }
}

async function resumePendingJobPolling() {
  try {
    const pages = await Promise.all(
      ["queued", "running"].map((status) => listInstallJobs({ status, page: 1, page_size: 100 })),
    );
    for (const page of pages) {
      for (const job of page.items) trackJob(job);
    }
  } catch {
    // Polling is supplemental; the table still reports a later API failure.
  }
}

function showError(error: unknown) {
  message.error(error instanceof Error ? error.message : "操作失败");
}

void loadSources();
void resumePendingJobPolling();
onUnmounted(stopJobPolling);
</script>

<template>
  <div class="page">
    <div class="page-head">
      <div>
        <h2>开发工具链</h2>
        <p>检测、安装和切换操作以异步任务执行；安装源按优先级依次回退。</p>
      </div>
      <u-button
        v-if="hasPermission('ops.toolchains:create')"
        type="primary"
        @click="openCreateToolchain"
      >
        新建自定义工具链
      </u-button>
    </div>

    <ProTable
      ref="toolchainList"
      url="/ops/toolchains"
      :columns="toolchainColumns"
      data-path="items"
      height="330px"
    >
      <template #column:action="{ rowData }">
        <u-action-group :max="5">
          <u-action @run="runDetect(rowData as ToolchainDefinition)">检测</u-action>
          <u-action @run="runOperation(rowData as ToolchainDefinition, 'install')">安装</u-action>
          <u-action @run="runOperation(rowData as ToolchainDefinition, 'upgrade')">升级</u-action>
          <u-action @run="runOperation(rowData as ToolchainDefinition, 'uninstall')">卸载</u-action>
          <u-action @run="runOperation(rowData as ToolchainDefinition, 'switch')">切版本</u-action>
          <u-action
            v-if="(rowData as ToolchainDefinition).kind === 'custom'"
            @run="openEditToolchain(rowData as ToolchainDefinition)"
            >编辑</u-action
          >
          <u-action
            v-if="(rowData as ToolchainDefinition).kind === 'custom'"
            @run="removeToolchain(rowData as ToolchainDefinition)"
            >删除</u-action
          >
        </u-action-group>
      </template>
    </ProTable>

    <section class="section">
      <div class="section-head">
        <h3>安装源</h3>
        <u-button @click="openCreateSource">新建安装源</u-button>
      </div>
      <div v-loading="sourceLoading" class="source-table">
        <u-table :columns="sourceColumns" :data="sourceList">
          <template #column:enabled="{ rowData }">
            {{ (rowData as InstallSource).enabled ? "是" : "否" }}
          </template>
          <template #column:action="{ rowData }">
            <u-action-group :max="3">
              <u-action @run="pingSource(rowData as InstallSource)">Ping</u-action>
              <u-action @run="openEditSource(rowData as InstallSource)">编辑</u-action>
              <u-action @run="removeSource(rowData as InstallSource)">删除</u-action>
            </u-action-group>
          </template>
        </u-table>
      </div>
    </section>

    <section class="section jobs">
      <h3>安装任务</h3>
      <ProTable
        ref="jobList"
        url="/ops/install-jobs"
        v-model:query="jobQuery"
        :columns="jobColumns"
        :auto-query-fields="['status']"
        pagination
      >
        <template #filters>
          <u-select
            v-model="jobQuery.status"
            clearable
            placeholder="状态"
            style="width: 140px"
            :options="[
              { label: 'queued', value: 'queued' },
              { label: 'running', value: 'running' },
              { label: 'success', value: 'success' },
              { label: 'failed', value: 'failed' },
              { label: 'interrupted', value: 'interrupted' },
            ]"
          />
        </template>
        <template #column:action="{ rowData }">
          <u-action-group :max="2">
            <u-action @run="showJobLog(rowData as ToolchainInstallJob)">查看日志</u-action>
            <u-action
              v-if="['failed', 'interrupted'].includes((rowData as ToolchainInstallJob).status)"
              @run="retryJob(rowData as ToolchainInstallJob)"
              >重试</u-action
            >
          </u-action-group>
        </template>
      </ProTable>
    </section>

    <FormDialog
      v-model="toolchainDialogOpen"
      :title="editingToolchain ? '编辑自定义工具链' : '新建自定义工具链'"
      :model="toolchainForm"
      label-width="120px"
      style="width: 700px"
      @submit="saveToolchain"
    >
      <div class="risk-warning">
        高风险：自定义命令会以 Bedrock 进程 UID
        直接执行，不是沙箱隔离。仅在完全理解命令及其权限影响时保存和执行。
      </div>
      <u-input label="名称" field="name" :rules="{ required: '必填' }" />
      <u-input label="可执行文件" field="executable" :rules="{ required: '必填' }" />
      <u-input label="描述" field="description" />
      <u-input label="检测命令" field="detect_command" />
      <u-textarea label="安装模板" field="install_template" />
      <u-textarea label="升级模板" field="upgrade_template" />
      <u-textarea label="卸载模板" field="uninstall_template" />
      <u-textarea label="切版本模板" field="switch_template" />
      <u-input label="默认版本" field="default_version" />
    </FormDialog>

    <FormDialog
      v-model="sourceDialogOpen"
      :title="editingSource ? '编辑安装源' : '新建安装源'"
      :model="sourceForm"
      label-width="100px"
      style="width: 560px"
      @submit="saveSource"
    >
      <u-input label="名称" field="name" :rules="{ required: '必填' }" />
      <u-input label="地址" field="base_url" :rules="{ required: '必填' }" />
      <u-input label="优先级" field="priority" type="number" />
      <u-select
        label="启用"
        field="enabled"
        :options="[
          { label: '启用', value: true },
          { label: '停用', value: false },
        ]"
      />
    </FormDialog>

    <u-dialog v-model="logViewerOpen" :title="jobLogTitle" style="width: 760px">
      <pre class="job-log">{{ jobLog || "暂无日志" }}</pre>
    </u-dialog>
  </div>
</template>

<style scoped>
.page,
.section {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.page-head,
.section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}
.page-head h2,
.section h3 {
  margin: 0;
  font-size: 18px;
}
.page-head p {
  margin: 6px 0 0;
  color: #6b7280;
}
.section {
  padding-top: 8px;
}
.source-table {
  min-height: 100px;
}
.jobs {
  min-height: 420px;
}
.risk-warning {
  margin-bottom: 12px;
  padding: 10px;
  border-radius: 6px;
  color: #92400e;
  background: #fef3c7;
  line-height: 1.6;
}
.job-log {
  max-height: 55vh;
  margin: 0;
  padding: 12px;
  overflow: auto;
  border-radius: 6px;
  color: #e5e7eb;
  background: #111827;
  white-space: pre-wrap;
}
</style>

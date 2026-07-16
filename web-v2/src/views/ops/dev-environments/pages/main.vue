<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref } from "vue";
import { o } from "@cat-kit/core";
import { message } from "@veltra/desktop";
import { Setting } from "@veltra/icons/normal";

import {
  createDevEnvSource,
  createDevEnvironment,
  deleteDevEnvSource,
  deleteDevEnvironment,
  detectDevEnvironment,
  enqueueDevEnvironmentOperation,
  getDevEnvJob,
  getDevEnvJobLogs,
  listDevEnvJobs,
  listDevEnvironments,
  pingDevEnvSource,
  retryDevEnvJob,
  updateDevEnvSource,
  updateDevEnvironment,
} from "@/api/ops";
import type { DevEnvInstallSource, DevEnvJob, DevEnvironment } from "@/api/types";
import defaultIcon from "@/assets/dev-env/default.svg";
import goIcon from "@/assets/dev-env/go.svg";
import javaIcon from "@/assets/dev-env/java.svg";
import nodeIcon from "@/assets/dev-env/nodejs.svg";
import pythonIcon from "@/assets/dev-env/python.svg";
import FormDialog from "@/components/form-dialog";
import { usePermission } from "@/composables/use-permission";

type DetectState = {
  status: "loading" | "detected" | "missing" | "error";
  version?: string;
  output?: string;
};

const ENV_ICONS: Record<string, string> = {
  go: goIcon,
  node: nodeIcon,
  java: javaIcon,
  python: pythonIcon,
  python3: pythonIcon,
};

const { hasPermission } = usePermission();

const loading = ref(false);
const environments = ref<DevEnvironment[]>([]);
const latestJobs = ref<Record<number, DevEnvJob | undefined>>({});
const detectStates = ref<Record<number, DetectState>>({});

const envDialogOpen = ref(false);
const editingEnv = ref<DevEnvironment | null>(null);
const scriptsDialogOpen = ref(false);
const scriptsEnv = ref<DevEnvironment | null>(null);
const sourcesDialogOpen = ref(false);
const sourcesEnvId = ref<number | null>(null);
const sourceDialogOpen = ref(false);
const editingSource = ref<DevEnvInstallSource | null>(null);
const sourceEnvID = ref<number | null>(null);

const logViewerOpen = ref(false);
const jobLog = ref("");
const jobLogTitle = ref("");
const viewedJob = ref<{ envId: number; jobId: number } | null>(null);

const pendingJobs = new Map<number, number>(); // jobId -> envId
let jobPollTimer: ReturnType<typeof setInterval> | undefined;

const envForm = reactive({
  name: "",
  executable: "",
  description: "",
  detect_script: "",
  install_script: "",
  upgrade_script: "",
  uninstall_script: "",
  versions_script: "",
  switch_script: "",
  default_version: "",
});

const sourceForm = reactive({
  name: "",
  base_url: "",
  priority: 10,
  enabled: true,
});

const scriptsReadOnly = computed(() => scriptsEnv.value?.kind === "builtin");
const sourcesEnv = computed(
  () => environments.value.find((item) => item.id === sourcesEnvId.value) ?? null,
);

function showError(error: unknown) {
  message.error(error instanceof Error ? error.message : "操作失败");
}

function envIcon(item: DevEnvironment): string {
  const exe = item.executable.toLowerCase();
  if (ENV_ICONS[exe]) return ENV_ICONS[exe];
  const name = item.name.toLowerCase();
  if (name.includes("node")) return nodeIcon;
  if (name.includes("python")) return pythonIcon;
  if (name.includes("java")) return javaIcon;
  if (name === "go" || name.includes("golang")) return goIcon;
  return defaultIcon;
}

function parseDetectedVersion(output: string): string {
  const text = output.trim();
  if (!text) return "已安装";
  const firstLine =
    text
      .split(/\r?\n/)
      .find((line) => line.trim())
      ?.trim() ?? text;

  const go = firstLine.match(/\bgo(\d+\.\d+(?:\.\d+)?)\b/i);
  if (go) return go[1];
  const python = firstLine.match(/\bPython\s+(\d+\.\d+(?:\.\d+)?)\b/i);
  if (python) return python[1];
  const java = text.match(/version\s+"([^"]+)"/i);
  if (java) return java[1];
  const node = firstLine.match(/\bv?(\d+\.\d+\.\d+)\b/);
  if (node) return node[0].startsWith("v") ? node[0] : `v${node[1]}`;
  const generic = firstLine.match(/\b(\d+\.\d+(?:\.\d+)?)\b/);
  if (generic) return generic[1];
  return firstLine.length > 48 ? `${firstLine.slice(0, 48)}…` : firstLine;
}

async function reload() {
  loading.value = true;
  try {
    const items = await listDevEnvironments();
    environments.value = items;
    await Promise.all(items.map((item) => refreshLatestJob(item.id)));
    void detectAll(items);
  } catch (error) {
    showError(error);
  } finally {
    loading.value = false;
  }
}

async function detectAll(items: DevEnvironment[]) {
  await Promise.all(items.map((item) => runDetect(item, { silent: true })));
}

async function refreshLatestJob(envId: number) {
  try {
    const page = await listDevEnvJobs(envId, { page: 1, page_size: 1 });
    const job = page.items[0];
    latestJobs.value = { ...latestJobs.value, [envId]: job };
    if (job && ["queued", "running"].includes(job.status)) {
      trackJob(envId, job.id);
    }
  } catch {
    // Supplemental; card still renders without a job.
  }
}

function openCreateEnv() {
  editingEnv.value = null;
  envDialogOpen.value = true;
}

function openEditEnv(item: DevEnvironment) {
  editingEnv.value = item;
  o(envForm).extend(item);
  envDialogOpen.value = true;
}

function openScripts(item: DevEnvironment) {
  scriptsEnv.value = item;
  o(envForm).extend(item);
  scriptsDialogOpen.value = true;
}

function openSourcesManager(item: DevEnvironment) {
  sourcesEnvId.value = item.id;
  sourcesDialogOpen.value = true;
}

async function saveEnv() {
  try {
    if (editingEnv.value) {
      await updateDevEnvironment(editingEnv.value.id, envForm);
      message.success("自定义开发环境已更新");
    } else {
      await createDevEnvironment(envForm);
      message.success("自定义开发环境已创建");
    }
    envDialogOpen.value = false;
    await reload();
  } catch (error) {
    showError(error);
  }
}

async function saveScripts() {
  if (!scriptsEnv.value) {
    scriptsDialogOpen.value = false;
    return;
  }
  if (scriptsEnv.value.kind !== "custom") {
    scriptsDialogOpen.value = false;
    return;
  }
  try {
    await updateDevEnvironment(scriptsEnv.value.id, envForm);
    message.success("命令行脚本已更新");
    scriptsDialogOpen.value = false;
    await reload();
  } catch (error) {
    showError(error);
  }
}

async function removeEnv(item: DevEnvironment) {
  if (!window.confirm(`删除自定义开发环境 ${item.name}？`)) return;
  try {
    await deleteDevEnvironment(item.id);
    message.success("已删除");
    await reload();
  } catch (error) {
    showError(error);
  }
}

async function runDetect(item: DevEnvironment, options?: { silent?: boolean }) {
  detectStates.value = {
    ...detectStates.value,
    [item.id]: { status: "loading", version: detectStates.value[item.id]?.version },
  };
  try {
    const result = await detectDevEnvironment(item.id);
    detectStates.value = {
      ...detectStates.value,
      [item.id]: result.detected
        ? {
            status: "detected",
            version: parseDetectedVersion(result.output),
            output: result.output,
          }
        : { status: "missing", output: result.output },
    };
    if (!options?.silent) {
      message[result.detected ? "success" : "warning"](
        result.detected ? `已检测到 ${item.name}` : `${item.name} 未检测到`,
      );
    }
  } catch (error) {
    detectStates.value = { ...detectStates.value, [item.id]: { status: "error" } };
    if (!options?.silent) showError(error);
  }
}

async function runOperation(
  item: DevEnvironment,
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
    const job = await enqueueDevEnvironmentOperation(item.id, operation, version);
    message.success("任务已排队");
    trackJob(item.id, job.id);
    latestJobs.value = { ...latestJobs.value, [item.id]: job };
  } catch (error) {
    showError(error);
  }
}

function openCreateSource(envId: number) {
  sourceEnvID.value = envId;
  editingSource.value = null;
  sourceDialogOpen.value = true;
}

function openEditSource(envId: number, source: DevEnvInstallSource) {
  sourceEnvID.value = envId;
  editingSource.value = source;
  o(sourceForm).extend(source);
  sourceDialogOpen.value = true;
}

async function saveSource() {
  if (sourceEnvID.value == null) return;
  try {
    if (editingSource.value) {
      await updateDevEnvSource(sourceEnvID.value, editingSource.value.id, sourceForm);
      message.success("安装源已更新");
    } else {
      await createDevEnvSource(sourceEnvID.value, sourceForm);
      message.success("安装源已创建");
    }
    sourceDialogOpen.value = false;
    await reload();
  } catch (error) {
    showError(error);
  }
}

async function removeSource(envId: number, source: DevEnvInstallSource) {
  if (!window.confirm(`删除安装源 ${source.name}？`)) return;
  try {
    await deleteDevEnvSource(envId, source.id);
    await reload();
  } catch (error) {
    showError(error);
  }
}

async function pingSource(envId: number, source: DevEnvInstallSource) {
  try {
    const result = await pingDevEnvSource(envId, source.id);
    message[result.ok ? "success" : "warning"](result.ok ? "连通性正常" : result.detail);
  } catch (error) {
    showError(error);
  }
}

async function showJobLog(envId: number, job: DevEnvJob) {
  try {
    jobLog.value = await getDevEnvJobLogs(envId, job.id);
    jobLogTitle.value = `${job.operation} #${job.id}`;
    viewedJob.value = { envId, jobId: job.id };
    logViewerOpen.value = true;
    if (["queued", "running"].includes(job.status)) trackJob(envId, job.id);
  } catch (error) {
    showError(error);
  }
}

async function retryJob(envId: number, job: DevEnvJob) {
  try {
    const next = await retryDevEnvJob(envId, job.id);
    message.success("已创建重试任务");
    trackJob(envId, next.id);
    latestJobs.value = { ...latestJobs.value, [envId]: next };
  } catch (error) {
    showError(error);
  }
}

function trackJob(envId: number, jobId: number) {
  pendingJobs.set(jobId, envId);
  ensureJobPolling();
}

function ensureJobPolling() {
  if (jobPollTimer || pendingJobs.size === 0) return;
  jobPollTimer = setInterval(() => {
    void pollPendingJobs();
  }, 2000);
  void pollPendingJobs();
}

async function pollPendingJobs() {
  if (pendingJobs.size === 0) {
    stopJobPolling();
    return;
  }
  try {
    const entries = [...pendingJobs.entries()];
    const jobs = await Promise.all(entries.map(([jobId, envId]) => getDevEnvJob(envId, jobId)));
    for (const job of jobs) {
      latestJobs.value = { ...latestJobs.value, [job.environment_id]: job };
      if (!["queued", "running"].includes(job.status)) {
        pendingJobs.delete(job.id);
        const env = environments.value.find((item) => item.id === job.environment_id);
        if (env && job.status === "success") {
          void runDetect(env, { silent: true });
        }
      }
      if (viewedJob.value?.jobId === job.id) {
        jobLog.value = await getDevEnvJobLogs(job.environment_id, job.id);
      }
    }
  } catch {
    // Transient poll failures should not stop monitoring.
  }
  if (pendingJobs.size === 0) stopJobPolling();
}

function stopJobPolling() {
  if (jobPollTimer) {
    clearInterval(jobPollTimer);
    jobPollTimer = undefined;
  }
}

function jobStatusLabel(status?: string) {
  return status || "暂无任务";
}

function versionTagType(state?: DetectState) {
  if (!state) return "info";
  if (state.status === "detected") return "success";
  if (state.status === "missing") return "warning";
  if (state.status === "error") return "danger";
  return "info";
}

function versionTagLabel(state?: DetectState) {
  if (!state || state.status === "loading") return "检测中…";
  if (state.status === "detected") return state.version || "已安装";
  if (state.status === "missing") return "未安装";
  return "检测失败";
}

onMounted(() => {
  void reload();
});
onUnmounted(stopJobPolling);
</script>

<template>
  <div v-loading="loading" class="page">
    <div class="page-head">
      <div>
        <h2>开发环境</h2>
        <p>管理宿主机语言运行时：安装源通过设置管理；进入页面时自动检测版本。</p>
      </div>
      <u-button
        v-if="hasPermission('ops.dev_environments:create')"
        type="primary"
        @click="openCreateEnv"
      >
        新建自定义环境
      </u-button>
    </div>

    <div class="cards">
      <article v-for="item in environments" :key="item.id" class="card">
        <header class="card-head">
          <div class="card-title">
            <div class="title-row">
              <img class="lang-icon" :src="envIcon(item)" :alt="item.name" width="28" height="28" />
              <h3>{{ item.name }}</h3>
              <u-tag size="small" :type="versionTagType(detectStates[item.id])">
                {{ versionTagLabel(detectStates[item.id]) }}
              </u-tag>
            </div>
            <p class="meta">
              <span>{{ item.kind === "builtin" ? "内置" : "自定义" }}</span>
              <span>{{ item.executable }}</span>
              <span v-if="item.default_version">默认 {{ item.default_version }}</span>
            </p>
            <p v-if="item.description" class="desc">{{ item.description }}</p>
          </div>
          <div class="actions">
            <u-button size="small" text :icon="Setting" @click="openSourcesManager(item)">
              设置
            </u-button>
            <u-action-group :max="5">
              <u-action @run="runDetect(item)">检测</u-action>
              <u-action @run="runOperation(item, 'install')">安装</u-action>
              <u-action @run="runOperation(item, 'upgrade')">升级</u-action>
              <u-action @run="runOperation(item, 'uninstall')">卸载</u-action>
              <u-action @run="runOperation(item, 'switch')">切版本</u-action>
              <u-action @run="openScripts(item)">脚本</u-action>
              <u-action v-if="item.kind === 'custom'" @run="openEditEnv(item)">编辑</u-action>
              <u-action v-if="item.kind === 'custom'" type="danger" @run="removeEnv(item)"
                >删除</u-action
              >
            </u-action-group>
          </div>
        </header>

        <section v-if="latestJobs[item.id]" class="block">
          <div class="block-head">
            <h4>最近任务</h4>
            <div class="actions">
              <span class="job-status">{{ jobStatusLabel(latestJobs[item.id]?.status) }}</span>
              <u-action-group :max="2">
                <u-action @run="showJobLog(item.id, latestJobs[item.id]!)">日志</u-action>
                <u-action
                  v-if="['failed', 'interrupted'].includes(latestJobs[item.id]?.status || '')"
                  @run="retryJob(item.id, latestJobs[item.id]!)"
                  >重试</u-action
                >
              </u-action-group>
            </div>
          </div>
          <p class="job-summary">
            {{ latestJobs[item.id]?.operation }}
            <template v-if="latestJobs[item.id]?.requested_version">
              · {{ latestJobs[item.id]?.requested_version }}
            </template>
          </p>
        </section>
      </article>
    </div>

    <FormDialog
      v-model="envDialogOpen"
      :title="editingEnv ? '编辑自定义开发环境' : '新建自定义开发环境'"
      :model="envForm"
      label-width="120px"
      style="width: 720px"
      @submit="saveEnv"
    >
      <div class="risk-warning">
        高风险：自定义命令会以 Bedrock 进程 UID
        直接执行，不是沙箱隔离。仅在完全理解命令及其权限影响时保存和执行。
      </div>
      <u-input label="名称" field="name" :rules="{ required: '必填' }" />
      <u-input label="可执行文件" field="executable" :rules="{ required: '必填' }" />
      <u-input label="描述" field="description" />
      <u-input label="默认版本" field="default_version" />
      <u-textarea label="检测脚本" field="detect_script" />
      <u-textarea label="安装脚本" field="install_script" />
      <u-textarea label="升级脚本" field="upgrade_script" />
      <u-textarea label="卸载脚本" field="uninstall_script" />
      <u-textarea label="切版本脚本" field="switch_script" />
      <u-textarea label="列版本脚本" field="versions_script" />
    </FormDialog>

    <FormDialog
      v-model="scriptsDialogOpen"
      :title="scriptsReadOnly ? `${scriptsEnv?.name} 命令行脚本（只读）` : '编辑命令行脚本'"
      :model="envForm"
      label-width="120px"
      style="width: 760px"
      :confirm-text="scriptsReadOnly ? '关闭' : '保存'"
      @submit="saveScripts"
    >
      <div v-if="!scriptsReadOnly" class="risk-warning">
        占位符：<span v-pre>{{ name }} / {{ executable }} / {{ version }} / {{ source_url }}</span>
      </div>
      <u-textarea label="检测脚本" field="detect_script" :readonly="scriptsReadOnly" />
      <u-textarea label="安装脚本" field="install_script" :readonly="scriptsReadOnly" />
      <u-textarea label="升级脚本" field="upgrade_script" :readonly="scriptsReadOnly" />
      <u-textarea label="卸载脚本" field="uninstall_script" :readonly="scriptsReadOnly" />
      <u-textarea label="切版本脚本" field="switch_script" :readonly="scriptsReadOnly" />
      <u-textarea label="列版本脚本" field="versions_script" :readonly="scriptsReadOnly" />
    </FormDialog>

    <u-dialog
      v-model="sourcesDialogOpen"
      :title="sourcesEnv ? `${sourcesEnv.name} · 安装源管理` : '安装源管理'"
      style="width: 640px"
    >
      <div class="sources-dialog">
        <div class="block-head">
          <h4>安装源列表</h4>
          <u-button
            v-if="sourcesEnv"
            size="small"
            text
            type="primary"
            @click="openCreateSource(sourcesEnv.id)"
          >
            添加
          </u-button>
        </div>
        <ul v-if="sourcesEnv?.sources?.length" class="source-list">
          <li v-for="source in sourcesEnv.sources" :key="source.id">
            <div class="source-info">
              <strong>{{ source.name }}</strong>
              <span class="source-url">{{ source.base_url }}</span>
              <span class="source-meta">
                优先级 {{ source.priority }} · {{ source.enabled ? "启用" : "停用" }}
              </span>
            </div>
            <div class="actions">
              <u-action-group :max="3">
                <u-action @run="pingSource(sourcesEnv.id, source)">Ping</u-action>
                <u-action @run="openEditSource(sourcesEnv.id, source)">编辑</u-action>
                <u-action type="danger" @run="removeSource(sourcesEnv.id, source)">删除</u-action>
              </u-action-group>
            </div>
          </li>
        </ul>
        <p v-else class="empty">尚未配置安装源</p>
      </div>
      <template #footer="{ close }">
        <u-button type="primary" @click="close()">关闭</u-button>
      </template>
    </u-dialog>

    <FormDialog
      v-model="sourceDialogOpen"
      :title="editingSource ? '编辑安装源' : '添加安装源'"
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
.page {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.page-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}
.page-head h2 {
  margin: 0;
  font-size: 18px;
}
.page-head p {
  margin: 6px 0 0;
  color: #6b7280;
  font-size: 13px;
  line-height: 1.5;
}
.cards {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(420px, 1fr));
  gap: 16px;
  align-items: start;
}
.card {
  display: flex;
  flex-direction: column;
  gap: 12px;
  min-width: 0;
  padding: 16px;
  border: 1px solid #e5e7eb;
  border-radius: 10px;
  background: #fff;
}
.card-head {
  display: flex;
  flex-direction: column;
  gap: 8px;
  min-width: 0;
}
.title-row {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
}
.lang-icon {
  flex-shrink: 0;
  width: 28px;
  height: 28px;
  object-fit: contain;
}
.card-title h3 {
  margin: 0;
  font-size: 16px;
  line-height: 1.4;
}
.actions {
  display: inline-flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 4px;
  min-width: 0;
}
.meta,
.desc,
.empty,
.job-summary,
.source-url,
.source-meta {
  margin: 4px 0 0;
  color: #6b7280;
  font-size: 13px;
  line-height: 1.4;
}
.meta {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
.block {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding-top: 12px;
  border-top: 1px solid #f3f4f6;
}
.block-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}
.block-head h4 {
  margin: 0;
  font-size: 14px;
}
.sources-dialog {
  display: flex;
  flex-direction: column;
  gap: 12px;
  min-height: 120px;
}
.source-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin: 0;
  padding: 0;
  list-style: none;
}
.source-list li {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  min-width: 0;
  padding: 10px 0;
  border-bottom: 1px solid #f3f4f6;
}
.source-list li:last-child {
  border-bottom: none;
}
.source-info {
  min-width: 0;
  flex: 1;
}
.source-info strong {
  display: block;
  font-size: 13px;
}
.source-url {
  display: block;
  word-break: break-all;
}
.job-status {
  font-size: 12px;
  color: #374151;
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

<script setup lang="ts">
defineOptions({ name: "AiClis" });

import { onMounted, reactive, ref } from "vue";
import { o } from "@cat-kit/core";
import { message } from "@veltra/desktop";

import {
  createCLISource,
  deleteCLISource,
  detectCLI,
  executeCLI,
  listCLIs,
  listCLISources,
  updateCLISource,
} from "@/api/ai";
import type { CliExecuteResult, CliInstallSource, CliRuntimeDefinition } from "@/api/types";
import FormDialog from "@/components/form-dialog";
import { usePermission } from "@/composables/use-permission";

type DetectState = {
  status: "loading" | "detected" | "missing" | "error";
  version?: string;
};

type Operation = "install" | "upgrade" | "uninstall";

const { hasPermission } = usePermission();

const loading = ref(false);
const items = ref<CliRuntimeDefinition[]>([]);
const riskNotice = ref("");
const detectStates = ref<Record<string, DetectState>>({});
const pendingOps = ref<Record<string, Operation | undefined>>({});

const sourcesDialogOpen = ref(false);
const sourcesCliKey = ref("");
const sourcesCliName = ref("");
const sourcesList = ref<CliInstallSource[]>([]);
const sourcesLoading = ref(false);

const sourceDialogOpen = ref(false);
const editingSource = ref<CliInstallSource | null>(null);
const sourceForm = reactive({
  cli_key: "claude_code",
  name: "",
  base_url: "",
  priority: 10,
  enabled: true,
});

const failureDialogOpen = ref(false);
const failureTitle = ref("");
const failureDetail = ref("");

function showError(error: unknown) {
  message.error(error instanceof Error ? error.message : "操作失败");
}

function versionTagType(state?: DetectState) {
  if (!state) return "info";
  if (state.status === "detected") return "success";
  if (state.status === "missing") return "warning";
  if (state.status === "error") return "danger";
  return "info";
}

function formatVersion(version?: string) {
  const trimmed = version?.trim();
  if (!trimmed || trimmed.includes("/") || trimmed.includes("\\")) return "";
  return trimmed;
}

function versionTagLabel(state?: DetectState) {
  if (!state || state.status === "loading") return "检测中…";
  if (state.status === "detected") return formatVersion(state.version) || "已安装";
  if (state.status === "missing") return "未安装";
  return "检测失败";
}

function formatFailureDetail(result: CliExecuteResult) {
  const parts = [result.output?.trim(), result.error?.trim()].filter(Boolean);
  return parts.join("\n\n") || "无输出";
}

async function reload() {
  loading.value = true;
  try {
    const data = await listCLIs();
    items.value = data.items ?? [];
    riskNotice.value = data.risk_notice ?? "";
    void detectAll(items.value);
  } catch (error) {
    showError(error);
  } finally {
    loading.value = false;
  }
}

async function detectAll(clis: CliRuntimeDefinition[]) {
  await Promise.all(clis.map((item) => runDetect(item, { silent: true })));
}

async function runDetect(item: CliRuntimeDefinition, options?: { silent?: boolean }) {
  detectStates.value = {
    ...detectStates.value,
    [item.key]: { status: "loading", version: detectStates.value[item.key]?.version },
  };
  try {
    const result = await detectCLI(item.key);
    detectStates.value = {
      ...detectStates.value,
      [item.key]: result.detected
        ? { status: "detected", version: formatVersion(result.version) || undefined }
        : { status: "missing" },
    };
    if (!options?.silent) {
      message[result.detected ? "success" : "warning"](
        result.detected ? `已检测到 ${item.name}` : `${item.name} 未安装`,
      );
    }
  } catch {
    detectStates.value = { ...detectStates.value, [item.key]: { status: "error" } };
    if (!options?.silent) {
      message.error("检测失败");
    }
  }
}

function isOpPending(key: string, op: Operation) {
  return pendingOps.value[key] === op;
}

async function runOperation(item: CliRuntimeDefinition, operation: Operation) {
  if (pendingOps.value[item.key]) return;

  let version = "";
  if (operation !== "uninstall") {
    const requested = window.prompt(`输入 ${item.name} 的目标版本（可留空）`);
    if (requested === null) return;
    version = requested;
  }
  if (operation === "uninstall" && !window.confirm(`确认卸载 ${item.name}？`)) return;

  pendingOps.value = { ...pendingOps.value, [item.key]: operation };
  try {
    const result = await executeCLI(item.key, operation, version);
    if (result.success) {
      message.success(`${item.name} ${operation} 完成`);
      await runDetect(item, { silent: true });
      return;
    }
    failureTitle.value = `${item.name} · ${operation} 失败`;
    failureDetail.value = formatFailureDetail(result);
    failureDialogOpen.value = true;
  } catch (error) {
    showError(error);
  } finally {
    const next = { ...pendingOps.value };
    delete next[item.key];
    pendingOps.value = next;
  }
}

async function openSourcesManager(item: CliRuntimeDefinition) {
  sourcesCliKey.value = item.key;
  sourcesCliName.value = item.name;
  sourcesDialogOpen.value = true;
  sourcesLoading.value = true;
  try {
    sourcesList.value = await listCLISources(item.key);
  } catch (error) {
    showError(error);
    sourcesList.value = [];
  } finally {
    sourcesLoading.value = false;
  }
}

async function refreshSources() {
  if (!sourcesCliKey.value) return;
  sourcesLoading.value = true;
  try {
    sourcesList.value = await listCLISources(sourcesCliKey.value);
  } catch (error) {
    showError(error);
  } finally {
    sourcesLoading.value = false;
  }
}

function openCreateSource() {
  editingSource.value = null;
  sourceForm.cli_key = sourcesCliKey.value;
  sourceDialogOpen.value = true;
}

function openEditSource(source: CliInstallSource) {
  editingSource.value = source;
  o(sourceForm).extend(source);
  sourceDialogOpen.value = true;
}

async function saveSource() {
  try {
    if (editingSource.value) {
      await updateCLISource(editingSource.value.id, sourceForm);
      message.success("安装源已更新");
    } else {
      await createCLISource({ ...sourceForm, cli_key: sourcesCliKey.value });
      message.success("安装源已创建");
    }
    sourceDialogOpen.value = false;
    await refreshSources();
  } catch (error) {
    showError(error);
  }
}

async function removeSource(source: CliInstallSource) {
  if (!window.confirm(`删除安装源 ${source.name}？`)) return;
  try {
    await deleteCLISource(source.id);
    await refreshSources();
  } catch (error) {
    showError(error);
  }
}

onMounted(() => {
  void reload();
});
</script>

<template>
  <div v-loading="loading" class="page">
    <div class="page-notice">
      <p class="risk">
        {{ riskNotice || "AI CLI 以 Bedrock 同 UID 执行，无 OS/容器沙箱。" }}
      </p>
      <p class="hint">进入页面时自动检测版本；安装源通过各 CLI 的设置管理。</p>
    </div>

    <div class="cards">
      <u-card v-for="item in items" :key="item.key" class="cli-card">
        <u-card-content>
          <header class="card-head">
            <div class="card-title">
              <div class="title-row">
                <h3>{{ item.name }}</h3>
                <u-tag size="small" :type="versionTagType(detectStates[item.key])">
                  {{ versionTagLabel(detectStates[item.key]) }}
                </u-tag>
              </div>
              <p class="meta">
                <span>{{ item.key }}</span>
                <span>{{ item.binary_name }}</span>
              </p>
              <p v-if="item.description" class="desc">{{ item.description }}</p>
            </div>
            <div class="actions">
              <u-action-group :max="5">
                <u-action v-if="hasPermission('ai.clis:view')" @run="openSourcesManager(item)">
                  设置
                </u-action>
                <u-action
                  v-if="hasPermission('ai.clis:execute')"
                  :disabled="!!pendingOps[item.key]"
                  @run="runDetect(item)"
                >
                  检测
                </u-action>
                <u-action
                  v-if="hasPermission('ai.clis:execute')"
                  :disabled="!!pendingOps[item.key]"
                  :loading="isOpPending(item.key, 'install')"
                  @run="runOperation(item, 'install')"
                >
                  安装
                </u-action>
                <u-action
                  v-if="hasPermission('ai.clis:execute')"
                  :disabled="!!pendingOps[item.key]"
                  :loading="isOpPending(item.key, 'upgrade')"
                  @run="runOperation(item, 'upgrade')"
                >
                  升级
                </u-action>
                <u-action
                  v-if="hasPermission('ai.clis:execute')"
                  :disabled="!!pendingOps[item.key]"
                  :loading="isOpPending(item.key, 'uninstall')"
                  type="danger"
                  @run="runOperation(item, 'uninstall')"
                >
                  卸载
                </u-action>
              </u-action-group>
            </div>
          </header>
        </u-card-content>
      </u-card>
    </div>

    <u-dialog
      v-model="sourcesDialogOpen"
      :title="sourcesCliName ? `${sourcesCliName} · 安装源管理` : '安装源管理'"
      style="width: 640px"
    >
      <div v-loading="sourcesLoading" class="sources-dialog">
        <div class="block-head">
          <h4>安装源列表</h4>
          <u-button
            v-if="hasPermission('ai.clis:create')"
            size="small"
            text
            type="primary"
            @click="openCreateSource"
          >
            添加
          </u-button>
        </div>
        <ul v-if="sourcesList.length" class="source-list">
          <li v-for="source in sourcesList" :key="source.id">
            <div class="source-info">
              <strong>{{ source.name }}</strong>
              <span class="source-url">{{ source.base_url }}</span>
              <span class="source-meta">
                优先级 {{ source.priority }} · {{ source.enabled ? "启用" : "停用" }}
              </span>
            </div>
            <div class="actions">
              <u-action-group :max="2">
                <u-action v-if="hasPermission('ai.clis:update')" @run="openEditSource(source)">
                  编辑
                </u-action>
                <u-action
                  v-if="hasPermission('ai.clis:delete')"
                  type="danger"
                  @run="removeSource(source)"
                >
                  删除
                </u-action>
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

    <u-dialog v-model="failureDialogOpen" :title="failureTitle" style="width: 760px">
      <pre class="failure-log">{{ failureDetail }}</pre>
      <template #footer="{ close }">
        <u-button type="primary" @click="close()">关闭</u-button>
      </template>
    </u-dialog>
  </div>
</template>

<style scoped lang="scss">
@use "pkg:@veltra/styles/functions" as fn;

.page {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.page-notice {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.risk {
  margin: 0;
  color: fn.use-var(color, warning);
  font-size: 13px;
  line-height: 1.5;
}
.hint {
  margin: 0;
  color: fn.use-var(text-color, second);
  font-size: 13px;
  line-height: 1.5;
}
.cards {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(420px, 1fr));
  gap: 16px;
  align-items: start;
}
.cli-card {
  min-width: 0;
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
.source-url,
.source-meta {
  margin: 4px 0 0;
  color: fn.use-var(text-color, second);
  font-size: 13px;
  line-height: 1.4;
}
.meta {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
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
  border-bottom: fn.use-var(border, muted);
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
.failure-log {
  max-height: 55vh;
  margin: 0;
  padding: 12px;
  overflow: auto;
  border-radius: fn.use-var(radius, small);
  color: fn.use-var(text-color, main);
  background: fn.use-var(bg-color, bottom);
  white-space: pre-wrap;
}
</style>

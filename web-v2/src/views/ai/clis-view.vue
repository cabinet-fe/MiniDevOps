<script setup lang="ts">
import { onMounted, reactive, ref, useTemplateRef } from "vue";
import { defineTableColumns, message } from "@veltra/desktop";

import {
  createCLISource,
  deleteCLISource,
  detectCLI,
  enqueueCLI,
  listCLIs,
  listCLISources,
  updateCLISource,
} from "@/api/ai";
import type { CliInstallJob, CliInstallSource, CliRuntimeDefinition } from "@/api/types";
import FormDialog from "@/components/form-dialog.vue";
import ProTable from "@/components/pro-table.vue";
import { usePermission } from "@/composables/use-permission";

const { hasPermission } = usePermission();
const items = ref<CliRuntimeDefinition[]>([]);
const riskNotice = ref("");
const loading = ref(false);
const jobList = useTemplateRef("jobList");
const jobQuery = reactive({ cli_key: "", status: "" });

const sourceList = ref<CliInstallSource[]>([]);
const sourceLoading = ref(false);
const sourceDialogOpen = ref(false);
const editingSource = ref<CliInstallSource | null>(null);
const sourceForm = reactive({
  cli_key: "claude_code",
  name: "",
  base_url: "",
  priority: 10,
  enabled: true,
});

const cliKeyOptions = [
  { label: "Claude Code", value: "claude_code" },
  { label: "OpenCode", value: "opencode" },
  { label: "Reasonix", value: "reasonix" },
  { label: "Codex", value: "codex" },
];

const columns = defineTableColumns([
  { key: "name", name: "名称", minWidth: 120 },
  { key: "key", name: "Key", width: 120 },
  { key: "binary_name", name: "二进制", width: 110 },
  { key: "installed_version", name: "版本", minWidth: 100 },
  { key: "install_status", name: "状态", width: 100 },
  { key: "healthy", name: "健康", width: 80 },
  { key: "action", name: "操作", width: 280, minWidth: 240 },
]);

const sourceColumns = defineTableColumns([
  { key: "cli_key", name: "CLI", width: 120 },
  { key: "name", name: "名称", minWidth: 120 },
  { key: "base_url", name: "地址", minWidth: 240 },
  { key: "priority", name: "优先级", width: 100, minWidth: 80 },
  { key: "enabled", name: "启用", width: 80, minWidth: 60 },
  { key: "action", name: "操作", width: 140, minWidth: 120 },
]);

const jobColumns = defineTableColumns([
  { key: "id", name: "ID", width: 70 },
  { key: "cli_key", name: "CLI", width: 120 },
  { key: "operation", name: "操作", width: 100 },
  { key: "status", name: "状态", width: 110 },
  { key: "created_at", name: "创建时间", minWidth: 160 },
]);

async function reload() {
  loading.value = true;
  try {
    const data = await listCLIs();
    items.value = data.items ?? [];
    riskNotice.value = data.risk_notice ?? "";
  } catch (error) {
    message.error(error instanceof Error ? error.message : "加载失败");
  } finally {
    loading.value = false;
  }
}

async function loadSources() {
  sourceLoading.value = true;
  try {
    sourceList.value = await listCLISources();
  } catch (error) {
    message.error(error instanceof Error ? error.message : "加载安装源失败");
  } finally {
    sourceLoading.value = false;
  }
}

async function onDetect(row: CliRuntimeDefinition) {
  try {
    const result = await detectCLI(row.key);
    message.info(
      result.detected ? `已检测到: ${result.version || result.path}` : result.output || "未安装",
    );
    await reload();
  } catch (error) {
    message.error(error instanceof Error ? error.message : "检测失败");
  }
}

async function onOp(row: CliRuntimeDefinition, op: "install" | "upgrade" | "uninstall") {
  try {
    await enqueueCLI(row.key, op);
    message.success("已提交任务（同 UID 执行，无沙箱）");
    jobList.value?.reload();
  } catch (error) {
    message.error(error instanceof Error ? error.message : "提交失败");
  }
}

function openCreateSource() {
  editingSource.value = null;
  Object.assign(sourceForm, {
    cli_key: "claude_code",
    name: "",
    base_url: "",
    priority: 10,
    enabled: true,
  });
  sourceDialogOpen.value = true;
}

function openEditSource(item: CliInstallSource) {
  editingSource.value = item;
  Object.assign(sourceForm, {
    cli_key: item.cli_key,
    name: item.name,
    base_url: item.base_url,
    priority: item.priority,
    enabled: item.enabled,
  });
  sourceDialogOpen.value = true;
}

async function saveSource() {
  try {
    if (editingSource.value) {
      await updateCLISource(editingSource.value.id, sourceForm);
      message.success("安装源已更新");
    } else {
      await createCLISource(sourceForm);
      message.success("安装源已创建");
    }
    sourceDialogOpen.value = false;
    await loadSources();
  } catch (error) {
    message.error(error instanceof Error ? error.message : "保存失败");
  }
}

async function removeSource(item: CliInstallSource) {
  if (!window.confirm(`删除安装源 ${item.name}？`)) return;
  try {
    await deleteCLISource(item.id);
    message.success("已删除");
    await loadSources();
  } catch (error) {
    message.error(error instanceof Error ? error.message : "删除失败");
  }
}

onMounted(() => {
  void reload();
  void loadSources();
});
</script>

<template>
  <div class="page">
    <header class="page-head">
      <h2>AI CLI</h2>
      <p class="risk">{{ riskNotice || "AI CLI 以 Bedrock 同 UID 执行，无 OS/容器沙箱。" }}</p>
    </header>

    <u-table :columns="columns" :data="items" v-loading="loading">
      <template #healthy="{ rowData }">
        {{ (rowData as CliRuntimeDefinition).healthy ? "是" : "否" }}
      </template>
      <template #action="{ rowData }">
        <u-action-group :max="4">
          <u-action
            v-if="hasPermission('ai.clis:execute')"
            @run="onDetect(rowData as CliRuntimeDefinition)"
          >
            检测
          </u-action>
          <u-action
            v-if="hasPermission('ai.clis:execute')"
            @run="onOp(rowData as CliRuntimeDefinition, 'install')"
          >
            安装
          </u-action>
          <u-action
            v-if="hasPermission('ai.clis:execute')"
            @run="onOp(rowData as CliRuntimeDefinition, 'upgrade')"
          >
            升级
          </u-action>
          <u-action
            v-if="hasPermission('ai.clis:execute')"
            type="danger"
            @run="onOp(rowData as CliRuntimeDefinition, 'uninstall')"
          >
            卸载
          </u-action>
        </u-action-group>
      </template>
    </u-table>

    <section class="section">
      <div class="section-head">
        <h3>安装源</h3>
        <u-button v-if="hasPermission('ai.clis:create')" type="primary" @click="openCreateSource">
          新建安装源
        </u-button>
      </div>
      <div v-loading="sourceLoading" class="source-table">
        <u-table :columns="sourceColumns" :data="sourceList">
          <template #enabled="{ rowData }">
            {{ (rowData as CliInstallSource).enabled ? "是" : "否" }}
          </template>
          <template #action="{ rowData }">
            <u-action-group :max="2">
              <u-action
                v-if="hasPermission('ai.clis:update')"
                @run="openEditSource(rowData as CliInstallSource)"
              >
                编辑
              </u-action>
              <u-action
                v-if="hasPermission('ai.clis:delete')"
                type="danger"
                @run="removeSource(rowData as CliInstallSource)"
              >
                删除
              </u-action>
            </u-action-group>
          </template>
        </u-table>
      </div>
    </section>

    <h3 class="section-title">安装任务</h3>
    <ProTable
      ref="jobList"
      url="/ai/cli-install-jobs"
      mode="pagination"
      :columns="jobColumns"
      v-model:query="jobQuery"
    >
      <template #filters>
        <u-input v-model="jobQuery.cli_key" placeholder="cli_key" clearable />
        <u-input v-model="jobQuery.status" placeholder="status" clearable />
      </template>
      <template #status="{ rowData }">
        {{ (rowData as CliInstallJob).status }}
      </template>
    </ProTable>

    <FormDialog
      v-model="sourceDialogOpen"
      :title="editingSource ? '编辑安装源' : '新建安装源'"
      :model="sourceForm"
      label-width="100px"
      style="width: 560px"
      @submit="saveSource"
    >
      <u-select
        label="CLI"
        field="cli_key"
        :options="cliKeyOptions"
        :rules="{ required: '必填' }"
        :disabled="!!editingSource"
      />
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
  </div>
</template>

<style scoped lang="scss">
.page {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.page-head h2 {
  margin: 0 0 8px;
}
.risk {
  margin: 0;
  color: var(--u-color-warning, #b45309);
  font-size: 13px;
}
.section {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}
.section-head h3,
.section-title {
  margin: 8px 0 0;
}
.source-table {
  min-height: 100px;
}
</style>

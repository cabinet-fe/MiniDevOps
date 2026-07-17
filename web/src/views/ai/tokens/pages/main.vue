<script setup lang="ts">
defineOptions({ name: "AiTokens" });

import { onMounted, reactive, ref } from "vue";
import { defineTableColumns, message } from "@veltra/desktop";

import { createToken, deleteToken, listTokens } from "@/api/ai";
import type { PersonalAccessToken } from "@/api/types";
import FormDialog from "@/components/form-dialog";
import { formatDateTime } from "@/lib/datetime";

const items = ref<PersonalAccessToken[]>([]);
const loading = ref(false);
const dialogOpen = ref(false);
const plaintext = ref("");
const form = reactive({
  name: "",
  scopeSkills: true,
  scopeAgents: false,
});

const columns = defineTableColumns([
  { key: "id", name: "ID", width: 70 },
  { key: "name", name: "名称", minWidth: 120 },
  { key: "token_prefix", name: "前缀", width: 140 },
  { key: "scopes", name: "Scope", minWidth: 160 },
  { key: "revoked_at", name: "状态", width: 100 },
  {
    key: "created_at",
    name: "创建时间",
    minWidth: 160,
    render: ({ val }) => formatDateTime(val),
  },
  { key: "action", name: "操作", width: 100, fixed: "right" },
]);

async function reload() {
  loading.value = true;
  try {
    items.value = await listTokens();
  } catch (error) {
    message.error(error instanceof Error ? error.message : "加载失败");
  } finally {
    loading.value = false;
  }
}

function openCreate() {
  form.name = "";
  form.scopeSkills = true;
  form.scopeAgents = false;
  plaintext.value = "";
  dialogOpen.value = true;
}

async function save() {
  const scopes: string[] = [];
  if (form.scopeSkills) scopes.push("skills:read");
  if (form.scopeAgents) scopes.push("agents:run");
  if (!scopes.length) {
    message.error("至少选择一个 scope");
    return;
  }
  try {
    const result = await createToken({ name: form.name, scopes });
    plaintext.value = result.token;
    message.success("令牌已创建，请立即复制明文（仅显示一次）");
    await reload();
  } catch (error) {
    message.error(error instanceof Error ? error.message : "创建失败");
  }
}

async function remove(row: PersonalAccessToken) {
  try {
    await deleteToken(row.id);
    message.success("已删除");
    await reload();
  } catch (error) {
    message.error(error instanceof Error ? error.message : "删除失败");
  }
}

onMounted(() => {
  void reload();
});
</script>

<template>
  <div>
    <div class="page-toolbar">
      <u-button type="primary" @click="openCreate">创建令牌</u-button>
    </div>

    <u-table :columns="columns" :data="items" v-loading="loading">
      <template #column:scopes="{ rowData }">
        <span class="tag-cell">
          <u-tag
            v-for="scope in (rowData as PersonalAccessToken).scopes || []"
            :key="scope"
            size="small"
            type="info"
          >
            {{ scope }}
          </u-tag>
        </span>
      </template>
      <template #column:revoked_at="{ rowData }">
        <u-tag
          size="small"
          :type="(rowData as PersonalAccessToken).revoked_at ? 'danger' : 'success'"
        >
          {{ (rowData as PersonalAccessToken).revoked_at ? "已吊销" : "有效" }}
        </u-tag>
      </template>
      <template #column:action="{ rowData }">
        <u-action-group :max="2">
          <u-action need-confirm type="danger" @run="remove(rowData as PersonalAccessToken)">
            删除
          </u-action>
        </u-action-group>
      </template>
    </u-table>

    <FormDialog
      v-model="dialogOpen"
      title="创建 PAT"
      :model="form"
      label-width="90px"
      style="width: 480px"
      @submit="save"
    >
      <u-input label="名称" field="name" :rules="{ required: '必填' }" />
      <u-form-item label="Scope">
        <div class="scope-row">
          <u-checkbox v-model="form.scopeSkills">skills:read</u-checkbox>
          <u-checkbox v-model="form.scopeAgents">agents:run</u-checkbox>
        </div>
      </u-form-item>
      <div v-if="plaintext" class="once">
        <strong>明文（仅此一次）：</strong>
        <code>{{ plaintext }}</code>
      </div>
    </FormDialog>
  </div>
</template>

<style scoped lang="scss">
.page-toolbar {
  display: flex;
  justify-content: flex-end;
}
.scope-row {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}
.tag-cell {
  display: inline-flex;
  flex-wrap: wrap;
  gap: 4px;
}
.once {
  margin-top: 12px;
  padding: 10px;
  background: var(--u-color-warning-bg, #fff7ed);
  border-radius: 6px;
  word-break: break-all;
}
</style>

<script setup lang="ts">
import { onMounted, reactive, ref } from "vue";
import { defineTableColumns, message } from "@veltra/desktop";

import { createToken, listTokens, revokeToken } from "@/api/ai";
import type { PersonalAccessToken } from "@/api/types";
import FormDialog from "@/components/form-dialog.vue";

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
  { key: "revoked_at", name: "吊销", width: 100 },
  { key: "created_at", name: "创建时间", minWidth: 160 },
  { key: "action", name: "操作", width: 100 },
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

async function revoke(row: PersonalAccessToken) {
  try {
    await revokeToken(row.id);
    await reload();
  } catch (error) {
    message.error(error instanceof Error ? error.message : "吊销失败");
  }
}

onMounted(() => {
  void reload();
});
</script>

<template>
  <div class="page">
    <header class="page-head">
      <h2>个人访问令牌</h2>
      <p>
        Scope 固定为 skills:read / agents:run；哈希存储，明文仅创建时回显一次。PAT 不替代
        HTTPS/TLS。
      </p>
      <u-button type="primary" @click="openCreate">创建令牌</u-button>
    </header>

    <u-table :columns="columns" :data="items" v-loading="loading">
      <template #scopes="{ rowData }">
        {{ ((rowData as PersonalAccessToken).scopes || []).join(", ") }}
      </template>
      <template #revoked_at="{ rowData }">
        {{ (rowData as PersonalAccessToken).revoked_at ? "已吊销" : "有效" }}
      </template>
      <template #action="{ rowData }">
        <u-action
          v-if="!(rowData as PersonalAccessToken).revoked_at"
          type="danger"
          @run="revoke(rowData as PersonalAccessToken)"
        >
          吊销
        </u-action>
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
.page {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.page-head p {
  margin: 4px 0 12px;
  font-size: 13px;
  color: var(--u-color-text-secondary, #666);
}
.scope-row {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}
.once {
  margin-top: 12px;
  padding: 10px;
  background: var(--u-color-warning-bg, #fff7ed);
  border-radius: 6px;
  word-break: break-all;
}
</style>

<script setup lang="ts">
import { reactive, ref, useTemplateRef } from "vue";
import { defineTableColumns, message } from "@veltra/desktop";

import { createCredential, deleteCredential, listCredentials, updateCredential } from "@/api/cicd";
import type { Credential } from "@/api/types";
import FormDialog from "@/components/form-dialog.vue";
import ResourceList from "@/components/resource-list.vue";
import { usePermission } from "@/composables/use-permission";

const { hasPermission } = usePermission();
const listRef = useTemplateRef("list");
const filters = reactive({ keyword: "" });
const dialogOpen = ref(false);
const editing = ref<Credential | null>(null);
const form = reactive({
  name: "",
  type: "token",
  username: "",
  secret: "",
  passphrase: "",
  description: "",
});

const columns = defineTableColumns([
  { key: "id", name: "ID", width: 80, minWidth: 60 },
  { key: "name", name: "名称", minWidth: 140 },
  { key: "type", name: "类型", width: 100, minWidth: 80 },
  { key: "username", name: "用户名", minWidth: 120 },
  { key: "has_secret", name: "密文", width: 80, minWidth: 60 },
  { key: "action", name: "操作", width: 160, minWidth: 120 },
]);

async function fetcher(params: { page: number; page_size: number }) {
  return listCredentials({ ...params, keyword: filters.keyword });
}

function openCreate() {
  editing.value = null;
  Object.assign(form, {
    name: "",
    type: "token",
    username: "",
    secret: "",
    passphrase: "",
    description: "",
  });
  dialogOpen.value = true;
}

function openEdit(row: Credential) {
  editing.value = row;
  Object.assign(form, {
    name: row.name,
    type: row.type,
    username: row.username || "",
    secret: "",
    passphrase: "",
    description: row.description || "",
  });
  dialogOpen.value = true;
}

async function save() {
  try {
    if (editing.value) {
      const body: Record<string, unknown> = {
        name: form.name,
        type: form.type,
        username: form.username,
        description: form.description,
      };
      if (form.secret) body.secret = form.secret;
      if (form.passphrase) body.passphrase = form.passphrase;
      await updateCredential(editing.value.id, body);
      message.success("已更新");
    } else {
      await createCredential({
        name: form.name,
        type: form.type,
        username: form.username,
        secret: form.secret,
        passphrase: form.passphrase || undefined,
        description: form.description,
      });
      message.success("已创建");
    }
    dialogOpen.value = false;
    await listRef.value?.refresh();
  } catch (err) {
    message.error(err instanceof Error ? err.message : "保存失败");
  }
}

async function remove(row: Credential) {
  try {
    await deleteCredential(row.id);
    message.success("已删除");
    await listRef.value?.refresh();
  } catch (err) {
    message.error(err instanceof Error ? err.message : "删除失败");
  }
}
</script>

<template>
  <div class="page">
    <div class="page-head">
      <h2>凭证</h2>
      <u-button v-if="hasPermission('cicd.credentials:create')" type="primary" @click="openCreate">
        新建凭证
      </u-button>
    </div>

    <ResourceList ref="list" :fetcher="fetcher" :columns="columns" :filters="filters">
      <template #filters="{ reload }">
        <u-input v-model="filters.keyword" placeholder="名称关键词" style="width: 200px" />
        <u-button @click="reload">刷新</u-button>
      </template>
      <template #column:has_secret="{ rowData }">
        {{ (rowData as Credential).has_secret ? "已设置" : "无" }}
      </template>
      <template #column:action="{ rowData }">
        <u-action-group :max="3">
          <u-action
            v-if="hasPermission('cicd.credentials:update')"
            @run="openEdit(rowData as Credential)"
          >
            编辑
          </u-action>
          <u-action
            v-if="hasPermission('cicd.credentials:delete')"
            @run="remove(rowData as Credential)"
          >
            删除
          </u-action>
        </u-action-group>
      </template>
    </ResourceList>

    <FormDialog
      v-model="dialogOpen"
      :title="editing ? '编辑凭证' : '新建凭证'"
      :model="form"
      label-width="110px"
      style="width: 520px"
      @submit="save"
    >
      <u-input label="名称" field="name" :rules="{ required: '必填' }" />
      <u-select
        label="类型"
        field="type"
        :options="[
          { label: 'password', value: 'password' },
          { label: 'token', value: 'token' },
          { label: 'ssh_key', value: 'ssh_key' },
          { label: 'api_key', value: 'api_key' },
        ]"
        :rules="{ required: '必填' }"
      />
      <u-input label="用户名" field="username" />
      <u-password-input
        :label="editing ? '密文（留空不改）' : '密文'"
        field="secret"
        autocomplete="new-password"
        :rules="editing ? undefined : { required: '必填' }"
      />
      <u-password-input
        v-if="form.type === 'ssh_key'"
        label="口令（留空不改）"
        field="passphrase"
        autocomplete="new-password"
      />
      <u-input label="描述" field="description" />
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
</style>

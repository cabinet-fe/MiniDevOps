<script setup lang="ts">
defineOptions({ name: "CicdCredentials" });

import { reactive, ref, useTemplateRef } from "vue";
import { o } from "@cat-kit/core";
import { message } from "@veltra/desktop";

import { createCredential, deleteCredential, updateCredential } from "@/api/cicd";
import type { Credential } from "@/api/types";
import FormDialog from "@/components/form-dialog";
import ProTable, { defineProTableColumns } from "@/components/pro-table";
import { usePermission } from "@/composables/use-permission";
import { tagType, type TagType } from "@/lib/tag";

const CRED_TYPE_TAG: Record<string, TagType> = {
  token: "primary",
  api_key: "primary",
  password: "warning",
  ssh_key: "info",
};

const { hasPermission } = usePermission();
const listRef = useTemplateRef("list");
const query = reactive({ keyword: "" });
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

const columns = defineProTableColumns([
  { key: "id", name: "ID", width: 80 },
  { key: "name", name: "名称" },
  { key: "type", name: "类型", width: 100 },
  { key: "username", name: "用户名" },
  { key: "has_secret", name: "密文", width: 80 },
  { key: "action", name: "操作", width: 160, align: "center", fixed: "right" },
]);

function openCreate() {
  editing.value = null;
  dialogOpen.value = true;
}

function openEdit(row: Credential) {
  editing.value = row;
  o(form).extend(row);
  form.secret = "";
  form.passphrase = "";
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
    await listRef.value?.reload();
  } catch (err) {
    message.error(err instanceof Error ? err.message : "保存失败");
  }
}

async function remove(row: Credential) {
  try {
    await deleteCredential(row.id);
    message.success("已删除");
    await listRef.value?.reload();
  } catch (err) {
    message.error(err instanceof Error ? err.message : "删除失败");
  }
}
</script>

<template>
  <div class="page">
    <ProTable ref="list" url="/credentials" v-model:query="query" :columns="columns" pagination>
      <template #filters="{ search }">
        <u-input v-model="query.keyword" placeholder="名称关键词" style="width: 200px" />
        <u-button type="primary" @click="search">查询</u-button>
        <u-button
          v-if="hasPermission('cicd.credentials:create')"
          type="primary"
          style="margin-left: auto"
          @click.prevent="openCreate"
        >
          新建凭证
        </u-button>
      </template>
      <template #column:type="{ rowData }">
        <u-tag size="small" :type="tagType((rowData as Credential).type, CRED_TYPE_TAG)">
          {{ (rowData as Credential).type }}
        </u-tag>
      </template>
      <template #column:has_secret="{ rowData }">
        <u-tag size="small" :type="(rowData as Credential).has_secret ? 'success' : 'warning'">
          {{ (rowData as Credential).has_secret ? "已设置" : "无" }}
        </u-tag>
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
    </ProTable>

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
</style>

<script setup lang="ts">
import { onMounted, reactive, ref, useTemplateRef } from "vue";
import { defineTableColumns, message } from "@veltra/desktop";

import {
  createServer,
  deleteServer,
  listCredentials,
  listServers,
  testServer,
  updateServer,
} from "@/api/cicd";
import type { Credential, Server } from "@/api/types";
import FormDialog from "@/components/form-dialog.vue";
import ResourceList from "@/components/resource-list.vue";
import { usePermission } from "@/composables/use-permission";

const { hasPermission } = usePermission();
const listRef = useTemplateRef("list");
const filters = reactive({ keyword: "" });
const dialogOpen = ref(false);
const editing = ref<Server | null>(null);
const credOptions = ref<{ label: string; value: number }[]>([]);
const form = reactive({
  name: "",
  host: "",
  port: 22,
  os_type: "linux",
  username: "",
  auth_type: "password",
  credential_id: undefined as number | undefined,
  agent_url: "",
  agent_credential_id: undefined as number | undefined,
  description: "",
  tags: "",
});

const columns = defineTableColumns([
  { key: "id", name: "ID", width: 80, minWidth: 60 },
  { key: "name", name: "名称", minWidth: 120 },
  { key: "host", name: "主机", minWidth: 140 },
  { key: "port", name: "端口", width: 80, minWidth: 60 },
  { key: "auth_type", name: "认证", width: 100, minWidth: 80 },
  { key: "status", name: "状态", width: 100, minWidth: 80 },
  { key: "action", name: "操作", width: 220, minWidth: 160 },
]);

async function fetcher(params: { page: number; page_size: number }) {
  return listServers({ ...params, keyword: filters.keyword });
}

onMounted(async () => {
  try {
    const res = await listCredentials({ page: 1, page_size: 100 });
    credOptions.value = (res.items ?? []).map((c: Credential) => ({
      label: `${c.name} (${c.type})`,
      value: c.id,
    }));
  } catch {
    /* ignore */
  }
});

function openCreate() {
  editing.value = null;
  Object.assign(form, {
    name: "",
    host: "",
    port: 22,
    os_type: "linux",
    username: "",
    auth_type: "password",
    credential_id: undefined,
    agent_url: "",
    agent_credential_id: undefined,
    description: "",
    tags: "",
  });
  dialogOpen.value = true;
}

function openEdit(row: Server) {
  editing.value = row;
  Object.assign(form, {
    name: row.name,
    host: row.host,
    port: row.port || 22,
    os_type: row.os_type || "linux",
    username: row.username || "",
    auth_type: row.auth_type || "password",
    credential_id: row.credential_id ?? undefined,
    agent_url: row.agent_url || "",
    agent_credential_id: row.agent_credential_id ?? undefined,
    description: row.description || "",
    tags: row.tags || "",
  });
  dialogOpen.value = true;
}

async function save() {
  try {
    const body: Record<string, unknown> = { ...form };
    if (!form.credential_id) {
      delete body.credential_id;
      if (editing.value) body.clear_credential = true;
    }
    if (!form.agent_credential_id) {
      delete body.agent_credential_id;
      if (editing.value) body.clear_agent_credential = true;
    }
    if (editing.value) {
      await updateServer(editing.value.id, body);
      message.success("已更新");
    } else {
      await createServer(body);
      message.success("已创建");
    }
    dialogOpen.value = false;
    await listRef.value?.refresh();
  } catch (err) {
    message.error(err instanceof Error ? err.message : "保存失败");
  }
}

async function remove(row: Server) {
  try {
    await deleteServer(row.id);
    message.success("已删除");
    await listRef.value?.refresh();
  } catch (err) {
    message.error(err instanceof Error ? err.message : "删除失败");
  }
}

async function onTest(row: Server) {
  try {
    const res = await testServer(row.id);
    message.success(res.output?.slice(0, 120) || "连接成功");
    await listRef.value?.refresh();
  } catch (err) {
    message.error(err instanceof Error ? err.message : "连接失败");
  }
}
</script>

<template>
  <div class="page">
    <div class="page-head">
      <h2>服务器</h2>
      <u-button v-if="hasPermission('cicd.servers:create')" type="primary" @click="openCreate">
        新建服务器
      </u-button>
    </div>

    <ResourceList ref="list" :fetcher="fetcher" :columns="columns" :filters="filters">
      <template #filters="{ reload }">
        <u-input v-model="filters.keyword" placeholder="名称/主机" style="width: 200px" />
        <u-button @click="reload">刷新</u-button>
      </template>
      <template #column:action="{ rowData }">
        <u-action-group :max="4">
          <u-action v-if="hasPermission('cicd.servers:update')" @run="openEdit(rowData as Server)">
            编辑
          </u-action>
          <u-action v-if="hasPermission('cicd.servers:view')" @run="onTest(rowData as Server)">
            测试
          </u-action>
          <u-action v-if="hasPermission('cicd.servers:delete')" @run="remove(rowData as Server)">
            删除
          </u-action>
        </u-action-group>
      </template>
    </ResourceList>

    <FormDialog
      v-model="dialogOpen"
      :title="editing ? '编辑服务器' : '新建服务器'"
      :model="form"
      label-width="110px"
      style="width: 560px"
      @submit="save"
    >
      <u-input label="名称" field="name" :rules="{ required: '必填' }" />
      <u-input label="主机" field="host" :rules="{ required: '必填' }" />
      <u-number-input label="端口" field="port" />
      <u-select
        label="OS"
        field="os_type"
        :options="[
          { label: 'linux', value: 'linux' },
          { label: 'windows', value: 'windows' },
        ]"
      />
      <u-input label="用户名" field="username" />
      <u-select
        label="认证方式"
        field="auth_type"
        :options="[
          { label: 'password', value: 'password' },
          { label: 'key', value: 'key' },
          { label: 'ssh_agent', value: 'ssh_agent' },
          { label: 'agent', value: 'agent' },
        ]"
      />
      <u-select label="凭证" field="credential_id" :options="credOptions" clearable />
      <u-input v-if="form.auth_type === 'agent'" label="Agent URL" field="agent_url" />
      <u-select
        v-if="form.auth_type === 'agent'"
        label="Agent 凭证"
        field="agent_credential_id"
        :options="credOptions"
        clearable
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

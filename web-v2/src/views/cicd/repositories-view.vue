<script setup lang="ts">
import { onMounted, reactive, ref, useTemplateRef } from "vue";
import { defineTableColumns, message } from "@veltra/desktop";

import {
  createRepository,
  deleteRepository,
  getWebhookSecret,
  listCredentials,
  rotateWebhookSecret,
  testRepository,
  updateRepository,
} from "@/api/cicd";
import type { Credential, Repository } from "@/api/types";
import FormDialog from "@/components/form-dialog.vue";
import ProTable from "@/components/pro-table.vue";
import { usePermission } from "@/composables/use-permission";

const { hasPermission } = usePermission();
const listRef = useTemplateRef("list");
const query = reactive({ keyword: "" });
const dialogOpen = ref(false);
const secretOpen = ref(false);
const editing = ref<Repository | null>(null);
const credOptions = ref<{ label: string; value: number }[]>([]);
const webhookInfo = reactive({ secret: "", url: "" });
const form = reactive({
  name: "",
  repo_url: "",
  default_branch: "main",
  description: "",
  tags: "",
  auth_type: "none",
  credential_id: undefined as number | undefined,
});

const columns = defineTableColumns([
  { key: "id", name: "ID", width: 80, minWidth: 60 },
  { key: "name", name: "名称", minWidth: 140 },
  { key: "repo_url", name: "URL", minWidth: 220 },
  { key: "default_branch", name: "分支", width: 100, minWidth: 80 },
  { key: "auth_type", name: "认证", width: 100, minWidth: 80 },
  { key: "action", name: "操作", width: 260, minWidth: 200 },
]);

onMounted(async () => {
  if (hasPermission("cicd.credentials:view") || hasPermission("cicd.credentials:use")) {
    try {
      const res = await listCredentials({ page: 1, page_size: 100 });
      credOptions.value = (res.items ?? []).map((c: Credential) => ({
        label: `${c.name} (${c.type})`,
        value: c.id,
      }));
    } catch {
      /* ignore */
    }
  }
});

function openCreate() {
  editing.value = null;
  Object.assign(form, {
    name: "",
    repo_url: "",
    default_branch: "main",
    description: "",
    tags: "",
    auth_type: "none",
    credential_id: undefined,
  });
  dialogOpen.value = true;
}

function openEdit(row: Repository) {
  editing.value = row;
  Object.assign(form, {
    name: row.name,
    repo_url: row.repo_url,
    default_branch: row.default_branch || "main",
    description: row.description || "",
    tags: row.tags || "",
    auth_type: row.auth_type || "none",
    credential_id: row.credential_id ?? undefined,
  });
  dialogOpen.value = true;
}

async function save() {
  try {
    const body: Record<string, unknown> = {
      name: form.name,
      repo_url: form.repo_url,
      default_branch: form.default_branch,
      description: form.description,
      tags: form.tags,
      auth_type: form.auth_type,
    };
    if (form.auth_type === "credential" && form.credential_id) {
      body.credential_id = form.credential_id;
    }
    if (editing.value) {
      if (form.auth_type === "none") body.clear_credential = true;
      await updateRepository(editing.value.id, body);
      message.success("已更新");
    } else {
      await createRepository(body);
      message.success("已创建");
    }
    dialogOpen.value = false;
    await listRef.value?.reload();
  } catch (err) {
    message.error(err instanceof Error ? err.message : "保存失败");
  }
}

async function remove(row: Repository) {
  try {
    await deleteRepository(row.id);
    message.success("已删除");
    await listRef.value?.reload();
  } catch (err) {
    message.error(err instanceof Error ? err.message : "删除失败");
  }
}

async function onTest(row: Repository) {
  try {
    const res = await testRepository(row.id);
    message.success(`拉取成功，分支 ${res.branches?.length ?? 0} 个`);
  } catch (err) {
    message.error(err instanceof Error ? err.message : "测试失败");
  }
}

async function showSecret(row: Repository) {
  try {
    const res = await getWebhookSecret(row.id);
    webhookInfo.secret = res.webhook_secret;
    webhookInfo.url = res.webhook_url;
    editing.value = row;
    secretOpen.value = true;
  } catch (err) {
    message.error(err instanceof Error ? err.message : "获取失败");
  }
}

async function rotateSecret() {
  if (!editing.value) return;
  try {
    const res = await rotateWebhookSecret(editing.value.id);
    webhookInfo.secret = res.webhook_secret;
    webhookInfo.url = res.webhook_url;
    message.success("已轮换");
  } catch (err) {
    message.error(err instanceof Error ? err.message : "轮换失败");
  }
}
</script>

<template>
  <div class="page">
    <div class="page-head">
      <h2>代码仓库</h2>
      <u-button v-if="hasPermission('cicd.repositories:create')" type="primary" @click="openCreate">
        新建仓库
      </u-button>
    </div>

    <ProTable ref="list" url="/repositories" v-model:query="query" :columns="columns" pagination>
      <template #filters="{ search }">
        <u-input v-model="query.keyword" placeholder="名称/URL" style="width: 200px" />
        <u-button type="primary" @click="search">查询</u-button>
      </template>
      <template #column:action="{ rowData }">
        <u-action-group :max="5">
          <u-action
            v-if="hasPermission('cicd.repositories:update')"
            @run="openEdit(rowData as Repository)"
          >
            编辑
          </u-action>
          <u-action
            v-if="hasPermission('cicd.repositories:view')"
            @run="onTest(rowData as Repository)"
          >
            测试
          </u-action>
          <u-action
            v-if="hasPermission('cicd.repositories:view')"
            @run="showSecret(rowData as Repository)"
          >
            Webhook
          </u-action>
          <u-action
            v-if="hasPermission('cicd.repositories:delete')"
            @run="remove(rowData as Repository)"
          >
            删除
          </u-action>
        </u-action-group>
      </template>
    </ProTable>

    <FormDialog
      v-model="dialogOpen"
      :title="editing ? '编辑仓库' : '新建仓库'"
      :model="form"
      label-width="110px"
      style="width: 560px"
      @submit="save"
    >
      <u-input label="名称" field="name" :rules="{ required: '必填' }" />
      <u-input label="Git URL" field="repo_url" :rules="{ required: '必填' }" />
      <u-input label="默认分支" field="default_branch" />
      <u-select
        label="认证"
        field="auth_type"
        :options="[
          { label: '无', value: 'none' },
          { label: '凭证', value: 'credential' },
        ]"
      />
      <u-select
        v-if="form.auth_type === 'credential'"
        label="凭证"
        field="credential_id"
        :options="credOptions"
      />
      <u-input label="标签" field="tags" />
      <u-input label="描述" field="description" />
    </FormDialog>

    <u-dialog v-model="secretOpen" title="Webhook Secret" style="width: 560px">
      <p class="mono">URL: {{ webhookInfo.url }}</p>
      <p class="mono">Secret: {{ webhookInfo.secret }}</p>
      <template #footer="{ close }">
        <u-button text @click="close()">关闭</u-button>
        <u-button
          v-if="hasPermission('cicd.repositories:update')"
          type="primary"
          @click="rotateSecret"
        >
          轮换
        </u-button>
      </template>
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
  align-items: center;
  justify-content: space-between;
}
.page-head h2 {
  margin: 0;
  font-size: 18px;
}
.mono {
  font-family: ui-monospace, monospace;
  word-break: break-all;
}
</style>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, useTemplateRef } from "vue";
import { defineTableColumns, message } from "@veltra/desktop";

import { createUser, deleteUser, listRoles, listUsers, updateUser } from "@/api/system";
import type { Role, User } from "@/api/types";
import FormDialog from "@/components/form-dialog.vue";
import ResourceList from "@/components/resource-list.vue";
import { usePermission } from "@/composables/use-permission";
import { useAuthStore } from "@/stores/auth";

const { hasPermission } = usePermission();
const auth = useAuthStore();
const listRef = useTemplateRef("list");
const filters = reactive({ keyword: "" });
const dialogOpen = ref(false);
const editing = ref<User | null>(null);
const roles = ref<Role[]>([]);
const form = reactive({
  username: "",
  password: "",
  display_name: "",
  email: "",
  is_active: true,
  role_ids: [] as number[],
});

const roleOptions = computed(() =>
  roles.value.map((r) => ({ label: `${r.name} (${r.code})`, value: r.id })),
);

const roleNameById = computed(() => {
  const map = new Map<number, string>();
  for (const r of roles.value) map.set(r.id, r.name);
  return map;
});

const columns = defineTableColumns([
  { key: "id", name: "ID", width: 80, minWidth: 60 },
  { key: "username", name: "用户名", minWidth: 120 },
  { key: "display_name", name: "显示名", minWidth: 120 },
  { key: "role_ids", name: "角色", minWidth: 160 },
  { key: "email", name: "邮箱", minWidth: 160 },
  { key: "is_active", name: "状态", width: 100, minWidth: 80 },
  { key: "is_super_admin", name: "超管", width: 80, minWidth: 60 },
  { key: "action", name: "操作", width: 160, minWidth: 120 },
]);

async function fetcher(params: { page: number; page_size: number }) {
  return listUsers(params);
}

onMounted(async () => {
  try {
    const res = await listRoles({ page: 1, page_size: 200 });
    roles.value = res.items ?? [];
  } catch {
    /* ignore */
  }
});

function roleLabels(ids?: number[]) {
  if (!ids?.length) return "—";
  return ids.map((id) => roleNameById.value.get(id) ?? `#${id}`).join(", ");
}

function openCreate() {
  editing.value = null;
  Object.assign(form, {
    username: "",
    password: "",
    display_name: "",
    email: "",
    is_active: true,
    role_ids: [],
  });
  dialogOpen.value = true;
}

function openEdit(row: User) {
  editing.value = row;
  Object.assign(form, {
    username: row.username,
    password: "",
    display_name: row.display_name || "",
    email: row.email || "",
    is_active: row.is_active,
    role_ids: [...(row.role_ids ?? [])],
  });
  dialogOpen.value = true;
}

async function save() {
  try {
    if (editing.value) {
      await updateUser(editing.value.id, {
        display_name: form.display_name,
        email: form.email,
        is_active: form.is_active,
        role_ids: form.role_ids,
        ...(form.password ? { password: form.password } : {}),
      });
      message.success("已更新");
      if (editing.value.id === auth.user?.id) {
        await auth.refreshMe(true);
      }
    } else {
      await createUser({
        username: form.username,
        password: form.password,
        display_name: form.display_name,
        email: form.email,
        is_active: form.is_active,
        role_ids: form.role_ids,
      });
      message.success("已创建");
    }
    dialogOpen.value = false;
    await listRef.value?.refresh();
  } catch (err) {
    message.error(err instanceof Error ? err.message : "保存失败");
  }
}

async function remove(row: User) {
  try {
    await deleteUser(row.id);
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
      <h2>用户</h2>
      <u-button v-if="hasPermission('system.users:create')" type="primary" @click="openCreate">
        新建用户
      </u-button>
    </div>

    <ResourceList ref="list" :fetcher="fetcher" :columns="columns" :filters="filters">
      <template #filters="{ reload }">
        <u-input v-model="filters.keyword" placeholder="用户名（本地提示）" style="width: 200px" />
        <u-button @click="reload">刷新</u-button>
      </template>
      <template #column:role_ids="{ rowData }">
        {{ roleLabels((rowData as User).role_ids) }}
      </template>
      <template #column:is_active="{ rowData }">
        {{ rowData.is_active ? "启用" : "禁用" }}
      </template>
      <template #column:is_super_admin="{ rowData }">
        {{ rowData.is_super_admin ? "是" : "否" }}
      </template>
      <template #column:action="{ rowData }">
        <u-action-group :max="3">
          <u-action v-if="hasPermission('system.users:update')" @run="openEdit(rowData as User)">
            编辑
          </u-action>
          <u-action
            v-if="hasPermission('system.users:delete') && !(rowData as User).is_super_admin"
            need-confirm
            type="danger"
            @run="remove(rowData as User)"
          >
            删除
          </u-action>
        </u-action-group>
      </template>
    </ResourceList>

    <FormDialog
      v-model="dialogOpen"
      :title="editing ? '编辑用户' : '新建用户'"
      :model="form"
      label-width="88px"
      style="width: 520px"
      @submit="save"
    >
      <u-input v-if="!editing" label="用户名" field="username" :rules="{ required: '必填' }" />
      <u-password-input
        label="密码"
        field="password"
        :placeholder="editing ? '留空则不修改' : '密码'"
        :rules="editing ? undefined : { required: '必填' }"
      />
      <u-input label="显示名" field="display_name" />
      <u-input label="邮箱" field="email" />
      <u-multi-select
        label="角色"
        field="role_ids"
        :options="roleOptions"
        :disabled="!!editing?.is_super_admin"
        placeholder="选择角色"
        filterable
        clearable
      />
      <u-switch label="启用" field="is_active" />
    </FormDialog>
  </div>
</template>

<style scoped>
.page-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-shrink: 0;
}
.page-head h2 {
  margin: 0;
  font-size: 20px;
}
</style>

<script setup lang="ts">
import { computed, reactive, ref, useTemplateRef } from "vue";
import { defineTableColumns, message } from "@veltra/desktop";

import { createRole, deleteRole, listMenus, setRolePermissions, updateRole } from "@/api/system";
import type { RbacResource, Role } from "@/api/types";
import FormDialog from "@/components/form-dialog.vue";
import ProTable from "@/components/pro-table.vue";
import { usePermission } from "@/composables/use-permission";
import { useAuthStore } from "@/stores/auth";

const ACTIONS = ["view", "create", "update", "delete", "execute", "use", "test"] as const;
const PROJECT_SCOPE_ACTIONS = ["view_all", "manage_all"] as const;

const { hasPermission } = usePermission();
const auth = useAuthStore();
const listRef = useTemplateRef("list");
const dialogOpen = ref(false);
const permOpen = ref(false);
const editing = ref<Role | null>(null);
const form = reactive({ name: "", code: "", description: "" });
const menuTree = ref<RbacResource[]>([]);
const checked = ref<Set<string>>(new Set());

const columns = defineTableColumns([
  { key: "id", name: "ID", width: 80, minWidth: 60 },
  { key: "name", name: "名称", minWidth: 120 },
  { key: "code", name: "编码", minWidth: 120 },
  { key: "description", name: "描述", minWidth: 160 },
  { key: "action", name: "操作", width: 200, minWidth: 160 },
]);

const flatMenus = computed(() => flattenMenus(menuTree.value));

function flattenMenus(nodes: RbacResource[], out: RbacResource[] = []): RbacResource[] {
  for (const n of nodes) {
    out.push(n);
    if (n.children?.length) flattenMenus(n.children, out);
  }
  return out;
}

function openCreate() {
  editing.value = null;
  Object.assign(form, { name: "", code: "", description: "" });
  dialogOpen.value = true;
}

function openEdit(row: Role) {
  editing.value = row;
  Object.assign(form, {
    name: row.name,
    code: row.code,
    description: row.description || "",
  });
  dialogOpen.value = true;
}

async function openPerms(row: Role) {
  editing.value = row;
  if (!menuTree.value.length) {
    const res = await listMenus();
    menuTree.value = res.items ?? [];
  }
  checked.value = new Set((row.permissions ?? []).map((p) => p.permission));
  permOpen.value = true;
}

function toggle(code: string, on: boolean) {
  const next = new Set(checked.value);
  if (on) next.add(code);
  else next.delete(code);
  checked.value = next;
}

function isChecked(code: string) {
  return checked.value.has(code);
}

function actionsForResource(resource: RbacResource) {
  return resource.path === "project.projects" ? [...ACTIONS, ...PROJECT_SCOPE_ACTIONS] : ACTIONS;
}

async function save() {
  try {
    if (editing.value) {
      await updateRole(editing.value.id, { name: form.name, description: form.description });
      message.success("已更新");
    } else {
      await createRole({ ...form, permissions: [] });
      message.success("已创建");
    }
    dialogOpen.value = false;
    await listRef.value?.reload();
  } catch (err) {
    message.error(err instanceof Error ? err.message : "保存失败");
  }
}

async function savePerms() {
  if (!editing.value) return;
  try {
    await setRolePermissions(editing.value.id, [...checked.value]);
    message.success("权限已保存");
    permOpen.value = false;
    await listRef.value?.reload();
    // Role permission changes affect current session menus/buttons.
    await auth.refreshMe(true);
  } catch (err) {
    message.error(err instanceof Error ? err.message : "保存失败");
  }
}

async function remove(row: Role) {
  try {
    await deleteRole(row.id);
    message.success("已删除");
    await listRef.value?.reload();
  } catch (err) {
    message.error(err instanceof Error ? err.message : "删除失败");
  }
}
</script>

<template>
  <div class="page">
    <div class="page-head">
      <h2>角色</h2>
      <u-button v-if="hasPermission('system.roles:create')" type="primary" @click="openCreate">
        新建角色
      </u-button>
    </div>

    <ProTable ref="list" url="/roles" :columns="columns" pagination>
      <template #column:action="{ rowData }">
        <u-action-group :max="4">
          <u-action v-if="hasPermission('system.roles:update')" @run="openEdit(rowData as Role)">
            编辑
          </u-action>
          <u-action v-if="hasPermission('system.roles:update')" @run="openPerms(rowData as Role)">
            权限
          </u-action>
          <u-action
            v-if="hasPermission('system.roles:delete')"
            need-confirm
            type="danger"
            @run="remove(rowData as Role)"
          >
            删除
          </u-action>
        </u-action-group>
      </template>
    </ProTable>

    <FormDialog
      v-model="dialogOpen"
      :title="editing ? '编辑角色' : '新建角色'"
      :model="form"
      label-width="72px"
      style="width: 480px"
      @submit="save"
    >
      <u-input label="名称" field="name" :rules="{ required: '必填' }" />
      <u-input label="编码" field="code" :disabled="!!editing" :rules="{ required: '必填' }" />
      <u-input label="描述" field="description" />
    </FormDialog>

    <u-dialog v-model="permOpen" title="角色权限" style="width: 720px">
      <div class="perm-grid">
        <div v-for="menu in flatMenus" :key="menu.id" class="perm-row">
          <div class="perm-path">
            {{ menu.menu_metadata?.title || menu.path }}
            <code>{{ menu.path }}</code>
          </div>
          <div class="perm-actions">
            <label v-for="action in actionsForResource(menu)" :key="action" class="perm-check">
              <input
                type="checkbox"
                :checked="isChecked(`${menu.path}:${action}`)"
                @change="
                  toggle(`${menu.path}:${action}`, ($event.target as HTMLInputElement).checked)
                "
              />
              {{ action }}
            </label>
          </div>
        </div>
      </div>
      <template #footer="{ close }">
        <u-button text @click="close()">取消</u-button>
        <u-button type="primary" @click="savePerms">保存权限</u-button>
      </template>
    </u-dialog>
  </div>
</template>

<style scoped>
.page-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}
.page-head h2 {
  margin: 0;
  font-size: 20px;
}
.perm-grid {
  max-height: 420px;
  overflow: auto;
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.perm-row {
  border-bottom: 1px solid #eee;
  padding-bottom: 8px;
}
.perm-path {
  font-weight: 500;
  margin-bottom: 4px;
}
.perm-path code {
  margin-left: 8px;
  color: #6b7280;
  font-size: 12px;
}
.perm-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}
.perm-check {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
}
</style>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue";
import { defineTableColumns, message } from "@veltra/desktop";

import { createResource, deleteResource, listResources, updateResource } from "@/api/system";
import type { RbacResource } from "@/api/types";
import FormDialog from "@/components/form-dialog.vue";
import { usePermission } from "@/composables/use-permission";

const TYPE_OPTIONS = [
  { label: "menu", value: "menu" },
  { label: "page", value: "page" },
  { label: "action", value: "action" },
  { label: "card", value: "card" },
];

const { hasPermission } = usePermission();
const tree = ref<RbacResource[]>([]);
const loading = ref(false);
const dialogOpen = ref(false);
const editing = ref<RbacResource | null>(null);
const form = reactive({
  path: "",
  type: "menu" as RbacResource["type"],
  parent_id: undefined as number | undefined,
  enabled: true,
  sort_key: 0,
  title: "",
  route: "",
});

const columns = defineTableColumns([
  { key: "path", name: "Path", minWidth: 220 },
  { key: "type", name: "类型", width: 100, minWidth: 80 },
  { key: "title", name: "菜单标题", minWidth: 140 },
  { key: "route", name: "路由", minWidth: 160 },
  { key: "sort_key", name: "排序", width: 80, minWidth: 60 },
  { key: "enabled", name: "状态", width: 90, minWidth: 70 },
  { key: "action", name: "操作", width: 220, minWidth: 180 },
]);

const parentOptions = computed(() =>
  flatten(tree.value)
    .filter((n) => !editing.value || n.id !== editing.value.id)
    .map((n) => ({
      label: `${n.path} (${n.type})`,
      value: n.id,
    })),
);

function flatten(nodes: RbacResource[], out: RbacResource[] = []): RbacResource[] {
  for (const n of nodes) {
    out.push(n);
    if (n.children?.length) flatten(n.children, out);
  }
  return out;
}

async function load() {
  loading.value = true;
  try {
    const res = await listResources();
    tree.value = res.items ?? [];
  } catch (err) {
    message.error(err instanceof Error ? err.message : "加载失败");
  } finally {
    loading.value = false;
  }
}

onMounted(() => {
  void load();
});

function openCreate(parent?: RbacResource) {
  editing.value = null;
  Object.assign(form, {
    path: "",
    type: "menu",
    parent_id: parent?.id,
    enabled: true,
    sort_key: 0,
    title: "",
    route: "",
  });
  dialogOpen.value = true;
}

function openEdit(row: RbacResource) {
  editing.value = row;
  Object.assign(form, {
    path: row.path,
    type: row.type,
    parent_id: row.parent_id ?? undefined,
    enabled: row.enabled,
    sort_key: row.sort_key,
    title: row.menu_metadata?.title || "",
    route: row.menu_metadata?.route || "",
  });
  dialogOpen.value = true;
}

async function save() {
  try {
    if (editing.value) {
      await updateResource(editing.value.id, {
        enabled: form.enabled,
        sort_key: form.sort_key,
        title: form.title,
        route: form.route,
      });
      message.success("已更新");
    } else {
      await createResource({
        path: form.path,
        type: form.type,
        parent_id: form.parent_id ?? null,
        enabled: form.enabled,
        sort_key: form.sort_key,
        title: form.title,
        route: form.route,
      });
      message.success("已创建");
    }
    dialogOpen.value = false;
    await load();
  } catch (err) {
    message.error(err instanceof Error ? err.message : "保存失败");
  }
}

async function remove(row: RbacResource) {
  try {
    await deleteResource(row.id);
    message.success("已删除");
    await load();
  } catch (err) {
    message.error(err instanceof Error ? err.message : "删除失败");
  }
}
</script>

<template>
  <div class="page">
    <div class="page-head">
      <h2>权限资源</h2>
      <div class="actions">
        <u-button @click="load">刷新</u-button>
        <u-button
          v-if="hasPermission('system.resources:create')"
          type="primary"
          @click="openCreate()"
        >
          新建资源
        </u-button>
      </div>
    </div>

    <div class="table-wrap" :class="{ 'is-loading': loading }">
      <u-table tree default-expand-all border stripe row-key="id" :columns="columns" :data="tree">
        <template #column:title="{ rowData }">
          {{ (rowData as RbacResource).menu_metadata?.title || "—" }}
        </template>
        <template #column:route="{ rowData }">
          {{ (rowData as RbacResource).menu_metadata?.route || "—" }}
        </template>
        <template #column:enabled="{ rowData }">
          {{ (rowData as RbacResource).enabled ? "启用" : "禁用" }}
        </template>
        <template #column:action="{ rowData }">
          <u-action-group :max="3">
            <u-action
              v-if="hasPermission('system.resources:create')"
              @run="openCreate(rowData as RbacResource)"
            >
              子节点
            </u-action>
            <u-action
              v-if="hasPermission('system.resources:update')"
              @run="openEdit(rowData as RbacResource)"
            >
              编辑
            </u-action>
            <u-action
              v-if="hasPermission('system.resources:delete')"
              need-confirm
              type="danger"
              @run="remove(rowData as RbacResource)"
            >
              删除
            </u-action>
          </u-action-group>
        </template>
        <template #empty>
          <u-empty :text="loading ? '加载中…' : '暂无资源'" />
        </template>
      </u-table>
    </div>

    <FormDialog
      v-model="dialogOpen"
      :title="editing ? '编辑资源' : '新建资源'"
      :model="form"
      label-width="88px"
      style="width: 520px"
      @submit="save"
    >
      <u-input
        label="Path"
        field="path"
        :disabled="!!editing"
        :rules="{ required: '必填' }"
        placeholder="如 system.users"
      />
      <u-select label="类型" field="type" :options="TYPE_OPTIONS" :disabled="!!editing" />
      <u-select
        v-if="!editing"
        label="父级"
        field="parent_id"
        :options="parentOptions"
        clearable
        placeholder="根节点"
      />
      <u-input
        v-if="form.type === 'menu' || editing?.type === 'menu'"
        label="菜单标题"
        field="title"
      />
      <u-input
        v-if="form.type === 'menu' || editing?.type === 'menu'"
        label="路由"
        field="route"
        placeholder="/system/users"
      />
      <u-number-input label="排序" field="sort_key" />
      <u-switch label="启用" field="enabled" />
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
.actions {
  display: flex;
  gap: 8px;
}
.table-wrap {
  flex: 1;
  min-height: 280px;
  position: relative;
}
.table-wrap :deep(.u-table) {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100% !important;
}
.table-wrap.is-loading::after {
  content: "";
  position: absolute;
  inset: 0;
  background: rgba(255, 255, 255, 0.35);
  pointer-events: none;
  z-index: 1;
}
</style>

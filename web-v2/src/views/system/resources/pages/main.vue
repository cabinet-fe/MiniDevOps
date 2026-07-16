<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue";
import { o } from "@cat-kit/core";
import { defineTableColumns, message } from "@veltra/desktop";

import {
  createResource,
  deleteResource,
  listResources,
  updateResource,
  updateResourceIcon,
} from "@/api/system";
import type { MenuMetadata, RbacResource } from "@/api/types";
import FormDialog from "@/components/form-dialog";
import { usePermission } from "@/composables/use-permission";
import { tagType, type TagType } from "@/lib/tag";

const TYPE_OPTIONS = [
  { label: "menu", value: "menu" },
  { label: "page", value: "page" },
  { label: "action", value: "action" },
  { label: "card", value: "card" },
];

const RESOURCE_TYPE_TAG: Record<string, TagType> = {
  menu: "primary",
  page: "info",
  action: undefined,
  card: "warning",
};

function resourceTypeTag(type: string) {
  return tagType(type, RESOURCE_TYPE_TAG);
}

const { hasPermission } = usePermission();
const tree = ref<RbacResource[]>([]);
const loading = ref(false);
const uploading = ref(false);
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

const isMenuType = computed(() => form.type === "menu" || editing.value?.type === "menu");

/** Level-1 menu icon: only when editing an existing top-level menu resource. */
const showIconUpload = computed(
  () => !!editing.value && editing.value.type === "menu" && !editing.value.parent_id,
);

const editingIconSrc = computed(() => iconSrc(editing.value?.menu_metadata));

function flatten(nodes: RbacResource[], out: RbacResource[] = []): RbacResource[] {
  for (const n of nodes) {
    out.push(n);
    if (n.children?.length) flatten(n.children, out);
  }
  return out;
}

function iconSrc(meta?: MenuMetadata): string | undefined {
  if (!meta?.icon_base64) return undefined;
  if (meta.icon_base64.startsWith("data:")) return meta.icon_base64;
  return `data:${meta.icon_mime || "image/png"};base64,${meta.icon_base64}`;
}

async function load() {
  loading.value = true;
  try {
    const res = await listResources();
    tree.value = res.items ?? [];
    if (editing.value) {
      const refreshed = findNode(tree.value, editing.value.id);
      if (refreshed) editing.value = refreshed;
    }
  } catch (err) {
    message.error(err instanceof Error ? err.message : "加载失败");
  } finally {
    loading.value = false;
  }
}

function findNode(nodes: RbacResource[], id: number): RbacResource | null {
  for (const node of nodes) {
    if (node.id === id) return node;
    if (node.children?.length) {
      const found = findNode(node.children, id);
      if (found) return found;
    }
  }
  return null;
}

onMounted(() => {
  void load();
});

function openCreate(parent?: RbacResource) {
  editing.value = null;
  form.parent_id = parent?.id;
  dialogOpen.value = true;
}

function openEdit(row: RbacResource) {
  editing.value = row;
  o(form).extend(row);
  form.title = row.menu_metadata?.title || "";
  form.route = row.menu_metadata?.route || "";
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

async function onIconPick(files: File[]) {
  if (!editing.value || !showIconUpload.value) return;
  if (!hasPermission("system.resources:update")) return;
  const file = files[0];
  if (!file) return;
  if (file.size > 32 * 1024) {
    message.error("图标原始体积不得超过 32KB");
    return;
  }
  uploading.value = true;
  try {
    const buf = await file.arrayBuffer();
    const bytes = new Uint8Array(buf);
    let binary = "";
    for (const b of bytes) binary += String.fromCharCode(b);
    const b64 = btoa(binary);
    const updated = await updateResourceIcon(editing.value.id, b64, file.type || "image/png");
    editing.value = updated;
    message.success("图标已更新");
    await load();
  } catch (err) {
    message.error(err instanceof Error ? err.message : "上传失败");
  } finally {
    uploading.value = false;
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
      <u-table tree default-expand-all :stripe="false" row-key="id" :columns="columns" :data="tree">
        <template #column:title="{ rowData }">
          {{ (rowData as RbacResource).menu_metadata?.title || "—" }}
        </template>
        <template #column:route="{ rowData }">
          {{ (rowData as RbacResource).menu_metadata?.route || "—" }}
        </template>
        <template #column:type="{ rowData }">
          <u-tag size="small" :type="resourceTypeTag((rowData as RbacResource).type)">
            {{ (rowData as RbacResource).type }}
          </u-tag>
        </template>
        <template #column:enabled="{ rowData }">
          <u-tag size="small" :type="(rowData as RbacResource).enabled ? 'success' : 'warning'">
            {{ (rowData as RbacResource).enabled ? "启用" : "禁用" }}
          </u-tag>
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
      <u-input v-if="isMenuType" label="菜单标题" field="title" />
      <u-input v-if="isMenuType" label="路由" field="route" placeholder="/system/users" />
      <u-form-item v-if="showIconUpload" label="一级图标">
        <div class="icon-field">
          <img v-if="editingIconSrc" class="icon-preview" :src="editingIconSrc" alt="menu icon" />
          <div class="icon-field__body">
            <u-file-picker
              accept="image/*"
              :disabled="!hasPermission('system.resources:update') || uploading"
              @pick="onIconPick"
            >
              <u-button :loading="uploading" :disabled="!hasPermission('system.resources:update')">
                {{ editingIconSrc ? "更换图标" : "上传图标" }}
              </u-button>
            </u-file-picker>
            <p class="icon-hint">仅一级菜单支持；原始体积 ≤ 32KB</p>
          </div>
        </div>
      </u-form-item>
      <u-number-input label="排序" field="sort_key" />
      <u-switch label="启用" field="enabled" />
    </FormDialog>
  </div>
</template>

<style scoped lang="scss">
@use "pkg:@veltra/styles/functions" as fn;

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

.icon-field {
  display: flex;
  align-items: flex-start;
  gap: fn.use-var(gap, default);
}

.icon-preview {
  width: 40px;
  height: 40px;
  object-fit: contain;
  border-radius: fn.use-var(radius, small);
  border: fn.use-var(border);
  background: fn.use-var(bg-color, middle);
  padding: 4px;
  flex-shrink: 0;
}

.icon-field__body {
  display: flex;
  flex-direction: column;
  gap: fn.use-var(gap, small);
}

.icon-hint {
  margin: 0;
  font-size: fn.use-var(font-size-assist, small);
  color: fn.use-var(text-color, assist);
}
</style>

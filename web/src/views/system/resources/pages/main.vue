<script setup lang="ts">
defineOptions({ name: "SystemResources" });

import { computed, reactive, ref, useTemplateRef } from "vue";
import { o } from "@cat-kit/core";
import { message } from "@veltra/desktop";

import {
  createResource,
  deleteResource,
  listResources,
  updateResource,
  updateResourceIcon,
} from "@/api/system";
import type { MenuMetadata, RbacResource } from "@/api/types";
import FormDialog from "@/components/form-dialog";
import ProTable, { defineProTableColumns } from "@/components/pro-table";
import { usePermission } from "@/composables/use-permission";
import { tagType, type TagType } from "@/lib/tag";

const TYPE_OPTIONS = [
  { label: "menu", value: "menu" },
  { label: "page", value: "page" },
  { label: "action", value: "action" },
  { label: "card", value: "card" },
];

const FILTER_TYPE_OPTIONS = [{ label: "全部类型", value: "" }, ...TYPE_OPTIONS];

const ENABLED_OPTIONS = [
  { label: "全部状态", value: "" },
  { label: "启用", value: "true" },
  { label: "禁用", value: "false" },
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
const listRef = useTemplateRef("list");
const query = reactive({ keyword: "", type: "", enabled: "" });
const parentTree = ref<RbacResource[]>([]);
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

const columns = defineProTableColumns([
  { key: "title", name: "菜单标题", minWidth: 140 },
  { key: "type", name: "类型" },
  { key: "path", name: "Code", minWidth: 220 },
  { key: "route", name: "路由", minWidth: 160 },
  { key: "sort_key", name: "排序" },
  { key: "enabled", name: "状态" },
  { key: "action", name: "操作", width: 220, fixed: "right" },
]);

const parentOptions = computed(() =>
  flatten(parentTree.value)
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

function onLoaded(items: Record<string, any>[]) {
  if (!editing.value) return;
  const refreshed = findNode(items as RbacResource[], editing.value.id);
  if (refreshed) editing.value = refreshed;
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

async function loadParentOptions() {
  try {
    const res = await listResources();
    parentTree.value = res.items ?? [];
  } catch {
    /* ignore — parent select stays empty */
  }
}

async function openCreate(parent?: RbacResource) {
  editing.value = null;
  form.parent_id = parent?.id;
  await loadParentOptions();
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
    await listRef.value?.reload();
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
    await listRef.value?.reload();
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
    await listRef.value?.reload();
  } catch (err) {
    message.error(err instanceof Error ? err.message : "删除失败");
  }
}
</script>

<template>
  <div>
    <ProTable
      ref="list"
      url="/rbac/resources"
      v-model:query="query"
      :columns="columns"
      tree
      default-expand-all
      :auto-query-fields="['type', 'enabled']"
      @loaded="onLoaded"
    >
      <template #filters="{ search }">
        <u-input v-model="query.keyword" placeholder="Code / 标题 / 路由" style="width: 220px" />
        <u-select
          v-model="query.type"
          placeholder="全部类型"
          :options="FILTER_TYPE_OPTIONS"
          style="width: 130px"
        />
        <u-select
          v-model="query.enabled"
          placeholder="全部状态"
          :options="ENABLED_OPTIONS"
          style="width: 130px"
        />
        <u-button type="primary" @click="search">查询</u-button>
        <u-button
          v-if="hasPermission('system.resources:create')"
          type="primary"
          style="margin-left: auto"
          @click.prevent="openCreate()"
        >
          新建资源
        </u-button>
      </template>
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
    </ProTable>

    <FormDialog
      v-model="dialogOpen"
      :title="editing ? '编辑资源' : '新建资源'"
      :model="form"
      label-width="88px"
      style="width: 520px"
      @submit="save"
    >
      <u-input
        label="Code"
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

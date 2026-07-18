<script setup lang="ts">
import { computed, onMounted, reactive, ref, useTemplateRef, watch } from "vue";
import { o } from "@cat-kit/core";
import { message } from "@veltra/desktop";

import {
  createResource,
  deleteResource,
  listMenuGroups,
  listResources,
  updateResource,
  updateResourceIcon,
} from "@/api/system";
import type { MenuGroup, RbacResource } from "@/api/types";
import FormDialog from "@/components/form-dialog";
import ProTable, { defineProTableColumns } from "@/components/pro-table";
import { usePermission } from "@/composables/use-permission";
import { useAuthStore } from "@/stores/auth";
import { tagType, type TagType } from "@/lib/tag";

const TYPE_OPTIONS = [
  { label: "menu", value: "menu" },
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
  action: undefined,
  card: "warning",
};

function resourceTypeTag(type: string) {
  return tagType(type, RESOURCE_TYPE_TAG);
}

const { hasPermission } = usePermission();
const auth = useAuthStore();
const listRef = useTemplateRef("list");
const query = reactive({ keyword: "", type: "", enabled: "", group_id: "" as string | number });
const groups = ref<MenuGroup[]>([]);
const menuParents = ref<RbacResource[]>([]);
const uploading = ref(false);
const dialogOpen = ref(false);
const editing = ref<RbacResource | null>(null);
const form = reactive({
  code: "",
  type: "menu" as RbacResource["type"],
  group_id: undefined as number | undefined,
  parent_id: undefined as number | undefined,
  enabled: true,
  sort_key: 0,
  title: "",
  route: "",
  hidden: false,
  super_admin_only: false,
});

const columns = defineProTableColumns([
  { key: "title", name: "标题", minWidth: 140 },
  { key: "type", name: "类型", width: 90 },
  { key: "code", name: "Code", minWidth: 140 },
  { key: "full_code", name: "Full Code", minWidth: 180 },
  { key: "route", name: "路由", minWidth: 140 },
  { key: "sort_key", name: "排序", width: 70 },
  { key: "flags", name: "标志", width: 140 },
  { key: "enabled", name: "状态", width: 80 },
  { key: "action", name: "操作", width: 220, fixed: "right" },
]);

const groupOptions = computed(() =>
  groups.value.map((g) => ({ label: `${g.name} (${g.code})`, value: g.id })),
);

const filterGroupOptions = computed(() => [
  { label: "全部分组", value: "" },
  ...groupOptions.value,
]);

const parentMenuOptions = computed(() =>
  menuParents.value
    .filter((n) => n.type === "menu" && (!editing.value || n.id !== editing.value.id))
    .map((n) => ({
      label: `${n.title || n.code} (${n.full_code})`,
      value: n.id,
    })),
);

const isMenuType = computed(() => form.type === "menu");
const isFeatureType = computed(() => form.type === "action" || form.type === "card");
const canEditSuperOnly = computed(() => !!auth.user?.is_super_admin);

const showIconUpload = computed(() => !!editing.value && editing.value.type === "menu");

const editingIconSrc = computed(() => iconSrc(editing.value));

const selectedGroup = computed(() => groups.value.find((g) => g.id === form.group_id));
/** Last autofilled route; keep updating until the user edits route manually. */
const lastAutoRoute = ref("");

function suggestedMenuRoute(code: string, prefix: string): string {
  const slug = code.includes("_") ? code.split("_").slice(1).join("-") || code : code;
  const base = prefix.replace(/\/$/, "");
  return base ? `${base}/${slug}` : `/${slug}`;
}

watch(
  () => [form.type, form.code, form.group_id] as const,
  () => {
    if (editing.value || form.type !== "menu") return;
    const code = form.code.trim();
    if (!code) {
      if (!form.route || form.route === lastAutoRoute.value) {
        form.route = "";
        lastAutoRoute.value = "";
      }
      return;
    }
    const next = suggestedMenuRoute(code, selectedGroup.value?.route_prefix ?? "");
    if (!form.route || form.route === lastAutoRoute.value) {
      form.route = next;
      lastAutoRoute.value = next;
    }
  },
);

function iconSrc(row?: RbacResource | null): string | undefined {
  if (!row?.icon_base64) return undefined;
  if (row.icon_base64.startsWith("data:")) return row.icon_base64;
  return `data:${row.icon_mime || "image/png"};base64,${row.icon_base64}`;
}

/** Menus are flat roots (features nest under them; parent picker only needs menus). */
function menuRoots(nodes: RbacResource[]): RbacResource[] {
  return nodes.filter((n) => n.type === "menu");
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

function onLoaded(items: Record<string, any>[]) {
  if (!editing.value) return;
  const refreshed = findNode(items as RbacResource[], editing.value.id);
  if (refreshed) editing.value = refreshed;
}

async function loadGroups() {
  try {
    const res = await listMenuGroups();
    groups.value = res.items ?? [];
  } catch {
    groups.value = [];
  }
}

async function loadMenuParents() {
  try {
    const res = await listResources({ type: "menu" });
    menuParents.value = menuRoots(res.items ?? []);
  } catch {
    menuParents.value = [];
  }
}

onMounted(() => {
  void loadGroups();
});

async function openCreate(parent?: RbacResource) {
  editing.value = null;
  lastAutoRoute.value = "";
  form.code = "";
  form.type = parent ? "action" : "menu";
  form.group_id = parent ? undefined : groups.value[0]?.id;
  form.parent_id = parent?.type === "menu" ? parent.id : undefined;
  form.enabled = true;
  form.sort_key = 0;
  form.title = "";
  form.route = "";
  form.hidden = false;
  form.super_admin_only = false;
  await loadGroups();
  await loadMenuParents();
  dialogOpen.value = true;
}

async function openEdit(row: RbacResource) {
  editing.value = row;
  lastAutoRoute.value = "";
  await loadGroups();
  await loadMenuParents();
  o(form).extend({
    code: row.code,
    type: row.type,
    group_id: row.group_id ?? undefined,
    parent_id: row.parent_id ?? undefined,
    enabled: row.enabled,
    sort_key: row.sort_key,
    title: row.title || "",
    route: row.route || "",
    hidden: row.hidden,
    super_admin_only: row.super_admin_only,
  });
  dialogOpen.value = true;
}

async function save() {
  try {
    if (editing.value) {
      const body: Record<string, unknown> = {
        enabled: form.enabled,
        sort_key: form.sort_key,
        title: form.title,
        route: form.route,
        hidden: form.hidden,
      };
      if (editing.value.type === "menu" && form.group_id) {
        body.group_id = form.group_id;
      }
      if (canEditSuperOnly.value) {
        body.super_admin_only = form.super_admin_only;
      }
      await updateResource(editing.value.id, body);
      message.success("已更新");
    } else {
      const body: Record<string, unknown> = {
        code: form.code,
        type: form.type,
        enabled: form.enabled,
        sort_key: form.sort_key,
        title: form.title,
      };
      if (form.type === "menu") {
        body.group_id = form.group_id;
        body.route = form.route;
        body.hidden = form.hidden;
      } else {
        body.parent_id = form.parent_id;
      }
      if (canEditSuperOnly.value) {
        body.super_admin_only = form.super_admin_only;
      }
      await createResource(body);
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
  if (!hasPermission("system_resources:update")) return;
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
      :query="query"
      :columns="columns"
      tree
      default-expand-all
      :auto-query-fields="['type', 'enabled', 'group_id']"
      @loaded="onLoaded"
    >
      <template #filters>
        <u-input v-model="query.keyword" placeholder="Code / 标题 / 路由" style="width: 200px" />
        <u-select
          v-model="query.group_id"
          placeholder="全部分组"
          :options="filterGroupOptions"
          style="width: 160px"
        />
        <u-select
          v-model="query.type"
          placeholder="全部类型"
          :options="FILTER_TYPE_OPTIONS"
          style="width: 120px"
        />
        <u-select
          v-model="query.enabled"
          placeholder="全部状态"
          :options="ENABLED_OPTIONS"
          style="width: 120px"
        />
        <u-button
          v-if="hasPermission('system_resources:create')"
          type="primary"
          style="margin-left: auto"
          @click.prevent="openCreate()"
        >
          新建资源
        </u-button>
      </template>
      <template #column:title="{ rowData }">
        {{ (rowData as RbacResource).title || "—" }}
      </template>
      <template #column:route="{ rowData }">
        {{ (rowData as RbacResource).route || "—" }}
      </template>
      <template #column:type="{ rowData }">
        <u-tag size="small" :type="resourceTypeTag((rowData as RbacResource).type)">
          {{ (rowData as RbacResource).type }}
        </u-tag>
      </template>
      <template #column:flags="{ rowData }">
        <span class="flag-cell">
          <u-tag v-if="(rowData as RbacResource).hidden" size="small" type="info">隐藏</u-tag>
          <u-tag v-if="(rowData as RbacResource).super_admin_only" size="small" type="warning">
            仅超管
          </u-tag>
          <template
            v-if="!(rowData as RbacResource).hidden && !(rowData as RbacResource).super_admin_only"
          >
            —
          </template>
        </span>
      </template>
      <template #column:enabled="{ rowData }">
        <u-tag size="small" :type="(rowData as RbacResource).enabled ? 'success' : 'warning'">
          {{ (rowData as RbacResource).enabled ? "启用" : "禁用" }}
        </u-tag>
      </template>
      <template #column:action="{ rowData }">
        <u-action-group :max="3">
          <u-action
            v-if="
              hasPermission('system_resources:create') && (rowData as RbacResource).type === 'menu'
            "
            @run="openCreate(rowData as RbacResource)"
          >
            子功能
          </u-action>
          <u-action
            v-if="hasPermission('system_resources:update')"
            @run="openEdit(rowData as RbacResource)"
          >
            编辑
          </u-action>
          <u-action
            v-if="hasPermission('system_resources:delete')"
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
      label-width="104px"
      style="width: 540px"
      @submit="save"
    >
      <u-input
        label="Code"
        field="code"
        :disabled="!!editing"
        :rules="{ required: '必填' }"
        placeholder="如 system_users（不含 .）"
      />
      <u-select label="类型" field="type" :options="TYPE_OPTIONS" :disabled="!!editing" />
      <u-select
        v-if="isMenuType"
        label="菜单分组"
        field="group_id"
        :options="groupOptions"
        :rules="{ required: '必填' }"
        placeholder="选择分组"
      />
      <u-select
        v-if="!editing && isFeatureType"
        label="所属菜单"
        field="parent_id"
        :options="parentMenuOptions"
        :rules="{ required: '必填' }"
        placeholder="选择菜单"
      />
      <u-input label="标题" field="title" />
      <u-input v-if="isMenuType" label="路由" field="route" placeholder="空则按分组前缀自动预填" />
      <u-form-item v-if="showIconUpload" label="菜单图标">
        <div class="icon-field">
          <img v-if="editingIconSrc" class="icon-preview" :src="editingIconSrc" alt="menu icon" />
          <div class="icon-field__body">
            <u-file-picker
              accept="image/*"
              :disabled="!hasPermission('system_resources:update') || uploading"
              @pick="onIconPick"
            >
              <u-button :loading="uploading" :disabled="!hasPermission('system_resources:update')">
                {{ editingIconSrc ? "更换图标" : "上传图标" }}
              </u-button>
            </u-file-picker>
            <p class="icon-hint">原始体积 ≤ 32KB</p>
          </div>
        </div>
      </u-form-item>
      <u-number-input label="排序" field="sort_key" />
      <u-switch v-if="isMenuType" label="隐藏" field="hidden" />
      <u-switch label="仅超管" field="super_admin_only" :disabled="!canEditSuperOnly" />
      <u-switch label="启用" field="enabled" />
    </FormDialog>
  </div>
</template>

<style scoped lang="scss">
@use "pkg:@veltra/styles/functions" as fn;

.flag-cell {
  display: inline-flex;
  flex-wrap: wrap;
  gap: 4px;
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

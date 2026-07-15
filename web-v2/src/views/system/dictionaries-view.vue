<script setup lang="ts">
import { reactive, ref, useTemplateRef } from "vue";
import { defineTableColumns, message } from "@veltra/desktop";

import {
  createDictionary,
  deleteDictionary,
  getDictionary,
  listDictionaries,
  updateDictionary,
} from "@/api/system";
import type { DictItem, Dictionary } from "@/api/types";
import FormDialog from "@/components/form-dialog.vue";
import ResourceList from "@/components/resource-list.vue";
import { usePermission } from "@/composables/use-permission";

const { hasPermission } = usePermission();
const listRef = useTemplateRef("list");
const dialogOpen = ref(false);
const editing = ref<Dictionary | null>(null);
const form = reactive({
  name: "",
  code: "",
  description: "",
  items: [] as DictItem[],
});

const columns = defineTableColumns([
  { key: "id", name: "ID", width: 80, minWidth: 60 },
  { key: "name", name: "名称", minWidth: 120 },
  { key: "code", name: "编码", minWidth: 120 },
  { key: "description", name: "描述", minWidth: 160 },
  { key: "action", name: "操作", width: 160, minWidth: 120 },
]);

async function fetcher(params: { page: number; page_size: number }) {
  return listDictionaries(params);
}

function openCreate() {
  editing.value = null;
  Object.assign(form, { name: "", code: "", description: "", items: [] });
  dialogOpen.value = true;
}

async function openEdit(row: Dictionary) {
  try {
    const full = await getDictionary(row.id);
    editing.value = full;
    Object.assign(form, {
      name: full.name,
      code: full.code,
      description: full.description || "",
      items: (full.items ?? []).map((it) => ({
        label: it.label,
        value: it.value,
        sort_order: it.sort_order ?? 0,
        enabled: it.enabled !== false,
      })),
    });
    dialogOpen.value = true;
  } catch (err) {
    message.error(err instanceof Error ? err.message : "加载失败");
  }
}

function addItem() {
  form.items.push({
    label: "",
    value: "",
    sort_order: form.items.length,
    enabled: true,
  });
}

function removeItem(idx: number) {
  form.items.splice(idx, 1);
}

async function save() {
  try {
    const items = form.items
      .map((it, i) => ({
        label: it.label.trim(),
        value: it.value.trim(),
        sort_order: it.sort_order ?? i,
        enabled: it.enabled !== false,
      }))
      .filter((it) => it.label && it.value);

    if (editing.value) {
      await updateDictionary(editing.value.id, {
        name: form.name,
        description: form.description,
        items,
      });
      message.success("已更新");
    } else {
      await createDictionary({
        name: form.name,
        code: form.code,
        description: form.description,
        items,
      });
      message.success("已创建");
    }
    dialogOpen.value = false;
    await listRef.value?.refresh();
  } catch (err) {
    message.error(err instanceof Error ? err.message : "保存失败");
  }
}

async function remove(row: Dictionary) {
  try {
    await deleteDictionary(row.id);
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
      <h2>字典</h2>
      <u-button
        v-if="hasPermission('system.dictionaries:create')"
        type="primary"
        @click="openCreate"
      >
        新建字典
      </u-button>
    </div>

    <ResourceList ref="list" :fetcher="fetcher" :columns="columns">
      <template #column:action="{ rowData }">
        <u-action-group :max="3">
          <u-action
            v-if="hasPermission('system.dictionaries:update')"
            @run="openEdit(rowData as Dictionary)"
          >
            编辑
          </u-action>
          <u-action
            v-if="hasPermission('system.dictionaries:delete')"
            need-confirm
            type="danger"
            @run="remove(rowData as Dictionary)"
          >
            删除
          </u-action>
        </u-action-group>
      </template>
    </ResourceList>

    <FormDialog
      v-model="dialogOpen"
      :title="editing ? '编辑字典' : '新建字典'"
      :model="form"
      label-width="72px"
      style="width: 640px"
      @submit="save"
    >
      <u-input label="名称" field="name" :rules="{ required: '必填' }" />
      <u-input label="编码" field="code" :disabled="!!editing" :rules="{ required: '必填' }" />
      <u-input label="描述" field="description" />

      <div class="items-head">
        <strong>字典项</strong>
        <u-button size="small" @click="addItem">添加项</u-button>
      </div>
      <div v-if="!form.items.length" class="items-empty">暂无字典项</div>
      <div v-for="(it, idx) in form.items" :key="idx" class="item-row">
        <u-input v-model="it.label" placeholder="标签" style="flex: 1" />
        <u-input v-model="it.value" placeholder="值" style="flex: 1" />
        <u-switch v-model="it.enabled" />
        <u-button size="small" @click="removeItem(idx)">删</u-button>
      </div>
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
.items-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin: 16px 0 8px;
}
.items-empty {
  color: #6b7280;
  font-size: 13px;
  margin-bottom: 8px;
}
.item-row {
  display: flex;
  gap: 8px;
  align-items: center;
  margin-bottom: 8px;
}
</style>

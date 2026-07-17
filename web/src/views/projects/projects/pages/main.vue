<script setup lang="ts">
defineOptions({ name: "Projects" });

import { reactive, ref, useTemplateRef } from "vue";
import { o } from "@cat-kit/core";
import { message } from "@veltra/desktop";
import { useRouter } from "vue-router";

import { archiveProject, createProject, deleteProject, updateProject } from "@/api/projects";
import type { ProductProject } from "@/api/types";
import FormDialog from "@/components/form-dialog";
import ProTable, { defineProTableColumns } from "@/components/pro-table";
import { usePermission } from "@/composables/use-permission";
import { formatDateTime } from "@/lib/datetime";

const router = useRouter();
const { hasPermission } = usePermission();
const tableRef = useTemplateRef("table");
const query = reactive({ keyword: "", status: "" });
const dialogOpen = ref(false);
const editing = ref<ProductProject | null>(null);
const form = reactive({
  name: "",
  slug: "",
  description: "",
  tags: "",
});

const columns = defineProTableColumns([
  { key: "name", name: "项目", sortable: true },
  { key: "slug", name: "Slug" },
  { key: "status", name: "状态", width: 100 },
  { key: "tags", name: "标签" },
  { key: "updated_at", name: "更新时间", sortable: true, render: ({ val }) => formatDateTime(val) },
  { key: "action", name: "操作", width: 220, align: "center", fixed: "right" },
]);

function openCreate() {
  editing.value = null;
  dialogOpen.value = true;
}

function openEdit(project: ProductProject) {
  editing.value = project;
  o(form).extend(project);
  dialogOpen.value = true;
}

async function save() {
  try {
    const input = {
      name: form.name,
      slug: form.slug,
      description: form.description,
      tags: form.tags,
    };
    if (editing.value) {
      await updateProject(editing.value.id, input);
      message.success("项目已更新");
    } else {
      await createProject(input);
      message.success("项目已创建");
    }
    dialogOpen.value = false;
    await tableRef.value?.reload();
  } catch (error) {
    message.error(error instanceof Error ? error.message : "保存失败");
  }
}

async function archive(project: ProductProject) {
  if (!window.confirm(`确认归档项目「${project.name}」？`)) return;
  try {
    await archiveProject(project.id);
    message.success("项目已归档");
    await tableRef.value?.reload();
  } catch (error) {
    message.error(error instanceof Error ? error.message : "归档失败");
  }
}

async function remove(project: ProductProject) {
  if (!window.confirm(`确认解散项目「${project.name}」？此操作不可撤销。`)) return;
  try {
    await deleteProject(project.id);
    message.success("项目已解散");
    await tableRef.value?.reload();
  } catch (error) {
    message.error(error instanceof Error ? error.message : "解散失败");
  }
}

function openProject(project: ProductProject) {
  void router.push({ name: "project-detail", params: { id: project.id } });
}

function splitTags(raw?: string | null): string[] {
  if (!raw?.trim()) return [];
  return raw
    .split(/[,，]/)
    .map((t) => t.trim())
    .filter(Boolean);
}
</script>

<template>
  <div>
    <ProTable
      ref="table"
      url="/projects"
      v-model:query="query"
      :columns="columns"
      pagination
      :auto-query-fields="['status']"
    >
      <template #filters="{ search }">
        <u-input v-model="query.keyword" placeholder="名称、Slug 或标签" style="width: 240px" />
        <u-select
          v-model="query.status"
          placeholder="全部状态"
          :options="[
            { label: '全部状态', value: '' },
            { label: '活跃', value: 'active' },
            { label: '已归档', value: 'archived' },
          ]"
          style="width: 130px"
        />
        <u-button type="primary" @click="search">查询</u-button>
        <u-button
          v-if="hasPermission('project.projects:create')"
          type="primary"
          style="margin-left: auto"
          @click.prevent="openCreate"
        >
          新建项目
        </u-button>
      </template>
      <template #column:name="{ rowData }">
        <u-action @run="openProject(rowData as ProductProject)">
          {{ (rowData as ProductProject).name }}
        </u-action>
      </template>
      <template #column:status="{ rowData }">
        <u-tag
          size="small"
          :type="(rowData as ProductProject).status === 'archived' ? 'warning' : 'success'"
        >
          {{ (rowData as ProductProject).status === "archived" ? "已归档" : "活跃" }}
        </u-tag>
      </template>
      <template #column:tags="{ rowData }">
        <span class="tag-cell">
          <u-tag v-for="tag in splitTags((rowData as ProductProject).tags)" :key="tag" size="small">
            {{ tag }}
          </u-tag>
        </span>
      </template>
      <template #column:action="{ rowData }">
        <u-action-group :max="4">
          <u-action @run="openProject(rowData as ProductProject)">进入</u-action>
          <u-action
            v-if="(rowData as ProductProject).permissions?.update"
            @run="openEdit(rowData as ProductProject)"
          >
            编辑
          </u-action>
          <u-action
            v-if="
              (rowData as ProductProject).permissions?.archive &&
              (rowData as ProductProject).status === 'active'
            "
            @run="archive(rowData as ProductProject)"
          >
            归档
          </u-action>
          <u-action
            v-if="(rowData as ProductProject).permissions?.delete"
            need-confirm
            type="danger"
            @run="remove(rowData as ProductProject)"
          >
            解散
          </u-action>
        </u-action-group>
      </template>
    </ProTable>

    <FormDialog
      v-model="dialogOpen"
      :title="editing ? '编辑项目' : '新建项目'"
      :model="form"
      label-width="100px"
      style="width: 560px"
      @submit="save"
    >
      <u-input label="名称" field="name" :rules="{ required: '必填' }" />
      <u-input label="Slug" field="slug" :rules="{ required: '必填' }" />
      <u-input label="标签" field="tags" placeholder="逗号分隔" />
      <u-textarea label="描述" field="description" :rows="4" />
    </FormDialog>
  </div>
</template>

<style scoped>
.tag-cell {
  display: inline-flex;
  flex-wrap: wrap;
  gap: 4px;
}
</style>

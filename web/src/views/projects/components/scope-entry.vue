<script setup lang="ts">
import { computed, reactive } from "vue";
import { useRoute, useRouter } from "vue-router";

import type { ProductProject } from "@/api/types";
import ProTable, { defineProTableColumns } from "@/components/pro-table";
import { formatDateTime } from "@/lib/datetime";

const route = useRoute();
const router = useRouter();
const query = reactive({ keyword: "", status: "active" });

const projectTab = computed(() => {
  const tab = route.meta.projectTab;
  return tab === "docs" ? "docs" : "requirements";
});

const actionLabel = computed(() => (projectTab.value === "docs" ? "进入文档" : "进入需求"));

const columns = defineProTableColumns([
  { key: "name", name: "项目", sortable: true },
  { key: "slug", name: "Slug" },
  { key: "status", name: "状态", width: 100 },
  { key: "updated_at", name: "更新时间", sortable: true, render: ({ val }) => formatDateTime(val) },
  { key: "action", name: "操作", width: 120, align: "center", fixed: "right" },
]);

function openProject(project: ProductProject) {
  void router.push({
    name: "project-detail",
    params: { id: project.id },
    query: { tab: projectTab.value },
  });
}
</script>

<template>
  <div>
    <ProTable
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
      <template #column:action="{ rowData }">
        <u-action @run="openProject(rowData as ProductProject)">{{ actionLabel }}</u-action>
      </template>
    </ProTable>
  </div>
</template>


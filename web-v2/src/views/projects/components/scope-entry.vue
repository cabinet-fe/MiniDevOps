<script setup lang="ts">
import { computed, reactive } from "vue";
import { useRoute, useRouter } from "vue-router";

import type { ProductProject } from "@/api/types";
import ProTable, { defineProTableColumns } from "@/components/pro-table";

const route = useRoute();
const router = useRouter();
const query = reactive({ keyword: "", status: "active" });

const projectTab = computed(() => {
  const tab = route.meta.projectTab;
  return tab === "docs" ? "docs" : "requirements";
});

const pageCopy = computed(() =>
  projectTab.value === "docs"
    ? {
        title: "接口文档",
        description: "选择一个产品项目，进入其接口文档。",
        action: "进入文档",
      }
    : {
        title: "需求管理",
        description: "选择一个产品项目，进入其需求列表。",
        action: "进入需求",
      },
);

const columns = defineProTableColumns([
  { key: "name", name: "项目", sortable: true },
  { key: "slug", name: "Slug" },
  { key: "status", name: "状态", width: 100 },
  { key: "updated_at", name: "更新时间", sortable: true },
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
  <div class="page">
    <div class="page-head">
      <div>
        <h2>{{ pageCopy.title }}</h2>
        <p>{{ pageCopy.description }}</p>
      </div>
    </div>

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
        <u-action @run="openProject(rowData as ProductProject)">{{ pageCopy.action }}</u-action>
      </template>
    </ProTable>
  </div>
</template>

<style scoped>
.page {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.page-head h2,
.page-head p {
  margin: 0;
}
.page-head p {
  margin-top: 6px;
  color: var(--u-text-color-assist, #7c8494);
  font-size: 13px;
}
</style>

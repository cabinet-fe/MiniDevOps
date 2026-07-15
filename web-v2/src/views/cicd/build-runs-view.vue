<script setup lang="ts">
import { reactive, useTemplateRef } from "vue";
import { useRouter } from "vue-router";
import { defineTableColumns } from "@veltra/desktop";

import { listBuildRuns } from "@/api/cicd";
import type { BuildRun } from "@/api/types";
import ResourceList from "@/components/resource-list.vue";

const router = useRouter();
const listRef = useTemplateRef("list");
const filters = reactive({ build_job_id: undefined as number | undefined, status: "" });

const columns = defineTableColumns([
  { key: "id", name: "ID", width: 80, minWidth: 60 },
  { key: "build_job_id", name: "任务", width: 90, minWidth: 70 },
  { key: "build_number", name: "#", width: 70, minWidth: 50 },
  { key: "status", name: "状态", width: 110, minWidth: 80 },
  { key: "stage", name: "阶段", width: 110, minWidth: 80 },
  { key: "distribution_summary", name: "分发", width: 120, minWidth: 90 },
  { key: "branch", name: "分支", minWidth: 100 },
  { key: "trigger_type", name: "触发", width: 100, minWidth: 80 },
  { key: "created_at", name: "创建时间", minWidth: 160 },
  { key: "action", name: "操作", width: 100, minWidth: 80 },
]);

async function fetcher(params: { page: number; page_size: number }) {
  return listBuildRuns({
    ...params,
    build_job_id: filters.build_job_id,
    status: filters.status || undefined,
  });
}

function openDetail(row: BuildRun) {
  void router.push({ name: "cicd-build-run-detail", params: { id: String(row.id) } });
}
</script>

<template>
  <div class="page">
    <div class="page-head">
      <h2>构建执行</h2>
    </div>

    <ResourceList ref="list" :fetcher="fetcher" :columns="columns" :filters="filters">
      <template #filters="{ reload }">
        <u-input
          v-model.number="filters.build_job_id"
          type="number"
          placeholder="任务 ID"
          style="width: 120px"
        />
        <u-select
          v-model="filters.status"
          clearable
          placeholder="状态"
          style="width: 140px"
          :options="[
            { label: 'queued', value: 'queued' },
            { label: 'running', value: 'running' },
            { label: 'success', value: 'success' },
            { label: 'failed', value: 'failed' },
            { label: 'cancelled', value: 'cancelled' },
            { label: 'interrupted', value: 'interrupted' },
          ]"
        />
        <u-button @click="reload">刷新</u-button>
      </template>
      <template #column:action="{ rowData }">
        <u-action @run="openDetail(rowData as BuildRun)">详情</u-action>
      </template>
    </ResourceList>
  </div>
</template>

<style scoped>
.page {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.page-head h2 {
  margin: 0;
  font-size: 18px;
}
</style>

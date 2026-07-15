<script setup lang="ts">
import { reactive } from "vue";
import { useRouter } from "vue-router";
import { defineTableColumns } from "@veltra/desktop";

import type { BuildRun } from "@/api/types";
import ProTable from "@/components/pro-table.vue";

const router = useRouter();
const query = reactive({ build_job_id: "", status: "" });

const columns = defineTableColumns([
  { key: "id", name: "ID", width: 80, minWidth: 60 },
  { key: "build_job_id", name: "任务", width: 90, minWidth: 70 },
  { key: "build_number", name: "#", width: 70, minWidth: 50 },
  { key: "status", name: "状态", width: 110, minWidth: 80 },
  { key: "stage", name: "阶段", width: 110, minWidth: 80 },
  { key: "distribution_summary", name: "分发", width: 120, minWidth: 90 },
  { key: "branch", name: "分支", minWidth: 100 },
  { key: "trigger_type", name: "触发", width: 100, minWidth: 80 },
  { key: "created_at", name: "创建时间", minWidth: 160, sortable: true },
  { key: "action", name: "操作", width: 100, minWidth: 80 },
]);

function openDetail(row: BuildRun) {
  void router.push({ name: "cicd-build-run-detail", params: { id: String(row.id) } });
}
</script>

<template>
  <div class="page">
    <div class="page-head">
      <h2>构建执行</h2>
    </div>

    <ProTable
      url="/build-runs"
      v-model:query="query"
      :columns="columns"
      :auto-query-fields="['status']"
      pagination
    >
      <template #filters="{ search }">
        <u-input
          v-model="query.build_job_id"
          type="number"
          placeholder="任务 ID"
          style="width: 120px"
        />
        <u-select
          v-model="query.status"
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
        <u-button type="primary" @click="search">查询</u-button>
      </template>
      <template #column:action="{ rowData }">
        <u-action @run="openDetail(rowData as BuildRun)">详情</u-action>
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
.page-head h2 {
  margin: 0;
  font-size: 18px;
}
</style>

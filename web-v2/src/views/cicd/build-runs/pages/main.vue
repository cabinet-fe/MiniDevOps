<script setup lang="ts">
import { reactive } from "vue";
import { useRouter } from "vue-router";

import type { BuildRun } from "@/api/types";
import ProTable, { defineProTableColumns } from "@/components/pro-table";
import { JOB_STATUS_TAG, TRIGGER_TYPE_TAG, tagType, type TagType } from "@/lib/tag";

const router = useRouter();
const query = reactive({ build_job_id: "", status: "" });

const STAGE_TAG: Record<string, TagType> = {
  pending: undefined,
  cloning: "primary",
  building: "primary",
  archiving: "primary",
  distributing: "warning",
  idle: "success",
};

const DISTRIBUTION_TAG: Record<string, TagType> = {
  none: undefined,
  running: "primary",
  all_success: "success",
  partial: "warning",
  all_failed: "danger",
  cancelled: "warning",
};

const columns = defineProTableColumns([
  { key: "id", name: "ID", width: 80 },
  { key: "build_job_id", name: "任务", width: 90 },
  { key: "build_number", name: "#", width: 70 },
  { key: "status", name: "状态", width: 100 },
  { key: "stage", name: "阶段", width: 100 },
  { key: "distribution_summary", name: "分发" },
  { key: "branch", name: "分支" },
  { key: "trigger_type", name: "触发", width: 100 },
  { key: "created_at", name: "创建时间", sortable: true },
  { key: "action", name: "操作", width: 100, align: "center", fixed: "right" },
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
      <template #column:status="{ rowData }">
        <u-tag size="small" :type="tagType((rowData as BuildRun).status, JOB_STATUS_TAG)">
          {{ (rowData as BuildRun).status }}
        </u-tag>
      </template>
      <template #column:stage="{ rowData }">
        <u-tag size="small" :type="tagType((rowData as BuildRun).stage, STAGE_TAG)">
          {{ (rowData as BuildRun).stage || "—" }}
        </u-tag>
      </template>
      <template #column:distribution_summary="{ rowData }">
        <u-tag
          size="small"
          :type="tagType((rowData as BuildRun).distribution_summary, DISTRIBUTION_TAG)"
        >
          {{ (rowData as BuildRun).distribution_summary || "—" }}
        </u-tag>
      </template>
      <template #column:trigger_type="{ rowData }">
        <u-tag size="small" :type="tagType((rowData as BuildRun).trigger_type, TRIGGER_TYPE_TAG)">
          {{ (rowData as BuildRun).trigger_type }}
        </u-tag>
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

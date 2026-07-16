<script setup lang="ts">
defineOptions({ name: "AiRuns" });

import { reactive, useTemplateRef } from "vue";
import { useRouter } from "vue-router";

import type { AgentRun } from "@/api/types";
import ProTable, { defineProTableColumns } from "@/components/pro-table";
import { formatDateTime } from "@/lib/datetime";
import { JOB_STATUS_TAG, TRIGGER_TYPE_TAG, tagType } from "@/lib/tag";

const router = useRouter();
const table = useTemplateRef("table");
const query = reactive({ agent_id: "", status: "" });

const columns = defineProTableColumns([
  { key: "id", name: "ID", width: 70 },
  { key: "agent_id", name: "Agent", width: 90 },
  { key: "trigger_type", name: "触发", width: 110 },
  { key: "status", name: "状态", width: 100 },
  { key: "created_at", name: "创建时间", render: ({ val }) => formatDateTime(val) },
  { key: "action", name: "操作", width: 100, align: "center", fixed: "right" },
]);

function openDetail(row: AgentRun) {
  void router.push(`/ai/runs/${row.id}`);
}
</script>

<template>
  <div class="page">
    <ProTable ref="table" url="/ai/runs" mode="pagination" :columns="columns" v-model:query="query">
      <template #filters>
        <u-input v-model="query.agent_id" placeholder="agent_id" clearable />
        <u-input v-model="query.status" placeholder="status" clearable />
      </template>
      <template #column:trigger_type="{ rowData }">
        <u-tag size="small" :type="tagType((rowData as AgentRun).trigger_type, TRIGGER_TYPE_TAG)">
          {{ (rowData as AgentRun).trigger_type }}
        </u-tag>
      </template>
      <template #column:status="{ rowData }">
        <u-tag size="small" :type="tagType((rowData as AgentRun).status, JOB_STATUS_TAG)">
          {{ (rowData as AgentRun).status }}
        </u-tag>
      </template>
      <template #column:action="{ rowData }">
        <u-action @run="openDetail(rowData as AgentRun)">详情</u-action>
      </template>
    </ProTable>
  </div>
</template>

<style scoped lang="scss">
.page {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
</style>

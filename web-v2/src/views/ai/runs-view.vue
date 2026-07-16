<script setup lang="ts">
import { reactive, useTemplateRef } from "vue";
import { useRouter } from "vue-router";
import { defineTableColumns } from "@veltra/desktop";

import type { AgentRun } from "@/api/types";
import ProTable from "@/components/pro-table.vue";

const router = useRouter();
const table = useTemplateRef("table");
const query = reactive({ agent_id: "", status: "" });

const columns = defineTableColumns([
  { key: "id", name: "ID", width: 70 },
  { key: "agent_id", name: "Agent", width: 90 },
  { key: "trigger_type", name: "触发", width: 120 },
  { key: "status", name: "状态", width: 110 },
  { key: "created_at", name: "创建时间", minWidth: 160 },
  { key: "action", name: "操作", width: 100 },
]);

function openDetail(row: AgentRun) {
  void router.push(`/ai/runs/${row.id}`);
}
</script>

<template>
  <div class="page">
    <h2>Agent 运行</h2>
    <ProTable ref="table" url="/ai/runs" mode="pagination" :columns="columns" v-model:query="query">
      <template #filters>
        <u-input v-model="query.agent_id" placeholder="agent_id" clearable />
        <u-input v-model="query.status" placeholder="status" clearable />
      </template>
      <template #action="{ rowData }">
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

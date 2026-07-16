<script setup lang="ts">
import { reactive, useTemplateRef } from "vue";
import { defineTableColumns, message } from "@veltra/desktop";

import { killProcess } from "@/api/ops";
import type { ProcessInfo } from "@/api/types";
import ProTable from "@/components/pro-table.vue";

const listRef = useTemplateRef("list");
const query = reactive({
  keyword: "",
  pid: "",
  port: "",
  sort: "cpu",
  order: "desc",
});

const columns = defineTableColumns([
  { key: "pid", name: "PID", width: 90, minWidth: 70 },
  { key: "name", name: "名称", minWidth: 140 },
  { key: "cpu_percent", name: "CPU", width: 100, minWidth: 80 },
  { key: "memory_bytes", name: "内存", width: 130, minWidth: 100 },
  { key: "username", name: "用户", minWidth: 120 },
  { key: "ports", name: "监听端口", minWidth: 120 },
  { key: "action", name: "操作", width: 90, minWidth: 70 },
]);

function formatBytes(value: number): string {
  if (!value) return "—";
  const units = ["B", "KB", "MB", "GB", "TB"];
  const index = Math.min(Math.floor(Math.log(value) / Math.log(1024)), units.length - 1);
  return `${(value / 1024 ** index).toFixed(index ? 1 : 0)} ${units[index]}`;
}

async function terminate(row: ProcessInfo) {
  if (!window.confirm(`确认终止进程 ${row.name}（PID ${row.pid}）？此操作不可撤销。`)) return;
  try {
    await killProcess(row.pid);
    message.success("进程终止请求已发送");
    await listRef.value?.reload();
  } catch (error) {
    message.error(error instanceof Error ? error.message : "终止进程失败");
  }
}
</script>

<template>
  <div class="page">
    <div class="page-head">
      <div>
        <h2>进程管理</h2>
        <p>仅超级管理员可操作。系统关键进程及 Bedrock 自身受保护。</p>
      </div>
    </div>

    <ProTable
      ref="list"
      url="/ops/processes"
      v-model:query="query"
      :columns="columns"
      :auto-query-fields="['sort', 'order']"
    >
      <template #filters="{ search }">
        <u-input v-model="query.keyword" placeholder="名称关键词" style="width: 160px" />
        <u-input v-model="query.pid" type="number" placeholder="PID" style="width: 110px" />
        <u-input v-model="query.port" type="number" placeholder="端口" style="width: 110px" />
        <u-select
          v-model="query.sort"
          style="width: 120px"
          :options="[
            { label: 'CPU', value: 'cpu' },
            { label: '内存', value: 'memory' },
            { label: '名称', value: 'name' },
          ]"
        />
        <u-select
          v-model="query.order"
          style="width: 100px"
          :options="[
            { label: '降序', value: 'desc' },
            { label: '升序', value: 'asc' },
          ]"
        />
        <u-button type="primary" @click="search">查询</u-button>
      </template>
      <template #column:cpu_percent="{ rowData }">
        {{ (rowData as ProcessInfo).cpu_percent.toFixed(1) }}%
      </template>
      <template #column:memory_bytes="{ rowData }">
        {{ formatBytes((rowData as ProcessInfo).memory_bytes) }}
      </template>
      <template #column:ports="{ rowData }">
        {{ (rowData as ProcessInfo).ports.join(", ") || "—" }}
      </template>
      <template #column:action="{ rowData }">
        <u-action danger @run="terminate(rowData as ProcessInfo)">终止</u-action>
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
.page-head p {
  margin: 6px 0 0;
  color: #6b7280;
}
</style>

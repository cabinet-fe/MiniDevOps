<script setup lang="ts">
import { reactive, useTemplateRef } from "vue";
import { message } from "@veltra/desktop";

import { killProcess } from "@/api/ops";
import type { ProcessInfo } from "@/api/types";
import ProTable, { defineProTableColumns } from "@/components/pro-table";
import { tagType, type TagType } from "@/lib/tag";

const PROCESS_STATUS_TAG: Record<string, TagType> = {
  R: "success",
  S: undefined,
  D: "warning",
  Z: "danger",
  T: "info",
  t: "info",
  X: "danger",
  I: undefined,
};

const listRef = useTemplateRef("list");
const query = reactive({
  keyword: "",
  pid: "",
  port: "",
  sort: "cpu_percent@desc",
});

const columns = defineProTableColumns([
  { key: "pid", name: "PID", width: 90 },
  { key: "name", name: "名称", sortable: true },
  { key: "cpu_percent", name: "CPU", width: 90, sortable: true },
  { key: "memory_bytes", name: "内存", width: 100, sortable: true },
  { key: "username", name: "用户", width: 100 },
  { key: "num_threads", name: "线程", width: 80 },
  { key: "status", name: "状态", width: 80 },
  { key: "start_time", name: "启动时间", width: 170 },
  { key: "ports", name: "监听端口", width: 120 },
  { key: "cmdline", name: "命令行" },
  { key: "action", name: "操作", width: 90, align: "center", fixed: "right" },
]);

function formatBytes(value: number): string {
  if (!value) return "—";
  const units = ["B", "KB", "MB", "GB", "TB"];
  const index = Math.min(Math.floor(Math.log(value) / Math.log(1024)), units.length - 1);
  return `${(value / 1024 ** index).toFixed(index ? 1 : 0)} ${units[index]}`;
}

function formatStartTime(ms: number): string {
  if (!ms) return "—";
  return new Date(ms).toLocaleString("zh-CN");
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
      row-key="pid"
    >
      <template #filters="{ search }">
        <u-input v-model="query.keyword" placeholder="名称 / 命令行" style="width: 180px" />
        <u-input v-model="query.pid" type="number" placeholder="PID" style="width: 110px" />
        <u-input v-model="query.port" type="number" placeholder="端口" style="width: 110px" />
        <u-button type="primary" @click="search">查询</u-button>
      </template>
      <template #column:cpu_percent="{ rowData }">
        {{ (rowData as ProcessInfo).cpu_percent.toFixed(1) }}%
      </template>
      <template #column:memory_bytes="{ rowData }">
        {{ formatBytes((rowData as ProcessInfo).memory_bytes) }}
      </template>
      <template #column:num_threads="{ rowData }">
        {{ (rowData as ProcessInfo).num_threads || "—" }}
      </template>
      <template #column:status="{ rowData }">
        <u-tag
          v-if="(rowData as ProcessInfo).status"
          size="small"
          :type="tagType((rowData as ProcessInfo).status, PROCESS_STATUS_TAG)"
        >
          {{ (rowData as ProcessInfo).status }}
        </u-tag>
        <span v-else>—</span>
      </template>
      <template #column:start_time="{ rowData }">
        {{ formatStartTime((rowData as ProcessInfo).start_time) }}
      </template>
      <template #column:ports="{ rowData }">
        {{ (rowData as ProcessInfo).ports.join(", ") || "—" }}
      </template>
      <template #column:cmdline="{ rowData }">
        <span class="cmdline" :title="(rowData as ProcessInfo).cmdline || undefined">
          {{ (rowData as ProcessInfo).cmdline || "—" }}
        </span>
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
.cmdline {
  display: block;
  max-width: 280px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 12px;
}
</style>

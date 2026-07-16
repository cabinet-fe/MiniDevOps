<script setup lang="ts">
import { reactive } from "vue";

import type { OperationLog } from "@/api/types";
import ProTable, { defineProTableColumns } from "@/components/pro-table";
import { tagType, type TagType } from "@/lib/tag";

const query = reactive({
  action: "",
  resource_type: "",
});

const ACTION_TAG: Record<string, TagType> = {
  GET: "info",
  POST: "primary",
  PUT: "warning",
  PATCH: "warning",
  DELETE: "danger",
};

const columns = defineProTableColumns([
  { key: "id", name: "ID", width: 80 },
  { key: "username", name: "用户" },
  { key: "action", name: "动作", width: 90 },
  { key: "resource_type", name: "资源" },
  { key: "resource_id", name: "资源ID", width: 90 },
  { key: "ip_address", name: "IP", width: 130 },
  { key: "created_at", name: "时间", sortable: true },
]);
</script>

<template>
  <div class="page">
    <div class="page-head">
      <h2>操作日志</h2>
    </div>

    <ProTable url="/operation-logs" v-model:query="query" :columns="columns" pagination>
      <template #filters="{ search }">
        <u-input v-model="query.action" placeholder="动作 (POST/PUT…)" style="width: 160px" />
        <u-input v-model="query.resource_type" placeholder="资源路径" style="width: 220px" />
        <u-button type="primary" @click="search">查询</u-button>
      </template>
      <template #column:action="{ rowData }">
        <u-tag size="small" :type="tagType((rowData as OperationLog).action, ACTION_TAG)">
          {{ (rowData as OperationLog).action }}
        </u-tag>
      </template>
    </ProTable>
  </div>
</template>

<style scoped>
.page-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}
.page-head h2 {
  margin: 0;
  font-size: 20px;
}
</style>

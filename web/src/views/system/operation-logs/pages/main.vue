<script setup lang="ts">
defineOptions({ name: "SystemOperationLogs" });

import { reactive } from "vue";

import type { OperationLog } from "@/api/types";
import ProTable, { defineProTableColumns } from "@/components/pro-table";
import { formatDateTime } from "@/lib/datetime";
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
  { key: "id", name: "ID" },
  { key: "username", name: "用户" },
  { key: "action", name: "动作", width: 90, align: "center" },
  { key: "resource_type", name: "资源" },
  { key: "resource_id", name: "资源ID" },
  { key: "ip_address", name: "IP" },
  {
    key: "created_at",
    name: "时间",
    width: 170,
    align: "center",
    sortable: true,
    render: ({ val }) => formatDateTime(val),
  },
]);
</script>

<template>
  <div>
    <ProTable url="/operation-logs" :query="query" :columns="columns" pagination>
      <template #filters>
        <u-input v-model="query.action" placeholder="动作 (POST/PUT…)" style="width: 160px" />
        <u-input v-model="query.resource_type" placeholder="资源路径" style="width: 220px" />
      </template>
      <template #column:action="{ rowData }">
        <u-tag size="small" :type="tagType((rowData as OperationLog).action, ACTION_TAG)">
          {{ (rowData as OperationLog).action }}
        </u-tag>
      </template>
    </ProTable>
  </div>
</template>

<script setup lang="ts">
import { reactive } from "vue";
import { defineTableColumns } from "@veltra/desktop";

import ProTable from "@/components/pro-table.vue";

const query = reactive({
  action: "",
  resource_type: "",
});

const columns = defineTableColumns([
  { key: "id", name: "ID", width: 80, minWidth: 60 },
  { key: "username", name: "用户", minWidth: 100 },
  { key: "action", name: "动作", minWidth: 80 },
  { key: "resource_type", name: "资源", minWidth: 160 },
  { key: "resource_id", name: "资源ID", width: 100, minWidth: 80 },
  { key: "ip_address", name: "IP", width: 140, minWidth: 100 },
  { key: "created_at", name: "时间", minWidth: 180, sortable: true },
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

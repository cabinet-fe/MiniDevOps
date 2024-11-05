<template>
  <u-table :columns="columns" />
</template>

<script lang="ts" setup>
import { defineTableColumns } from 'ultra-ui'

const columns = defineTableColumns([{ name: '名称', key: 'name' }])
</script>

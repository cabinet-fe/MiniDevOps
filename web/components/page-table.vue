<template>
  <div>
    <div></div>
    <u-table :columns="ctx.columns" :data="data" :slots="slots">
      <template #column:action="s"></template>
    </u-table>
    <u-paginator
      v-model:page-number="params.page"
      v-model:page-size="params.pageSize"
      :total="total"
    />

    <slot name="dialog" />
  </div>
</template>

<script setup lang="ts">
import type { PageParams, PageTableCtx } from '@/hooks'
import type { TableColumnSlotsScope } from 'ultra-ui'
import { shallowReactive, shallowRef } from 'vue'

defineOptions({
  name: 'PageTable'
})

const { ctx } = defineProps<{
  ctx: PageTableCtx
}>()

const slots = defineSlots<{
  dialog: () => any
  [key: `column:${string}`]: (scope: TableColumnSlotsScope) => any
}>()

const params = shallowReactive<PageParams>({
  page: 1,
  pageSize: 20
})

const data = shallowRef<Record<string, any>[]>([])
const total = shallowRef(0)

ctx.getData(params).then(res => {
  data.value = res.data
  total.value = res.total
})
</script>

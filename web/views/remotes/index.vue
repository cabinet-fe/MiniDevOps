<template>
  <m-table-pro :columns="columns" :server="server" pagination ref="tableRef">
    <template #searcher>
      <u-input label="名称" v-model="queries.name" />
    </template>

    <template #column:action="{ rowData }">
      <u-action-group>
        <u-action circle @run="formRef?.open('update', rowData)">
          <u-icon><EditPen /></u-icon>
        </u-action>
        <u-action
          need-confirm
          type="danger"
          circle
          @run="handleDelete(rowData)"
        >
          <u-icon><Delete /></u-icon>
        </u-action>
      </u-action-group>
    </template>

    <template #tools>
      <RemoteForm ref="formRef" @success="tableRef?.fetchData()" />
    </template>
  </m-table-pro>
</template>

<script lang="ts" setup>
import { defineTableProServer } from '@meta/components'
import { defineTableColumns } from 'ultra-ui'
import { shallowReactive, useTemplateRef } from 'vue'
import RemoteForm from './form.vue'
import { Delete, EditPen } from 'icon-ultra'
import { http } from '@/utils/http'

const columns = defineTableColumns(
  [
    { name: '目录名称', key: 'name' },
    { name: '主机地址', key: 'host' },
    { name: '目录', key: 'path' },
    { name: '操作', key: 'action', width: 100 }
  ],
  { align: 'center' }
)

const queries = shallowReactive({
  name: ''
})
const server = defineTableProServer({
  api: '/remotes/page',
  queries,
  dataPath: 'data.rows',
  totalPath: 'data.total'
})

const formRef = useTemplateRef('formRef')
const tableRef = useTemplateRef('tableRef')

const handleDelete = async (row: any) => {
  await http.delete(`/remotes/${row.id}`)
  tableRef.value?.fetchData()
}
</script>

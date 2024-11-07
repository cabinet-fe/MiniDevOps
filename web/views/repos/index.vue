<template>
  <m-table-pro :columns="columns" :server="server" pagination ref="tableRef">
    <template #searcher>
      <u-input label="名称" v-model="queries.name" />
    </template>

    <template #column:actions="{ rowData }">
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
      <RepoForm ref="formRef" @success="tableRef?.fetchData()" />
    </template>
  </m-table-pro>
</template>

<script lang="ts" setup>
import { defineTableProServer } from '@meta/components'
import { defineTableColumns, message } from 'ultra-ui'
import { shallowReactive, useTemplateRef } from 'vue'
import RepoForm from './form.vue'
import { Delete, EditPen } from 'icon-ultra'
import { http } from '@/utils/http'

const columns = defineTableColumns(
  [
    { name: '名称', key: 'name' },
    { name: '仓库地址', key: 'address' },
    { name: '存放目录', key: 'codePath' },
    { name: '操作', key: 'actions', width: 120, align: 'center' }
  ],
  { align: 'center' }
)

const queries = shallowReactive({
  name: ''
})
const server = defineTableProServer({
  api: '/repos/page',
  queries,
  dataPath: 'data.rows',
  totalPath: 'data.total'
})

const tableRef = useTemplateRef('tableRef')

const formRef = useTemplateRef('formRef')

const handleDelete = async (row: any) => {
  await http.delete(`/repos/${row.id}`)
  tableRef.value?.fetchData()
  message.success('删除成功')
}
</script>

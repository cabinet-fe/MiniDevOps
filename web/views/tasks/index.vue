<template>
  <m-table-pro :columns="columns" :server="server" pagination ref="tableRef">
    <template #searcher>
      <u-input label="名称" v-model="queries.name" />
    </template>

    <template #column:action="{ rowData }">
      <u-action-group :max="4">
        <u-action @run="handleBuild(rowData)">构建</u-action>
        <u-action @run="formRef?.open('update', rowData)">编辑</u-action>
        <u-action type="danger" @run="handleDelete(rowData)">删除</u-action>
      </u-action-group>
    </template>

    <template #tools>
      <TaskForm ref="formRef" @success="tableRef?.fetchData()" />
    </template>
  </m-table-pro>
</template>

<script lang="ts" setup>
import { defineTableProServer } from '@meta/components'
import { defineTableColumns, message } from 'ultra-ui'
import { shallowReactive, useTemplateRef } from 'vue'
import TaskForm from './form.vue'
import { http } from '@/utils/http'

const columns = defineTableColumns(
  [
    { name: '任务名称', key: 'name' },
    { name: '代码仓库', key: 'repo.name' },
    { name: '代码分支', key: 'branch' },
    { name: '构建进度', key: 'progress' },
    { name: '上次构建状态', key: 'status' },
    { name: '构建物目录', key: 'bundlerDir' },
    { name: '远程目录', key: 'remote.path' },
    { name: '操作', key: 'action', width: 150, fixed: 'right' }
  ],
  { align: 'center' }
)

const queries = shallowReactive({
  name: ''
})
const server = defineTableProServer({
  api: '/tasks/page',
  queries,
  dataPath: 'data.rows',
  totalPath: 'data.total'
})

const tableRef = useTemplateRef('tableRef')
const formRef = useTemplateRef('formRef')

const handleBuild = (row: Record<string, any>) => {
  console.log(row)
}

const handleDelete = async (row: Record<string, any>) => {
  await http.delete(`/tasks/${row.id}`)
  tableRef.value?.fetchData()
  message.success('删除成功')
}
</script>

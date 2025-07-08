<template>
  <m-table-pro :columns="columns" :server="server" pagination ref="table">
    <template #searcher>
      <u-input label="名称" v-model="queries.name" />
    </template>

    <template #column:status="{ rowData }">
      <u-tag type="success" v-if="buildingTasks.has(rowData.id)">
        构建中
      </u-tag>
      <u-tag v-else> 空闲 </u-tag>
    </template>

    <template #column:action="{ rowData }">
      <u-action-group :max="4">
        <u-action
          v-if="!buildingTasks.has(rowData.id)"
          @run="handleBuild(rowData)"
          title="构建"
        >
          <u-icon><VideoPlay /></u-icon>
        </u-action>

        <u-action @run="formRef?.open('update', rowData)">
          <u-icon><EditPen /></u-icon>
        </u-action>
        <u-action type="danger" @run="handleDelete(rowData)">
          <u-icon><Delete /></u-icon>
        </u-action>
      </u-action-group>
    </template>

    <template #tools>
      <TaskForm ref="form" @success="tableRef?.fetchData()" />
    </template>
  </m-table-pro>
</template>

<script lang="ts" setup>
import { defineTableProServer } from '@meta/components'
import { defineTableColumns, message } from 'ultra-ui'
import {
  shallowReactive,
  useTemplateRef,
  onBeforeUnmount,
  shallowRef
} from 'vue'
import TaskForm from './form.vue'
import { http } from '@/utils/http'
import { VideoPlay, VideoPause, Delete, EditPen } from 'icon-ultra'

const columns = defineTableColumns(
  [
    { name: '任务名称', key: 'name' },
    { name: '代码仓库', key: 'repo.name' },
    { name: '代码分支', key: 'branch' },
    { name: '状态', key: 'status' },
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

const tableRef = useTemplateRef('table')
const formRef = useTemplateRef('form')

const handleBuild = async (row: Record<string, any>) => {
  await http.post(`/tasks/${row.id}/build`)
  message.success('构建中')
}

const handleStopBuild = async (row: Record<string, any>) => {
  await http.post(`/tasks/${row.id}/stop-build`)
  message.success('停止构建')
}

const handleDelete = async (row: Record<string, any>) => {
  await http.delete(`/tasks/${row.id}`)
  tableRef.value?.fetchData()
  message.success('删除成功')
}

// const socket = new WebSocket(`ws://${location.host}/ws/tasks/progress`)

// socket.addEventListener('open', () => {
//   socket.send('connect')
// })

// socket.addEventListener('error', e => {
//   console.log(e)
// })

const buildingTasks = shallowRef<Set<number>>(new Set())

// socket.addEventListener('message', event => {
//   const msg = JSON.parse(event.data)

//   if (msg.type === 'progress') {
//     buildingTasks.value = new Set(msg.data)
//   } else if (msg.type === 'result') {
//     const { status, taskName } = msg.data

//     if (status === 'success') {
//       message.success(`${taskName} 构建成功`)
//     } else {
//       message.error(`${taskName} 构建失败`)
//     }
//   }
// })

onBeforeUnmount(() => {
  // socket.close()
})
</script>

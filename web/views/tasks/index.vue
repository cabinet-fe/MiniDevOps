<template>
  <div>
    <u-button type="primary" @click="handleAdd()">新增</u-button>
  </div>
  <u-table
    :columns="columns"
    :request="taskService.getTasks"
    row-key="id"
    ref="tableRef"
  >
    <template #column:action="{ rowData }">
      <u-action-group>
        <u-action @run="handleRun(rowData.id)">运行</u-action>
        <u-action type="primary" @run="handleEdit(rowData)">编辑</u-action>
        <u-action type="danger" need-confirm @run="handleDelete(rowData.id)">
          删除
        </u-action>
      </u-action-group>
    </template>
  </u-table>

  <u-dialog
    :title="dialogType === 'edit' ? '编辑任务' : '新增任务'"
    v-model="visible"
    @confirm="handleSubmit"
  >
    <u-form :model="model">
      <u-input label="任务名称" field="name" />
      <u-select
        label="关联仓库"
        field="repoId"
        :request="repoService.getRepos"
        value-key="id"
        label-key="name"
      />
    </u-form>
  </u-dialog>
</template>

<script setup lang="ts">
import { taskService } from '@/apis/task'
import { repoService } from '@/apis/repo'
import { useTable, useFormDialog } from '@/hooks'
import { defineTableColumns } from 'ultra-ui'

const { tableRef, reload } = useTable()
const { open, visible, dialogType, model } = useFormDialog({
  name: {
    value: '',
    required: true
  },
  repoId: { value: undefined }
})

const columns = defineTableColumns([
  { key: 'name', name: '任务名称' },
  { key: 'repo.name', name: '关联仓库' },
  { key: 'status', name: '状态' },
  { key: 'createdAt', name: '创建时间' },
  { key: 'actions', name: '操作', width: 240 }
])

function handleAdd() {
  open('create')
}

function handleEdit(row: Record<string, any>) {
  open('edit', { data: row })
}

async function handleRun(id: number) {
  await taskService.runTask(id)
  reload()
}

async function handleDelete(id: number) {
  await taskService.deleteTask(id)
  reload()
}

async function handleSubmit() {
  if (dialogType.value === 'edit') {
    await taskService.updateTask(model.data)
  } else {
    await taskService.createTask(model.data)
  }
  reload()
}
</script>

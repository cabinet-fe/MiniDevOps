<template>
  <u-table
    :columns="columns"
    :request="taskService.getTasks"
    row-key="id"
    ref="tableRef"
  >
    <template #tool>
      <u-button type="primary" @click="handleAdd">新增</u-button>
    </template>
  </u-table>

  <u-dialog
    :title="isEdit ? '编辑任务' : '新增任务'"
    ref="dialogRef"
    @confirm="handleSubmit"
  >
    <u-form :model="model">
      <u-form-item label="任务名称" field="name">
        <u-input field="name" />
      </u-form-item>
      <u-form-item label="关联仓库" field="repoId">
        <u-select
          :request="repoService.getRepos"
          field="repoId"
          value-key="id"
          label-key="name"
        />
      </u-form-item>
    </u-form>
  </u-dialog>
</template>

<script setup lang="ts">
import { taskService } from '@/apis/task'
import { repoService } from '@/apis/repo'
import type { Task } from '@/types'
import { useTable, useDialog, useForm } from '@/hooks'

const { tableRef, reload } = useTable()
const { dialogRef, open, isEdit } = useDialog()
const { model, create, update } = useForm<Task>({ name: '', repoId: null })

const columns = [
  { prop: 'name', label: '任务名称' },
  { prop: 'repo.name', label: '关联仓库' },
  { prop: 'status', label: '状态' },
  { prop: 'createdAt', label: '创建时间' },
  {
    prop: 'actions',
    label: '操作',
    width: 240,
    render: (_: any, row: Task) => {
      return (
        <u-action-group>
          <u-action @click="handleRun(row.id)">运行</u-action>
          <u-action @click="handleEdit(row)">编辑</u-action>
          <u-action type="danger" @click="handleDelete(row.id)">
            删除
          </u-action>
        </u-action-group>
      )
    }
  }
]

function handleAdd() {
  create()
  isEdit.value = false
  open()
}

function handleEdit(row: Task) {
  update(row)
  isEdit.value = true
  open()
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
  if (isEdit.value) {
    await taskService.updateTask(model.data)
  } else {
    await taskService.createTask(model.data)
  }
  reload()
}
</script>

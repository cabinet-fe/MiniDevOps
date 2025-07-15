<template>
  <div>
    <u-button type="primary" @click="handleAdd()">新增</u-button>
  </div>
  <u-table
    :columns="columns"
    :request="remoteService.getRemotes"
    row-key="id"
    ref="tableRef"
  >
    <template #column:action="{ rowData }">
      <u-action-group>
        <u-action type="primary" @run="handleEdit(rowData)">编辑</u-action>
        <u-action type="danger" need-confirm @run="handleDelete(rowData.id)">
          删除
        </u-action>
      </u-action-group>
    </template>
  </u-table>

  <u-dialog
    :title="dialogType === 'edit' ? '编辑远程目录' : '新增远程目录'"
    v-model="visible"
    @confirm="handleSubmit"
  >
    <u-form :model="model">
      <u-input label="名称" field="name" />
      <u-input label="用户" field="user" />
      <u-input label="地址" field="addr" />
    </u-form>
  </u-dialog>
</template>

<script setup lang="ts">
import { remoteService } from '@/apis/remote'
import { useTable, useFormDialog } from '@/hooks'
import { defineTableColumns } from 'ultra-ui'

const { tableRef, reload } = useTable()
const { open, visible, dialogType, model } = useFormDialog({
  name: {
    value: '',
    required: true
  },
  user: { value: '' },
  addr: { value: '' }
})

const columns = defineTableColumns([
  { key: 'name', name: '名称' },
  { key: 'user', name: '用户' },
  { key: 'addr', name: '地址' },
  { key: 'createdAt', name: '创建时间' },
  { key: 'actions', name: '操作', width: 180 }
])

function handleAdd() {
  open('create')
}

function handleEdit(row: Record<string, any>) {
  open('edit', { data: row })
}

async function handleDelete(id: number) {
  await remoteService.deleteRemote(id)
  reload()
}

async function handleSubmit() {
  if (dialogType.value === 'edit') {
    await remoteService.updateRemote(model.data)
  } else {
    await remoteService.createRemote(model.data)
  }
  reload()
}
</script>

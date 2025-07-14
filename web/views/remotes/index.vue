<template>
  <u-table
    :columns="columns"
    :request="remoteService.getRemotes"
    row-key="id"
    ref="tableRef"
  >
    <template #tool>
      <u-button type="primary" @click="handleAdd">新增</u-button>
    </template>
  </u-table>

  <u-dialog
    :title="isEdit ? '编辑远程目录' : '新增远程目录'"
    ref="dialogRef"
    @confirm="handleSubmit"
  >
    <u-form :model="model">
      <u-form-item label="名称" field="name">
        <u-input field="name" />
      </u-form-item>
      <u-form-item label="用户" field="user">
        <u-input field="user" />
      </u-form-item>
      <u-form-item label="地址" field="addr">
        <u-input field="addr" />
      </u-form-item>
    </u-form>
  </u-dialog>
</template>

<script setup lang="ts">
import { remoteService } from '@/apis/remote'
import type { Remote } from '@/types'
import { useTable, useDialog, useForm } from '@/hooks'

const { tableRef, reload } = useTable()
const { dialogRef, open, isEdit } = useDialog()
const { model, create, update } = useForm<Remote>({
  name: '',
  user: '',
  addr: ''
})

const columns = [
  { prop: 'name', label: '名称' },
  { prop: 'user', label: '用户' },
  { prop: 'addr', label: '地址' },
  { prop: 'createdAt', label: '创建时间' },
  {
    prop: 'actions',
    label: '操作',
    width: 180,
    render: (_: any, row: Remote) => {
      return (
        <u-action-group>
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

function handleEdit(row: Remote) {
  update(row)
  isEdit.value = true
  open()
}

async function handleDelete(id: number) {
  await remoteService.deleteRemote(id)
  reload()
}

async function handleSubmit() {
  if (isEdit.value) {
    await remoteService.updateRemote(model.data)
  } else {
    await remoteService.createRemote(model.data)
  }
  reload()
}
</script>

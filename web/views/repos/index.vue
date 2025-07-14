<template>
  <div>
    <u-button type="primary" @click="handleAdd()">新增</u-button>
  </div>
  <u-table
    :columns="columns"
    :request="repoService.getRepos"
    row-key="id"
    ref="tableRef"
  >
    <template #column:action="{ row }">
      <u-button type="primary" @click="handleEdit(row)">编辑</u-button>
    </template>
  </u-table>

  <u-dialog
    :title="dialogType === 'edit' ? '编辑仓库' : '新增仓库'"
    v-model="visible"
  >
    <u-form :model="formModel">
      <u-input label="仓库名称" field="name" />
      <u-input label="仓库地址" field="url" />
    </u-form>
  </u-dialog>
</template>

<script setup lang="ts">
import { repoService } from '@/apis/repo'
import type { Repo } from '@/types'
import { useTable, useDialog } from '@/hooks'
import { defineTableColumns, FormModel, type TableRow } from 'ultra-ui'

const { tableRef, reload } = useTable()
const { open, visible, dialogType } = useDialog()

const columns = defineTableColumns([
  { key: 'name', name: '仓库名称' },
  { key: 'url', name: '仓库地址' },
  { key: 'createdAt', name: '创建时间' },
  { key: 'actions', name: '操作', width: 180 }
])

const formModel = new FormModel({
  name: {
    value: '',
    required: true
  },
  url: {
    value: ''
  }
})

function handleAdd() {
  open('create')
}

function handleEdit(row: TableRow) {
  open('edit', { data: row })
}

async function handleDelete(id: number) {
  await repoService.deleteRepo(id)
  reload()
}

async function handleSubmit() {
  if (dialogType.value === 'edit') {
    await repoService.updateRepo(formModel.data)
  } else {
    await repoService.createRepo(formModel.data)
  }
  reload()
}
</script>

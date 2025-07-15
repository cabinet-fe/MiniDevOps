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
    :title="dialogType === 'edit' ? '编辑仓库' : '新增仓库'"
    v-model="visible"
  >
    <u-form :model="model">
      <u-input label="仓库名称" field="name" />
      <u-input label="仓库地址" field="url" />
    </u-form>
  </u-dialog>
</template>

<script setup lang="ts">
import { repoService } from '@/apis/repo'
import { useTable, useFormDialog } from '@/hooks'
import { defineTableColumns } from 'ultra-ui'

const { tableRef, reload } = useTable()
const { open, visible, dialogType, model } = useFormDialog({
  name: {
    value: '',
    required: true
  },
  url: { value: '' }
})

const columns = defineTableColumns([
  { key: 'name', name: '仓库名称' },
  { key: 'url', name: '仓库地址' },
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
  await repoService.deleteRepo(id)
  reload()
}

async function handleSubmit() {
  if (dialogType.value === 'edit') {
    await repoService.updateRepo(model.data)
  } else {
    await repoService.createRepo(model.data)
  }
  reload()
}
</script>

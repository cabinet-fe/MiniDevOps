<template>
  <m-page-table :ctx="ctx">
    <template #dialog>
      <m-dialog
        :visible="visible"
        :title="dialogType === 'edit' ? '编辑仓库' : '新增仓库'"
        :model="model"
      >
        <u-input label="仓库名称" field="name" />
        <u-input label="仓库地址" field="url" />
      </m-dialog>
    </template>

    <template #column:action="{ rowData }">
      <u-action-group>
        <u-action type="primary" @run="handleEdit(rowData)">编辑</u-action>
        <u-action type="danger" need-confirm @run="handleDelete(rowData.id)">
          删除
        </u-action>
      </u-action-group>
    </template>
  </m-page-table>
</template>

<script setup lang="ts">
import { repoService } from '@/apis/repo'
import { useTable, useFormDialog } from '@/hooks'

const { open, visible, dialogType, model } = useFormDialog({
  name: {
    value: '',
    required: true
  },
  url: { value: '' }
})

const ctx = useTable({
  columns: [
    { key: 'name', name: '仓库名称' },
    { key: 'url', name: '仓库地址' },
    { key: 'createdAt', name: '创建时间' },
    { key: 'actions', name: '操作', width: 180 }
  ],
  getData: repoService.getRepoPage
})

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

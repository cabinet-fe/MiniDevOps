<template>
  <m-dialog-pro :ctx="ctx" style="width: 500px">
    <u-form :model="model">
      <u-input label="目录名称" field="name" />
      <u-input label="主机地址" field="host" />
      <u-input label="目录" field="path" />
    </u-form>

    <template #trigger>
      <u-button type="primary" @click="ctx.open({ action: 'create' })">
        新增
      </u-button>
    </template>
  </m-dialog-pro>
</template>

<script lang="ts" setup>
import { http } from '@/utils/http'
import { useDialogPro } from '@meta/components'
import { FormModel } from 'ultra-ui'

const emit = defineEmits(['success'])

const model = new FormModel({
  name: { required: true },
  host: { required: true },
  path: { required: true },
  id: {}
})

const ctx = useDialogPro({
  models: [model],
  async submit(state) {
    if (state.action === 'create') {
      await http.post('/remotes', model.data)
    } else {
      await http.put(`/remotes/${model.data.id}`, model.data)
    }
    emit('success')
  },
  afterOpen(state) {
    if (!state.action) return
    state.title = {
      create: '新增远程目录',
      update: '修改远程目录'
    }[state.action]
  }
})

defineExpose({
  open(action: 'update', data: Record<string, any>) {
    model.setData(data)
    ctx.open({ action })
  }
})
</script>

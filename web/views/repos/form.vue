<template>
  <m-dialog-pro :ctx="ctx" style="width: 500px">
    <u-form :model="model">
      <u-input label="名称" field="name" />
      <u-input label="仓库地址" field="address" />
      <u-input
        label="存放目录"
        field="codePath"
        tips="代码拉取后存放的目标目录"
      />
      <template v-if="ctx.actionOf(['create'])">
        <u-input label="用户名" field="username" />
        <u-password-input label="密码" field="pwd" />
      </template>
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
  address: { required: true },
  codePath: { required: true },
  username: { required: true },
  pwd: { required: true },
  id: {}
})

const ctx = useDialogPro({
  models: [model],
  async submit(state) {
    if (state.action === 'create') {
      await http.post('/repos', model.data)
    } else {
      await http.put(`/repos/${model.data.id}`, model.data)
    }

    emit('success')
  },
  afterOpen(state) {
    if (!state.action) return
    state.title = {
      create: '新增仓库',
      update: '修改仓库'
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

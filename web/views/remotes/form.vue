<template>
  <u-dialog v-model="visible" style="width: 500px">
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
  </u-dialog>
</template>

<script lang="ts" setup>
import { http } from '@/utils/http'
import { FormModel } from 'ultra-ui'
import { ref } from 'vue'

const emit = defineEmits(['success'])

const model = new FormModel({
  name: { required: true },
  host: { required: true },
  path: { required: true },
  id: {}
})

const visible = ref(false)

defineExpose({
  open(action: 'update', data: Record<string, any>) {
    model.setData(data)
  }
})
</script>

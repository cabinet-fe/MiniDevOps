<template>
  <u-dialog
    :model-value="visible"
    @update:model-value="emit('update:visible', $event)"
    :title="title"
    style="width: 900px"
  >
    <u-form :model="model">
      <slot />

      <div class="flex justify-end">
        <u-button @click="emit('update:visible', false)">取消</u-button>
        <u-button type="primary" @click="emit('submit')">提交</u-button>
      </div>
    </u-form>
  </u-dialog>
</template>

<script setup lang="ts">
import type { FormModel } from 'ultra-ui'

defineOptions({
  name: 'Dialog'
})

defineProps<{
  visible: boolean
  title: string
  model: FormModel
}>()

const emit = defineEmits<{
  (e: 'submit'): void
  (e: 'update:visible', visible: boolean): void
}>()
</script>

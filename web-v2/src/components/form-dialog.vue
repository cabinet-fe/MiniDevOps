<script setup lang="ts">
import { ref, useTemplateRef, watch } from "vue";

const props = withDefaults(
  defineProps<{
    modelValue?: boolean;
    title?: string;
    /** 表单 model；打开弹框前写入初始值，关闭时自动 reset 回打开时快照 */
    model: Record<string, any>;
    labelWidth?: string | number;
    cols?: number;
    confirmText?: string;
    cancelText?: string;
    confirmLoading?: boolean;
  }>(),
  {
    modelValue: false,
    cols: 1,
    labelWidth: "88px",
    confirmText: "保存",
    cancelText: "取消",
    confirmLoading: false,
  },
);

const emit = defineEmits<{
  "update:modelValue": [value: boolean];
  /** 校验通过后触发 */
  submit: [];
  closed: [];
}>();

const formRef = useTemplateRef("form");
/** 每次打开递增，让表单按当前 model 重新快照 */
const sessionKey = ref(0);

watch(
  () => props.modelValue,
  (open) => {
    if (open) sessionKey.value += 1;
  },
);

function close() {
  emit("update:modelValue", false);
}

async function onConfirm() {
  const ok = await formRef.value?.validate();
  if (!ok) return;
  emit("submit");
}

function onClosed() {
  formRef.value?.reset();
  emit("closed");
}

defineExpose({
  validate: () => formRef.value?.validate() ?? Promise.resolve(false),
  reset: () => formRef.value?.reset(),
  close,
});
</script>

<template>
  <u-dialog
    :model-value="modelValue"
    :title="title"
    @update:model-value="emit('update:modelValue', $event)"
    @closed="onClosed"
  >
    <u-form :key="sessionKey" ref="form" :model="model" :label-width="labelWidth" :cols="cols">
      <slot />
    </u-form>

    <template #footer="{ close: dialogClose }">
      <slot name="footer" :close="dialogClose" :submit="onConfirm">
        <u-button text @click="dialogClose()">{{ cancelText }}</u-button>
        <u-button type="primary" :loading="confirmLoading" @click="onConfirm">
          {{ confirmText }}
        </u-button>
      </slot>
    </template>
  </u-dialog>
</template>

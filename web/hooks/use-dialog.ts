import { ref } from 'vue'

type DialogType = 'create' | 'edit'

export function useDialog() {
  const visible = ref(false)
  const dialogType = ref<DialogType>('create')

  function open(
    type: DialogType,
    ctx?: {
      data
    }
  ) {
    dialogType.value = type
    visible.value = true
  }

  function close() {
    visible.value = false
  }

  return { open, close, dialogType, visible }
}

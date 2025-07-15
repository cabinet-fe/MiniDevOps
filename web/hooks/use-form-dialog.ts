import { ref } from 'vue'
import { FormModel, type FormModelItem } from 'ultra-ui'
type DialogType = 'create' | 'edit'

export function useFormDialog<T extends Record<string, FormModelItem<any>>>(
  fields: T
) {
  const visible = ref(false)
  const dialogType = ref<DialogType>('create')
  const model = new FormModel(fields)

  function open(
    type: DialogType,
    ctx?: {
      data?: Record<string, any>
    }
  ) {
    dialogType.value = type
    visible.value = true

    if (ctx) {
      const { data } = ctx
      // @ts-ignore
      data && model.setData(data)
    }
  }

  function close() {
    visible.value = false
    model.resetData()
  }

  return { open, close, dialogType, visible, model }
}

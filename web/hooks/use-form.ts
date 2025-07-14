import { FormModel } from 'ultra-ui'

export function useForm<T extends object>(initialData: T) {
  const model = new FormModel(initialData)

  function create() {
    model.reset(initialData)
  }

  function update(data: T) {
    model.reset(data)
  }

  return { model, create, update }
}

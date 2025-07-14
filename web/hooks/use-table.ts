import { ref } from 'vue'

export function useTable() {
  const tableRef = ref()

  function reload() {
    tableRef.value.reload()
  }

  return { tableRef, reload }
}

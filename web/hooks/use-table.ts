import type { TableColumn } from 'ultra-ui'

export type PageParams = {
  page: number
  pageSize: number
} & Record<string, any>

interface TableOptions {
  columns: TableColumn[]
  getData: (params?: PageParams) => Promise<{
    data: Record<string, any>[]
    total: number
  }>
}

export type PageTableCtx = TableOptions

export function useTable(options: TableOptions): PageTableCtx {
  const { columns, getData } = options

  return { columns, getData }
}

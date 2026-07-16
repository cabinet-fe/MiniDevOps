import { defineTableColumns, type TableColumn } from "@veltra/desktop";

export type ProTableQuery = Record<string, unknown>;

/** Column config; `sortable` adds a header sort control (UTable has no built-in sort). */
export type ProTableColumn = TableColumn & {
  sortable?: boolean;
  children?: ProTableColumn[];
};

export function defineProTableColumns(
  columns: ProTableColumn[],
  commonProps?: Partial<Pick<TableColumn, "align" | "minWidth">>,
): ProTableColumn[] {
  return defineTableColumns(columns as TableColumn[], commonProps) as ProTableColumn[];
}

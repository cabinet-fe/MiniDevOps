import { date } from "@cat-kit/core";

const DATETIME_FORMAT = "yyyy-MM-dd HH:mm:ss";

/** Format a timestamp for table/list display; empty/invalid → "". */
export function formatDateTime(value: string | number | Date | null | undefined): string {
  if (value == null || value === "") return "";
  const d = date(value);
  if (Number.isNaN(d.timestamp)) return "";
  return d.format(DATETIME_FORMAT);
}

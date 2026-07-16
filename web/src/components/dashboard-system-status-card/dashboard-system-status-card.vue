<script setup lang="ts">
defineOptions({ name: "DashboardSystemStatusCard" });

import { computed } from "vue";
import type { ColorType } from "@veltra/utils";
import { Monitor, Folder } from "@veltra/icons/normal";

import type { DiskStatus, SystemStatus } from "@/api/types";
import { formatDateTime } from "@/lib/datetime";

const props = defineProps<{
  data: SystemStatus | null;
}>();

const HEALTH_META: Record<string, { label: string; type: ColorType }> = {
  ok: { label: "正常", type: "success" },
  degraded: { label: "降级", type: "warning" },
};

const healthMeta = computed(() => {
  const key = props.data?.health ?? "";
  return HEALTH_META[key] ?? { label: key || "—", type: "info" as ColorType };
});

const cpuPercent = computed(() => props.data?.cpu_usage_percent ?? 0);
const memPercent = computed(() => props.data?.memory_usage_percent ?? 0);

function loadType(percentage: number): ColorType {
  if (percentage >= 90) return "danger";
  if (percentage >= 70) return "warning";
  return "primary";
}

function formatBytes(value: number | undefined): string {
  if (value == null || !Number.isFinite(value)) return "—";
  if (value === 0) return "0 B";
  const units = ["B", "KB", "MB", "GB", "TB"];
  const index = Math.min(Math.floor(Math.log(value) / Math.log(1024)), units.length - 1);
  return `${(value / 1024 ** index).toFixed(index ? 1 : 0)} ${units[index]}`;
}

function diskFreeLabel(dir: DiskStatus): string {
  return `${formatBytes(dir.free_bytes)} 可用 / ${formatBytes(dir.total_bytes)}`;
}

function shortPath(path: string): string {
  if (path.length <= 28) return path;
  return `…${path.slice(-26)}`;
}
</script>

<template>
  <u-card class="tile">
    <u-card-header class="tile__header">
      <div class="tile__title-row">
        <span class="tile__icon" aria-hidden="true">
          <u-icon :size="18" color="primary"><Monitor /></u-icon>
        </span>
        <div class="tile__titles">
          <h3 class="tile__title">系统状态</h3>
          <p class="tile__subtitle">
            {{
              data?.collected_at ? `采集于 ${formatDateTime(data.collected_at)}` : "实时资源占用"
            }}
          </p>
        </div>
        <u-tag class="tile__health" size="small" dark :type="healthMeta.type">
          {{ healthMeta.label }}
        </u-tag>
      </div>
    </u-card-header>

    <u-card-content class="tile__body">
      <div class="gauges">
        <div class="gauge">
          <u-progress circle :size="112" :percentage="cpuPercent" :type="loadType">
            <template #default="{ percentage }">
              <span class="gauge__pct">{{ percentage.toFixed(0) }}%</span>
            </template>
          </u-progress>
          <span class="gauge__label">CPU</span>
        </div>
        <div class="gauge">
          <u-progress circle :size="112" :percentage="memPercent" :type="loadType">
            <template #default="{ percentage }">
              <span class="gauge__pct">{{ percentage.toFixed(0) }}%</span>
            </template>
          </u-progress>
          <span class="gauge__label">内存</span>
          <span class="gauge__hint">
            {{ formatBytes(data?.memory_used_bytes) }} / {{ formatBytes(data?.memory_total_bytes) }}
          </span>
        </div>
      </div>

      <div class="disks">
        <div class="disks__head">
          <u-icon :size="14"><Folder /></u-icon>
          <span>磁盘占用</span>
        </div>
        <ul v-if="data?.directories?.length" class="disks__list">
          <li v-for="dir in data.directories" :key="dir.path" class="disk">
            <div class="disk__meta">
              <span class="disk__path" :title="dir.path">{{ shortPath(dir.path) }}</span>
              <span class="disk__pct">{{ dir.used_percent.toFixed(1) }}%</span>
            </div>
            <u-progress :percentage="dir.used_percent" :type="loadType" />
            <p class="disk__free">{{ diskFreeLabel(dir) }}</p>
          </li>
        </ul>
        <p v-else class="disks__empty">暂无磁盘采样</p>
      </div>
    </u-card-content>
  </u-card>
</template>

<style scoped lang="scss">
@use "pkg:@veltra/styles/functions" as fn;

.tile {
  height: 100%;
  min-height: 320px;
  display: flex;
  flex-direction: column;
  background: color-mix(in srgb, fn.use-var(bg-color, top) 88%, fn.use-var(color, primary) 4%);
}

.tile__header {
  padding-bottom: 0;
}

.tile__title-row {
  display: flex;
  align-items: flex-start;
  gap: 12px;
}

.tile__icon {
  flex-shrink: 0;
  display: grid;
  place-items: center;
  width: 36px;
  height: 36px;
  border-radius: fn.use-var(radius, default);
  background: color-mix(in srgb, fn.use-var(color, primary) 22%, transparent);
}

.tile__titles {
  flex: 1;
  min-width: 0;
}

.tile__title {
  margin: 0;
  color: fn.use-var(text-color, title);
  font-size: 16px;
  font-weight: 600;
  letter-spacing: 0.02em;
}

.tile__subtitle {
  margin: 4px 0 0;
  color: fn.use-var(text-color, assist);
  font-size: 12px;
}

.tile__health {
  flex-shrink: 0;
  margin-top: 4px;
}

.tile__body {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 22px;
}

.gauges {
  display: flex;
  justify-content: space-around;
  gap: 16px;
  padding: 8px 0 4px;
}

.gauge {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.gauge__pct {
  color: fn.use-var(text-color, title);
  font-size: 22px;
  font-weight: 650;
  font-variant-numeric: tabular-nums;
  letter-spacing: -0.02em;
}

.gauge__label {
  color: fn.use-var(text-color, second);
  font-size: 13px;
  font-weight: 600;
  letter-spacing: 0.06em;
}

.gauge__hint {
  color: fn.use-var(text-color, assist);
  font-size: 11px;
  font-variant-numeric: tabular-nums;
  text-align: center;
}

.disks {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.disks__head {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: fn.use-var(text-color, second);
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.disks__list {
  margin: 0;
  padding: 0;
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.disk__meta {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 6px;
}

.disk__path {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: fn.use-var(text-color, title);
  font-size: 13px;
  font-weight: 550;
}

.disk__pct {
  flex-shrink: 0;
  color: fn.use-var(text-color, second);
  font-size: 13px;
  font-variant-numeric: tabular-nums;
  font-weight: 600;
}

.disk__free {
  margin: 6px 0 0;
  color: fn.use-var(text-color, assist);
  font-size: 12px;
  font-variant-numeric: tabular-nums;
}

.disks__empty {
  margin: 0;
  color: fn.use-var(text-color, assist);
  font-size: 13px;
}
</style>

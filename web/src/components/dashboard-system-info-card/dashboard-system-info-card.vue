<script setup lang="ts">
defineOptions({ name: "DashboardSystemInfoCard" });

import { computed } from "vue";
import { Server, Time, Internet, Variable } from "@veltra/icons/normal";

import type { SystemInfo } from "@/api/types";
import { formatDateTime } from "@/lib/datetime";

const props = defineProps<{
  data: SystemInfo | null;
}>();

const platform = computed(() => {
  if (!props.data) return "—";
  return `${props.data.os} / ${props.data.arch}`;
});

const uptime = computed(() => {
  const start = props.data?.start_time;
  if (!start) return "—";
  const ms = Date.now() - new Date(start).getTime();
  if (!Number.isFinite(ms) || ms < 0) return "—";
  const totalHours = Math.floor(ms / 3_600_000);
  const days = Math.floor(totalHours / 24);
  const hours = totalHours % 24;
  const mins = Math.floor((ms % 3_600_000) / 60_000);
  if (days > 0) return `${days} 天 ${hours} 小时`;
  if (totalHours > 0) return `${totalHours} 小时 ${mins} 分`;
  return `${mins} 分钟`;
});
</script>

<template>
  <u-card class="tile">
    <u-card-header class="tile__header">
      <div class="tile__title-row">
        <span class="tile__icon" aria-hidden="true">
          <u-icon :size="18" color="primary"><Server /></u-icon>
        </span>
        <div class="tile__titles">
          <h3 class="tile__title">系统信息</h3>
          <p class="tile__subtitle">只读主机与运行时快照</p>
        </div>
      </div>
    </u-card-header>

    <u-card-content class="tile__body">
      <div class="hero">
        <p class="hero__label">版本</p>
        <p class="hero__version">{{ data?.version || "—" }}</p>
        <p class="hero__host">
          <u-icon :size="14"><Internet /></u-icon>
          <span>{{ data?.hostname || "—" }}</span>
        </p>
      </div>

      <div class="facts">
        <div class="fact">
          <span class="fact__icon" aria-hidden="true">
            <u-icon :size="14"><Server /></u-icon>
          </span>
          <span class="fact__label">平台</span>
          <span class="fact__value">{{ platform }}</span>
        </div>
        <div class="fact">
          <span class="fact__icon" aria-hidden="true">
            <u-icon :size="14"><Variable /></u-icon>
          </span>
          <span class="fact__label">运行时</span>
          <span class="fact__value">{{ data?.runtime || "—" }}</span>
        </div>
        <div class="fact">
          <span class="fact__icon" aria-hidden="true">
            <u-icon :size="14"><Time /></u-icon>
          </span>
          <span class="fact__label">已运行</span>
          <span class="fact__value">{{ uptime }}</span>
        </div>
        <div class="fact">
          <span class="fact__icon" aria-hidden="true">
            <u-icon :size="14"><Time /></u-icon>
          </span>
          <span class="fact__label">启动时间</span>
          <span class="fact__value">{{ formatDateTime(data?.start_time) || "—" }}</span>
        </div>
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

.tile__body {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.hero {
  padding: 18px 16px;
  border-radius: fn.use-var(radius, large);
  background: linear-gradient(
    145deg,
    color-mix(in srgb, fn.use-var(color, primary) 18%, transparent),
    color-mix(in srgb, fn.use-var(bg-color, bottom) 80%, transparent) 55%
  );
}

.hero__label {
  margin: 0;
  color: fn.use-var(text-color, assist);
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.hero__version {
  margin: 8px 0 0;
  color: fn.use-var(text-color, title);
  font-size: clamp(28px, 3.4vw, 40px);
  font-weight: 650;
  letter-spacing: -0.03em;
  line-height: 1.1;
  overflow-wrap: anywhere;
}

.hero__host {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  margin: 12px 0 0;
  color: fn.use-var(text-color, second);
  font-size: 13px;
}

.facts {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.fact {
  display: flex;
  flex-direction: column;
  gap: 6px;
  min-width: 0;
  padding: 14px 12px;
  border-radius: fn.use-var(radius, default);
  background: color-mix(in srgb, fn.use-var(bg-color, bottom) 70%, transparent);
}

.fact__icon {
  color: fn.use-var(text-color, assist);
}

.fact__label {
  color: fn.use-var(text-color, second);
  font-size: 12px;
  letter-spacing: 0.04em;
}

.fact__value {
  color: fn.use-var(text-color, title);
  font-size: 14px;
  font-weight: 550;
  line-height: 1.35;
  overflow-wrap: anywhere;
}
</style>

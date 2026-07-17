<script setup lang="ts">
defineOptions({ name: "AiRunDetail" });

import { onMounted, onUnmounted, ref } from "vue";
import { useRoute } from "vue-router";
import { message } from "@veltra/desktop";

import { getRun } from "@/api/ai";
import { getAccessToken } from "@/api/http";
import type { AgentRun } from "@/api/types";

const route = useRoute();
const run = ref<AgentRun | null>(null);
const lines = ref<string[]>([]);
let socket: WebSocket | null = null;

async function load() {
  const id = Number(route.params.id);
  run.value = await getRun(id);
  connectWS(id);
}

function connectWS(id: number) {
  const token = getAccessToken();
  if (!token) return;
  const proto = location.protocol === "https:" ? "wss" : "ws";
  socket = new WebSocket(
    `${proto}://${location.host}/ws/ai/runs/${id}/logs?token=${encodeURIComponent(token)}`,
  );
  socket.onmessage = (ev) => {
    const text = String(ev.data);
    if (text.startsWith("__TERMINAL__:")) {
      void load();
      return;
    }
    lines.value.push(text);
  };
  socket.onerror = () => message.error("日志 WS 连接失败");
}

onMounted(() => {
  void load().catch((error: unknown) => {
    message.error(error instanceof Error ? error.message : "加载失败");
  });
});

onUnmounted(() => {
  socket?.close();
});
</script>

<template>
  <div v-if="run">
    <p>
      Agent {{ run.agent_id }} · {{ run.trigger_type }} ·
      <strong>{{ run.status }}</strong>
    </p>
    <p v-if="run.error_message" class="error">{{ run.error_message }}</p>
    <pre class="log">{{ lines.join("\n") || run.output_text || "等待日志…" }}</pre>
  </div>
</template>

<style scoped lang="scss">
.error {
  color: var(--u-color-danger, #c00);
}
.log {
  margin: 0;
  padding: 12px;
  background: #0f172a;
  color: #e2e8f0;
  border-radius: 8px;
  min-height: 320px;
  overflow: auto;
  font-size: 12px;
  line-height: 1.5;
}
</style>

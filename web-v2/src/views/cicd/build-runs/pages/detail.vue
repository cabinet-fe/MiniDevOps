<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { message } from "@veltra/desktop";

import {
  buildRunArtifactURL,
  cancelBuildRun,
  getBuildRun,
  redeployBuildRun,
  retryBuildRun,
} from "@/api/cicd";
import { getAccessToken } from "@/api/http";
import type { BuildRun } from "@/api/types";
import BuildLogViewer, { resolveBuildLogStatus } from "@/components/build-log-viewer";
import { usePermission } from "@/composables/use-permission";

const route = useRoute();
const router = useRouter();
const { hasPermission } = usePermission();

const run = ref<BuildRun | null>(null);
const loading = ref(true);
const acting = ref(false);
const logViewerRef = ref<InstanceType<typeof BuildLogViewer> | null>(null);

let pollTimer: ReturnType<typeof setInterval> | null = null;

const canExecute = computed(() => hasPermission("cicd.build_jobs:execute"));
const runId = computed(() => Number(route.params.id));

const isLive = computed(() => {
  const s = run.value?.status;
  return s === "queued" || s === "running" || run.value?.distribution_summary === "running";
});

const logViewerStatus = computed(() =>
  resolveBuildLogStatus(run.value?.status, run.value?.distribution_summary),
);

const canCancel = computed(() => {
  if (!canExecute.value || !run.value) return false;
  return (
    run.value.status === "queued" ||
    run.value.status === "running" ||
    (run.value.status === "success" && run.value.distribution_summary === "running")
  );
});

const canRetry = computed(
  () =>
    canExecute.value &&
    !!run.value &&
    ["failed", "cancelled", "interrupted", "success"].includes(run.value.status),
);

const canRedeploy = computed(
  () => canExecute.value && run.value?.status === "success" && !!run.value.artifact_path,
);

async function load() {
  const id = runId.value;
  if (!id) {
    message.error("无效 ID");
    return;
  }
  try {
    run.value = await getBuildRun(id);
  } catch (err) {
    message.error(err instanceof Error ? err.message : "加载失败");
  } finally {
    loading.value = false;
  }
}

function startPolling() {
  stopPolling();
  pollTimer = setInterval(async () => {
    try {
      run.value = await getBuildRun(runId.value);
      if (!isLive.value) stopPolling();
    } catch {
      /* ignore */
    }
  }, 2500);
}

function stopPolling() {
  if (pollTimer) {
    clearInterval(pollTimer);
    pollTimer = null;
  }
}

async function onCancel() {
  if (!run.value || acting.value) return;
  acting.value = true;
  try {
    run.value = await cancelBuildRun(run.value.id);
    message.success("已取消");
  } catch (err) {
    message.error(err instanceof Error ? err.message : "取消失败");
  } finally {
    acting.value = false;
  }
}

async function onRetry() {
  if (!run.value || acting.value) return;
  acting.value = true;
  try {
    const next = await retryBuildRun(run.value.id);
    message.success(`已创建重试 #${next.build_number}`);
    await router.push({ name: "cicd-build-run-detail", params: { id: String(next.id) } });
  } catch (err) {
    message.error(err instanceof Error ? err.message : "重试失败");
  } finally {
    acting.value = false;
  }
}

async function onRedeploy() {
  if (!run.value || acting.value) return;
  acting.value = true;
  try {
    run.value = await redeployBuildRun(run.value.id);
    message.success("已开始重新分发");
    logViewerRef.value?.appendLine("=== Redeploy requested ===");
    logViewerRef.value?.reconnect();
    startPolling();
  } catch (err) {
    message.error(err instanceof Error ? err.message : "重新分发失败");
  } finally {
    acting.value = false;
  }
}

async function onDownloadArtifact() {
  const token = getAccessToken();
  if (!token || !run.value) return;
  try {
    const res = await fetch(buildRunArtifactURL(run.value.id), {
      headers: { Authorization: `Bearer ${token}` },
    });
    if (!res.ok) {
      throw new Error((await res.text()) || `HTTP ${res.status}`);
    }
    const blob = await res.blob();
    const cd = res.headers.get("Content-Disposition") || "";
    const m = /filename="?([^"]+)"?/.exec(cd);
    const name = m?.[1] || `build-${run.value.build_number}.tar.gz`;
    const a = document.createElement("a");
    a.href = URL.createObjectURL(blob);
    a.download = name;
    a.click();
    URL.revokeObjectURL(a.href);
  } catch (err) {
    message.error(err instanceof Error ? err.message : "下载失败");
  }
}

watch(runId, async () => {
  loading.value = true;
  stopPolling();
  await load();
  if (isLive.value) startPolling();
});

onMounted(async () => {
  await load();
  if (isLive.value) startPolling();
});

onUnmounted(() => {
  stopPolling();
});
</script>

<template>
  <div class="page">
    <div class="page-head">
      <h2>执行详情 #{{ run?.build_number ?? route.params.id }}</h2>
      <div class="actions">
        <u-button v-if="canCancel" :disabled="acting" variant="outline" @click="onCancel">
          取消
        </u-button>
        <u-button v-if="canRetry" :disabled="acting" @click="onRetry">重试</u-button>
        <u-button v-if="canRedeploy" :disabled="acting" @click="onRedeploy">重新分发</u-button>
        <u-button v-if="run?.status === 'success'" variant="outline" @click="onDownloadArtifact">
          下载制品
        </u-button>
        <u-button @click="router.push({ name: 'cicd-build-runs' })">返回列表</u-button>
      </div>
    </div>

    <p class="risk-note">
      构建脚本与 Bedrock 以同一操作系统用户执行，无沙箱隔离；请仅让可信人员编辑脚本。
    </p>

    <div v-if="loading" class="state">加载中…</div>
    <template v-else-if="run">
      <dl class="meta">
        <div>
          <dt>状态</dt>
          <dd>{{ run.status }}</dd>
        </div>
        <div>
          <dt>阶段</dt>
          <dd>{{ run.stage }}</dd>
        </div>
        <div>
          <dt>分发汇总</dt>
          <dd>{{ run.distribution_summary }}</dd>
        </div>
        <div>
          <dt>任务 ID</dt>
          <dd>{{ run.build_job_id }}</dd>
        </div>
        <div>
          <dt>分支</dt>
          <dd>{{ run.branch || "—" }}</dd>
        </div>
        <div>
          <dt>Commit</dt>
          <dd>{{ run.commit_hash || "—" }}</dd>
        </div>
        <div>
          <dt>触发</dt>
          <dd>{{ run.trigger_type }}</dd>
        </div>
      </dl>

      <h3>构建日志</h3>
      <BuildLogViewer ref="logViewerRef" :run-id="runId" :live="isLive" :status="logViewerStatus" />

      <h3>部署尝试</h3>
      <u-empty v-if="!run.deploy_attempts?.length" text="暂无部署尝试" />
      <ul v-else class="attempts">
        <li v-for="a in run.deploy_attempts" :key="a.id">
          batch {{ a.batch_no }} · target {{ a.deploy_target_id ?? "—" }} · {{ a.status }}
          <span v-if="a.error_message"> — {{ a.error_message }}</span>
        </li>
      </ul>

      <h3>配置快照</h3>
      <pre class="snap">{{ run.snapshot_json || "（空）" }}</pre>
    </template>
  </div>
</template>

<style scoped>
.page {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.page-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
}
.page-head h2 {
  margin: 0;
  font-size: 18px;
}
.actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
.risk-note {
  margin: 0;
  font-size: 12px;
  opacity: 0.75;
}
.meta {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
  gap: 12px;
  margin: 0;
}
.meta div {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.meta dt {
  font-size: 12px;
  opacity: 0.7;
}
.meta dd {
  margin: 0;
  font-weight: 500;
}
.snap {
  margin: 0;
  padding: 12px;
  overflow: auto;
  max-height: 200px;
  font-size: 12px;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  background: color-mix(in srgb, currentColor 6%, transparent);
  border-radius: 8px;
  white-space: pre-wrap;
  word-break: break-all;
}
.attempts {
  margin: 0;
  padding-left: 20px;
}
.state {
  opacity: 0.7;
}
</style>

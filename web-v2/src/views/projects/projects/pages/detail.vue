<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { message } from "@veltra/desktop";
import { useRoute, useRouter } from "vue-router";

import { getProject, listProjectMembers } from "@/api/projects";
import type { ProductProject, ProjectMember, ProjectRole } from "@/api/types";
import { usePermission } from "@/composables/use-permission";
import { useAuthStore } from "@/stores/auth";

import DocsPanel from "../../components/docs-panel.vue";
import MembersPanel from "../../components/members-panel.vue";
import RequirementsPanel from "../../components/requirements-panel.vue";

const route = useRoute();
const router = useRouter();
const { hasPermission } = usePermission();
const auth = useAuthStore();
const project = ref<ProductProject | null>(null);
const members = ref<ProjectMember[]>([]);
const tab = ref("requirements");

const projectID = computed(() => Number(route.params.id));
const projectRole = computed<ProjectRole | undefined>(
  () => members.value.find((member) => member.user_id === auth.user?.id)?.role,
);
const canManageAll = computed(() => hasPermission("project.projects:manage_all"));
const tabs = computed(
  () =>
    [
      hasPermission("project.requirements:view") ? { key: "requirements", name: "需求" } : null,
      hasPermission("project.docs:view") ? { key: "docs", name: "接口文档" } : null,
      { key: "members", name: "成员" },
    ].filter(Boolean) as { key: string; name: string }[],
);

function resolveTab(preferred?: unknown): string {
  const key = typeof preferred === "string" ? preferred : "";
  if (key && tabs.value.some((item) => item.key === key)) return key;
  return tabs.value[0]?.key ?? "members";
}

async function loadMembership() {
  if (!project.value) {
    members.value = [];
    return;
  }
  try {
    members.value = await listProjectMembers(project.value.id);
  } catch (error) {
    members.value = [];
    message.error(error instanceof Error ? error.message : "成员权限加载失败");
  }
}

async function load() {
  if (!Number.isSafeInteger(projectID.value) || projectID.value <= 0) {
    project.value = null;
    members.value = [];
    return;
  }
  try {
    project.value = await getProject(projectID.value);
    await loadMembership();
    tab.value = resolveTab(route.query.tab);
  } catch (error) {
    project.value = null;
    message.error(error instanceof Error ? error.message : "读取项目失败");
  }
}

watch(projectID, () => void load(), { immediate: true });

watch(
  () => route.query.tab,
  (next) => {
    const resolved = resolveTab(next);
    if (tab.value !== resolved) tab.value = resolved;
  },
);

watch(tab, (next) => {
  if (route.query.tab === next) return;
  void router.replace({ query: { ...route.query, tab: next } });
});
</script>

<template>
  <div class="page">
    <div class="page-head">
      <div>
        <u-button text @click="router.push({ name: 'projects' })">返回项目列表</u-button>
        <h2 v-if="project">{{ project.name }}</h2>
        <p v-if="project">
          <span class="slug">{{ project.slug }}</span>
          <u-tag size="small" :type="project.status === 'archived' ? 'warning' : 'success'">
            {{ project.status === "archived" ? "已归档" : "活跃" }}
          </u-tag>
        </p>
      </div>
      <p v-if="project?.description" class="description">{{ project.description }}</p>
    </div>

    <template v-if="project">
      <u-tabs v-model="tab" :items="tabs" />
      <RequirementsPanel
        v-if="tab === 'requirements' && hasPermission('project.requirements:view')"
        :project="project"
        :project-role="projectRole"
        :manage-all="canManageAll"
      />
      <DocsPanel
        v-else-if="tab === 'docs' && hasPermission('project.docs:view')"
        :project="project"
        :project-role="projectRole"
        :manage-all="canManageAll"
      />
      <MembersPanel
        v-else-if="tab === 'members'"
        :project="project"
        @members-changed="loadMembership"
        @owner-transferred="load"
      />
    </template>
    <u-empty v-else text="项目不存在或无权访问" />
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
  align-items: flex-start;
  justify-content: space-between;
  gap: 24px;
}
.page-head h2 {
  margin: 6px 0;
}
.page-head p {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0;
}
.slug {
  color: var(--u-text-color-assist, #7c8494);
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
}
.description {
  max-width: 48%;
  color: var(--u-text-color-second, #626b7d);
  line-height: 1.6;
}
</style>

import { createRouter, createWebHistory } from "vue-router";

import { getAccessToken } from "@/api/http";
import { useAuthStore } from "@/stores/auth";

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: "/login",
      name: "login",
      component: () => import("@/pages/login/login.vue"),
      meta: { public: true },
    },
    {
      path: "/",
      component: () => import("@/pages/layout.vue"),
      children: [
        {
          path: "",
          name: "home",
          component: () => import("@/pages/home.vue"),
        },
        {
          path: "system/users",
          name: "system-users",
          component: () => import("@/views/system/users/pages/main.vue"),
          meta: { permission: "system.users:view" },
        },
        {
          path: "system/roles",
          name: "system-roles",
          component: () => import("@/views/system/roles/pages/main.vue"),
          meta: { permission: "system.roles:view" },
        },
        {
          path: "system/resources",
          name: "system-resources",
          component: () => import("@/views/system/resources/pages/main.vue"),
          meta: { permission: "system.resources:view" },
        },
        {
          path: "system/dictionaries",
          name: "system-dictionaries",
          component: () => import("@/views/system/dictionaries/pages/main.vue"),
          meta: { permission: "system.dictionaries:view" },
        },
        {
          path: "system/operation-logs",
          name: "system-operation-logs",
          component: () => import("@/views/system/operation-logs/pages/main.vue"),
          meta: { permission: "system.operation_logs:view" },
        },
        {
          path: "cicd/repositories",
          name: "cicd-repositories",
          component: () => import("@/views/cicd/repositories/pages/main.vue"),
          meta: { permission: "cicd.repositories:view" },
        },
        {
          path: "cicd/build-jobs",
          name: "cicd-build-jobs",
          component: () => import("@/views/cicd/build-jobs/pages/main.vue"),
          meta: { permission: "cicd.build_jobs:view" },
        },
        {
          path: "cicd/build-runs",
          name: "cicd-build-runs",
          component: () => import("@/views/cicd/build-runs/pages/main.vue"),
          meta: { permission: "cicd.build_runs:view" },
        },
        {
          path: "cicd/build-runs/:id",
          name: "cicd-build-run-detail",
          component: () => import("@/views/cicd/build-runs/pages/detail.vue"),
          meta: { permission: "cicd.build_runs:view" },
        },
        {
          path: "cicd/servers",
          name: "cicd-servers",
          component: () => import("@/views/cicd/servers/pages/main.vue"),
          meta: { permission: "cicd.servers:view" },
        },
        {
          path: "cicd/credentials",
          name: "cicd-credentials",
          component: () => import("@/views/cicd/credentials/pages/main.vue"),
          meta: { permission: "cicd.credentials:view" },
        },
        {
          path: "project",
          redirect: "/project/projects",
        },
        {
          path: "project/projects",
          name: "projects",
          component: () => import("@/views/projects/projects/pages/main.vue"),
          meta: { permission: "project.projects:view" },
        },
        {
          path: "project/projects/:id",
          name: "project-detail",
          component: () => import("@/views/projects/projects/pages/detail.vue"),
          meta: { permission: "project.projects:view" },
        },
        {
          // Requirements/docs are scoped to a project; menu entries pick a project first.
          path: "project/requirements",
          name: "project-requirements",
          component: () => import("@/views/projects/requirements/pages/main.vue"),
          meta: { permission: "project.requirements:view", projectTab: "requirements" },
        },
        {
          path: "project/docs",
          name: "project-docs",
          component: () => import("@/views/projects/docs/pages/main.vue"),
          meta: { permission: "project.docs:view", projectTab: "docs" },
        },
        {
          path: "projects",
          redirect: "/project/projects",
        },
        {
          path: "projects/:id",
          redirect: (to) => `/project/projects/${String(to.params.id)}`,
        },
        {
          path: "cicd/:rest(.*)*",
          name: "cicd-placeholder",
          component: () => import("@/pages/placeholder.vue"),
        },
        {
          path: "ops/processes",
          name: "ops-processes",
          component: () => import("@/views/ops/processes/pages/main.vue"),
          meta: { permission: "ops.processes:view" },
        },
        {
          path: "ops/dev-environments",
          name: "ops-dev-environments",
          component: () => import("@/views/ops/dev-environments/pages/main.vue"),
          meta: { permission: "ops.dev_environments:view" },
        },
        {
          path: "ai/clis",
          name: "ai-clis",
          component: () => import("@/views/ai/clis/pages/main.vue"),
          meta: { permission: "ai.clis:view" },
        },
        {
          path: "ai/agents",
          name: "ai-agents",
          component: () => import("@/views/ai/agents/pages/main.vue"),
          meta: { permission: "ai.agents:view" },
        },
        {
          path: "ai/runs",
          name: "ai-runs",
          component: () => import("@/views/ai/runs/pages/main.vue"),
          meta: { permission: "ai.agents:view" },
        },
        {
          path: "ai/runs/:id",
          name: "ai-run-detail",
          component: () => import("@/views/ai/runs/pages/detail.vue"),
          meta: { permission: "ai.agents:view" },
        },
        {
          path: "ai/skills",
          name: "ai-skills",
          component: () => import("@/views/ai/skills/pages/main.vue"),
          meta: { permission: "ai.skills:view" },
        },
        {
          path: "ai/tokens",
          name: "ai-tokens",
          component: () => import("@/views/ai/tokens/pages/main.vue"),
        },
      ],
    },
  ],
});

router.beforeEach(async (to) => {
  const isPublic = to.meta.public === true;
  const hasToken = !!getAccessToken();

  if (!isPublic && !hasToken) {
    return { name: "login", query: { redirect: to.fullPath } };
  }
  if (to.name === "login" && hasToken) {
    return { name: "home" };
  }

  if (!isPublic && hasToken) {
    const auth = useAuthStore();
    await auth.refreshMe();
    if (!auth.user) {
      return { name: "login", query: { redirect: to.fullPath } };
    }
    const needed = typeof to.meta.permission === "string" ? to.meta.permission : "";
    if (needed && !auth.hasPermission(needed)) {
      return { name: "home" };
    }
  }
  return true;
});

export default router;

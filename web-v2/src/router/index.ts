import { createRouter, createWebHistory } from "vue-router";

import { getAccessToken } from "@/api/http";
import { useAuthStore } from "@/stores/auth";

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: "/login",
      name: "login",
      component: () => import("@/views/login-view.vue"),
      meta: { public: true },
    },
    {
      path: "/",
      component: () => import("@/layouts/app-layout.vue"),
      children: [
        {
          path: "",
          name: "home",
          component: () => import("@/views/home-view.vue"),
        },
        {
          path: "system/users",
          name: "system-users",
          component: () => import("@/views/system/users-view.vue"),
          meta: { permission: "system.users:view" },
        },
        {
          path: "system/roles",
          name: "system-roles",
          component: () => import("@/views/system/roles-view.vue"),
          meta: { permission: "system.roles:view" },
        },
        {
          path: "system/resources",
          name: "system-resources",
          component: () => import("@/views/system/resources-view.vue"),
          meta: { permission: "system.resources:view" },
        },
        {
          path: "system/menus",
          name: "system-menus",
          component: () => import("@/views/system/menus-view.vue"),
          meta: { permission: "system.menus:view" },
        },
        {
          path: "system/dictionaries",
          name: "system-dictionaries",
          component: () => import("@/views/system/dictionaries-view.vue"),
          meta: { permission: "system.dictionaries:view" },
        },
        {
          path: "system/operation-logs",
          name: "system-operation-logs",
          component: () => import("@/views/system/operation-logs-view.vue"),
          meta: { permission: "system.operation_logs:view" },
        },
        {
          path: "cicd/repositories",
          name: "cicd-repositories",
          component: () => import("@/views/cicd/repositories-view.vue"),
          meta: { permission: "cicd.repositories:view" },
        },
        {
          path: "cicd/build-jobs",
          name: "cicd-build-jobs",
          component: () => import("@/views/cicd/build-jobs-view.vue"),
          meta: { permission: "cicd.build_jobs:view" },
        },
        {
          path: "cicd/build-runs",
          name: "cicd-build-runs",
          component: () => import("@/views/cicd/build-runs-view.vue"),
          meta: { permission: "cicd.build_runs:view" },
        },
        {
          path: "cicd/build-runs/:id",
          name: "cicd-build-run-detail",
          component: () => import("@/views/cicd/build-run-detail-view.vue"),
          meta: { permission: "cicd.build_runs:view" },
        },
        {
          path: "cicd/servers",
          name: "cicd-servers",
          component: () => import("@/views/cicd/servers-view.vue"),
          meta: { permission: "cicd.servers:view" },
        },
        {
          path: "cicd/credentials",
          name: "cicd-credentials",
          component: () => import("@/views/cicd/credentials-view.vue"),
          meta: { permission: "cicd.credentials:view" },
        },
        {
          path: "cicd/:rest(.*)*",
          name: "cicd-placeholder",
          component: () => import("@/views/placeholder-view.vue"),
        },
        {
          path: "ops/:rest(.*)*",
          name: "ops-placeholder",
          component: () => import("@/views/placeholder-view.vue"),
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

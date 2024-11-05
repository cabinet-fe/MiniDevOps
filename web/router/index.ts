import { createRouter, createWebHistory, type RouteComponent } from 'vue-router'

export const router = createRouter({
  routes: [
    {
      path: '/registries',
      alias: '/',
      component: () => import('@/views/registries/index.vue')
    }
  ],
  history: createWebHistory('/')
})

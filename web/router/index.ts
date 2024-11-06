import {
  createRouter,
  createWebHistory,
  type RouteRecordSingleView
} from 'vue-router'

const views = import.meta.glob<boolean, string>('../views/**/index.vue')

const routes = ['repos', 'tasks', 'remotes'].map((name, index) => {
  return {
    path: `/${name}`,
    component: views[`../views/${name}/index.vue`],
    meta: { index }
  }
}) as RouteRecordSingleView[]

export const router = createRouter({
  routes: [
    {
      path: '/',
      redirect: '/repos'
    },
    ...routes
  ],
  history: createWebHistory('/')
})

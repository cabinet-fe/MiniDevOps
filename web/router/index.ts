import {
  createRouter,
  createWebHistory,
  type RouteLocationNormalized,
  type RouteRecordSingleView
} from 'vue-router'
import Login from '@/pages/login.vue'
import { session, TOKEN } from '@/utils/cache'
import Layout from '@/pages/layout.vue'
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
      path: '/login',
      component: Login
    },

    {
      component: Layout,
      path: '/',
      redirect: '/repos',
      children: routes
    }
  ],
  history: createWebHistory('/')
})

const whiteList = new Set(['/login'])

function access(to: RouteLocationNormalized) {
  if (whiteList.has(to.path)) return true
  if (session.get(TOKEN)) return true
  return false
}

router.beforeEach((to, _, next) => {
  if (access(to)) return next()
  next('/login')
})

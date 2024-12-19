import {
  createRouter,
  createWebHistory,
  type RouteRecordSingleView
} from 'vue-router'
import Login from '@/pages/login.vue'
import { WebCache } from 'cat-kit/fe'
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
      children: routes
    }
  ],
  history: createWebHistory('/')
})

function hasLogin() {
  return WebCache.create('session').get('token')
}

router.beforeEach((to, from, next) => {
  if (to.path === '/login') {
    next()
  } else {
    if (hasLogin()) {
      next()
    } else {
      next('/login')
    }
  }
})

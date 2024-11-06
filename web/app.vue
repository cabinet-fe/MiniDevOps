<template>
  <div
    class="flex justify-between h-full bg-gray-100 p-8 gap-8 overflow-hidden"
  >
    <aside
      class="w-[240px] shrink-0 h-full bg-white rounded-3xl shadow-lg p-4 z-2"
    >
      <div
        class="h-[40px] font-size-[16px] flex items-center hover:bg-gray-200 rounded-lg transition-all duration-300 cursor-pointer px-2"
        v-for="(menu, i) in menus"
        :class="{
          'bg-gray-200': router.currentRoute.value.path === menu.path,
          'mt-1': i !== 0
        }"
        @click="router.push(menu.path)"
      >
        {{ menu.name }}
      </div>
    </aside>
    <main class="grow-1 h-full">
      <router-view v-slot="{ Component }">
        <transition :name="transitionName" mode="out-in">
          <component :is="Component" style="height: 100%" />
        </transition>
      </router-view>
    </main>
  </div>
</template>

<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router'
import { watch, shallowRef } from 'vue'

const router = useRouter()
const route = useRoute()

const menus = [
  { path: '/repos', name: '仓库' },
  { path: '/tasks', name: '任务' },
  { path: '/remotes', name: '远程目录' }
]

let transitionName = shallowRef('fade')

watch(
  () => route.meta.index as number,
  (i, oi) => {
    if (oi === undefined || oi === null) return 'fade'
    if (i > oi) {
      transitionName.value = 'to-left'
    } else {
      transitionName.value = 'to-right'
    }
  }
)
</script>

<style>
.to-left-enter-active,
.to-left-leave-active,
.to-right-enter-active,
.to-right-leave-active {
  transition: all 0.15s ease-out;
}

.to-left-enter-active,
.to-right-enter-active {
  transition-delay: 0.2s;
}

.to-left-enter-from,
.to-right-leave-to {
  opacity: 0;
  transform: translate3d(80px, 0, 0);
}

.to-left-leave-to,
.to-right-enter-from {
  opacity: 0;
  transform: translate3d(-80px, 0, 0);
}

/* .m-table-pro__tool {
  border-radius: 1.5rem !important;
  overflow: hidden;
}

.u-table {
  border-radius: 1.5rem 1.5rem 0 0 !important;
} */
</style>

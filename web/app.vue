<template>
  <div class="flex justify-between h-full bg-gray-100 p-8 gap-8">
    <aside class="w-[240px] shrink-0 h-full bg-white rounded-3xl shadow-lg">
      <div class="h-[50px]"></div>
    </aside>
    <main class="grow-1 h-full">
      <router-view />
    </main>
  </div>
</template>

<script setup lang="ts"></script>

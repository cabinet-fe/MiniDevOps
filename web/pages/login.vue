<template>
  <div
    class="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-50 to-indigo-100"
  >
    <div class="w-full max-w-md p-8 bg-white rounded-2xl shadow-lg">
      <div class="text-center mb-8">
        <h1 class="text-3xl font-bold text-gray-800">欢迎登录</h1>
        <p class="text-gray-600 mt-2">请输入您的账号和密码</p>
      </div>

      <u-form :model="model" size="large">
        <u-input placeholder="请输入账号" field="name" />
        <u-input placeholder="请输入密码" field="password" />
      </u-form>

      <u-button
        type="primary"
        :loading="loading"
        class="w-full"
        @click="handleLogin"
      >
        登录
      </u-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { FormModel } from 'ultra-ui'
import { ref } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()
const loading = ref(false)

const model = new FormModel({
  name: {},
  password: {}
})

async function handleLogin() {
  loading.value = true

  try {
    router.push('/repos')
  } catch (err) {}

  loading.value = false
}
</script>

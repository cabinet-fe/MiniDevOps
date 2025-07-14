<template>
  <div
    class="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-50 to-indigo-100"
  >
    <div
      class="w-full max-w-md p-8 bg-white rounded-2xl shadow-lg transform transition-all hover:shadow-2xl"
    >
      <div class="text-center mb-8">
        <h1 class="text-3xl font-bold text-gray-800">欢迎登录</h1>
        <p class="text-gray-600 mt-2">Mini DevOps</p>
      </div>

      <u-form :model="model" size="large" class="space-y-4">
        <u-input placeholder="请输入账号" field="username" />
        <u-password-input
          placeholder="请输入密码"
          field="password"
          show-password
        />
      </u-form>

      <u-button
        type="primary"
        :loading="loading"
        class="w-full mt-8"
        size="large"
        @click="handleLogin"
      >
        登录
      </u-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { FormModel, message } from 'ultra-ui'
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { userService } from '@/apis/user'
import { session, TOKEN } from '@/utils/cache'

const router = useRouter()
const loading = ref(false)

const model = new FormModel({
  username: {
    value: 'admin',
    required: true
  },
  password: {
    value: 'admin123',
    required: true
  }
})

async function handleLogin() {
  const valid = await model.validate()
  if (!valid) return

  loading.value = true
  try {
    const res = await userService.login(model.data)
    if (res.data.token) {
      session.set(TOKEN, res.data.token)
      message.success('登录成功')
      await router.push('/')
    }
  } finally {
    loading.value = false
  }
}
</script>

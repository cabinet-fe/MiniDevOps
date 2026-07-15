<script setup lang="ts">
import { reactive, ref, useTemplateRef } from "vue";
import { useRoute, useRouter } from "vue-router";
import { message } from "@veltra/desktop";

import { useAuthStore } from "@/stores/auth";

const auth = useAuthStore();
const router = useRouter();
const route = useRoute();

const formRef = useTemplateRef("form");
const loading = ref(false);
const formData = reactive({
  username: "",
  password: "",
});

async function handleSubmit() {
  const valid = await formRef.value?.validate();
  if (!valid) return;

  loading.value = true;
  try {
    await auth.login(formData.username, formData.password);
    const redirect = typeof route.query.redirect === "string" ? route.query.redirect : "/";
    await router.replace(redirect || "/");
  } catch (err) {
    const msg = err instanceof Error ? err.message : "登录失败";
    message.error(msg);
  } finally {
    loading.value = false;
  }
}
</script>

<template>
  <div class="login-page">
    <div class="login-panel">
      <h1 class="title">Bedrock</h1>
      <p class="subtitle">项目开发基石平台</p>

      <u-form
        ref="form"
        :model="formData"
        label-position="top"
        label-width="auto"
        :cols="1"
        class="login-form"
      >
        <u-input
          label="用户名"
          field="username"
          placeholder="请输入用户名"
          :rules="{ required: '请输入用户名' }"
        />
        <u-password-input
          label="密码"
          field="password"
          placeholder="请输入密码"
          :rules="{ required: '请输入密码' }"
        />
      </u-form>

      <u-button type="primary" class="submit-btn" :loading="loading" @click="handleSubmit">
        登录
      </u-button>
    </div>
  </div>
</template>

<style scoped lang="scss">
@use "pkg:@veltra/styles/functions" as fn;

.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background:
    radial-gradient(
      ellipse at 20% 20%,
      color-mix(in srgb, fn.use-var(color, primary) 12%, transparent),
      transparent 50%
    ),
    radial-gradient(
      ellipse at 80% 80%,
      color-mix(in srgb, fn.use-var(color, info) 10%, transparent),
      transparent 45%
    ),
    fn.use-var(bg-color, bottom);
}

.login-panel {
  width: min(400px, calc(100vw - 32px));
  padding: 40px 32px;
  border-radius: fn.use-var(radius, large);
  background: fn.use-var(bg-color, top);
  border: fn.use-var(border);
  box-shadow: fn.use-var(shadow);
}

.title {
  margin: 0;
  font-size: fn.use-var(font-size-title, large);
  font-weight: 700;
  color: fn.use-var(text-color, title);
  letter-spacing: 0.04em;
}

.subtitle {
  margin: 8px 0 28px;
  color: fn.use-var(text-color, assist);
  font-size: fn.use-var(font-size-main, default);
}

.login-form {
  margin-bottom: fn.use-var(gap, large);
}

.submit-btn {
  width: 100%;
}
</style>

<script setup lang="ts">
import { nextTick, reactive, ref, useTemplateRef } from "vue";
import { useRoute, useRouter } from "vue-router";
import { message } from "@veltra/desktop";

import { useAuthStore } from "@/stores/auth";

import Atmosphere from "./components/atmosphere";

const auth = useAuthStore();
const router = useRouter();
const route = useRoute();

const formRef = useTemplateRef("form");
const pageRef = useTemplateRef<HTMLElement>("page");
const loading = ref(false);
const formData = reactive({
  username: "",
  password: "",
});

function pinPageScroll() {
  const page = pageRef.value;
  if (!page) return;
  page.scrollTop = 0;
  void nextTick(() => {
    page.scrollTop = 0;
  });
}

async function handleSubmit() {
  if (loading.value) return;

  const valid = await formRef.value?.validate();
  pinPageScroll();
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
    pinPageScroll();
  }
}
</script>

<template>
  <div ref="page" class="login-page">
    <Atmosphere />

    <div class="stage">
      <header class="brand">
        <p class="brand-en">BEDROCK</p>
        <h1 class="brand-cn">磐石<span class="seal" aria-hidden="true">磐</span></h1>
        <p class="brand-tag">磐石者，万物之基也</p>
        <p class="brand-intro">代码托管 · 持续集成 · 部署运维 · 智能协同，诸事归一</p>
      </header>

      <div class="paper">
        <span class="paper-corner paper-corner--tl" aria-hidden="true" />
        <span class="paper-corner paper-corner--tr" aria-hidden="true" />
        <span class="paper-corner paper-corner--bl" aria-hidden="true" />
        <span class="paper-corner paper-corner--br" aria-hidden="true" />

        <p class="paper-caption"><span class="paper-caption-text">准入</span></p>

        <u-form
          ref="form"
          :model="formData"
          label-position="top"
          label-width="auto"
          :cols="1"
          class="login-form"
          @keyup.enter="handleSubmit"
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
  </div>
</template>

<style scoped lang="scss">
.login-page {
  --paper: #f6f2e6;
  --paper-deep: #ede7d5;
  --ink: #2b2a26;
  --ink-soft: #7a7264;
  --pine: #3d6b58;
  --cinnabar: #b3452e;
  --line: #d8cfb6;

  position: fixed;
  inset: 0;
  isolation: isolate;
  height: 100dvh;
  overflow: hidden;
  overscroll-behavior: none;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: flex-start;
  padding: clamp(28px, 7vh, 72px) 20px 28px;
  box-sizing: border-box;
  background: var(--paper);
  color: var(--ink);
}

.stage {
  position: relative;
  z-index: 1;
  width: min(400px, 100%);
  display: flex;
  flex-direction: column;
  gap: 24px;
  flex-shrink: 0;
}

.brand {
  text-align: center;
  animation: brand-rise 0.9s cubic-bezier(0.22, 1, 0.36, 1) both;
}

.brand-en {
  margin: 0 0 6px;
  font-size: 11px;
  font-weight: 500;
  letter-spacing: 0.42em;
  text-indent: 0.42em;
  color: var(--ink-soft);
}

.brand-cn {
  position: relative;
  display: inline-block;
  margin: 0;
  font-family: "Songti SC", "STSong", "SimSun", "Noto Serif CJK SC", serif;
  font-size: clamp(52px, 11vw, 72px);
  font-weight: 700;
  line-height: 1;
  letter-spacing: 0.18em;
  text-indent: 0.18em;
  color: var(--ink);
}

/* 朱砂小印，缀于题名之侧 */
.seal {
  position: absolute;
  right: -34px;
  bottom: 4px;
  display: grid;
  place-items: center;
  width: 26px;
  height: 26px;
  border-radius: 4px;
  background: var(--cinnabar);
  color: #f8f3e6;
  font-family: "Songti SC", "STSong", "SimSun", "Noto Serif CJK SC", serif;
  font-size: 15px;
  font-weight: 700;
  letter-spacing: 0;
  text-indent: 0;
  box-shadow: 0 1px 3px rgb(43 42 38 / 25%);
  transform: rotate(3deg);
}

.brand-tag {
  margin: 14px 0 0;
  font-family: "Songti SC", "STSong", "SimSun", "Noto Serif CJK SC", serif;
  font-size: 14px;
  letter-spacing: 0.32em;
  text-indent: 0.32em;
  color: var(--ink);
}

.brand-intro {
  margin: 8px 0 0;
  font-size: 12px;
  letter-spacing: 0.08em;
  color: var(--ink-soft);
}

/* 笺纸面板：双线边框 + 四角回纹 */
.paper {
  position: relative;
  padding: 26px 22px 22px;
  background:
    linear-gradient(180deg, rgb(255 255 255 / 55%) 0%, transparent 30%), var(--paper-deep);
  border: 1px solid var(--line);
  box-shadow:
    0 1px 0 rgb(255 255 255 / 60%) inset,
    0 12px 32px rgb(64 54 32 / 12%);
  animation: paper-settle 1.05s cubic-bezier(0.22, 1, 0.36, 1) 0.12s both;

  &::before {
    content: "";
    position: absolute;
    inset: 5px;
    border: 1px solid rgb(61 107 88 / 22%);
    pointer-events: none;
  }
}

.paper-corner {
  position: absolute;
  width: 14px;
  height: 14px;
  border: 2px solid var(--pine);
  opacity: 0.55;
  pointer-events: none;

  &--tl {
    top: -2px;
    left: -2px;
    border-right: none;
    border-bottom: none;
  }

  &--tr {
    top: -2px;
    right: -2px;
    border-left: none;
    border-bottom: none;
  }

  &--bl {
    bottom: -2px;
    left: -2px;
    border-right: none;
    border-top: none;
  }

  &--br {
    bottom: -2px;
    right: -2px;
    border-left: none;
    border-top: none;
  }
}

.paper-caption {
  display: flex;
  align-items: center;
  gap: 12px;
  margin: 0 0 18px;

  &::before,
  &::after {
    content: "";
    flex: 1;
    height: 1px;
    background: linear-gradient(90deg, transparent, var(--line));
  }

  &::after {
    background: linear-gradient(270deg, transparent, var(--line));
  }
}

.paper-caption-text {
  font-family: "Songti SC", "STSong", "SimSun", "Noto Serif CJK SC", serif;
  font-size: 14px;
  letter-spacing: 0.42em;
  text-indent: 0.42em;
  color: var(--pine);
}

.login-form {
  margin-bottom: 16px;

  :deep(.u-form-item) {
    margin-bottom: 4px;
  }

  :deep(.u-form-item__label),
  :deep(.u-form-item__label-text) {
    color: var(--ink-soft) !important;
  }

  :deep(.u-form-item__error) {
    min-height: 18px;
  }

  :deep(.u-form-item:not(.is-error) .u-form-item__content)::after {
    content: "";
    display: block;
    height: 18px;
  }
}

.submit-btn {
  width: 100%;
  letter-spacing: 0.32em;
  text-indent: 0.32em;
}

@keyframes brand-rise {
  from {
    opacity: 0;
    transform: translateY(18px);
  }

  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes paper-settle {
  from {
    opacity: 0;
    transform: translateY(22px);
  }

  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@media (max-width: 480px) {
  .login-page {
    padding: 32px 16px 24px;
  }

  .brand-cn {
    letter-spacing: 0.12em;
    text-indent: 0.12em;
  }

  .seal {
    right: -30px;
  }

  .paper {
    padding: 22px 16px 16px;
  }
}
</style>

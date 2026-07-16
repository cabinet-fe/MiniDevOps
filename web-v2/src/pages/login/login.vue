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
        <h1 class="brand-cn">磐石</h1>
        <p class="brand-tag">开发基石 · 承压而立</p>
      </header>

      <div class="vault">
        <span class="vault-wing vault-wing--left" aria-hidden="true" />
        <span class="vault-wing vault-wing--right" aria-hidden="true" />
        <span class="vault-bolt vault-bolt--tl" aria-hidden="true" />
        <span class="vault-bolt vault-bolt--tr" aria-hidden="true" />
        <span class="vault-bolt vault-bolt--bl" aria-hidden="true" />
        <span class="vault-bolt vault-bolt--br" aria-hidden="true" />
        <div class="vault-rail vault-rail--top" aria-hidden="true" />
        <div class="vault-rail vault-rail--bottom" aria-hidden="true" />

        <div class="plate">
          <p class="plate-caption">ACCESS · 准入</p>

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
  </div>
</template>

<style scoped lang="scss">
.login-page {
  --ink: #070a08;
  --recess: #141b17;
  --verdigris: #4a7a64;
  --verdigris-bright: #6b9a82;
  --brass: #b08a4a;
  --brass-dim: #7a6238;
  --bone: #d4cfc3;
  --ash: #8e968e;
  --rust-mist: rgb(138 79 58 / 22%);

  --u-color-primary: var(--verdigris);
  --u-color-primary-light-1: #527d6a;
  --u-color-primary-dark-1: #355a4a;
  --u-color-primary-dark-3: #2a483c;
  --u-color-primary-a-10: rgb(74 122 100 / 10%);
  --u-color-primary-a-16: rgb(74 122 100 / 16%);
  --u-color-primary-a-22: rgb(74 122 100 / 22%);
  --u-color-primary-a-28: rgb(74 122 100 / 28%);
  --u-color-primary-a-40: rgb(74 122 100 / 40%);
  --u-color-primary-a-50: rgb(74 122 100 / 50%);
  --u-color-primary-a-60: rgb(74 122 100 / 60%);

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
  padding: clamp(28px, 8vh, 88px) 20px 28px;
  box-sizing: border-box;
  background: var(--ink);
  color: var(--bone);
  font-family: ui-monospace, "SF Mono", Menlo, Consolas, "Courier New", monospace;
}

.stage {
  position: relative;
  z-index: 1;
  width: min(420px, 100%);
  display: flex;
  flex-direction: column;
  gap: 22px;
  flex-shrink: 0;

  &::before {
    content: "";
    position: absolute;
    z-index: -1;
    left: -18%;
    right: -18%;
    top: 18%;
    bottom: -8%;
    background:
      radial-gradient(ellipse 55% 50% at 50% 55%, rgb(74 122 100 / 12%), transparent 72%),
      radial-gradient(ellipse 70% 60% at 50% 58%, rgb(0 0 0 / 55%), transparent 75%);
    filter: blur(2px);
  }
}

.brand {
  text-align: center;
  animation: brand-rise 0.9s cubic-bezier(0.22, 1, 0.36, 1) both;
}

.brand-en {
  margin: 0 0 4px;
  font-size: 11px;
  font-weight: 500;
  letter-spacing: 0.42em;
  color: var(--brass);
  text-indent: 0.42em;
  text-shadow: 0 0 20px rgb(176 138 74 / 25%);
}

.brand-cn {
  margin: 0;
  font-family: "Songti SC", "STSong", "SimSun", "Noto Serif CJK SC", serif;
  font-size: clamp(52px, 11vw, 76px);
  font-weight: 700;
  line-height: 1;
  letter-spacing: 0.18em;
  text-indent: 0.18em;
  color: var(--bone);
  text-shadow:
    0 1px 0 rgb(0 0 0 / 65%),
    0 14px 40px rgb(0 0 0 / 55%),
    0 0 24px rgb(74 122 100 / 18%);
}

.brand-tag {
  margin: 14px 0 0;
  font-size: 12px;
  letter-spacing: 0.28em;
  text-indent: 0.28em;
  color: var(--ash);
}

.vault {
  position: relative;
  padding: 14px;
  background:
    linear-gradient(165deg, rgb(120 140 128 / 14%) 0%, transparent 42%),
    linear-gradient(
      180deg,
      rgb(42 52 46 / 78%) 0%,
      rgb(22 30 26 / 82%) 45%,
      rgb(12 16 14 / 88%) 100%
    );
  border: 1px solid rgb(176 138 74 / 38%);
  backdrop-filter: blur(1.5px);
  box-shadow:
    0 0 0 1px rgb(0 0 0 / 60%),
    0 1px 0 rgb(255 255 255 / 7%) inset,
    0 24px 48px rgb(0 0 0 / 45%),
    0 0 60px rgb(74 122 100 / 10%),
    0 -40px 80px rgb(0 0 0 / 35%);
  animation: plate-settle 1.05s cubic-bezier(0.22, 1, 0.36, 1) 0.12s both;

  &::before {
    content: "";
    position: absolute;
    inset: 5px;
    border: 1px solid rgb(74 122 100 / 28%);
    box-shadow:
      0 0 28px rgb(74 122 100 / 8%) inset,
      inset 0 0 40px rgb(0 0 0 / 20%);
    pointer-events: none;
  }

  &::after {
    content: "";
    position: absolute;
    inset: 0;
    pointer-events: none;
    background: repeating-linear-gradient(
      8deg,
      transparent 0 48px,
      rgb(255 255 255 / 1.5%) 48px 49px,
      transparent 49px 100px,
      rgb(0 0 0 / 10%) 100px 102px
    );
    opacity: 0.55;
    mix-blend-mode: soft-light;
  }
}

.vault-wing {
  position: absolute;
  top: 18%;
  bottom: 18%;
  width: min(18vw, 96px);
  pointer-events: none;
  box-shadow: 0 0 24px rgb(0 0 0 / 35%);
  opacity: 0.7;

  &--left {
    right: 100%;
    margin-right: -1px;
    background:
      linear-gradient(90deg, transparent, rgb(30 40 34 / 75%) 40%, rgb(42 52 46 / 55%)),
      linear-gradient(
        180deg,
        transparent 0%,
        rgb(176 138 74 / 28%) 14%,
        rgb(74 122 100 / 20%) 50%,
        rgb(176 138 74 / 28%) 86%,
        transparent 100%
      );
    mask-image: linear-gradient(90deg, transparent, #000 55%);
  }

  &--right {
    left: 100%;
    margin-left: -1px;
    background:
      linear-gradient(270deg, transparent, rgb(30 40 34 / 75%) 40%, rgb(42 52 46 / 55%)),
      linear-gradient(
        180deg,
        transparent 0%,
        rgb(176 138 74 / 28%) 14%,
        rgb(74 122 100 / 20%) 50%,
        rgb(176 138 74 / 28%) 86%,
        transparent 100%
      );
    mask-image: linear-gradient(270deg, transparent, #000 55%);
  }
}

.vault-rail {
  position: absolute;
  left: 18%;
  right: 18%;
  height: 2px;
  background: linear-gradient(
    90deg,
    transparent,
    var(--brass-dim) 20%,
    var(--brass) 50%,
    var(--brass-dim) 80%,
    transparent
  );
  opacity: 0.65;
  animation: brass-sweep 8s ease-in-out infinite;

  &--top {
    top: 0;
  }

  &--bottom {
    bottom: 0;
    animation-delay: -4s;
  }
}

.vault-bolt {
  position: absolute;
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: radial-gradient(circle at 32% 28%, #d2b87a, var(--brass-dim) 52%, #2a2418 100%);
  box-shadow:
    0 1px 2px rgb(0 0 0 / 60%),
    inset 0 1px 0 rgb(255 255 255 / 28%),
    0 0 0 1px rgb(0 0 0 / 40%);

  &--tl {
    top: 8px;
    left: 8px;
  }

  &--tr {
    top: 8px;
    right: 8px;
  }

  &--bl {
    bottom: 8px;
    left: 8px;
  }

  &--br {
    bottom: 8px;
    right: 8px;
  }
}

.plate {
  position: relative;
  padding: 22px 18px 18px;
  color: var(--bone);
  background:
    linear-gradient(180deg, rgb(255 255 255 / 3%) 0%, transparent 28%),
    repeating-linear-gradient(90deg, transparent 0 10px, rgb(255 255 255 / 1.5%) 10px 11px),
    linear-gradient(180deg, #1c2621 0%, var(--recess) 100%);
  border: 1px solid rgb(0 0 0 / 55%);
  box-shadow:
    0 0 0 1px rgb(74 122 100 / 16%),
    0 2px 0 rgb(255 255 255 / 4%) inset,
    0 -18px 36px rgb(0 0 0 / 35%) inset,
    inset 0 0 40px rgb(0 0 0 / 25%);
}

.plate-caption {
  margin: 0 0 16px;
  font-size: 10px;
  letter-spacing: 0.32em;
  color: var(--verdigris-bright);
  font-weight: 500;
  text-shadow: 0 0 12px rgb(74 122 100 / 35%);
}

.login-form {
  margin-bottom: 16px;

  :deep(.u-form-item) {
    margin-bottom: 4px;
  }

  :deep(.u-form-item__label),
  :deep(.u-form-item__label-text) {
    color: var(--ash) !important;
  }

  :deep(.u-form-item__error) {
    min-height: 18px;
  }

  :deep(.u-form-item:not(.is-error) .u-form-item__content)::after {
    content: "";
    display: block;
    height: 18px;
  }

  :deep(.u-input) {
    background: linear-gradient(180deg, #c5cdc0 0%, #a8b3a6 100%) !important;
    border: 1px solid rgb(0 0 0 / 45%) !important;
    border-radius: 2px;
    box-shadow:
      inset 0 1px 3px rgb(0 0 0 / 28%),
      0 0 0 1px rgb(74 122 100 / 18%) !important;
    color: #141b17;
  }

  :deep(.u-input__native) {
    color: #141b17;

    &::placeholder {
      color: rgb(20 27 23 / 45%);
    }
  }
}

.submit-btn {
  width: 100%;
  letter-spacing: 0.18em;
  text-indent: 0.18em;
  font-family: inherit;
  border-radius: 2px;
  box-shadow: 0 0 20px rgb(74 122 100 / 22%);
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

@keyframes plate-settle {
  from {
    opacity: 0;
    transform: translateY(22px);
  }

  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes brass-sweep {
  0%,
  100% {
    opacity: 0.3;
    filter: brightness(0.9);
  }

  50% {
    opacity: 0.85;
    filter: brightness(1.2);
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

  .vault {
    padding: 12px;
  }

  .plate {
    padding: 18px 14px 14px;
  }
}
</style>

<script setup lang="ts">
import { useAuthStore } from "@/stores/auth";

const auth = useAuthStore();
</script>

<template>
  <div class="home">
    <h2>欢迎使用 Bedrock 2.0</h2>
    <p class="hint">侧栏菜单由 <code>/auth/me</code> 下发，按角色权限裁剪。</p>
    <p v-if="auth.user" class="meta">
      当前用户：{{ auth.user.display_name || auth.user.username }}
      <span v-if="auth.user.is_super_admin">（超级管理员）</span>
    </p>
  </div>
</template>

<style scoped>
.home h2 {
  margin: 0 0 8px;
  font-size: 22px;
  font-weight: 600;
}

.hint {
  margin: 0;
  color: #6b7280;
}

.meta {
  margin-top: 16px;
  color: #374151;
}

code {
  font-family: ui-monospace, monospace;
}
</style>

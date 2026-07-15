<script setup lang="ts">
import { useRoute } from "vue-router";

const route = useRoute();
</script>

<template>
  <div class="placeholder">
    <h2>功能即将就绪</h2>
    <p>
      路径 <code>{{ route.path }}</code> 将在后续阶段实现。
    </p>
  </div>
</template>

<style scoped>
.placeholder h2 {
  margin: 0 0 8px;
  font-size: 20px;
}
p {
  color: #6b7280;
}
code {
  font-family: ui-monospace, monospace;
}
</style>

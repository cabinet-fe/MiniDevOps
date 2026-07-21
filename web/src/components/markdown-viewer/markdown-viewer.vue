<script setup lang="ts">
import MarkdownRender from "markstream-vue";
import "markstream-vue/index.css";

defineProps<{
  content: string;
}>();
</script>

<template>
  <!-- markstream-vue 的暗色变量挂靠在 .dark 类上（而非 data-theme 属性），
       此处显式补类名，否则暗色主题下按浅色配色渲染导致文字看不清 -->
  <div class="markdown-viewer dark">
    <MarkdownRender :content="content" />
  </div>
</template>

<style scoped lang="scss">
.markdown-viewer {
  min-width: 0;
  padding: 16px 20px;
  border-radius: 8px;
  background: var(--u-bg-color-top, #fff);
  color-scheme: dark;
  line-height: 1.65;

  :deep(h1),
  :deep(h2),
  :deep(h3) {
    margin-top: 1.25em;
    margin-bottom: 0.5em;
  }

  :deep(h1:first-child),
  :deep(h2:first-child),
  :deep(h3:first-child) {
    margin-top: 0;
  }

  :deep(pre) {
    overflow: auto;
    padding: 12px;
    border-radius: 6px;
    background: var(--u-bg-color-middle, #f6f7f9);
  }

  :deep(code) {
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
    font-size: 0.92em;
  }

  :deep(table) {
    width: 100%;
    border-collapse: collapse;
    font-size: 13px;
  }

  :deep(th),
  :deep(td) {
    padding: 8px 10px;
    border: 1px solid var(--u-border-color, #e5e7eb);
    text-align: left;
  }

  :deep(th) {
    background: var(--u-bg-color-middle, #f6f7f9);
  }
}
</style>

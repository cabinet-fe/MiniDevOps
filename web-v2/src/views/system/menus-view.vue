<script setup lang="ts">
import { onMounted, ref } from "vue";
import { message } from "@veltra/desktop";

import { listMenus, updateMenu, updateMenuIcon } from "@/api/system";
import type { RbacResource } from "@/api/types";
import { usePermission } from "@/composables/use-permission";

const { hasPermission } = usePermission();
const tree = ref<RbacResource[]>([]);
const loading = ref(false);
const selected = ref<RbacResource | null>(null);
const title = ref("");
const routePath = ref("");

async function load() {
  loading.value = true;
  try {
    const res = await listMenus();
    tree.value = res.items ?? [];
  } catch (err) {
    message.error(err instanceof Error ? err.message : "加载失败");
  } finally {
    loading.value = false;
  }
}

function select(node: RbacResource) {
  selected.value = node;
  title.value = node.menu_metadata?.title || "";
  routePath.value = node.menu_metadata?.route || "";
}

async function saveMeta() {
  if (!selected.value) return;
  try {
    await updateMenu(selected.value.id, { title: title.value, route: routePath.value });
    message.success("已保存");
    await load();
  } catch (err) {
    message.error(err instanceof Error ? err.message : "保存失败");
  }
}

async function onIconFile(ev: Event) {
  if (!selected.value || selected.value.parent_id) {
    message.error("仅一级菜单可上传图标");
    return;
  }
  const input = ev.target as HTMLInputElement;
  const file = input.files?.[0];
  if (!file) return;
  if (file.size > 32 * 1024) {
    message.error("图标原始体积不得超过 32KB");
    input.value = "";
    return;
  }
  const buf = await file.arrayBuffer();
  const bytes = new Uint8Array(buf);
  let binary = "";
  for (const b of bytes) binary += String.fromCharCode(b);
  const b64 = btoa(binary);
  try {
    await updateMenuIcon(selected.value.id, b64, file.type || "image/png");
    message.success("图标已更新");
    await load();
  } catch (err) {
    message.error(err instanceof Error ? err.message : "上传失败");
  } finally {
    input.value = "";
  }
}

onMounted(() => {
  void load();
});
</script>

<template>
  <div class="page">
    <div class="page-head">
      <h2>菜单</h2>
      <u-button @click="load">刷新</u-button>
    </div>

    <div class="layout">
      <div class="tree-pane">
        <p v-if="loading" class="hint">加载中…</p>
        <ul v-else class="menu-tree">
          <li v-for="root in tree" :key="root.id">
            <button
              type="button"
              class="node-btn"
              :class="{ active: selected?.id === root.id }"
              @click="select(root)"
            >
              {{ root.menu_metadata?.title || root.path }}
            </button>
            <ul v-if="root.children?.length">
              <li v-for="child in root.children" :key="child.id">
                <button
                  type="button"
                  class="node-btn"
                  :class="{ active: selected?.id === child.id }"
                  @click="select(child)"
                >
                  {{ child.menu_metadata?.title || child.path }}
                </button>
              </li>
            </ul>
          </li>
        </ul>
      </div>

      <div v-if="selected" class="editor">
        <h3>{{ selected.path }}</h3>
        <u-form :model="{ title, routePath }" label-width="72px" :cols="1">
          <u-input v-model="title" label="标题" field="title" />
          <u-input v-model="routePath" label="路由" field="routePath" />
        </u-form>
        <div v-if="!selected.parent_id" class="icon-row">
          <label class="icon-label">
            一级图标（≤32KB）
            <input
              type="file"
              accept="image/*"
              :disabled="!hasPermission('system.menus:update')"
              @change="onIconFile"
            />
          </label>
          <img
            v-if="selected.menu_metadata?.icon_base64"
            class="icon-preview"
            :src="
              selected.menu_metadata.icon_base64.startsWith('data:')
                ? selected.menu_metadata.icon_base64
                : `data:${selected.menu_metadata.icon_mime || 'image/png'};base64,${selected.menu_metadata.icon_base64}`
            "
            alt="icon"
          />
        </div>
        <u-button v-if="hasPermission('system.menus:update')" type="primary" @click="saveMeta">
          保存
        </u-button>
      </div>
      <u-empty v-else text="选择菜单节点进行编辑" />
    </div>
  </div>
</template>

<style scoped>
.page-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}
.page-head h2 {
  margin: 0;
  font-size: 20px;
}
.layout {
  display: grid;
  grid-template-columns: 280px 1fr;
  gap: 24px;
}
.menu-tree,
.menu-tree ul {
  list-style: none;
  margin: 0;
  padding: 0;
}
.menu-tree ul {
  padding-left: 16px;
}
.node-btn {
  display: block;
  width: 100%;
  text-align: left;
  border: none;
  background: transparent;
  padding: 6px 8px;
  border-radius: 6px;
  cursor: pointer;
}
.node-btn.active,
.node-btn:hover {
  background: #eff6ff;
}
.editor h3 {
  margin: 0 0 12px;
  font-size: 16px;
}
.icon-row {
  margin: 12px 0 16px;
  display: flex;
  align-items: center;
  gap: 12px;
}
.icon-label {
  font-size: 13px;
  color: #374151;
}
.icon-preview {
  width: 32px;
  height: 32px;
  object-fit: contain;
}
.hint {
  color: #6b7280;
}
</style>

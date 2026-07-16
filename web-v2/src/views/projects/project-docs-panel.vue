<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from "vue";
import { message } from "@veltra/desktop";

import { listAgents } from "@/api/ai";
import {
  createDocNode,
  deleteDocNode,
  generateDocs,
  getDocDiff,
  getDocNode,
  importDocsZIP,
  listDocTree,
  moveDocNode,
  publishDocNode,
  updateDocNode,
  uploadMarkdown,
} from "@/api/projects";
import type { ApiDocDiff, ApiDocNode, ProductProject, ProjectRole } from "@/api/types";
import FormDialog from "@/components/form-dialog.vue";
import { usePermission } from "@/composables/use-permission";

const props = defineProps<{
  project: ProductProject;
  projectRole?: ProjectRole;
  manageAll: boolean;
}>();
const { hasPermission } = usePermission();

const tree = ref<ApiDocNode[]>([]);
const selectedID = ref<number>();
const selected = ref<ApiDocNode | null>(null);
const draftContent = ref("");
const previewMode = ref<"draft" | "published">("draft");
const diff = ref<ApiDocDiff | null>(null);
const nodeDialogOpen = ref(false);
const moveDialogOpen = ref(false);
const creatingKind = ref<"dir" | "doc">("doc");
const nodeForm = reactive({ name: "" });
const moveForm = reactive({ parent_id: undefined as number | undefined, sort_order: 0 });
const generateAgentID = ref<number>();
const agentOptions = ref<{ label: string; value: number }[]>([]);
const generating = ref(false);

const canEditProjectContent = computed(
  () =>
    props.manageAll ||
    props.projectRole === "owner" ||
    props.projectRole === "admin" ||
    props.projectRole === "member",
);
const canAdminProjectContent = computed(
  () => props.manageAll || props.projectRole === "owner" || props.projectRole === "admin",
);
const canCreate = computed(
  () => hasPermission("project.docs:create") && canEditProjectContent.value,
);
const canUpdate = computed(
  () => hasPermission("project.docs:update") && canEditProjectContent.value,
);
const canDelete = computed(
  () => hasPermission("project.docs:delete") && canAdminProjectContent.value,
);
const canGenerate = computed(
  () => hasPermission("project.docs:execute") && canEditProjectContent.value,
);
// Markdown is intentionally rendered as interpolated text below, never v-html.
// This keeps raw HTML and javascript: links inert until a vetted renderer exists.
const renderedContent = computed(() =>
  previewMode.value === "published"
    ? (selected.value?.published_content ?? "")
    : draftContent.value,
);

async function loadTree() {
  try {
    tree.value = await listDocTree(props.project.id);
  } catch (error) {
    message.error(error instanceof Error ? error.message : "文档树加载失败");
  }
}

async function selectNode(id?: number) {
  selectedID.value = id;
  diff.value = null;
  if (!id) {
    selected.value = null;
    draftContent.value = "";
    return;
  }
  try {
    const node = await getDocNode(props.project.id, id);
    selected.value = node;
    draftContent.value = node.draft_content ?? "";
    previewMode.value = node.draft_updated_at ? "draft" : "published";
    if (node.kind === "doc") diff.value = await getDocDiff(props.project.id, node.id);
  } catch (error) {
    message.error(error instanceof Error ? error.message : "读取文档失败");
  }
}

function openCreate(kind: "dir" | "doc") {
  creatingKind.value = kind;
  nodeForm.name = "";
  nodeDialogOpen.value = true;
}

function selectedDirectoryID() {
  if (!selected.value) return null;
  return selected.value.kind === "dir" ? selected.value.id : (selected.value.parent_id ?? null);
}

async function createNode() {
  try {
    const node = await createDocNode(props.project.id, {
      kind: creatingKind.value,
      name: nodeForm.name,
      parent_id: selectedDirectoryID(),
    });
    nodeDialogOpen.value = false;
    await loadTree();
    await selectNode(node.id);
    message.success(creatingKind.value === "dir" ? "目录已创建" : "文档草稿已创建");
  } catch (error) {
    message.error(error instanceof Error ? error.message : "创建失败");
  }
}

async function saveDraft() {
  if (!selected.value || selected.value.kind !== "doc") return;
  try {
    const node = await updateDocNode(props.project.id, selected.value.id, {
      draft_content: draftContent.value,
    });
    selected.value = node;
    diff.value = await getDocDiff(props.project.id, node.id);
    await loadTree();
    message.success("草稿已保存");
  } catch (error) {
    message.error(error instanceof Error ? error.message : "草稿保存失败");
  }
}

async function runGenerate() {
  if (!selected.value || selected.value.kind !== "doc") return;
  if (!generateAgentID.value) {
    message.error("请选择智能体");
    return;
  }
  generating.value = true;
  try {
    const result = await generateDocs(props.project.id, {
      agent_id: generateAgentID.value,
      node_id: selected.value.id,
    });
    message.success(`已创建 AgentRun #${result.agent_run_id}；成功后仅写入草稿，请人工发布`);
  } catch (error) {
    message.error(error instanceof Error ? error.message : "生成失败");
  } finally {
    generating.value = false;
  }
}

async function publish() {
  if (!selected.value || selected.value.kind !== "doc") return;
  if (!window.confirm(`确认发布文档「${selected.value.name}」的草稿？`)) return;
  try {
    const node = await publishDocNode(
      props.project.id,
      selected.value.id,
      selected.value.content_version,
    );
    selected.value = node;
    draftContent.value = "";
    diff.value = await getDocDiff(props.project.id, node.id);
    await loadTree();
    message.success("文档已发布");
  } catch (error) {
    const text = error instanceof Error ? error.message : "发布失败";
    message.error(text.includes("版本冲突") ? "版本冲突：请刷新文档后重试" : text);
    if (text.includes("版本冲突")) await selectNode(selected.value.id);
  }
}

async function removeNode() {
  if (!selected.value || !window.confirm(`确认删除「${selected.value.name}」及其子节点？`)) return;
  try {
    await deleteDocNode(props.project.id, selected.value.id);
    await selectNode();
    await loadTree();
    message.success("节点已删除");
  } catch (error) {
    message.error(error instanceof Error ? error.message : "删除失败");
  }
}

function openMove() {
  if (!selected.value) return;
  moveForm.parent_id = selected.value.parent_id ?? undefined;
  moveForm.sort_order = selected.value.sort_order;
  moveDialogOpen.value = true;
}

async function move() {
  if (!selected.value) return;
  try {
    await moveDocNode(props.project.id, selected.value.id, {
      parent_id: moveForm.parent_id ?? null,
      sort_order: moveForm.sort_order,
    });
    moveDialogOpen.value = false;
    await loadTree();
    await selectNode(selected.value.id);
    message.success("节点已移动");
  } catch (error) {
    message.error(error instanceof Error ? error.message : "移动失败");
  }
}

async function uploadMarkdownFile(files: File[]) {
  const file = files[0];
  if (!file) return;
  try {
    const node = await uploadMarkdown(props.project.id, selectedDirectoryID(), file);
    await loadTree();
    await selectNode(node.id);
    message.success("Markdown 已导入为草稿");
  } catch (error) {
    message.error(error instanceof Error ? error.message : "Markdown 导入失败");
  }
}

async function importZIPFile(files: File[]) {
  const file = files[0];
  if (!file) return;
  try {
    const items = await importDocsZIP(props.project.id, selectedDirectoryID(), file);
    await loadTree();
    if (items[0]) await selectNode(items[0].id);
    message.success(`已导入 ${items.length} 个 Markdown 草稿`);
  } catch (error) {
    message.error(error instanceof Error ? error.message : "ZIP 导入失败");
  }
}

watch(
  () => props.project.id,
  () => {
    selected.value = null;
    selectedID.value = undefined;
    void loadTree();
  },
  { immediate: true },
);

onMounted(() => {
  void loadTree();
  void listAgents({ page: 1, page_size: 100 })
    .then((page) => {
      agentOptions.value = (page.items ?? []).map((a) => ({ label: a.name, value: a.id }));
    })
    .catch(() => {
      /* AI menu may be unavailable */
    });
});
</script>

<template>
  <section class="docs">
    <aside class="tree-panel">
      <div class="tree-head">
        <strong>文档树</strong>
        <u-action-group v-if="canCreate" :max="4">
          <u-action @run="openCreate('dir')">新建目录</u-action>
          <u-action @run="openCreate('doc')">新建文档</u-action>
        </u-action-group>
      </div>
      <u-tree
        v-model:selected="selectedID"
        :data="tree"
        label-key="name"
        value-key="id"
        children-key="children"
        selectable
        expand-all
        @update:selected="selectNode"
      />
      <div v-if="canCreate" class="uploads">
        <u-file-picker accept=".md,text/markdown" @pick="uploadMarkdownFile" />
        <u-file-picker accept=".zip,application/zip" @pick="importZIPFile" />
      </div>
    </aside>

    <section class="editor-panel">
      <u-empty v-if="!selected" text="从左侧选择文档节点" />
      <template v-else>
        <div class="editor-head">
          <div>
            <h3>{{ selected.name }}</h3>
            <p>
              <u-tag :type="selected.kind === 'dir' ? 'default' : 'primary'">{{
                selected.kind
              }}</u-tag>
              <span v-if="selected.kind === 'doc'">已发布版本 {{ selected.content_version }}</span>
            </p>
          </div>
          <u-action-group :max="4">
            <u-action v-if="canUpdate" @run="openMove">移动</u-action>
            <u-action v-if="canDelete" type="danger" @run="removeNode">删除</u-action>
          </u-action-group>
        </div>

        <template v-if="selected.kind === 'doc'">
          <div class="doc-actions">
            <u-button v-if="canUpdate" type="primary" @click="saveDraft">保存草稿</u-button>
            <u-button v-if="canUpdate && selected.draft_updated_at" type="success" @click="publish">
              发布草稿
            </u-button>
            <template v-if="canGenerate">
              <u-select
                v-model="generateAgentID"
                :options="agentOptions"
                placeholder="选择智能体"
                style="width: 160px"
              />
              <u-button :loading="generating" @click="runGenerate">AI 生成草稿</u-button>
            </template>
          </div>
          <p v-if="selected.draft_source_run_id" class="gen-hint">
            草稿来源 AgentRun #{{ selected.draft_source_run_id }}（需人工发布）
          </p>
          <u-textarea
            v-if="canUpdate"
            v-model="draftContent"
            :rows="16"
            placeholder="Markdown 草稿"
          />
          <pre class="markdown-preview">{{ renderedContent }}</pre>
          <div v-if="diff" class="diff-summary">
            草稿 {{ diff.draft_lines }} 行，已发布 {{ diff.published_lines }} 行，新增
            {{ diff.added_lines }} 行，移除 {{ diff.removed_lines }} 行。
          </div>
        </template>
        <u-empty v-else text="目录不包含 Markdown 内容" />
      </template>
    </section>

    <FormDialog
      v-model="nodeDialogOpen"
      :title="creatingKind === 'dir' ? '新建目录' : '新建文档'"
      :model="nodeForm"
      label-width="80px"
      style="width: 420px"
      @submit="createNode"
    >
      <u-input label="名称" field="name" :rules="{ required: '必填' }" />
    </FormDialog>
    <FormDialog
      v-model="moveDialogOpen"
      title="移动节点"
      :model="moveForm"
      label-width="100px"
      style="width: 420px"
      @submit="move"
    >
      <u-number-input label="父目录 ID" field="parent_id" placeholder="留空表示根目录" />
      <u-number-input label="排序" field="sort_order" />
    </FormDialog>
  </section>
</template>

<style scoped>
.docs {
  display: grid;
  min-height: 460px;
  grid-template-columns: 280px minmax(0, 1fr);
  gap: 16px;
}
.tree-panel,
.editor-panel {
  min-width: 0;
  padding: 14px;
  border-radius: 8px;
  background: var(--u-bg-color-top, #fff);
}
.tree-panel {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.tree-head,
.editor-head,
.editor-head p,
.doc-actions,
.uploads {
  display: flex;
  align-items: center;
}
.tree-head,
.editor-head {
  justify-content: space-between;
  gap: 12px;
}
.uploads {
  flex-wrap: wrap;
  gap: 8px;
}
.editor-panel {
  display: flex;
  flex-direction: column;
  gap: 14px;
}
.editor-head h3,
.editor-head p {
  margin: 0;
}
.editor-head p {
  gap: 8px;
  margin-top: 5px;
  color: var(--u-text-color-assist, #7c8494);
  font-size: 13px;
}
.doc-actions {
  flex-wrap: wrap;
  gap: 8px;
}
.gen-hint {
  margin: 0;
  font-size: 12px;
  color: var(--u-color-text-secondary, #666);
}
.markdown-preview {
  max-height: 260px;
  margin: 0;
  overflow: auto;
  padding: 12px;
  border-radius: 6px;
  background: var(--u-bg-color-middle, #f6f7f9);
  white-space: pre-wrap;
  word-break: break-word;
}
.diff-summary {
  color: var(--u-text-color-second, #626b7d);
  font-size: 13px;
}
@media (max-width: 900px) {
  .docs {
    grid-template-columns: 1fr;
  }
}
</style>

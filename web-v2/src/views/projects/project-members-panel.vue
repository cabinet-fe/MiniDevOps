<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue";
import { defineTableColumns, message } from "@veltra/desktop";

import {
  addProjectMember,
  listProjectMembers,
  removeProjectMember,
  transferProjectOwner,
  updateProjectMember,
} from "@/api/projects";
import type { ProductProject, ProjectMember, ProjectRole } from "@/api/types";
import FormDialog from "@/components/form-dialog.vue";
import { usePermission } from "@/composables/use-permission";
import { useAuthStore } from "@/stores/auth";

const props = defineProps<{ project: ProductProject }>();
const emit = defineEmits<{ membersChanged: []; ownerTransferred: [] }>();

const auth = useAuthStore();
const { hasPermission } = usePermission();
const members = ref<ProjectMember[]>([]);
const loading = ref(false);
const dialogOpen = ref(false);
const editing = ref<ProjectMember | null>(null);
const form = reactive({ user_id: 0, role: "member" as Exclude<ProjectRole, "owner"> });

const selfMember = computed(() => members.value.find((member) => member.user_id === auth.user?.id));
const canManageAll = computed(() => hasPermission("project.projects:manage_all"));
const canManageMembers = computed(
  () =>
    hasPermission("project.projects:update") &&
    (canManageAll.value ||
      selfMember.value?.role === "owner" ||
      selfMember.value?.role === "admin"),
);
const canTransferOwner = computed(
  () =>
    hasPermission("project.projects:update") &&
    (canManageAll.value || selfMember.value?.role === "owner"),
);

const columns = defineTableColumns([
  { key: "user_id", name: "用户 ID", minWidth: 120 },
  { key: "role", name: "项目角色", minWidth: 140 },
  { key: "created_at", name: "加入时间", minWidth: 180 },
  { key: "action", name: "操作", width: 220, minWidth: 180 },
]);

async function load() {
  loading.value = true;
  try {
    members.value = await listProjectMembers(props.project.id);
  } catch (error) {
    message.error(error instanceof Error ? error.message : "成员加载失败");
  } finally {
    loading.value = false;
  }
}

function openAdd() {
  editing.value = null;
  Object.assign(form, { user_id: 0, role: "member" });
  dialogOpen.value = true;
}

function openEdit(member: ProjectMember) {
  editing.value = member;
  Object.assign(form, {
    user_id: member.user_id,
    role: member.role as Exclude<ProjectRole, "owner">,
  });
  dialogOpen.value = true;
}

async function save() {
  try {
    if (editing.value) {
      await updateProjectMember(props.project.id, editing.value.user_id, form.role);
      message.success("角色已更新");
    } else {
      await addProjectMember(props.project.id, form.user_id, form.role);
      message.success("成员已添加");
    }
    dialogOpen.value = false;
    await load();
    emit("membersChanged");
  } catch (error) {
    message.error(error instanceof Error ? error.message : "保存失败");
  }
}

async function remove(member: ProjectMember) {
  if (!window.confirm(`确认移除用户 #${member.user_id}？`)) return;
  try {
    await removeProjectMember(props.project.id, member.user_id);
    message.success("成员已移除");
    await load();
    emit("membersChanged");
  } catch (error) {
    message.error(error instanceof Error ? error.message : "移除失败");
  }
}

async function transfer(member: ProjectMember) {
  if (!window.confirm(`确认将 Owner 转让给用户 #${member.user_id}？`)) return;
  try {
    await transferProjectOwner(props.project.id, member.user_id);
    message.success("Owner 已转让");
    emit("ownerTransferred");
    await load();
  } catch (error) {
    message.error(error instanceof Error ? error.message : "转让失败");
  }
}

onMounted(() => void load());
</script>

<template>
  <section class="panel">
    <div class="panel-head">
      <div>
        <h3>项目成员</h3>
        <p>Owner 可转让负责人；Admin 可管理非 Owner 成员。</p>
      </div>
      <u-button v-if="canManageMembers" type="primary" @click="openAdd">添加成员</u-button>
    </div>

    <div v-loading="loading" class="table-wrap">
      <u-table :columns="columns" :data="members" row-key="id">
        <template #column:role="{ rowData }">
          <u-tag :type="(rowData as ProjectMember).role === 'owner' ? 'primary' : 'default'">
            {{ (rowData as ProjectMember).role }}
          </u-tag>
        </template>
        <template #column:action="{ rowData }">
          <u-action-group :max="3">
            <u-action
              v-if="canManageMembers && (rowData as ProjectMember).role !== 'owner'"
              @run="openEdit(rowData as ProjectMember)"
            >
              修改角色
            </u-action>
            <u-action
              v-if="canTransferOwner && (rowData as ProjectMember).role !== 'owner'"
              @run="transfer(rowData as ProjectMember)"
            >
              转让 Owner
            </u-action>
            <u-action
              v-if="canManageMembers && (rowData as ProjectMember).role !== 'owner'"
              need-confirm
              type="danger"
              @run="remove(rowData as ProjectMember)"
            >
              移除
            </u-action>
          </u-action-group>
        </template>
      </u-table>
    </div>

    <FormDialog
      v-model="dialogOpen"
      :title="editing ? '修改成员角色' : '添加成员'"
      :model="form"
      label-width="100px"
      style="width: 420px"
      @submit="save"
    >
      <u-number-input
        v-if="!editing"
        label="用户 ID"
        field="user_id"
        :rules="{ required: '必填' }"
      />
      <u-select
        label="角色"
        field="role"
        :options="[
          { label: '项目管理员', value: 'admin' },
          { label: '成员', value: 'member' },
          { label: '只读', value: 'readonly' },
        ]"
      />
    </FormDialog>
  </section>
</template>

<style scoped>
.panel {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.panel-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}
.panel-head h3,
.panel-head p {
  margin: 0;
}
.panel-head p {
  margin-top: 4px;
  color: var(--u-text-color-assist, #7c8494);
  font-size: 13px;
}
.table-wrap {
  min-height: 180px;
  padding: 12px;
  border-radius: 8px;
  background: var(--u-bg-color-top, #fff);
}
</style>

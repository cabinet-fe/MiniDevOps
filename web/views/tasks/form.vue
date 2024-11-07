<template>
  <m-dialog-pro :ctx="ctx" style="width: 800px">
    <u-form :model="model" label-width="100px">
      <u-input label="任务名称" field="name" />
      <u-select
        label="仓库"
        field="repoId"
        :options="repoList"
        value-key="id"
        label-key="name"
        @update:model-value="getBranchList"
      />
      <u-select label="分支" field="branch" :options="branchList" />
      <u-input
        label="构建物路径"
        field="bundlerDir"
        tips="构建后的构建物相对于代码根目录的路径"
      />
      <u-input
        label="部署目录"
        field="deployPath"
        tips="指定部署到本机服务器上的目录"
      />
      <u-multi-select
        label="远程目录"
        field="remoteIds"
        :options="remoteList"
        value-key="id"
        label-key="name"
        span="full"
      />

      <m-code-editor span="full" label="构建脚本" field="script" />
    </u-form>

    <template #trigger>
      <u-button type="primary" @click="ctx.open({ action: 'create' })">
        新增
      </u-button>
    </template>
  </m-dialog-pro>
</template>

<script lang="ts" setup>
import { http } from '@/utils/http'
import { useDialogPro } from '@meta/components'
import { FormModel } from 'ultra-ui'
import { shallowRef } from 'vue'

const emit = defineEmits(['success'])

const model = new FormModel({
  name: { required: true },
  repoId: { required: true },
  branch: { required: true },
  address: { required: true },
  bundlerDir: { required: true },
  deployPath: {},
  remoteIds: { value: () => [] as number[] },
  script: { value: '' },
  id: {}
})

const remoteList = shallowRef<any[]>([])
const repoList = shallowRef<any[]>([])

http.get('/remotes/list').then(({ data }) => {
  remoteList.value = data
})

http.get('/repos/list').then(({ data }) => {
  repoList.value = data
})

const branchList = shallowRef<any[]>([])

function getBranchList(repoId?: number) {
  if (!repoId) {
    branchList.value = []
    return
  }
  http.get(`/repos/${repoId}/branch`).then(({ data }) => {
    branchList.value = data.map(v => ({ label: v, value: v }))
  })
}

const ctx = useDialogPro({
  models: [model],
  async submit(state) {
    const { remoteIds, id, ...data } = model.data
    const body = {
      ...data,
      remoteIds: remoteIds.join(',')
    }
    if (state.action === 'create') {
      await http.post('/tasks', body)
    } else {
      await http.put(`/tasks/${id}`, body)
    }

    emit('success')
  },
  afterOpen(state) {
    if (!state.action) return
    state.title = {
      create: '新增构建任务',
      update: '修改构建任务'
    }[state.action]
  }
})

defineExpose({
  open(action: 'update', data: Record<string, any>) {
    const { remoteIds, ...rest } = data
    model.setData({
      ...rest,
      remoteIds: !remoteIds ? [] : remoteIds.split(',').map(Number)
    })
    getBranchList(rest.repoId)
    ctx.open({ action })
  }
})
</script>

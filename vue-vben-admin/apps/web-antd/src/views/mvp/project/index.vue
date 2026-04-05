<script setup lang="ts">
import { useRouter } from 'vue-router';
import type { VbenFormProps } from '#/adapter/form';
import type { VxeGridProps } from '#/adapter/vxe-table';

import { Page, useVbenModal } from '@vben/common-ui';
import { Button, message, Modal, Progress, Tag } from 'ant-design-vue';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import { getProjectList, deleteProject } from '#/api/mvp/project';
import { confirmPlan, resumeProject } from '#/api/mvp/workflow';
import type { ProjectItem } from '#/api/mvp/project/types';
import { engineVersionMap } from '../consts';

import FormModal from './modules/form.vue';
import PauseModal from './modules/pause-modal.vue';
import StatusPanel from './modules/status-panel.vue';

const router = useRouter();

/** 状态中文名称映射 */
const STATUS_LABELS: Record<string, string> = {
  designing: '设计中',
  reviewing: '方案审核中',
  running: '执行中',
  executing: '执行中',
  reworking: '返工中',
  paused: '已暂停',
  completed: '已完成',
  failed: '已失败',
  canceled: '已取消',
};

/** 状态 Tag 颜色映射 */
const STATUS_COLORS: Record<string, string> = {
  designing: 'blue',
  reviewing: 'cyan',
  running: 'green',
  executing: 'green',
  reworking: 'orange',
  paused: 'orange',
  completed: 'default',
  failed: 'red',
  canceled: 'default',
};

/** 新建项目弹窗 */
const [FormModalComp, formModalApi] = useVbenModal({
  connectedComponent: FormModal,
  destroyOnClose: true,
});

/** 暂停弹窗 */
const [PauseModalComp, pauseModalApi] = useVbenModal({
  connectedComponent: PauseModal,
  destroyOnClose: true,
});

/** 项目状态详情抽屉 */
const [StatusPanelComp, statusPanelApi] = useVbenModal({
  connectedComponent: StatusPanel,
  destroyOnClose: true,
});

/** 搜索表单配置 */
const formOptions: VbenFormProps = {
  collapsed: false,
  showCollapseButton: false,
  submitOnEnter: true,
  schema: [
    {
      component: 'Input',
      componentProps: { placeholder: '请输入项目名称', allowClear: true },
      fieldName: 'name',
      label: '项目名称',
    },
    {
      component: 'Select',
      fieldName: 'status',
      label: '项目状态',
      componentProps: {
        placeholder: '全部状态',
        allowClear: true,
        options: [
          { label: '设计中', value: 'designing' },
          { label: '方案审核中', value: 'reviewing' },
          { label: '执行中', value: 'running' },
          { label: '已暂停', value: 'paused' },
          { label: '已完成', value: 'completed' },
        ],
      },
    },
  ],
};

/** 表格列配置 */
const gridOptions: VxeGridProps<ProjectItem> = {
  columns: [
    { title: '序号', type: 'seq', width: 60 },
    {
      field: 'name',
      title: '项目名称',
      minWidth: 160,
      slots: { default: 'col-name' },
    },
    {
      field: 'projectCategory',
      title: '项目分类',
      width: 120,
    },
    {
      field: 'engineVersion',
      title: '引擎版本',
      width: 120,
      slots: { default: 'col-engine-version' },
    },
    {
      field: 'status',
      title: '状态',
      width: 100,
      slots: { default: 'col-status' },
    },
    {
      field: 'taskProgress',
      title: '任务进度',
      width: 180,
      slots: { default: 'col-progress' },
    },
    {
      field: 'pauseReason',
      title: '暂停原因',
      minWidth: 160,
      showOverflow: 'tooltip',
    },
    {
      field: 'createdAt',
      title: '创建时间',
      width: 180,
      formatter: 'formatDateTime',
      sortable: true,
    },
    {
      title: '操作',
      width: 280,
      fixed: 'right',
      slots: { default: 'action' },
    },
  ],
  height: 'auto',
  pagerConfig: {},
  proxyConfig: {
    ajax: {
      query: async ({ page, sorts }, formValues) => {
        const params: Record<string, any> = {
          pageNum: page.currentPage,
          pageSize: page.pageSize,
          ...formValues,
        };
        if (sorts && sorts.length > 0) {
          const sort = sorts[0];
          if (sort?.field && sort?.order) {
            params.orderBy = sort.field;
            params.orderDir = sort.order;
          }
        }
        const res = await getProjectList(params as any);
        return { items: res?.list ?? [], total: res?.total ?? 0 };
      },
    },
  },
  sortConfig: { remote: true, trigger: 'cell' },
  toolbarConfig: { custom: true, refresh: true, search: true },
};

const [Grid, gridApi] = useVbenVxeGrid({ formOptions, gridOptions });

/** 新建项目 */
function handleCreate() {
  formModalApi.setData(null).open();
}

/** 编辑项目 */
function handleEdit(row: ProjectItem) {
  formModalApi.setData({ id: row.id }).open();
}

/** 进入对话（与AI架构师对话） */
function handleChat(row: ProjectItem) {
  router.push({ path: '/mvp/chat', query: { projectId: row.id } });
}

/** 查看详情（打开状态面板） */
function handleViewStatus(row: ProjectItem) {
  statusPanelApi.setData({ project: row }).open();
}

/** 打开执行控制台 */
function handleExecution(row: ProjectItem) {
  router.push({ path: '/mvp/workflow/execution', query: { projectId: row.id } });
}

/** 确认方案（设计阶段 -> 执行阶段） */
function handleConfirmPlan(row: ProjectItem) {
  Modal.confirm({
    title: '确认实施方案',
    content: '确认后系统将先执行方案审核（系统预检 + AI审核 + 调度优化），审核通过后自动开始执行。',
    okText: '确认方案',
    async onOk() {
      await confirmPlan(row.id);
      message.success('方案已提交审核，请稍候...');
      gridApi.reload();
    },
  });
}

/** 暂停项目 */
function handlePause(row: ProjectItem) {
  pauseModalApi.setData({ projectID: row.id }).open();
}

/** 恢复项目 */
function handleResume(row: ProjectItem) {
  Modal.confirm({
    title: '恢复项目',
    content: '确定要恢复该项目的执行吗？',
    async onOk() {
      await resumeProject(row.id);
      message.success('项目已恢复执行');
      gridApi.reload();
    },
  });
}

/** 删除项目 */
function handleDelete(row: ProjectItem) {
  Modal.confirm({
    title: '确认删除',
    content: `确定要删除项目「${row.name}」吗？此操作不可恢复。`,
    okType: 'danger',
    async onOk() {
      await deleteProject(row.id);
      message.success('删除成功');
      gridApi.reload();
    },
  });
}
</script>

<template>
  <Page auto-content-height>
    <!-- 弹窗组件挂载 -->
    <FormModalComp @success="() => gridApi.reload()" />
    <PauseModalComp @success="() => gridApi.reload()" />
    <StatusPanelComp @refresh="() => gridApi.reload()" />

    <Grid>
      <!-- 工具栏操作区 -->
      <template #toolbar-actions>
        <Button type="primary" @click="handleCreate">新建项目</Button>
      </template>

      <!-- 项目名称列 -->
      <template #col-name="{ row }">
        <span class="font-medium text-gray-800">{{ row.name }}</span>
        <p v-if="row.description" class="text-xs text-gray-400 mt-0.5 truncate max-w-xs">
          {{ row.description }}
        </p>
      </template>

      <!-- 引擎版本列 -->
      <template #col-engine-version="{ row }">
        <Tag :color="engineVersionMap[row.engineVersion ?? 'legacy']?.color ?? 'default'">
          {{ engineVersionMap[row.engineVersion ?? 'legacy']?.label ?? 'Legacy' }}
        </Tag>
      </template>

      <!-- 状态 Tag 列 -->
      <template #col-status="{ row }">
        <Tag :color="STATUS_COLORS[row.status] ?? 'default'">
          {{ STATUS_LABELS[row.status] ?? row.status ?? '-' }}
        </Tag>
      </template>

      <!-- 任务进度列 -->
      <template #col-progress="{ row }">
        <div v-if="row.totalTasks > 0" class="flex items-center gap-2">
          <Progress
            :percent="
              row.totalTasks
                ? Math.floor(((row.completedTasks ?? 0) / row.totalTasks) * 100)
                : 0
            "
            size="small"
            class="flex-1"
          />
          <span class="text-xs text-gray-400 whitespace-nowrap">
            {{ row.completedTasks ?? 0 }}/{{ row.totalTasks }}
          </span>
        </div>
        <span v-else class="text-xs text-gray-300">-</span>
      </template>

      <!-- 操作列：根据状态动态显示按钮 -->
      <template #action="{ row }">
        <!-- 设计中：进入对话 + 确认方案 -->
        <template v-if="row.status === 'designing'">
          <Button type="link" size="small" @click="handleChat(row)">进入对话</Button>
          <Button type="link" size="small" @click="handleConfirmPlan(row)">确认方案</Button>
        </template>

        <!-- 审核中：查看详情（只读等待） -->
        <template v-else-if="row.status === 'reviewing'">
          <Button type="link" size="small" @click="handleViewStatus(row)">查看详情</Button>
          <Tag color="cyan" class="ml-1">审核中...</Tag>
        </template>

        <!-- 执行中：查看详情 + 执行控制台 + 暂停 -->
        <template v-else-if="row.status === 'running' || row.status === 'executing'">
          <Button type="link" size="small" @click="handleViewStatus(row)">查看详情</Button>
          <Button v-if="row.engineVersion === 'workflow_v2'" type="link" size="small" @click="handleExecution(row)">执行控制台</Button>
          <Button type="link" size="small" class="text-orange-500" @click="handlePause(row)">暂停</Button>
        </template>

        <!-- 返工中：查看详情 + 执行控制台 + 暂停 -->
        <template v-else-if="row.status === 'reworking'">
          <Button type="link" size="small" @click="handleViewStatus(row)">查看详情</Button>
          <Button type="link" size="small" @click="handleExecution(row)">执行控制台</Button>
          <Button type="link" size="small" class="text-orange-500" @click="handlePause(row)">暂停</Button>
        </template>

        <!-- 暂停中：恢复 + 进入对话 -->
        <template v-else-if="row.status === 'paused'">
          <Button type="link" size="small" class="text-green-600" @click="handleResume(row)">恢复</Button>
          <Button type="link" size="small" @click="handleChat(row)">进入对话</Button>
        </template>

        <!-- 已完成：查看详情 -->
        <template v-else-if="row.status === 'completed'">
          <Button type="link" size="small" @click="handleViewStatus(row)">查看详情</Button>
        </template>

        <!-- 编辑按钮（任何状态都可以编辑） -->
        <Button type="link" size="small" @click="handleEdit(row)">编辑</Button>

        <!-- 删除按钮（任何状态都可以删除） -->
        <Button
          type="link"
          size="small"
          danger
          @click="handleDelete(row)"
        >删除</Button>
      </template>
    </Grid>
  </Page>
</template>

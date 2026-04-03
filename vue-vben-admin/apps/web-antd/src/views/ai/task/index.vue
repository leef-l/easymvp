<script setup lang="ts">
import type { VbenFormProps } from '#/adapter/form';
import type { VxeGridProps } from '#/adapter/vxe-table';

import dayjs from 'dayjs';
import { Button, message, Modal, Tag } from 'ant-design-vue';
import { Page, useVbenModal } from '@vben/common-ui';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import { cancelTask, getTaskDetail, getTaskList, getTaskLogs } from '#/api/ai/task';
import type { TaskItem } from '#/api/ai/task/types';

import ExecuteForm from './modules/execute-form.vue';

const [ExecuteFormComp, executeModalApi] = useVbenModal({
  connectedComponent: ExecuteForm,
  destroyOnClose: true,
});

const statusOptions = [
  { label: '待执行', value: 'pending' },
  { label: '执行中', value: 'running' },
  { label: '已完成', value: 'success' },
  { label: '失败', value: 'failed' },
  { label: '已取消', value: 'cancelled' },
];

const formOptions: VbenFormProps = {
  collapsed: false,
  submitOnEnter: true,
  schema: [
    {
      component: 'Select',
      fieldName: 'engineCode',
      label: '执行引擎',
      componentProps: {
        options: [
          { label: 'Aider', value: 'aider' },
          { label: 'OpenHands', value: 'openhands' },
        ],
        allowClear: true,
        class: 'w-full',
      },
    },
    {
      component: 'Select',
      fieldName: 'status',
      label: '状态',
      componentProps: { options: statusOptions, allowClear: true, class: 'w-full' },
    },
  ],
};

function getStatusColor(status: string) {
  switch (status) {
    case 'success':
      return 'green';
    case 'failed':
      return 'red';
    case 'running':
      return 'blue';
    case 'cancelled':
      return 'default';
    default:
      return 'orange';
  }
}

function getActivityColor(row: TaskItem) {
  if (row.status !== 'running') {
    return 'default';
  }
  return row.stalled ? 'error' : row.isActuallyWorking ? 'success' : 'warning';
}

function getActivityText(row: TaskItem) {
  if (row.status !== 'running') {
    return '-';
  }
  return row.stalled ? '疑似卡住' : row.isActuallyWorking ? '活跃中' : '待判定';
}

const gridOptions: VxeGridProps<TaskItem> = {
  columns: [
    { field: 'title', title: '任务标题', minWidth: 180 },
    { field: 'engineCode', title: '执行引擎', width: 120 },
    { field: 'status', title: '状态', width: 120, slots: { default: 'status_cell' } },
    {
      field: 'lastActiveAt',
      title: '最后活跃',
      width: 180,
      formatter({ cellValue }) {
        return cellValue ? dayjs(cellValue).format('YYYY-MM-DD HH:mm:ss') : '-';
      },
    },
    { field: 'repoPath', title: '仓库路径', minWidth: 220, showOverflow: 'tooltip' },
    { field: 'createdAt', title: '创建时间', width: 180, formatter: 'formatDateTime' },
    { title: '操作', width: 240, fixed: 'right', slots: { default: 'action' } },
  ],
  proxyConfig: {
    ajax: {
      query: async ({ page }, formValues) => {
        const res = await getTaskList({
          pageNum: page.currentPage,
          pageSize: page.pageSize,
          engineCode: formValues.engineCode,
          status: formValues.status,
        });
        return { items: res?.list ?? [], total: res?.total ?? 0 };
      },
    },
  },
  toolbarConfig: {
    refresh: true,
    search: true,
    custom: true,
  },
};

const [Grid, gridApi] = useVbenVxeGrid({
  formOptions,
  gridOptions,
});

function handleCreate() {
  executeModalApi.open();
}

async function handleView(row: TaskItem) {
  const detail = await getTaskDetail(row.id);
  Modal.info({
    title: detail.title,
    width: 720,
    content: detail.responseSummary || detail.errorMessage || detail.instruction || '暂无详细内容',
  });
}

async function handleLogs(row: TaskItem) {
  const res = await getTaskLogs(row.id);
  const content = (res?.list ?? [])
    .map((item) => `[${item.logType}] ${item.content}`)
    .join('\n\n');
  Modal.info({
    title: `任务日志 - ${row.title}`,
    width: 800,
    content: content || '暂无日志',
  });
}

async function handleCancel(row: TaskItem) {
  await cancelTask(row.id);
  message.success('任务已取消');
  gridApi.reload();
}
</script>

<template>
  <Page auto-content-height>
    <ExecuteFormComp @success="() => gridApi.reload()" />
    <Grid>
      <template #toolbar-actions>
        <Button type="primary" @click="handleCreate">新建任务</Button>
      </template>
      <template #status_cell="{ row }">
        <div class="flex flex-wrap gap-1">
          <Tag :color="getStatusColor(row.status)">
            {{ row.status }}
          </Tag>
          <Tag v-if="row.status === 'running'" :color="getActivityColor(row)">
            {{ getActivityText(row) }}
          </Tag>
        </div>
      </template>
      <template #action="{ row }">
        <Button type="link" size="small" @click="handleView(row)">查看</Button>
        <Button type="link" size="small" @click="handleLogs(row)">日志</Button>
        <Button
          v-if="row.status === 'pending' || row.status === 'running'"
          type="link"
          danger
          size="small"
          @click="handleCancel(row)"
        >
          取消
        </Button>
      </template>
    </Grid>
  </Page>
</template>

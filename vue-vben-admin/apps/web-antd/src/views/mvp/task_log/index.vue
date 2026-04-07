<script setup lang="ts">
import type { VbenFormProps } from '#/adapter/form';
import type { VxeGridProps } from '#/adapter/vxe-table';

import { Page, useVbenModal } from '@vben/common-ui';
import { Button, message, Modal, Tag } from 'ant-design-vue';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import { getTaskLogList, deleteTaskLog, batchDeleteTaskLog, exportTaskLog, importTaskLog, downloadImportTemplateTaskLog } from '#/api/mvp/task_log';
import type { TaskLogItem } from '#/api/mvp/task_log/types';
import FormModal from './modules/form.vue';
import DetailDrawer from './modules/detail-drawer.vue';

/** 标签颜色池 */
const TAG_COLORS = ['green', 'red', 'blue', 'orange', 'cyan', 'purple', 'geekblue', 'magenta'];


/** 表单弹窗 */
const [FormModalComp, formModalApi] = useVbenModal({
  connectedComponent: FormModal,
  destroyOnClose: true,
});

/** 详情抽屉 */
const [DetailDrawerComp, detailDrawerApi] = useVbenModal({
  connectedComponent: DetailDrawer,
  destroyOnClose: true,
});

/** 搜索表单配置 */
const formOptions: VbenFormProps = {
  collapsed: false,
  showCollapseButton: true,
  submitOnChange: false,
  submitOnEnter: true,
  schema: [
    {
      component: 'RangePicker',
      fieldName: 'timeRange',
      label: '创建时间',
      componentProps: {
        showTime: true,
        format: 'YYYY-MM-DD HH:mm:ss',
        valueFormat: 'YYYY-MM-DD HH:mm:ss',
        class: 'w-full',
      },
    },
  ],
};

/** 表格列配置 */
const gridOptions: VxeGridProps<TaskLogItem> = {
  columns: [
    { type: 'checkbox', width: 50 },
    { title: '序号', type: 'seq', width: 50 },
    { field: 'taskName', title: '任务ID' },
    { field: 'action', title: '动作' },
    { field: 'fromStatus', title: '原状态' },
    { field: 'toStatus', title: '新状态' },
    { field: 'message', title: '日志内容' },
    { field: 'operator', title: '操作者' },
    { field: 'createdAt', title: '创建时间', width: 180, formatter: 'formatDateTime', sortable: true },
    { title: '操作', width: 240, fixed: 'right', slots: { default: 'action' } },
  ],
  height: 'auto',
  pagerConfig: {},
  proxyConfig: {
    ajax: {
      query: async ({ page, sorts }, formValues) => {
        const { timeRange, ...rest } = formValues;
        const params: Record<string, any> = {
          pageNum: page.currentPage,
          pageSize: page.pageSize,
          ...rest,
        };
        if (timeRange && timeRange.length === 2) {
          params.startTime = timeRange[0];
          params.endTime = timeRange[1];
        }
        if (sorts && sorts.length > 0) {
          const sort = sorts[0];
          if (sort && sort.field && sort.order) {
            params.orderBy = sort.field;
            params.orderDir = sort.order;
          }
        }
        const res = await getTaskLogList(params as any);
        return { items: res?.list ?? [], total: res?.total ?? 0 };
      },
    },
  },
  sortConfig: {
    remote: true,
    trigger: 'cell',
  },
  toolbarConfig: {
    custom: true,
    refresh: true,
    search: true,
  },
};

const [Grid, gridApi] = useVbenVxeGrid({
  formOptions,
  gridOptions,
});

/** 新建 */
function handleCreate() {
  formModalApi.setData(null).open();
}

/** 查看 */
function handleView(row: TaskLogItem) {
  detailDrawerApi.setData({ id: row.id }).open();
}

/** 编辑 */
function handleEdit(row: TaskLogItem) {
  formModalApi.setData({ id: row.id }).open();
}

/** 删除 */
function handleDelete(row: TaskLogItem) {
  Modal.confirm({
    title: '确认删除',
    content: '确定要删除该任务日志表吗？',
    okType: 'danger',
    async onOk() {
      await deleteTaskLog(row.id);
      message.success('删除成功');
      gridApi.reload();
    },
  });
}
/** 批量删除 */
function handleBatchDelete() {
  const rows = gridApi.grid.getCheckboxRecords();
  if (rows.length === 0) {
    message.warning('请先选择要删除的数据');
    return;
  }
  Modal.confirm({
    title: '确认批量删除',
    content: `确定要删除选中的 ${rows.length} 条任务日志表吗？`,
    okType: 'danger',
    async onOk() {
      await batchDeleteTaskLog(rows.map((r: TaskLogItem) => r.id));
      message.success('批量删除成功');
      gridApi.reload();
    },
  });
}

/** 导出 */
async function handleExport() {
  try {
    const formValues = await gridApi.formApi.getValues();
    const params: Record<string, any> = { ...formValues };
    if (params.timeRange && params.timeRange.length === 2) {
      params.startTime = params.timeRange[0];
      params.endTime = params.timeRange[1];
      delete params.timeRange;
    }
    const blob = await exportTaskLog(params);
    const url = URL.createObjectURL(blob as any);
    const a = document.createElement('a');
    a.href = url;
    a.download = '任务日志表.csv';
    a.click();
    URL.revokeObjectURL(url);
    message.success('导出成功');
  } catch {
    message.error('导出失败');
  }
}

/** 导入 */
async function handleImport() {
  const input = document.createElement('input');
  input.type = 'file';
  input.accept = '.csv,.xlsx,.xls';
  input.onchange = async () => {
    const file = input.files?.[0];
    if (!file) { input.remove(); return; }
    const formData = new FormData();
    formData.append('file', file);
    try {
      const res = await importTaskLog(formData);
      message.success(`导入完成：成功 ${res?.success ?? 0} 条，失败 ${res?.fail ?? 0} 条`);
      gridApi.reload();
    } catch {
      message.error('导入失败');
    } finally {
      input.remove();
    }
  };
  document.body.appendChild(input);
  input.click();
}

/** 下载导入模板 */
async function handleDownloadTemplate() {
  try {
    const blob = await downloadImportTemplateTaskLog();
    const url = URL.createObjectURL(blob as any);
    const a = document.createElement('a');
    a.href = url;
    a.download = '任务日志表导入模板.csv';
    a.click();
    URL.revokeObjectURL(url);
  } catch {
    message.error('下载模板失败');
  }
}
</script>

<template>
  <Page auto-content-height>
    <FormModalComp @success="() => gridApi.reload()" />
    <DetailDrawerComp />
    <Grid>
      <template #toolbar-actions>
        <Button v-auth="['mvp:task_log:create']" type="primary" @click="handleCreate">新建</Button>
        <Button v-auth="['mvp:task_log:batch-delete']" danger class="ml-2" @click="handleBatchDelete">批量删除</Button>
        <Button v-auth="['mvp:task_log:export']" class="ml-2" @click="handleExport">导出</Button>
        <Button v-auth="['mvp:task_log:import']" class="ml-2" @click="handleImport">导入</Button>
        <Button class="ml-2" @click="handleDownloadTemplate">模板下载</Button>
      </template>
      <template #action="{ row }">
        <Button v-auth="['mvp:task_log:detail']" type="link" size="small" @click="handleView(row)">查看</Button>
        <Button v-auth="['mvp:task_log:update']" type="link" size="small" @click="handleEdit(row)">编辑</Button>
        <Button v-auth="['mvp:task_log:delete']" type="link" danger size="small" @click="handleDelete(row)">删除</Button>
      </template>
    </Grid>
  </Page>
</template>

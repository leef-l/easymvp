<script setup lang="ts">
import type { VbenFormProps } from '#/adapter/form';
import type { VxeGridProps } from '#/adapter/vxe-table';

import { ref } from 'vue';
import { Page, useVbenModal } from '@vben/common-ui';
import { Button, message, Modal, Tag } from 'ant-design-vue';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import { getConversationList, deleteConversation, batchDeleteConversation, exportConversation, importConversation, downloadImportTemplateConversation, batchUpdateConversation } from '#/api/mvp/conversation';
import type { ConversationItem } from '#/api/mvp/conversation/types';
import { getProjectList } from '#/api/mvp/project';
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

/** 项目选项 */
const projectOptions = ref<{ label: string; value: string }[]>([]);
async function loadProjectOptions() {
  try {
    const res = await getProjectList({ pageNum: 1, pageSize: 200 } as any);
    projectOptions.value = (res?.list || []).map((p: any) => ({
      label: p.name,
      value: String(p.id),
    }));
  } catch {
    // ignore
  }
}
loadProjectOptions().catch((e) => console.warn('[conversation] loadProjectOptions 失败:', e));

/** 搜索表单配置 */
const formOptions: VbenFormProps = {
  collapsed: false,
  showCollapseButton: true,
  submitOnChange: false,
  submitOnEnter: true,
  schema: [
    {
      component: 'ApiSelect',
      componentProps: {
        placeholder: '请选择项目',
        allowClear: true,
        options: projectOptions,
      },
      fieldName: 'projectID',
      label: '所属项目',
    },
    {
      component: 'Select',
      componentProps: {
        placeholder: '请选择角色类型',
        allowClear: true,
        options: [
          { label: '架构师', value: 'architect' },
          { label: '实施员', value: 'implementer' },
          { label: '审���员', value: 'auditor' },
        ],
      },
      fieldName: 'roleType',
      label: '角色类型',
    },
    {
      component: 'Input',
      componentProps: { placeholder: '请输入对话标题', allowClear: true },
      fieldName: 'title',
      label: '对话标题',
    },
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
const gridOptions: VxeGridProps<ConversationItem> = {
  columns: [
    { type: 'checkbox', width: 50 },
    { title: '序号', type: 'seq', width: 50 },
    { field: 'projectName', title: '项目ID' },
    { field: 'taskName', title: '关联任务ID，NULL=项目级对话' },
    { field: 'title', title: '对话标题' },
    { field: 'roleType', title: '对话角色类型' },
    { field: 'status', title: '状态' },
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
        const res = await getConversationList(params as any);
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
function handleView(row: ConversationItem) {
  detailDrawerApi.setData({ id: row.id }).open();
}

/** 编辑 */
function handleEdit(row: ConversationItem) {
  formModalApi.setData({ id: row.id }).open();
}

/** 删除 */
function handleDelete(row: ConversationItem) {
  Modal.confirm({
    title: '确认删除',
    content: '确定要删除该MVP对话表吗？',
    okType: 'danger',
    async onOk() {
      await deleteConversation(row.id);
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
    content: `确定要删除选中的 ${rows.length} 条MVP对话表吗？`,
    okType: 'danger',
    async onOk() {
      await batchDeleteConversation(rows.map((r: ConversationItem) => r.id));
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
    const blob = await exportConversation(params);
    const url = URL.createObjectURL(blob as any);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'MVP对话表.csv';
    a.click();
    setTimeout(() => URL.revokeObjectURL(url), 1000);
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
      const res = await importConversation(formData);
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
    const blob = await downloadImportTemplateConversation();
    const url = URL.createObjectURL(blob as any);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'MVP对话表导入模板.csv';
    a.click();
    setTimeout(() => URL.revokeObjectURL(url), 1000);
  } catch {
    message.error('下载模板失败');
  }
}

/** 批量修改状态 */
function handleBatchUpdateStatus() {
  const rows = gridApi.grid.getCheckboxRecords();
  if (rows.length === 0) {
    message.warning('请先选择要修改的数据');
    return;
  }
  Modal.confirm({
    title: '批量修改状态',
    content: `确定要将选中的 ${rows.length} 条数据的状态切换吗？`,
    async onOk() {
      const newStatus = String(rows[0]?.status) === '1' ? '0' : '1';
      await batchUpdateConversation({ ids: rows.map((r: ConversationItem) => r.id), status: newStatus });
      message.success('批量修改成功');
      gridApi.reload();
    },
  });
}
</script>

<template>
  <Page auto-content-height>
    <FormModalComp @success="() => gridApi.reload()" />
    <DetailDrawerComp />
    <Grid>
      <template #toolbar-actions>
        <Button v-auth="['mvp:conversation:create']" type="primary" @click="handleCreate">新建</Button>
        <Button v-auth="['mvp:conversation:batch-delete']" danger class="ml-2" @click="handleBatchDelete">批量删除</Button>
        <Button v-auth="['mvp:conversation:export']" class="ml-2" @click="handleExport">导出</Button>
        <Button v-auth="['mvp:conversation:import']" class="ml-2" @click="handleImport">导入</Button>
        <Button class="ml-2" @click="handleDownloadTemplate">模板下载</Button>
        <Button v-auth="['mvp:conversation:batch-update']" class="ml-2" @click="handleBatchUpdateStatus">批量修改状态</Button>
      </template>
      <template #action="{ row }">
        <Button v-auth="['mvp:conversation:detail']" type="link" size="small" @click="handleView(row)">查看</Button>
        <Button v-auth="['mvp:conversation:update']" type="link" size="small" @click="handleEdit(row)">编辑</Button>
        <Button v-auth="['mvp:conversation:delete']" type="link" danger size="small" @click="handleDelete(row)">删除</Button>
      </template>
    </Grid>
  </Page>
</template>

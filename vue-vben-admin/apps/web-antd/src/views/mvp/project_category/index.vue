<script setup lang="ts">
import type { VbenFormProps } from '#/adapter/form';
import type { VxeGridProps } from '#/adapter/vxe-table';
import type { ProjectCategoryItem } from '#/api/mvp/project_category/types';

import { Page, useVbenModal } from '@vben/common-ui';

import { Button, message, Modal, Tag } from 'ant-design-vue';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import { batchDeleteProjectCategory, batchUpdateProjectCategory, deleteProjectCategory, downloadImportTemplateProjectCategory, exportProjectCategory, getProjectCategoryList, importProjectCategory } from '#/api/mvp/project_category';

import DetailDrawer from './modules/detail-drawer.vue';
import FormModal from './modules/form.vue';

function hasVerificationProfile(row: ProjectCategoryItem) {
  return !!row.verificationProfileJson?.trim();
}

function hasVerificationGate(row: ProjectCategoryItem) {
  return !!row.verificationGateJson?.trim();
}

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
      component: 'Input',
      componentProps: { placeholder: '请输入展示名称', allowClear: true },
      fieldName: 'displayName',
      label: '展示名称',
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
const gridOptions: VxeGridProps<ProjectCategoryItem> = {
  columns: [
    { type: 'checkbox', width: 50 },
    { title: '序号', type: 'seq', width: 50 },
    { field: 'categoryCode', title: '稳定分类编码' },
    { field: 'displayName', title: '展示名称' },
    { field: 'familyCode', title: '能力家族编码' },
    { field: 'description', title: '分类说明' },
    { field: 'verificationProfileJson', title: '默认验证模板', width: 120, slots: { default: 'verification_profile' } },
    { field: 'verificationGateJson', title: '放行规则', width: 120, slots: { default: 'verification_gate' } },
    { field: 'status', title: '1启用 0停用' },
    { field: 'sort', title: '排序' },
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
        const res = await getProjectCategoryList(params as any);
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
function handleView(row: ProjectCategoryItem) {
  detailDrawerApi.setData({ id: row.id }).open();
}

/** 编辑 */
function handleEdit(row: ProjectCategoryItem) {
  formModalApi.setData({ id: row.id }).open();
}

/** 删除 */
function handleDelete(row: ProjectCategoryItem) {
  Modal.confirm({
    title: '确认删除',
    content: '确定要删除该项目分类配置表吗？',
    okType: 'danger',
    async onOk() {
      await deleteProjectCategory(row.id);
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
    content: `确定要删除选中的 ${rows.length} 条项目分类配置表吗？`,
    okType: 'danger',
    async onOk() {
      await batchDeleteProjectCategory(rows.map((r: ProjectCategoryItem) => r.id));
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
    const blob = await exportProjectCategory(params);
    const url = URL.createObjectURL(blob as any);
    const a = document.createElement('a');
    a.href = url;
    a.download = '项目分类配置表.csv';
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
  input.addEventListener('change', async () => {
    const file = input.files?.[0];
    if (!file) { input.remove(); return; }
    const formData = new FormData();
    formData.append('file', file);
    try {
      const res = await importProjectCategory(formData);
      message.success(`导入完成：成功 ${res?.success ?? 0} 条，失败 ${res?.fail ?? 0} 条`);
      gridApi.reload();
    } catch {
      message.error('导入失败');
    } finally {
      input.remove();
    }
  });
  document.body.append(input);
  input.click();
}

/** 下载导入模板 */
async function handleDownloadTemplate() {
  try {
    const blob = await downloadImportTemplateProjectCategory();
    const url = URL.createObjectURL(blob as any);
    const a = document.createElement('a');
    a.href = url;
    a.download = '项目分类配置表导入模板.csv';
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
      const newStatus = rows[0]?.status === 1 ? 0 : 1;
      await batchUpdateProjectCategory({ ids: rows.map((r: ProjectCategoryItem) => r.id), status: newStatus });
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
        <Button v-auth="['mvp:project_category:create']" type="primary" @click="handleCreate">新建</Button>
        <Button v-auth="['mvp:project_category:batch-delete']" danger class="ml-2" @click="handleBatchDelete">批量删除</Button>
        <Button v-auth="['mvp:project_category:export']" class="ml-2" @click="handleExport">导出</Button>
        <Button v-auth="['mvp:project_category:import']" class="ml-2" @click="handleImport">导入</Button>
        <Button class="ml-2" @click="handleDownloadTemplate">模板下载</Button>
        <Button v-auth="['mvp:project_category:batch-update']" class="ml-2" @click="handleBatchUpdateStatus">批量修改状态</Button>
      </template>
      <template #verification_profile="{ row }">
        <Tag :color="hasVerificationProfile(row) ? 'green' : 'default'">
          {{ hasVerificationProfile(row) ? '已配置' : '自动探测' }}
        </Tag>
      </template>
      <template #verification_gate="{ row }">
        <Tag :color="hasVerificationGate(row) ? 'blue' : 'orange'">
          {{ hasVerificationGate(row) ? '已配置' : '未配置' }}
        </Tag>
      </template>
      <template #action="{ row }">
        <Button v-auth="['mvp:project_category:detail']" type="link" size="small" @click="handleView(row)">查看</Button>
        <Button v-auth="['mvp:project_category:update']" type="link" size="small" @click="handleEdit(row)">编辑</Button>
        <Button v-auth="['mvp:project_category:delete']" type="link" danger size="small" @click="handleDelete(row)">删除</Button>
      </template>
    </Grid>
  </Page>
</template>

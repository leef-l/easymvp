<script setup lang="ts">
import type { VbenFormProps } from '#/adapter/form';
import type { VxeGridProps } from '#/adapter/vxe-table';

import { Page, useVbenModal } from '@vben/common-ui';
import { Button, message, Modal, Switch, Tag } from 'ant-design-vue';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import { getProviderList, deleteProvider, batchDeleteProvider, updateProvider } from '#/api/ai/provider';
import type { ProviderItem } from '#/api/ai/provider/types';
import FormModal from './modules/form.vue';
import DetailDrawer from './modules/detail-drawer.vue';

/** 供应商类型配置 */
const providerTypeConfig: Record<string, { color: string; label: string }> = {
  openai_compatible: { color: 'blue', label: 'OpenAI 兼容' },
  anthropic: { color: 'purple', label: 'Anthropic' },
  google: { color: 'green', label: 'Google' },
};

/** 获取供应商类型显示配置 */
function getProviderTypeConfig(type: string) {
  return providerTypeConfig[type] ?? { color: 'default', label: type };
}

/** 状态选项 */
const statusOptions = [
  { label: '禁用', value: 0 },
  { label: '启用', value: 1 },
];

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
      componentProps: { placeholder: '请输入供应商名称', allowClear: true },
      fieldName: 'name',
      label: '供应商名称',
    },
    {
      component: 'Select',
      componentProps: {
        allowClear: true,
        options: statusOptions,
        placeholder: '请选择状态',
        class: 'w-full',
      },
      fieldName: 'status',
      label: '状态',
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
const gridOptions: VxeGridProps<ProviderItem> = {
  columns: [
    { type: 'checkbox', width: 50 },
    { title: '序号', type: 'seq', width: 50 },
    { field: 'name', title: '供应商名称', minWidth: 120 },
    { field: 'providerType', title: '供应商类型', width: 150, slots: { default: 'providerType_cell' } },
    { field: 'baseURL', title: 'Base URL', minWidth: 200, showOverflow: 'tooltip' },
    { field: 'status', title: '状态', width: 100, slots: { default: 'status_cell' } },
    { field: 'createdAt', title: '创建时间', width: 180, formatter: 'formatDateTime', sortable: true },
    { title: '操作', width: 200, fixed: 'right', slots: { default: 'action' } },
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
        const res = await getProviderList(params as any);
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
function handleView(row: ProviderItem) {
  detailDrawerApi.setData({ id: row.id }).open();
}

/** 编辑 */
function handleEdit(row: ProviderItem) {
  formModalApi.setData({ id: row.id }).open();
}

/** 删除 */
function handleDelete(row: ProviderItem) {
  Modal.confirm({
    title: '确认删除',
    content: `确定要删除供应商「${row.name}」吗？`,
    okType: 'danger',
    async onOk() {
      await deleteProvider(row.id);
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
    content: `确定要删除选中的 ${rows.length} 条供应商吗？`,
    okType: 'danger',
    async onOk() {
      await batchDeleteProvider(rows.map((r: ProviderItem) => r.id));
      message.success('批量删除成功');
      gridApi.reload();
    },
  });
}

/** 切换状态 */
async function handleStatusChange(row: ProviderItem, checked: boolean) {
  try {
    await updateProvider({
      id: row.id,
      name: row.name,
      code: row.code,
      providerType: row.providerType,
      baseURL: row.baseURL,
      icon: row.icon,
      status: checked ? 1 : 0,
      sort: row.sort,
    });
    message.success('状态更新成功');
    gridApi.reload();
  } catch {
    message.error('状态更新失败');
    gridApi.reload();
  }
}
</script>

<template>
  <Page auto-content-height>
    <FormModalComp @success="() => gridApi.reload()" />
    <DetailDrawerComp />
    <Grid>
      <template #toolbar-actions>
        <Button v-auth="['ai:provider:create']" type="primary" @click="handleCreate">新建供应商</Button>
        <Button v-auth="['ai:provider:batch-delete']" danger class="ml-2" @click="handleBatchDelete">批量删除</Button>
      </template>
      <template #providerType_cell="{ row }">
        <Tag :color="getProviderTypeConfig(row.providerType).color">
          {{ getProviderTypeConfig(row.providerType).label }}
        </Tag>
      </template>
      <template #status_cell="{ row }">
        <Switch
          v-auth="['ai:provider:update']"
          :checked="row.status === 1"
          :checked-children="'启用'"
          :un-checked-children="'禁用'"
          size="small"
          @change="(checked) => handleStatusChange(row, checked as boolean)"
        />
      </template>
      <template #action="{ row }">
        <Button v-auth="['ai:provider:detail']" type="link" size="small" @click="handleView(row)">查看</Button>
        <Button v-auth="['ai:provider:update']" type="link" size="small" @click="handleEdit(row)">编辑</Button>
        <Button v-auth="['ai:provider:delete']" type="link" danger size="small" @click="handleDelete(row)">删除</Button>
      </template>
    </Grid>
  </Page>
</template>

<script setup lang="ts">
import type { VbenFormProps } from '#/adapter/form';
import type { VxeGridProps } from '#/adapter/vxe-table';

import { Page, useVbenModal } from '@vben/common-ui';
import { Button, message, Modal, Switch } from 'ant-design-vue';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import { getPlanList, deletePlan, batchDeletePlan, updatePlan } from '#/api/ai/plan';
import type { PlanItem } from '#/api/ai/plan/types';
import FormModal from './modules/form.vue';
import DetailDrawer from './modules/detail-drawer.vue';

/** 状态选项 */
const statusOptions = [
  { label: '禁用', value: 0 },
  { label: '启用', value: 1 },
];

/** API Key 脱敏显示：显示前6位 + *** */
function maskApiKey(key?: string): string {
  if (!key) return '-';
  if (key.length <= 6) return key + '***';
  return key.slice(0, 6) + '***';
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
      componentProps: { placeholder: '请输入套餐名称', allowClear: true },
      fieldName: 'name',
      label: '套餐名称',
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
const gridOptions: VxeGridProps<PlanItem> = {
  columns: [
    { type: 'checkbox', width: 50 },
    { title: '序号', type: 'seq', width: 50 },
    { field: 'name', title: '套餐名称', minWidth: 120 },
    { field: 'providerName', title: '所属供应商', minWidth: 120 },
    { field: 'apiKey', title: 'API Key', minWidth: 160, slots: { default: 'apiKey_cell' } },
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
        const res = await getPlanList(params as any);
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
function handleView(row: PlanItem) {
  detailDrawerApi.setData({ id: row.id }).open();
}

/** 编辑 */
function handleEdit(row: PlanItem) {
  formModalApi.setData({ id: row.id }).open();
}

/** 删除 */
function handleDelete(row: PlanItem) {
  Modal.confirm({
    title: '确认删除',
    content: `确定要删除套餐「${row.name}」吗？`,
    okType: 'danger',
    async onOk() {
      await deletePlan(row.id);
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
    content: `确定要删除选中的 ${rows.length} 条套餐吗？`,
    okType: 'danger',
    async onOk() {
      await batchDeletePlan(rows.map((r: PlanItem) => r.id));
      message.success('批量删除成功');
      gridApi.reload();
    },
  });
}

/** 切换状态 */
async function handleStatusChange(row: PlanItem, checked: boolean) {
  try {
    await updatePlan({ id: row.id, providerID: row.providerID, name: row.name, code: row.code, apiKey: row.apiKey, apiSecret: row.apiSecret, status: checked ? 1 : 0, sort: row.sort });
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
        <Button v-auth="['ai:plan:create']" type="primary" @click="handleCreate">新建套餐</Button>
        <Button v-auth="['ai:plan:batch-delete']" danger class="ml-2" @click="handleBatchDelete">批量删除</Button>
      </template>
      <template #apiKey_cell="{ row }">
        <span class="font-mono text-sm text-gray-600">{{ maskApiKey(row.apiKey) }}</span>
      </template>
      <template #status_cell="{ row }">
        <Switch
          v-auth="['ai:plan:update']"
          :checked="row.status === 1"
          :checked-children="'启用'"
          :un-checked-children="'禁用'"
          size="small"
          @change="(checked) => handleStatusChange(row, checked as boolean)"
        />
      </template>
      <template #action="{ row }">
        <Button v-auth="['ai:plan:detail']" type="link" size="small" @click="handleView(row)">查看</Button>
        <Button v-auth="['ai:plan:update']" type="link" size="small" @click="handleEdit(row)">编辑</Button>
        <Button v-auth="['ai:plan:delete']" type="link" danger size="small" @click="handleDelete(row)">删除</Button>
      </template>
    </Grid>
  </Page>
</template>

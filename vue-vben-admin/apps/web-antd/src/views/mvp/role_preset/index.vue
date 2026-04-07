<script setup lang="ts">
import { h } from 'vue';
import type { VbenFormProps } from '#/adapter/form';
import type { VxeGridProps } from '#/adapter/vxe-table';

import { Page, useVbenModal } from '@vben/common-ui';
import { Button, message, Modal, Tag, Tooltip } from 'ant-design-vue';
import { QuestionCircleOutlined } from '@ant-design/icons-vue';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import { getRolePresetList, deleteRolePreset, batchDeleteRolePreset, exportRolePreset, importRolePreset, downloadImportTemplateRolePreset, batchUpdateRolePreset } from '#/api/mvp/role_preset';
import type { RolePresetItem } from '#/api/mvp/role_preset/types';
import FormModal from './modules/form.vue';
import DetailDrawer from './modules/detail-drawer.vue';
import { roleTypeMap, roleLevelMap, executionModeMap } from '../consts';

/** 标签颜色池 */
const TAG_COLORS = ['green', 'red', 'blue', 'orange', 'cyan', 'purple', 'geekblue', 'magenta'];

/** 状态选项 */
const statusOptions = [
  { label: '禁用', value: '0' },
  { label: '启用', value: '1' },
];

/** 状态映射 */
const statusMap: Record<string, string> = {
  '0': '禁用',
  '1': '启用',
};

/** 状态颜色 */
function getStatusColor(val: string | number): string {
  const colorMap: Record<string, string> = { '0': 'red', '1': 'green' };
  return colorMap[String(val)] ?? 'default';
}

/** 渲染带 Tooltip 的列标题 */
function tooltipHeader(label: string, tip: string) {
  return () => h('span', {}, [
    label + ' ',
    h(Tooltip, { title: tip }, {
      default: () => h(QuestionCircleOutlined, { style: { color: '#999', marginLeft: '4px' } }),
    }),
  ]);
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
  showCollapseButton: false,
  submitOnChange: false,
  submitOnEnter: true,
  schema: [
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
  ],
};

/** 表格列配置 */
const gridOptions: VxeGridProps<RolePresetItem> = {
  columns: [
    { type: 'checkbox', width: 50 },
    { title: '序号', type: 'seq', width: 50 },
    { field: 'projectCategory', title: '项目分类', width: 100 },
    { field: 'roleType', title: '角色类型', width: 100, slots: { default: 'roleType_cell' } },
    { field: 'roleLevel', title: '角色等级', width: 100, slots: { default: 'roleLevel_cell' } },
    { field: 'modelName', title: 'AI模型', minWidth: 160 },
    { field: 'systemPrompt', title: '默认系统提示词', minWidth: 200, showOverflow: 'tooltip', slots: { header: tooltipHeader('默认系统提示词', '角色设定') } },
    { field: 'executionMode', title: '执行方式', width: 100, slots: { default: 'executionMode_cell' } },
    { field: 'status', title: '状态', width: 80, slots: { default: 'status_cell' } },
    { field: 'sort', title: '排序', width: 70 },
    { title: '操作', width: 180, fixed: 'right', slots: { default: 'action' } },
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
        const res = await getRolePresetList(params as any);
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
function handleView(row: RolePresetItem) {
  detailDrawerApi.setData({ id: row.id }).open();
}

/** 编辑 */
function handleEdit(row: RolePresetItem) {
  formModalApi.setData({ id: row.id }).open();
}

/** 删除 */
function handleDelete(row: RolePresetItem) {
  Modal.confirm({
    title: '确认删除',
    content: '确定要删除该角色预设吗？',
    okType: 'danger',
    async onOk() {
      await deleteRolePreset(row.id);
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
    content: `确定要删除选中的 ${rows.length} 条角色预设吗？`,
    okType: 'danger',
    async onOk() {
      await batchDeleteRolePreset(rows.map((r: RolePresetItem) => r.id));
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
    const blob = await exportRolePreset(params);
    const url = URL.createObjectURL(blob as any);
    const a = document.createElement('a');
    a.href = url;
    a.download = '角色预设.csv';
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
      const res = await importRolePreset(formData);
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
    const blob = await downloadImportTemplateRolePreset();
    const url = URL.createObjectURL(blob as any);
    const a = document.createElement('a');
    a.href = url;
    a.download = '角色预设导入模板.csv';
    a.click();
    URL.revokeObjectURL(url);
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
      await batchUpdateRolePreset({ ids: rows.map((r: RolePresetItem) => r.id), status: newStatus });
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
        <Button v-auth="['mvp:role_preset:create']" type="primary" @click="handleCreate">新建</Button>
        <Button v-auth="['mvp:role_preset:batch-delete']" danger class="ml-2" @click="handleBatchDelete">批量删除</Button>
        <Button v-auth="['mvp:role_preset:export']" class="ml-2" @click="handleExport">导出</Button>
        <Button v-auth="['mvp:role_preset:import']" class="ml-2" @click="handleImport">导入</Button>
        <Button class="ml-2" @click="handleDownloadTemplate">模板下载</Button>
        <Button v-auth="['mvp:role_preset:batch-update']" class="ml-2" @click="handleBatchUpdateStatus">批量修改状态</Button>
      </template>
      <template #roleType_cell="{ row }">
        <Tag v-if="roleTypeMap[row.roleType]" :color="roleTypeMap[row.roleType].color">
          {{ roleTypeMap[row.roleType].label }}
        </Tag>
        <span v-else>{{ row.roleType || '-' }}</span>
      </template>
      <template #roleLevel_cell="{ row }">
        <Tag v-if="roleLevelMap[row.roleLevel]" :color="roleLevelMap[row.roleLevel].color">
          {{ roleLevelMap[row.roleLevel].label }}
        </Tag>
        <span v-else>{{ row.roleLevel || '-' }}</span>
      </template>
      <template #executionMode_cell="{ row }">
        <Tag v-if="executionModeMap[row.executionMode]" :color="executionModeMap[row.executionMode].color">
          {{ executionModeMap[row.executionMode].label }}
        </Tag>
        <span v-else>{{ row.executionMode || 'Chat' }}</span>
      </template>
      <template #status_cell="{ row }">
        <Tag :color="getStatusColor(row.status)">
          {{ statusMap[row.status] || row.status }}
        </Tag>
      </template>
      <template #action="{ row }">
        <Button v-auth="['mvp:role_preset:detail']" type="link" size="small" @click="handleView(row)">查看</Button>
        <Button v-auth="['mvp:role_preset:update']" type="link" size="small" @click="handleEdit(row)">编辑</Button>
        <Button v-auth="['mvp:role_preset:delete']" type="link" danger size="small" @click="handleDelete(row)">删除</Button>
      </template>
    </Grid>
  </Page>
</template>

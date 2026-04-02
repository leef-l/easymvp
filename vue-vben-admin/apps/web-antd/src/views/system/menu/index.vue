<script setup lang="ts">
import type { VbenFormProps } from '#/adapter/form';
import type { VxeGridProps } from '#/adapter/vxe-table';

import { Page, useVbenModal } from '@vben/common-ui';
import { Button, message, Modal, Tag } from 'ant-design-vue';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import { getMenuTree, deleteMenu } from '#/api/system/menu';
import type { MenuItem } from '#/api/system/menu/types';
import FormModal from './modules/form.vue';

/** 标签颜色池 */
const TAG_COLORS = ['green', 'red', 'blue', 'orange', 'cyan', 'purple', 'geekblue', 'magenta'];

/** 类型选项 */
const typeOptions = [
  { label: '目录', value: 1 },
  { label: '菜单', value: 2 },
  { label: '按钮', value: 3 },
  { label: '外链', value: 4 },
  { label: '内链', value: 5 },
];

/** 类型映射 */
const typeMap: Record<number, string> = {
  1: '目录',
  2: '菜单',
  3: '按钮',
  4: '外链',
  5: '内链',
};

/** 类型颜色 */
function getTypeColor(val: number): string {
  const keys = [1, 2, 3, 4, 5];
  const idx = keys.indexOf(val);
  return TAG_COLORS[idx >= 0 ? idx % TAG_COLORS.length : 0] ?? 'default';
}

/** 是否显示选项 */
const isShowOptions = [
  { label: '隐藏', value: 0 },
  { label: '显示', value: 1 },
];

/** 是否显示映射 */
const isShowMap: Record<number, string> = {
  0: '隐藏',
  1: '显示',
};

/** 是否显示颜色 */
function getIsShowColor(val: number): string {
  const keys = [0, 1];
  const idx = keys.indexOf(val);
  return TAG_COLORS[idx >= 0 ? idx % TAG_COLORS.length : 0] ?? 'default';
}

/** 是否缓存选项 */
const isCacheOptions = [
  { label: '不缓存', value: 0 },
  { label: '缓存', value: 1 },
];

/** 是否缓存映射 */
const isCacheMap: Record<number, string> = {
  0: '不缓存',
  1: '缓存',
};

/** 是否缓存颜色 */
function getIsCacheColor(val: number): string {
  const keys = [0, 1];
  const idx = keys.indexOf(val);
  return TAG_COLORS[idx >= 0 ? idx % TAG_COLORS.length : 0] ?? 'default';
}

/** 状态选项 */
const statusOptions = [
  { label: '关闭', value: 0 },
  { label: '开启', value: 1 },
];

/** 状态映射 */
const statusMap: Record<number, string> = {
  0: '关闭',
  1: '开启',
};

/** 状态颜色 */
function getStatusColor(val: number): string {
  const keys = [0, 1];
  const idx = keys.indexOf(val);
  return TAG_COLORS[idx >= 0 ? idx % TAG_COLORS.length : 0] ?? 'default';
}

/** 表单弹窗 */
const [FormModalComp, formModalApi] = useVbenModal({
  connectedComponent: FormModal,
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
      component: 'Select',
      componentProps: {
        allowClear: true,
        options: typeOptions,
        placeholder: '请选择类型',
        class: 'w-full',
      },
      fieldName: 'type',
      label: '类型',
    },
    {
      component: 'Select',
      componentProps: {
        allowClear: true,
        options: isShowOptions,
        placeholder: '请选择是否显示',
        class: 'w-full',
      },
      fieldName: 'isShow',
      label: '是否显示',
    },
    {
      component: 'Select',
      componentProps: {
        allowClear: true,
        options: isCacheOptions,
        placeholder: '请选择是否缓存',
        class: 'w-full',
      },
      fieldName: 'isCache',
      label: '是否缓存',
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
  ],
};

/** 表格列配置 */
const gridOptions: VxeGridProps<MenuItem> = {
  columns: [
    // { title: '序号', type: 'seq', width: 50 },
    { field: 'title', title: '菜单名称', width:200, treeNode: true },
    { field: 'type', title: '类型', width: 60, slots: { default: 'type_cell' } },
    { field: 'path', title: '前端路由路径' },
    { field: 'component', title: '前端组件路径' },
    { field: 'permission', title: '权限标识' },
    { field: 'icon', title: '图标' },
    { field: 'sort', title: '排序（升序）' },
    { field: 'isShow', title: '是否显示', width: 80, slots: { default: 'isShow_cell' } },
    { field: 'isCache', title: '是否缓存', width: 80, slots: { default: 'isCache_cell' } },
    // { field: 'linkURL', title: '外链地址' },
    { field: 'status', title: '状态', width: 80, slots: { default: 'status_cell' } },
    { field: 'createdAt', title: '创建时间', width: 180, formatter: 'formatDateTime' },
    { title: '操作', width: 130, fixed: 'right', slots: { default: 'action' } },
  ],
  pagerConfig: { enabled: false },
  treeConfig: {
    childrenField: 'children',
    expandAll: false,
  },
  proxyConfig: {
    ajax: {
      query: async (_params, formValues) => {
        const res = await getMenuTree(formValues);
        return res ?? [];
      },
    },
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

/** 编辑 */
function handleEdit(row: MenuItem) {
  formModalApi.setData({ id: row.id }).open();
}

/** 删除 */
function handleDelete(row: MenuItem) {
  Modal.confirm({
    title: '确认删除',
    content: '确定要删除该菜单表吗？',
    okType: 'danger',
    async onOk() {
      await deleteMenu(row.id);
      message.success('删除成功');
      gridApi.reload();
    },
  });
}
</script>

<template>
  <Page auto-content-height>
    <FormModalComp @success="() => gridApi.reload()" />
    <Grid>
      <template #toolbar-actions>
        <Button type="primary" @click="handleCreate">新建</Button>
      </template>
      <template #type_cell="{ row }">
        <Tag :color="getTypeColor(row.type)">
          {{ typeMap[row.type] || row.type }}
        </Tag>
      </template>
      <template #isShow_cell="{ row }">
        <Tag :color="getIsShowColor(row.isShow)">
          {{ isShowMap[row.isShow] || row.isShow }}
        </Tag>
      </template>
      <template #isCache_cell="{ row }">
        <Tag :color="getIsCacheColor(row.isCache)">
          {{ isCacheMap[row.isCache] || row.isCache }}
        </Tag>
      </template>
      <template #status_cell="{ row }">
        <Tag :color="getStatusColor(row.status)">
          {{ statusMap[row.status] || row.status }}
        </Tag>
      </template>
      <template #action="{ row }">
        <Button type="link" size="small" @click="handleEdit(row)">编辑</Button>
        <Button type="link" danger size="small" @click="handleDelete(row)">删除</Button>
      </template>
    </Grid>
  </Page>
</template>

<script setup lang="ts">
import type { VbenFormProps } from '#/adapter/form';
import type { VxeGridProps } from '#/adapter/vxe-table';

import { Page, useVbenModal } from '@vben/common-ui';
import { Button, message, Modal, Tag } from 'ant-design-vue';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import { getRoleTree, deleteRole } from '#/api/system/role';
import type { RoleItem } from '#/api/system/role/types';
import FormModal from './modules/form.vue';
import GrantMenuModal from './modules/grant-menu.vue';
import GrantDeptModal from './modules/grant-dept.vue';
import { ref } from 'vue';

/** 标签颜色池 */
const TAG_COLORS = ['green', 'red', 'blue', 'orange', 'cyan', 'purple', 'geekblue', 'magenta'];

/** 数据范围选项 */
const dataScopeOptions = [
  { label: '全部', value: 1 },
  { label: '本部门及以下', value: 2 },
  { label: '本部门', value: 3 },
  { label: '仅本人', value: 4 },
  { label: '自定义', value: 5 },
];

/** 数据范围映射 */
const dataScopeMap: Record<number, string> = {
  1: '全部',
  2: '本部门及以下',
  3: '本部门',
  4: '仅本人',
  5: '自定义',
};

/** 数据范围颜色 */
function getDataScopeColor(val: number): string {
  const keys = [1, 2, 3, 4, 5];
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

/** 超级管理员映射 */
const isAdminMap: Record<number, string> = {
  0: '否',
  1: '是',
};

/** 超级管理员颜色 */
function getIsAdminColor(val: number): string {
  return val === 1 ? 'red' : 'default';
}

const grantMenuRef = ref();
const grantDeptRef = ref();

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
        options: dataScopeOptions,
        placeholder: '请选择数据范围',
        class: 'w-full',
      },
      fieldName: 'dataScope',
      label: '数据范围',
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
const gridOptions: VxeGridProps<RoleItem> = {
  columns: [
    // { title: '序号', type: 'seq', width: 50 },
    { field: 'title', title: '角色名称', treeNode: true},
    { field: 'dataScope', title: '数据范围', width: 120, slots: { default: 'dataScope_cell' } },
    { field: 'sort', title: '排序（升序）' },
    { field: 'status', title: '状态', width: 120, slots: { default: 'status_cell' } },
    { field: 'isAdmin', title: '超级管理员', width: 120, slots: { default: 'isAdmin_cell' } },
    { field: 'createdAt', title: '创建时间', width: 180, formatter: 'formatDateTime' },
    { title: '操作', width: 320, fixed: 'right', slots: { default: 'action' } },
  ],
  pagerConfig: { enabled: false },
  treeConfig: {
    childrenField: 'children',
    expandAll: true,
  },
  proxyConfig: {
    ajax: {
      query: async (_params, formValues) => {
        const res = await getRoleTree(formValues);
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
function handleEdit(row: RoleItem) {
  formModalApi.setData({ id: row.id }).open();
}

/** 删除 */
function handleDelete(row: RoleItem) {
  Modal.confirm({
    title: '确认删除',
    content: '确定要删除该角色吗？',
    okType: 'danger',
    async onOk() {
      await deleteRole(row.id);
      message.success('删除成功');
      gridApi.reload();
    },
  });
}

/** 授权菜单 */
function handleGrantMenu(row: RoleItem) {
  grantMenuRef.value?.open(row.id);
}

/** 数据权限 */
function handleGrantDept(row: RoleItem) {
  grantDeptRef.value?.open(row.id, row.dataScope);
}
</script>

<template>
  <Page auto-content-height>
    <FormModalComp @success="() => gridApi.reload()" />
    <Grid>
      <template #toolbar-actions>
        <Button type="primary" @click="handleCreate">新建</Button>
      </template>
      <template #dataScope_cell="{ row }">
        <Tag :color="getDataScopeColor(row.dataScope)">
          {{ dataScopeMap[row.dataScope] || row.dataScope }}
        </Tag>
      </template>
      <template #status_cell="{ row }">
        <Tag :color="getStatusColor(row.status)">
          {{ statusMap[row.status] || row.status }}
        </Tag>
      </template>
      <template #isAdmin_cell="{ row }">
        <Tag :color="getIsAdminColor(row.isAdmin)">
          {{ isAdminMap[row.isAdmin] || '否' }}
        </Tag>
      </template>
      <template #action="{ row }">
        <Button type="link" size="small" @click="handleEdit(row)">编辑</Button>
        <Button type="link" size="small" @click="handleGrantMenu(row)">菜单权限</Button>
        <Button type="link" size="small" @click="handleGrantDept(row)">数据权限</Button>
        <Button type="link" danger size="small" @click="handleDelete(row)">删除</Button>
      </template>
    </Grid>
    <GrantMenuModal ref="grantMenuRef" @success="() => gridApi.reload()" />
    <GrantDeptModal ref="grantDeptRef" @success="() => gridApi.reload()" />
  </Page>
</template>

<script setup lang="ts">
import type { VbenFormProps } from '#/adapter/form';
import type { VxeGridProps } from '#/adapter/vxe-table';

import { h, onMounted, ref } from 'vue';

import { Page, useVbenModal } from '@vben/common-ui';
import { Button, Card, Input, message, Modal, Tag, Tree } from 'ant-design-vue';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import { getDeptTree } from '#/api/system/dept';
import type { DeptItem } from '#/api/system/dept/types';
import { getUsersList, deleteUsers, resetUsersPassword } from '#/api/system/users';
import type { UsersItem } from '#/api/system/users/types';
import FormModal from './modules/form.vue';

/** 标签颜色池 */
const TAG_COLORS = ['green', 'red', 'blue', 'orange', 'cyan', 'purple', 'geekblue', 'magenta'];

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

/** 部门树状态 */
const deptTree = ref<DeptItem[]>([]);
const searchValue = ref('');
const selectedDeptId = ref<string>('');
const deptExpandedKeys = ref<string[]>([]);

/** 递归收集所有节点 key */
function collectDeptKeys(nodes: DeptItem[]): string[] {
  const keys: string[] = [];
  for (const node of nodes) {
    keys.push(node.id);
    if (node.children?.length) {
      keys.push(...collectDeptKeys(node.children));
    }
  }
  return keys;
}

/** 加载部门树 */
onMounted(async () => {
  try {
    const res = await getDeptTree();
    deptTree.value = res ?? [];
    deptExpandedKeys.value = collectDeptKeys(deptTree.value);
  } catch {
    // ignore
  }
});

/** 过滤树节点 */
function filterTreeNode(node: any): boolean {
  if (!searchValue.value) return true;
  return String(node.title ?? '').toLowerCase().includes(searchValue.value.toLowerCase());
}

/** 选择部门节点 */
function handleDeptSelect(selectedKeys: string[]) {
  selectedDeptId.value = selectedKeys[0] ?? '';
  gridApi.reload();
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
      component: 'Input',
      componentProps: {
        allowClear: true,
        placeholder: '请输入登录用户名',
      },
      fieldName: 'username',
      label: '用户名',
    },
    {
      component: 'Input',
      componentProps: {
        allowClear: true,
        placeholder: '请输入昵称',
      },
      fieldName: 'nickname',
      label: '昵称',
    },
    {
      component: 'Input',
      componentProps: {
        allowClear: true,
        placeholder: '请输入邮箱',
      },
      fieldName: 'email',
      label: '邮箱',
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
const gridOptions: VxeGridProps<UsersItem> = {
  columns: [
    { title: '序号', type: 'seq', width: 50 },
    { field: 'username', title: '登录用户名' },
    { field: 'nickname', title: '昵称' },
    { field: 'deptTitle', title: '所属部门' },
    { field: 'roleTitles', title: '角色', width: 200, slots: { default: 'roleTitles_cell' } },
    { field: 'email', title: '邮箱' },
    { field: 'status', title: '状态', width: 120, slots: { default: 'status_cell' } },
    { field: 'createdAt', title: '创建时间', width: 180, formatter: 'formatDateTime' },
    { title: '操作', width: 260, fixed: 'right', slots: { default: 'action' } },
  ],
  height: 'auto',
  pagerConfig: {},
  proxyConfig: {
    ajax: {
      query: async ({ page }, formValues) => {
        const res = await getUsersList({
          pageNum: page.currentPage,
          pageSize: page.pageSize,
          ...formValues,
          ...(selectedDeptId.value ? { deptId: selectedDeptId.value } : {}),
        });
        return { items: res?.list ?? [], total: res?.total ?? 0 };
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
function handleEdit(row: UsersItem) {
  formModalApi.setData({ id: row.id }).open();
}

/** 删除 */
function handleDelete(row: UsersItem) {
  Modal.confirm({
    title: '确认删除',
    content: '确定要删除该用户吗？',
    okType: 'danger',
    async onOk() {
      await deleteUsers(row.id);
      message.success('删除成功');
      gridApi.reload();
    },
  });
}

/** 重置密码 */
function handleResetPassword(row: UsersItem) {
  let newPassword = '';
  Modal.confirm({
    title: '重置密码',
    content: () =>
      h(Input.Password, {
        placeholder: '请输入新密码',
        onChange: (e: any) => {
          newPassword = e.target.value;
        },
      }),
    async onOk() {
      if (!newPassword) {
        message.warning('请输入新密码');
        return Promise.reject();
      }
      await resetUsersPassword({ id: row.id, password: newPassword });
      message.success('密码重置成功');
    },
  });
}
</script>

<template>
  <Page auto-content-height>
    <FormModalComp @success="() => gridApi.reload()" />
    <div class="flex h-full gap-4">
      <Card class="w-[240px] shrink-0 overflow-auto" title="部门" size="small">
        <template #extra>
          <Input.Search
            v-model:value="searchValue"
            placeholder="搜索部门"
            size="small"
            allow-clear
            style="width: 140px"
          />
        </template>
        <Tree
          :tree-data="deptTree"
          :field-names="{ title: 'title', key: 'id', children: 'children' }"
          :selected-keys="selectedDeptId ? [selectedDeptId] : []"
          v-model:expanded-keys="deptExpandedKeys"
          :filter-tree-node="filterTreeNode"
          @select="handleDeptSelect"
        >
          <template #title="{ title }">
            <template v-if="searchValue && title.toLowerCase().includes(searchValue.toLowerCase())">
              <span>{{ title.slice(0, title.toLowerCase().indexOf(searchValue.toLowerCase())) }}</span>
              <span style="color: #f50; font-weight: 600">{{ title.slice(title.toLowerCase().indexOf(searchValue.toLowerCase()), title.toLowerCase().indexOf(searchValue.toLowerCase()) + searchValue.length) }}</span>
              <span>{{ title.slice(title.toLowerCase().indexOf(searchValue.toLowerCase()) + searchValue.length) }}</span>
            </template>
            <span v-else>{{ title }}</span>
          </template>
        </Tree>
      </Card>
      <div class="flex-1 overflow-hidden">
        <Grid>
          <template #toolbar-actions>
            <Button type="primary" @click="handleCreate">新建</Button>
          </template>
          <template #status_cell="{ row }">
            <Tag :color="getStatusColor(row.status)">
              {{ statusMap[row.status] || row.status }}
            </Tag>
          </template>
          <template #roleTitles_cell="{ row }">
            <Tag v-for="(name, idx) in (row.roleTitles || [])" :key="idx" :color="TAG_COLORS[idx % TAG_COLORS.length]">
              {{ name }}
            </Tag>
          </template>
          <template #action="{ row }">
            <Button type="link" size="small" @click="handleEdit(row)">编辑</Button>
            <Button type="link" size="small" @click="handleResetPassword(row)">重置密码</Button>
            <Button v-if="row.username !== 'admin'" type="link" danger size="small" @click="handleDelete(row)">删除</Button>
          </template>
        </Grid>
      </div>
    </div>
  </Page>
</template>

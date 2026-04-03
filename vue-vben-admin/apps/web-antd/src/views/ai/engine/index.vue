<script setup lang="ts">
import type { VxeGridProps } from '#/adapter/vxe-table';

import { Button, message, Tag } from 'ant-design-vue';
import { Page, useVbenModal } from '@vben/common-ui';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import { getEngineList, testEngineConnection } from '#/api/ai/engine';
import type { EngineItem } from '#/api/ai/engine/types';

import FormModal from './modules/form.vue';

const [FormModalComp, formModalApi] = useVbenModal({
  connectedComponent: FormModal,
  destroyOnClose: true,
});

function getStatusColor(status?: number) {
  return status === 1 ? 'green' : 'default';
}

const gridOptions: VxeGridProps<EngineItem> = {
  columns: [
    { field: 'name', title: '引擎名称', minWidth: 140 },
    { field: 'code', title: '编码', width: 140 },
    { field: 'status', title: '引擎状态', width: 120, slots: { default: 'status_cell' } },
    { field: 'configStatus', title: '配置状态', width: 120, slots: { default: 'configStatus_cell' } },
    { field: 'updatedAt', title: '更新时间', width: 180, formatter: 'formatDateTime' },
    { title: '操作', width: 220, fixed: 'right', slots: { default: 'action' } },
  ],
  pagerConfig: { enabled: false },
  proxyConfig: {
    ajax: {
      query: async () => {
        const res = await getEngineList();
        return res?.list ?? [];
      },
    },
  },
  toolbarConfig: {
    custom: true,
    refresh: true,
  },
};

const [Grid, gridApi] = useVbenVxeGrid({
  gridOptions,
});

function handleEdit(row: EngineItem) {
  formModalApi.setData({ engineCode: row.code, name: row.name }).open();
}

async function handleTest(row: EngineItem) {
  const res = await testEngineConnection(row.code);
  if (res?.success) {
    message.success(res.message || '连接成功');
  } else {
    message.error(res?.message || '连接失败');
  }
}
</script>

<template>
  <Page auto-content-height>
    <FormModalComp @success="() => gridApi.reload()" />
    <Grid>
      <template #status_cell="{ row }">
        <Tag :color="getStatusColor(row.status)">
          {{ row.status === 1 ? '启用' : '禁用' }}
        </Tag>
      </template>
      <template #configStatus_cell="{ row }">
        <Tag :color="getStatusColor(row.configStatus)">
          {{ row.configStatus === 1 ? '已配置' : '未启用' }}
        </Tag>
      </template>
      <template #action="{ row }">
        <Button type="link" size="small" @click="handleEdit(row)">配置</Button>
        <Button type="link" size="small" @click="handleTest(row)">测试连接</Button>
      </template>
    </Grid>
  </Page>
</template>

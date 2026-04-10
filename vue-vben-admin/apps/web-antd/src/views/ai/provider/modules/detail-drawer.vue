<script setup lang="ts">
import type { ProviderItem } from '#/api/ai/provider/types';

import { ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import { Descriptions, DescriptionsItem, Tag } from 'ant-design-vue';

import { getProviderDetail } from '#/api/ai/provider';

/** 状态映射 */
const statusMap: Record<number, string> = {
  0: '禁用',
  1: '启用',
};

function getStatusLabel(status?: number) {
  return status === undefined ? '-' : (statusMap[status] ?? status);
}

const detail = ref<null | ProviderItem>(null);

const [Modal, modalApi] = useVbenModal({
  fullscreenButton: false,
  footer: false,
  async onOpenChange(isOpen: boolean) {
    if (isOpen) {
      const data = modalApi.getData<{ id: string }>();
      if (data?.id) {
        modalApi.setState({ title: 'AI供应商表详情' });
        try {
          detail.value = await getProviderDetail(data.id);
        } catch {
          detail.value = null;
        }
      }
    } else {
      detail.value = null;
    }
  },
});
</script>

<template>
  <Modal class="w-[600px]">
    <Descriptions v-if="detail" bordered :column="1" size="small">
      <DescriptionsItem label="ID">{{ detail.id }}</DescriptionsItem>
      <DescriptionsItem label="供应商名称">{{ detail.name || '-' }}</DescriptionsItem>
      <DescriptionsItem label="供应商代码">{{ detail.code || '-' }}</DescriptionsItem>
      <DescriptionsItem label="Provider类型">{{ detail.providerType || '-' }}</DescriptionsItem>
      <DescriptionsItem label="API基础地址">{{ detail.baseURL || '-' }}</DescriptionsItem>
      <DescriptionsItem label="图标URL">{{ detail.icon || '-' }}</DescriptionsItem>
      <DescriptionsItem label="状态">
        <Tag>{{ getStatusLabel(detail.status) }}</Tag>
      </DescriptionsItem>
      <DescriptionsItem label="排序">{{ detail.sort || '-' }}</DescriptionsItem>
      <DescriptionsItem label="创建时间">{{ detail.createdAt || '-' }}</DescriptionsItem>
      <DescriptionsItem label="更新时间">{{ detail.updatedAt || '-' }}</DescriptionsItem>
    </Descriptions>
  </Modal>
</template>

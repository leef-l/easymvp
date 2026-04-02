<script setup lang="ts">
import { ref } from 'vue';
import { useVbenModal } from '@vben/common-ui';
import { Descriptions, DescriptionsItem, Tag } from 'ant-design-vue';
import { getPlanDetail } from '#/api/ai/plan';
import type { PlanItem } from '#/api/ai/plan/types';

/** 状态映射 */
const statusMap: Record<number, string> = {
  0: '禁用',
  1: '启用',
};

const detail = ref<PlanItem | null>(null);

const [Modal, modalApi] = useVbenModal({
  fullscreenButton: false,
  footer: false,
  async onOpenChange(isOpen: boolean) {
    if (isOpen) {
      const data = modalApi.getData<{ id: string }>();
      if (data?.id) {
        modalApi.setState({ title: 'AI套餐表详情' });
        try {
          detail.value = await getPlanDetail(data.id);
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
      <DescriptionsItem label="供应商ID">{{ detail.providerName || '-' }}</DescriptionsItem>
      <DescriptionsItem label="套餐名称">{{ detail.name || '-' }}</DescriptionsItem>
      <DescriptionsItem label="套餐代码">{{ detail.code || '-' }}</DescriptionsItem>
      <DescriptionsItem label="API Key">{{ detail.apiKey || '-' }}</DescriptionsItem>
      <DescriptionsItem label="API Secret">{{ detail.apiSecret || '-' }}</DescriptionsItem>
      <DescriptionsItem label="状态">
        <Tag>{{ statusMap[detail.status] || detail.status }}</Tag>
      </DescriptionsItem>
      <DescriptionsItem label="排序">{{ detail.sort || '-' }}</DescriptionsItem>
      <DescriptionsItem label="创建时间">{{ detail.createdAt || '-' }}</DescriptionsItem>
      <DescriptionsItem label="更新时间">{{ detail.updatedAt || '-' }}</DescriptionsItem>
    </Descriptions>
  </Modal>
</template>

<script setup lang="ts">
import type { ConfigItem } from '#/api/mvp/config/types';

import { ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import { Descriptions, DescriptionsItem } from 'ant-design-vue';

import { getConfigDetail } from '#/api/mvp/config';

const detail = ref<ConfigItem | null>(null);

const [Modal, modalApi] = useVbenModal({
  fullscreenButton: false,
  footer: false,
  async onOpenChange(isOpen: boolean) {
    if (isOpen) {
      const data = modalApi.getData<{ id: string }>();
      if (data?.id) {
        modalApi.setState({ title: 'MVP配置表详情' });
        try {
          detail.value = await getConfigDetail(data.id);
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
      <DescriptionsItem label="配置键">{{ detail.configKey || '-' }}</DescriptionsItem>
      <DescriptionsItem label="配置值">{{ detail.configValue || '-' }}</DescriptionsItem>
      <DescriptionsItem label="值类型">{{ detail.configType || '-' }}</DescriptionsItem>
      <DescriptionsItem label="分类">{{ detail.category || '-' }}</DescriptionsItem>
      <DescriptionsItem label="配置说明">{{ detail.description || '-' }}</DescriptionsItem>
      <DescriptionsItem label="创建时间">{{ detail.createdAt || '-' }}</DescriptionsItem>
      <DescriptionsItem label="更新时间">{{ detail.updatedAt || '-' }}</DescriptionsItem>
    </Descriptions>
  </Modal>
</template>

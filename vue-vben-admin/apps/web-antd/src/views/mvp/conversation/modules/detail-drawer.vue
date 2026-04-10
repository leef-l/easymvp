<script setup lang="ts">
import type { ConversationItem } from '#/api/mvp/conversation/types';

import { ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import { Descriptions, DescriptionsItem } from 'ant-design-vue';

import { getConversationDetail } from '#/api/mvp/conversation';

const detail = ref<ConversationItem | null>(null);

const [Modal, modalApi] = useVbenModal({
  fullscreenButton: false,
  footer: false,
  async onOpenChange(isOpen: boolean) {
    if (isOpen) {
      const data = modalApi.getData<{ id: string }>();
      if (data?.id) {
        modalApi.setState({ title: 'MVP对话表详情' });
        try {
          detail.value = await getConversationDetail(data.id);
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
      <DescriptionsItem label="项目ID">{{ detail.projectName || '-' }}</DescriptionsItem>
      <DescriptionsItem label="关联任务ID，NULL=项目级对话">{{ detail.taskName || '-' }}</DescriptionsItem>
      <DescriptionsItem label="对话标题">{{ detail.title || '-' }}</DescriptionsItem>
      <DescriptionsItem label="对话角色类型">{{ detail.roleType || '-' }}</DescriptionsItem>
      <DescriptionsItem label="状态">{{ detail.status || '-' }}</DescriptionsItem>
      <DescriptionsItem label="创建时间">{{ detail.createdAt || '-' }}</DescriptionsItem>
      <DescriptionsItem label="更新时间">{{ detail.updatedAt || '-' }}</DescriptionsItem>
    </Descriptions>
  </Modal>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useVbenModal } from '@vben/common-ui';
import { Descriptions, DescriptionsItem, Tag } from 'ant-design-vue';
import { getMessageDetail } from '#/api/mvp/message';
import type { MessageItem } from '#/api/mvp/message/types';

const detail = ref<MessageItem | null>(null);
const messageTypeMeta: Record<string, { color: string; text: string }> = {
  chat_user: { color: 'blue', text: '普通对话-用户' },
  chat_reply: { color: 'cyan', text: '普通对话-AI' },
  task_prompt: { color: 'gold', text: '任务指令' },
  task_reply: { color: 'green', text: '任务回复' },
  system_notice: { color: 'purple', text: '系统通知' },
  poison: { color: 'red', text: '毒药消息' },
  general: { color: 'default', text: '通用' },
};

const [Modal, modalApi] = useVbenModal({
  fullscreenButton: false,
  footer: false,
  async onOpenChange(isOpen: boolean) {
    if (isOpen) {
      const data = modalApi.getData<{ id: string }>();
      if (data?.id) {
        modalApi.setState({ title: 'MVP消息表详情' });
        try {
          detail.value = await getMessageDetail(data.id);
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
      <DescriptionsItem label="对话ID">{{ detail.conversationTitle || '-' }}</DescriptionsItem>
      <DescriptionsItem label="消息角色">{{ detail.role || '-' }}</DescriptionsItem>
      <DescriptionsItem label="消息类型">
        <Tag :color="messageTypeMeta[detail.messageType || 'general']?.color || 'default'">
          {{ messageTypeMeta[detail.messageType || 'general']?.text || detail.messageType || '-' }}
        </Tag>
      </DescriptionsItem>
      <DescriptionsItem label="消息内容">{{ detail.content || '-' }}</DescriptionsItem>
      <DescriptionsItem label="使用的AI模型ID">{{ detail.modelID || '-' }}</DescriptionsItem>
      <DescriptionsItem label="token消耗">{{ detail.tokenUsage || '-' }}</DescriptionsItem>
      <DescriptionsItem label="状态">{{ detail.status || '-' }}</DescriptionsItem>
      <DescriptionsItem label="创建时间">{{ detail.createdAt || '-' }}</DescriptionsItem>
      <DescriptionsItem label="更新时间">{{ detail.updatedAt || '-' }}</DescriptionsItem>
    </Descriptions>
  </Modal>
</template>

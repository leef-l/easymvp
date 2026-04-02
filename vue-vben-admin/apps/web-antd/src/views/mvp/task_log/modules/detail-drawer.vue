<script setup lang="ts">
import { ref } from 'vue';
import { useVbenModal } from '@vben/common-ui';
import { Descriptions, DescriptionsItem, Tag } from 'ant-design-vue';
import { getTaskLogDetail } from '#/api/mvp/task_log';
import type { TaskLogItem } from '#/api/mvp/task_log/types';

const detail = ref<TaskLogItem | null>(null);

const [Modal, modalApi] = useVbenModal({
  fullscreenButton: false,
  footer: false,
  async onOpenChange(isOpen: boolean) {
    if (isOpen) {
      const data = modalApi.getData<{ id: string }>();
      if (data?.id) {
        modalApi.setState({ title: '任务日志表详情' });
        try {
          detail.value = await getTaskLogDetail(data.id);
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
      <DescriptionsItem label="任务ID">{{ detail.taskName || '-' }}</DescriptionsItem>
      <DescriptionsItem label="动作">{{ detail.action || '-' }}</DescriptionsItem>
      <DescriptionsItem label="原状态">{{ detail.fromStatus || '-' }}</DescriptionsItem>
      <DescriptionsItem label="新状态">{{ detail.toStatus || '-' }}</DescriptionsItem>
      <DescriptionsItem label="日志内容">{{ detail.message || '-' }}</DescriptionsItem>
      <DescriptionsItem label="操作者">{{ detail.operator || '-' }}</DescriptionsItem>
      <DescriptionsItem label="创建时间">{{ detail.createdAt || '-' }}</DescriptionsItem>
      <DescriptionsItem label="更新时间">{{ detail.updatedAt || '-' }}</DescriptionsItem>
    </Descriptions>
  </Modal>
</template>

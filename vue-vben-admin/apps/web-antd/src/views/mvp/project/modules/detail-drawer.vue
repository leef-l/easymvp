<script setup lang="ts">
import type { ProjectItem } from '#/api/mvp/project/types';

import { ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import { Descriptions, DescriptionsItem } from 'ant-design-vue';

import { getProjectDetail } from '#/api/mvp/project';

const detail = ref<null | ProjectItem>(null);

const [Modal, modalApi] = useVbenModal({
  fullscreenButton: false,
  footer: false,
  async onOpenChange(isOpen: boolean) {
    if (isOpen) {
      const data = modalApi.getData<{ id: string }>();
      if (data?.id) {
        modalApi.setState({ title: 'MVP项目表详情' });
        try {
          detail.value = await getProjectDetail(data.id);
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
      <DescriptionsItem label="项目名称">{{ detail.name || '-' }}</DescriptionsItem>
      <DescriptionsItem label="项目简介">{{ detail.description || '-' }}</DescriptionsItem>
      <DescriptionsItem label="状态">{{ detail.status || '-' }}</DescriptionsItem>
      <DescriptionsItem label="暂停原因">{{ detail.pauseReason || '-' }}</DescriptionsItem>
      <DescriptionsItem label="项目全局上下文">{{ detail.globalContext || '-' }}</DescriptionsItem>
      <DescriptionsItem label="架构师AI模型">{{ detail.architectModelName || detail.architectModelID || '-' }}</DescriptionsItem>
      <DescriptionsItem label="创建时间">{{ detail.createdAt || '-' }}</DescriptionsItem>
      <DescriptionsItem label="更新时间">{{ detail.updatedAt || '-' }}</DescriptionsItem>
    </Descriptions>
  </Modal>
</template>

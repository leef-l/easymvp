<script setup lang="ts">
import type { ModelItem } from '#/api/ai/model/types';

import { ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import { Descriptions, DescriptionsItem, Tag } from 'ant-design-vue';

import { getModelDetail } from '#/api/ai/model';
import { roleTypeMap as capabilityMap } from '#/views/mvp/consts';

/** 是否支持流式输出映射 */
const supportsStreamMap: Record<number, string> = {
  0: '否',
  1: '是',
};

/** 状态映射 */
const statusMap: Record<number, string> = {
  0: '禁用',
  1: '启用',
};

function getCapabilityMeta(capability?: string) {
  return capability ? capabilityMap[capability] : undefined;
}

function getStatusLabel(status?: number) {
  return status === undefined ? '-' : (statusMap[status] ?? status);
}

function getSupportsStreamLabel(supportsStream?: number) {
  return supportsStream === undefined
    ? '-'
    : (supportsStreamMap[supportsStream] ?? supportsStream);
}

const detail = ref<ModelItem | null>(null);

const [Modal, modalApi] = useVbenModal({
  fullscreenButton: false,
  footer: false,
  async onOpenChange(isOpen: boolean) {
    if (isOpen) {
      const data = modalApi.getData<{ id: string }>();
      if (data?.id) {
        modalApi.setState({ title: 'AI模型表详情' });
        try {
          detail.value = await getModelDetail(data.id);
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
      <DescriptionsItem label="套餐ID">{{ detail.planName || '-' }}</DescriptionsItem>
      <DescriptionsItem label="供应商ID">{{ detail.providerName || '-' }}</DescriptionsItem>
      <DescriptionsItem label="模型显示名称">{{ detail.name || '-' }}</DescriptionsItem>
      <DescriptionsItem label="模型代码">{{ detail.modelCode || '-' }}</DescriptionsItem>
      <DescriptionsItem label="项目角色">
        <Tag v-if="getCapabilityMeta(detail.capability)" :color="getCapabilityMeta(detail.capability)?.color">
          {{ getCapabilityMeta(detail.capability)?.label }}
        </Tag>
        <span v-else>{{ detail.capability || '-' }}</span>
      </DescriptionsItem>
      <DescriptionsItem label="最大输出token">{{ detail.maxTokens || '-' }}</DescriptionsItem>
      <DescriptionsItem label="上下文窗口大小">{{ detail.contextWindow || '-' }}</DescriptionsItem>
      <DescriptionsItem label="是否支持流式输出">
        <Tag>{{ getSupportsStreamLabel(detail.supportsStream) }}</Tag>
      </DescriptionsItem>
      <DescriptionsItem label="默认角色提示词">{{ detail.rolePrompt || '-' }}</DescriptionsItem>
      <DescriptionsItem label="状态">
        <Tag>{{ getStatusLabel(detail.status) }}</Tag>
      </DescriptionsItem>
      <DescriptionsItem label="排序">{{ detail.sort || '-' }}</DescriptionsItem>
      <DescriptionsItem label="创建时间">{{ detail.createdAt || '-' }}</DescriptionsItem>
      <DescriptionsItem label="更新时间">{{ detail.updatedAt || '-' }}</DescriptionsItem>
    </Descriptions>
  </Modal>
</template>

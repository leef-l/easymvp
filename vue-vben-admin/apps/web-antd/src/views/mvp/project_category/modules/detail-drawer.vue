<script setup lang="ts">
import type { ProjectCategoryItem } from '#/api/mvp/project_category/types';

import { ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import { Descriptions, DescriptionsItem, Divider } from 'ant-design-vue';

import { getProjectCategoryDetail } from '#/api/mvp/project_category';

const detail = ref<null | ProjectCategoryItem>(null);

const [Modal, modalApi] = useVbenModal({
  fullscreenButton: false,
  footer: false,
  async onOpenChange(isOpen: boolean) {
    if (isOpen) {
      const data = modalApi.getData<{ id: string }>();
      if (data?.id) {
        modalApi.setState({ title: '项目分类配置表详情' });
        try {
          detail.value = await getProjectCategoryDetail(data.id);
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
  <Modal class="w-[960px]">
    <template v-if="detail">
      <Descriptions bordered :column="1" size="small">
        <DescriptionsItem label="ID">{{ detail.id }}</DescriptionsItem>
        <DescriptionsItem label="稳定分类编码">{{ detail.categoryCode || '-' }}</DescriptionsItem>
        <DescriptionsItem label="展示名称">{{ detail.displayName || '-' }}</DescriptionsItem>
        <DescriptionsItem label="能力家族编码">{{ detail.familyCode || '-' }}</DescriptionsItem>
        <DescriptionsItem label="分类说明">{{ detail.description || '-' }}</DescriptionsItem>
        <DescriptionsItem label="1启用 0停用">{{ detail.status || '-' }}</DescriptionsItem>
        <DescriptionsItem label="排序">{{ detail.sort || '-' }}</DescriptionsItem>
        <DescriptionsItem label="创建时间">{{ detail.createdAt || '-' }}</DescriptionsItem>
        <DescriptionsItem label="更新时间">{{ detail.updatedAt || '-' }}</DescriptionsItem>
      </Descriptions>

      <Divider orientation="left" class="mt-4">分类默认验证配置</Divider>
      <div class="rounded bg-gray-50 p-3 text-sm text-gray-700 whitespace-pre-wrap break-all min-h-[72px] font-mono">
        {{ detail.verificationProfileJson || '（未配置，回退项目级 .easymvp/verification.json 或自动探测）' }}
      </div>

      <Divider orientation="left">分类验证放行规则</Divider>
      <div class="rounded bg-gray-50 p-3 text-sm text-gray-700 whitespace-pre-wrap break-all min-h-[72px] font-mono">
        {{ detail.verificationGateJson || '（未配置，使用系统默认放行逻辑）' }}
      </div>
    </template>
  </Modal>
</template>

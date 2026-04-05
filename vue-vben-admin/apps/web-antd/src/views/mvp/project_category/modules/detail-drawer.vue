<script setup lang="ts">
import { ref } from 'vue';
import { useVbenModal } from '@vben/common-ui';
import { Descriptions, DescriptionsItem, Tag } from 'ant-design-vue';
import { getProjectCategoryDetail } from '#/api/mvp/project_category';
import type { ProjectCategoryItem } from '#/api/mvp/project_category/types';

/** 状态映射 */
const statusMap: Record<number, string> = {
  0: '禁用',
  1: '启用',
};

function getStatusColor(val: number): string {
  const colorMap: Record<number, string> = { 0: 'red', 1: 'green' };
  return colorMap[val] ?? 'default';
}

const detail = ref<ProjectCategoryItem | null>(null);

const [Modal, modalApi] = useVbenModal({
  fullscreenButton: false,
  footer: false,
  async onOpenChange(isOpen: boolean) {
    if (isOpen) {
      const data = modalApi.getData<{ id: string }>();
      if (data?.id) {
        modalApi.setState({ title: '项目分类详情' });
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
  <Modal class="w-[560px]">
    <Descriptions v-if="detail" bordered :column="1" size="small">
      <DescriptionsItem label="ID">{{ detail.id }}</DescriptionsItem>
      <DescriptionsItem label="分类代码">{{ detail.categoryCode || '-' }}</DescriptionsItem>
      <DescriptionsItem label="显示名称">{{ detail.displayName || '-' }}</DescriptionsItem>
      <DescriptionsItem label="所属系">{{ detail.familyCode || '-' }}</DescriptionsItem>
      <DescriptionsItem label="描述">{{ detail.description || '-' }}</DescriptionsItem>
      <DescriptionsItem label="状态">
        <Tag :color="getStatusColor(detail.status)">
          {{ statusMap[detail.status] ?? detail.status }}
        </Tag>
      </DescriptionsItem>
      <DescriptionsItem label="排序">{{ detail.sort ?? '-' }}</DescriptionsItem>
      <DescriptionsItem label="创建时间">{{ detail.createdAt || '-' }}</DescriptionsItem>
      <DescriptionsItem label="更新时间">{{ detail.updatedAt || '-' }}</DescriptionsItem>
    </Descriptions>
  </Modal>
</template>

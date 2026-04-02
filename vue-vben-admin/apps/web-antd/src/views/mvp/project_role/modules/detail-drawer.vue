<script setup lang="ts">
import { ref } from 'vue';
import { useVbenModal } from '@vben/common-ui';
import { Descriptions, DescriptionsItem, Tag } from 'ant-design-vue';
import { getProjectRoleDetail } from '#/api/mvp/project_role';
import type { ProjectRoleItem } from '#/api/mvp/project_role/types';
import { roleTypeMap, roleLevelMap } from '../../consts';

/** 状态映射 */
const statusMap: Record<number, string> = {
  0: '禁用',
  1: '启用',
};

const detail = ref<ProjectRoleItem | null>(null);

const [Modal, modalApi] = useVbenModal({
  fullscreenButton: false,
  footer: false,
  async onOpenChange(isOpen: boolean) {
    if (isOpen) {
      const data = modalApi.getData<{ id: string }>();
      if (data?.id) {
        modalApi.setState({ title: '项目角色配置表详情' });
        try {
          detail.value = await getProjectRoleDetail(data.id);
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
      <DescriptionsItem label="所属项目">{{ detail.projectName || '-' }}</DescriptionsItem>
      <DescriptionsItem label="角色类型">
        <Tag v-if="roleTypeMap[detail.roleType]" :color="roleTypeMap[detail.roleType].color">
          {{ roleTypeMap[detail.roleType].label }}
        </Tag>
        <span v-else>{{ detail.roleType || '-' }}</span>
      </DescriptionsItem>
      <DescriptionsItem label="角色等级">
        <Tag v-if="roleLevelMap[detail.roleLevel]" :color="roleLevelMap[detail.roleLevel].color">
          {{ roleLevelMap[detail.roleLevel].label }}
        </Tag>
        <span v-else>{{ detail.roleLevel || '-' }}</span>
      </DescriptionsItem>
      <DescriptionsItem label="AI模型">{{ detail.modelName || '-' }}</DescriptionsItem>
      <DescriptionsItem label="系统提示词">{{ detail.systemPrompt || '-' }}</DescriptionsItem>
      <DescriptionsItem label="状态">
        <Tag>{{ statusMap[detail.status] || detail.status }}</Tag>
      </DescriptionsItem>
      <DescriptionsItem label="创建时间">{{ detail.createdAt || '-' }}</DescriptionsItem>
      <DescriptionsItem label="更新时间">{{ detail.updatedAt || '-' }}</DescriptionsItem>
    </Descriptions>
  </Modal>
</template>

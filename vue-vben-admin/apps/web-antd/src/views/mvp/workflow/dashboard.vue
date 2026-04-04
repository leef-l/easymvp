<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useRoute } from 'vue-router';

import { Page } from '@vben/common-ui';
import { Card, Descriptions, DescriptionsItem, Tag, Empty, Spin } from 'ant-design-vue';

import { workflowRunStatusMap, stageTypeMap } from '../consts';

const route = useRoute();
const loading = ref(false);
const projectId = ref<string>('');
const workflowRunId = ref<string>('');

// 工作流概览数据（M2 阶段接入真实 API）
const workflowDetail = ref<Record<string, any> | null>(null);

onMounted(() => {
  projectId.value = (route.query.projectId as string) ?? '';
  workflowRunId.value = (route.query.workflowRunId as string) ?? '';

  if (workflowRunId.value) {
    loadWorkflowDetail();
  }
});

async function loadWorkflowDetail() {
  loading.value = true;
  try {
    // TODO: M2 阶段接入真实 API
    // const res = await getWorkflowRunDetail(workflowRunId.value);
    // workflowDetail.value = res;
    workflowDetail.value = null; // 占位
  } finally {
    loading.value = false;
  }
}
</script>

<template>
  <Page auto-content-height>
    <Spin :spinning="loading">
      <!-- 无数据占位 -->
      <Empty
        v-if="!workflowDetail"
        description="Workflow Dashboard — 工作流控制台（M2 阶段实现）"
        class="mt-20"
      />

      <!-- 工作流概览（M2 阶段渲染真实数据） -->
      <template v-else>
        <Card title="工作流概览" class="mb-4">
          <Descriptions :column="3" bordered size="small">
            <DescriptionsItem label="工作流ID">
              {{ workflowDetail.workflowRunId }}
            </DescriptionsItem>
            <DescriptionsItem label="状态">
              <Tag :color="workflowRunStatusMap[workflowDetail.status]?.color ?? 'default'">
                {{ workflowRunStatusMap[workflowDetail.status]?.label ?? workflowDetail.status }}
              </Tag>
            </DescriptionsItem>
            <DescriptionsItem label="当前阶段">
              <Tag :color="stageTypeMap[workflowDetail.currentStage]?.color ?? 'default'">
                {{ stageTypeMap[workflowDetail.currentStage]?.label ?? workflowDetail.currentStage }}
              </Tag>
            </DescriptionsItem>
          </Descriptions>
        </Card>
      </template>
    </Spin>
  </Page>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { useRoute } from 'vue-router';

import { Page } from '@vben/common-ui';
import {
  Card,
  Descriptions,
  DescriptionsItem,
  Tag,
  Empty,
  Spin,
  Progress,
  Statistic,
  Row,
  Col,
} from 'ant-design-vue';

import {
  getProjectStatus,
  type ProjectStatusResult,
} from '../../../api/mvp/workflow';
import { workflowRunStatusMap, stageTypeMap } from '../consts';

const route = useRoute();
const loading = ref(false);
const projectId = ref<string>('');

const statusData = ref<ProjectStatusResult | null>(null);

const isV2 = computed(() => statusData.value?.engineVersion === 'workflow_v2');

onMounted(() => {
  projectId.value = (route.query.projectId as string) ?? '';
  if (projectId.value) {
    loadStatus();
  }
});

async function loadStatus() {
  loading.value = true;
  try {
    const res = await getProjectStatus(projectId.value);
    statusData.value = res;
  } finally {
    loading.value = false;
  }
}
</script>

<template>
  <Page auto-content-height>
    <Spin :spinning="loading">
      <Empty
        v-if="!statusData"
        description="请从项目列表进入查看工作流状态"
        class="mt-20"
      />

      <template v-else>
        <!-- 工作流概览 -->
        <Card title="工作流概览" class="mb-4">
          <Descriptions :column="3" bordered size="small">
            <DescriptionsItem label="项目ID">
              {{ projectId }}
            </DescriptionsItem>
            <DescriptionsItem label="引擎版本">
              <Tag :color="isV2 ? 'purple' : 'default'">
                {{ isV2 ? 'Workflow V2' : 'Legacy' }}
              </Tag>
            </DescriptionsItem>
            <DescriptionsItem label="状态">
              <Tag
                :color="
                  workflowRunStatusMap[statusData.workflowStatus || statusData.status]?.color ??
                  'default'
                "
              >
                {{
                  workflowRunStatusMap[statusData.workflowStatus || statusData.status]?.label ??
                  statusData.status
                }}
              </Tag>
            </DescriptionsItem>
            <DescriptionsItem v-if="isV2 && statusData.currentStage" label="当前阶段">
              <Tag :color="stageTypeMap[statusData.currentStage]?.color ?? 'default'">
                {{ stageTypeMap[statusData.currentStage]?.label ?? statusData.currentStage }}
              </Tag>
            </DescriptionsItem>
            <DescriptionsItem v-if="statusData.pauseReason" label="暂停原因">
              {{ statusData.pauseReason }}
            </DescriptionsItem>
          </Descriptions>
        </Card>

        <!-- 任务统计 -->
        <Card title="任务进度" class="mb-4">
          <Row :gutter="16" class="mb-4">
            <Col :span="6">
              <Statistic title="总任务数" :value="statusData.totalTasks" />
            </Col>
            <Col :span="6">
              <Statistic
                title="已完成"
                :value="statusData.statusCounts?.['domain_completed'] ?? statusData.statusCounts?.['completed'] ?? 0"
                value-style="color: #3f8600"
              />
            </Col>
            <Col :span="6">
              <Statistic
                title="运行中"
                :value="statusData.statusCounts?.['domain_running'] ?? statusData.activeRunningTasks ?? 0"
                value-style="color: #1890ff"
              />
            </Col>
            <Col :span="6">
              <Statistic
                title="失败"
                :value="statusData.statusCounts?.['domain_failed'] ?? statusData.statusCounts?.['failed'] ?? 0"
                value-style="color: #cf1322"
              />
            </Col>
          </Row>

          <Progress
            v-if="statusData.totalTasks > 0"
            :percent="statusData.progressPercent ?? Math.round(((statusData.statusCounts?.['domain_completed'] ?? statusData.statusCounts?.['completed'] ?? 0) / statusData.totalTasks) * 100)"
            :status="statusData.status === 'completed' ? 'success' : statusData.status === 'failed' ? 'exception' : 'active'"
          />
        </Card>

        <!-- Legacy 额外信息 -->
        <Card v-if="!isV2 && statusData.activeBatch > 0" title="调度信息" class="mb-4">
          <Descriptions :column="3" bordered size="small">
            <DescriptionsItem label="活跃批次">
              {{ statusData.activeBatch }}
            </DescriptionsItem>
            <DescriptionsItem label="卡住任务数">
              {{ statusData.stalledTaskCount }}
            </DescriptionsItem>
            <DescriptionsItem label="实际工作中">
              <Tag :color="statusData.isActuallyWorking ? 'green' : 'default'">
                {{ statusData.isActuallyWorking ? '是' : '否' }}
              </Tag>
            </DescriptionsItem>
          </Descriptions>
        </Card>
      </template>
    </Spin>
  </Page>
</template>

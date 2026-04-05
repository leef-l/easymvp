<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';

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
  Steps,
  Step,
  Button,
  Space,
  Alert,
  Divider,
} from 'ant-design-vue';

import {
  getProjectStatus,
  getStageHistory,
  getCompletionSummary,
  type ProjectStatusResult,
  type StageHistoryItem,
  type CompletionSummaryResult,
} from '../../../api/mvp/workflow';
import { workflowRunStatusMap, stageTypeMap } from '../consts';

const route = useRoute();
const router = useRouter();
const loading = ref(false);
const projectId = ref<string>('');

const statusData = ref<ProjectStatusResult | null>(null);
const stages = ref<StageHistoryItem[]>([]);
const summary = ref<CompletionSummaryResult | null>(null);

const isV2 = computed(() => statusData.value?.engineVersion === 'workflow_v2');
const currentStatus = computed(() => statusData.value?.workflowStatus || statusData.value?.status || '');
const isCompleted = computed(() => currentStatus.value === 'completed');

// 阶段进度 Steps 的 current index
const stageStepCurrent = computed(() => {
  if (!stages.value.length) return 0;
  const order = ['design', 'review', 'execute', 'rework', 'accept', 'complete'];
  const currentStage = statusData.value?.currentStage || '';
  const idx = order.indexOf(currentStage);
  return idx >= 0 ? idx : stages.value.length - 1;
});

// 阶段步骤状态
function stageStepStatus(stage: string): 'error' | 'finish' | 'process' | 'wait' {
  const stageRun = stages.value.find((s) => s.stageType === stage);
  if (!stageRun) return 'wait';
  if (stageRun.status === 'completed') return 'finish';
  if (stageRun.status === 'failed') return 'error';
  if (stageRun.status === 'running') return 'process';
  return 'wait';
}

onMounted(async () => {
  projectId.value = (route.query.projectId as string) ?? '';
  if (projectId.value) {
    await loadAll();
  }
});

async function loadAll() {
  loading.value = true;
  try {
    const [statusRes, stageRes] = await Promise.all([
      getProjectStatus(projectId.value),
      getStageHistory(projectId.value),
    ]);
    statusData.value = statusRes;
    stages.value = stageRes.stages || [];

    // 已完成时加载总结
    if (statusRes.workflowStatus === 'completed' || statusRes.status === 'completed') {
      try {
        summary.value = await getCompletionSummary(projectId.value);
      } catch {
        // 总结可能尚未生成
      }
    }
  } finally {
    loading.value = false;
  }
}

function goToReview() {
  router.push({ path: '/mvp/workflow/review', query: { projectId: projectId.value } });
}
function goToExecution() {
  router.push({ path: '/mvp/workflow/execution', query: { projectId: projectId.value } });
}
function goToRework() {
  router.push({ path: '/mvp/workflow/rework', query: { projectId: projectId.value } });
}
function goToAccept() {
  router.push({ path: '/mvp/workflow/accept', query: { projectId: projectId.value } });
}
function goToAutonomy() {
  router.push({ path: '/mvp/workflow/autonomy', query: { projectId: projectId.value } });
}
function goToTimeline() {
  router.push({ path: '/mvp/workflow/timeline', query: { projectId: projectId.value } });
}
function goToFeishu() {
  router.push({ path: '/mvp/workflow/feishu' });
}
function goToChat() {
  router.push({ path: '/mvp/chat', query: { projectId: projectId.value } });
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
        <Card class="mb-4">
          <template #title>
            <Space>
              <span>工作流仪表板</span>
              <Tag :color="isV2 ? 'green' : 'default'" size="small">
                {{ isV2 ? 'V2' : 'Legacy' }}
              </Tag>
            </Space>
          </template>
          <template #extra>
            <Button type="link" size="small" @click="goToTimeline">事件时间线</Button>
          </template>

          <Row :gutter="24">
            <Col :span="8">
              <Descriptions :column="1" size="small">
                <DescriptionsItem label="项目状态">
                  <Tag
                    :color="workflowRunStatusMap[currentStatus]?.color ?? 'default'"
                  >
                    {{ workflowRunStatusMap[currentStatus]?.label ?? currentStatus }}
                  </Tag>
                </DescriptionsItem>
                <DescriptionsItem v-if="statusData.currentStage" label="当前阶段">
                  <Tag :color="stageTypeMap[statusData.currentStage]?.color ?? 'default'">
                    {{ stageTypeMap[statusData.currentStage]?.label ?? statusData.currentStage }}
                  </Tag>
                </DescriptionsItem>
                <DescriptionsItem v-if="statusData.pauseReason" label="暂停原因">
                  <span class="text-orange-500">{{ statusData.pauseReason }}</span>
                </DescriptionsItem>
              </Descriptions>
            </Col>
            <Col :span="16">
              <Row :gutter="16">
                <Col :span="6">
                  <Statistic title="总任务" :value="statusData.totalTasks" />
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
                :status="isCompleted ? 'success' : currentStatus === 'failed' ? 'exception' : 'active'"
                class="mt-2"
              />
            </Col>
          </Row>
        </Card>

        <!-- V2 阶段流程图 -->
        <Card v-if="isV2 && stages.length > 0" title="阶段流程" class="mb-4">
          <Steps :current="stageStepCurrent" size="small">
            <Step title="设计" :status="stageStepStatus('design')" />
            <Step title="审核" :status="stageStepStatus('review')" />
            <Step title="执行" :status="stageStepStatus('execute')" />
            <Step
              v-if="stages.some((s) => s.stageType === 'rework')"
              title="返工"
              :status="stageStepStatus('rework')"
            />
            <Step title="验收" :status="stageStepStatus('accept')" />
            <Step title="完成" :status="stageStepStatus('complete')" />
          </Steps>

          <Divider />

          <!-- 阶段详情表 -->
          <Descriptions :column="{ xs: 1, sm: 2, md: 3 }" size="small" bordered>
            <DescriptionsItem
              v-for="stage in stages"
              :key="stage.id"
              :label="(stageTypeMap[stage.stageType]?.label ?? stage.stageType) + ' #' + stage.stageNo"
            >
              <Tag :color="stage.status === 'completed' ? 'green' : stage.status === 'failed' ? 'red' : stage.status === 'running' ? 'blue' : 'default'" size="small">
                {{ stage.status }}
              </Tag>
              <span v-if="stage.startedAt" class="text-gray-400 text-xs ml-2">
                {{ stage.startedAt }}
              </span>
              <span v-if="stage.error" class="text-red-500 text-xs ml-2">
                {{ stage.error }}
              </span>
            </DescriptionsItem>
          </Descriptions>
        </Card>

        <!-- 快速操作 -->
        <Card v-if="isV2" title="快速操作" class="mb-4" size="small">
          <Space wrap>
            <Button @click="goToChat">
              进入对话
            </Button>
            <Button
              v-if="currentStatus === 'reviewing'"
              type="primary"
              @click="goToReview"
            >
              审核工作台
            </Button>
            <Button
              v-if="currentStatus === 'executing' || currentStatus === 'reworking'"
              type="primary"
              @click="goToExecution"
            >
              执行控制台
            </Button>
            <Button
              v-if="stages.some((s) => s.stageType === 'rework')"
              @click="goToRework"
            >
              返工记录
            </Button>
            <Button
              v-if="currentStatus === 'accepting' || stages.some((s) => s.stageType === 'accept')"
              type="primary"
              @click="goToAccept"
            >
              验收控制台
            </Button>
            <Button @click="goToAutonomy">
              自治控制台
            </Button>
            <Button @click="goToFeishu">
              飞书协作
            </Button>
            <Button @click="goToTimeline">
              事件时间线
            </Button>
          </Space>
        </Card>

        <!-- 完成总结 -->
        <Card v-if="isCompleted && summary" title="完成总结" class="mb-4">
          <Row :gutter="16" class="mb-4">
            <Col :span="4">
              <Statistic title="总任务" :value="summary.totalTasks" />
            </Col>
            <Col :span="4">
              <Statistic title="成功" :value="summary.completedTasks" value-style="color: #3f8600" />
            </Col>
            <Col :span="4">
              <Statistic title="失败" :value="summary.failedTasks" value-style="color: #cf1322" />
            </Col>
            <Col :span="4">
              <Statistic title="跳过" :value="summary.skippedTasks" />
            </Col>
            <Col :span="4">
              <Statistic title="成功率" :value="summary.successRate" suffix="%" />
            </Col>
            <Col :span="4">
              <Statistic title="返工轮次" :value="summary.reworkRounds" />
            </Col>
          </Row>

          <Descriptions :column="3" size="small" bordered>
            <DescriptionsItem label="总耗时">{{ summary.totalDuration }}</DescriptionsItem>
            <DescriptionsItem label="平均任务耗时">{{ summary.avgTaskDuration }}</DescriptionsItem>
            <DescriptionsItem label="开始时间">{{ summary.startedAt }}</DescriptionsItem>
            <DescriptionsItem label="完成时间">{{ summary.finishedAt }}</DescriptionsItem>
            <DescriptionsItem label="升级任务数">{{ summary.escalatedTasks }}</DescriptionsItem>
            <DescriptionsItem label="交接次数">{{ summary.handoffCount }}</DescriptionsItem>
          </Descriptions>

          <div v-if="summary.stageDurations && Object.keys(summary.stageDurations).length" class="mt-4">
            <h4 class="text-sm font-medium mb-2">各阶段耗时</h4>
            <Space wrap>
              <Tag v-for="(duration, stage) in summary.stageDurations" :key="stage" color="blue">
                {{ stageTypeMap[stage]?.label ?? stage }}: {{ duration }}
              </Tag>
            </Space>
          </div>
        </Card>

        <!-- Legacy 信息 -->
        <Card v-if="!isV2" title="Legacy 引擎信息" class="mb-4">
          <Alert
            message="此项目使用旧版引擎"
            description="旧版引擎不支持阶段化流程。新项目已默认使用 Workflow V2。"
            type="info"
            show-icon
            class="mb-4"
          />
          <Descriptions :column="3" bordered size="small">
            <DescriptionsItem v-if="statusData.activeBatch > 0" label="活跃批次">
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

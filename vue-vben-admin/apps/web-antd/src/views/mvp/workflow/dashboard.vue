<script setup lang="ts">
// @ts-nocheck
import { ref, computed, watch, defineAsyncComponent } from 'vue';
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
  Divider,
  message,
  Modal,
  Tabs,
  TabPane,
} from 'ant-design-vue';

import {
  getProjectStatus,
  getStageHistory,
  getCompletionSummary,
  pauseProject,
  resumeProject,
  confirmPlan,
  manualApprove,
  manualReject,
  retryTask,
  skipTask,
  acceptApprove,
  acceptRework,
  triggerReplan,
  resetToDesign,
  getExecutionStatus,
  type ProjectStatusResult,
  type StageHistoryItem,
  type CompletionSummaryResult,
  type DomainTaskItem,
} from '../../../api/mvp/workflow';
import { workflowRunStatusMap, stageTypeMap } from '../consts';

const route = useRoute();
const router = useRouter();
const loading = ref(false);
const projectId = ref<string>('');

const statusData = ref<ProjectStatusResult | null>(null);
const stages = ref<StageHistoryItem[]>([]);
const summary = ref<CompletionSummaryResult | null>(null);
let dashboardLoadVersion = 0;

const currentStatus = computed(() => statusData.value?.workflowStatus || statusData.value?.status || '');
const isCompleted = computed(() => currentStatus.value === 'completed');

// Failed tasks for retry/skip operations
const failedTasks = ref<DomainTaskItem[]>([]);

// Load failed tasks when in execute/rework/paused state
async function loadFailedTasks(requestVersion?: number) {
  const currentProjectId = projectId.value;
  if (!currentProjectId) {
    failedTasks.value = [];
    return;
  }
  try {
    const res = await getExecutionStatus(currentProjectId);
    if (
      currentProjectId !== projectId.value
      || (requestVersion !== undefined && requestVersion !== dashboardLoadVersion)
    ) return;
    failedTasks.value = (res.tasks || []).filter(
      (t: DomainTaskItem) => t.status === 'failed' || t.status === 'escalated'
    );
  } catch {
    if (
      currentProjectId === projectId.value
      && (requestVersion === undefined || requestVersion === dashboardLoadVersion)
    ) {
      failedTasks.value = [];
    }
  }
}

// Status guide text
const statusGuide = computed(() => {
  const s = currentStatus.value;
  const failCount = failedTasks.value.length;
  const guides: Record<string, string> = {
    designing: '当前处于设计阶段，架构师正在与你沟通需求并拆解任务。\n建议操作：进入对话与架构师沟通，方案满意后点击「确认方案」。',
    reviewing: 'AI 审计员正在审核方案质量。通常 1-2 分钟完成。\n建议操作：等待审核结果。如果审核不合理或超时，可「手动通过审核」跳过。',
    executing: `任务自动执行中，无需人工参与。${failCount > 0 ? `\n当前有 ${failCount} 个失败任务，建议：查看失败原因后「重试」或「跳过」。` : ''}\n建议操作：观察进度，出现问题时可「暂停项目」排查。`,
    reworking: `返工阶段，系统正在自动分析失败原因并修复。${failCount > 0 ? `\n当前有 ${failCount} 个失败/升级任务。` : ''}\n建议操作：等待自动修复。如果卡住，可「恢复执行」或「重试」失败任务。也可「触发重规划」让架构师重新安排。`,
    paused: `项目已暂停，所有任务停止执行。${failCount > 0 ? `\n有 ${failCount} 个失败/升级任务需要处理。` : ''}\n建议操作：点击「恢复执行」继续。如有大量失败任务，可先「全部重试」或「全部跳过」再恢复。`,
    accepting: '验收阶段，AI 正在检查项目交付质量。\n建议操作：查看验收结果。满意可「手动放行」，不满意可「驳回返工」。',
    completed: '项目已完成，所有任务执行结束。可查看完成总结。',
  };
  return guides[s] || '当前状态：' + s;
});

// Human operations
async function handlePause() {
  Modal.confirm({
    title: '暂停项目',
    content: '请输入暂停原因（可选）',
    async onOk() {
      await pauseProject({ projectID: projectId.value, pauseReason: '人工暂停' });
      message.success('项目已暂停');
      await loadAll();
    },
  });
}

async function handleResume() {
  await resumeProject(projectId.value);
  message.success('项目已恢复执行');
  await loadAll();
}

async function handleConfirmPlan() {
  Modal.confirm({
    title: '确认方案',
    content: '确认后进入审核阶段，是否继续？',
    async onOk() {
      await confirmPlan(projectId.value);
      message.success('已提交审核');
      await loadAll();
    },
  });
}

async function handleManualApprove() {
  Modal.confirm({
    title: '手动通过审核',
    content: '跳过 AI 审核，直接通过方案进入执行阶段。确定？',
    okType: 'primary',
    async onOk() {
      await manualApprove(projectId.value);
      message.success('审核已通过，开始执行');
      await loadAll();
    },
  });
}

async function handleManualReject() {
  Modal.confirm({
    title: '驳回方案',
    content: '驳回后回到设计阶段，架构师可重新修改方案。',
    okType: 'danger',
    async onOk() {
      await manualReject(projectId.value, '人工驳回');
      message.success('方案已驳回');
      await loadAll();
    },
  });
}

async function handleRetryTask(task: DomainTaskItem) {
  await retryTask({ projectID: projectId.value, taskID: task.id });
  message.success(`任务「${task.name}」已重新排队`);
  await loadAll();
}

async function handleSkipTask(task: DomainTaskItem) {
  Modal.confirm({
    title: '跳过任务',
    content: `确定跳过「${task.name}」？跳过后该任务标记为完成，不再执行。`,
    okType: 'danger',
    async onOk() {
      await skipTask({ projectID: projectId.value, taskID: task.id, reason: '人工跳过' });
      message.success(`任务「${task.name}」已跳过`);
      await loadAll();
    },
  });
}

async function handleRetryAll() {
  Modal.confirm({
    title: '全部重试',
    content: `将重置 ${failedTasks.value.length} 个失败/升级任务为待执行状态。调度器会按批次顺序、依赖关系依次执行，不会同时启动所有任务。`,
    async onOk() {
      const hide = message.loading('正在重置任务...', 0);
      let count = 0;
      for (const task of failedTasks.value) {
        try {
          await retryTask({ projectID: projectId.value, taskID: task.id });
          count++;
        } catch {
          // 部分任务可能已经不在 failed/escalated 状态
        }
      }
      hide();
      message.success(`已重置 ${count} 个任务，调度器将按批次顺序依次执行`);
      await loadAll();
    },
  });
}

async function handleSkipAll() {
  Modal.confirm({
    title: '跳过全部失败任务',
    content: `确定跳过全部 ${failedTasks.value.length} 个失败/升级任务？跳过后这些任务标记为完成，调度器将推进到下一批次。`,
    okType: 'danger',
    async onOk() {
      const hide = message.loading('正在跳过任务...', 0);
      let count = 0;
      for (const task of failedTasks.value) {
        try {
          await skipTask({ projectID: projectId.value, taskID: task.id, reason: '人工批量跳过' });
          count++;
        } catch {
          // ignore
        }
      }
      hide();
      message.success(`已跳过 ${count} 个任务`);
      await loadAll();
    },
  });
}

async function handleAcceptApprove() {
  await acceptApprove(projectId.value, '人工放行');
  message.success('验收已通过');
  await loadAll();
}

async function handleAcceptRework() {
  Modal.confirm({
    title: '驳回返工',
    content: '驳回验收并触发返工流程。',
    okType: 'danger',
    async onOk() {
      await acceptRework(projectId.value, '人工驳回返工');
      message.success('已触发返工');
      await loadAll();
    },
  });
}

async function handleTriggerReplan() {
  Modal.confirm({
    title: '触发重规划',
    content: '将由架构师 AI 重新分析失败原因并生成新的任务方案。此过程可能需要 1-2 分钟，请在时间线中查看进度。',
    async onOk() {
      const hide = message.loading('正在触发重规划...', 0);
      try {
        await triggerReplan(projectId.value);
        message.success('重规划已触发，请在时间线中查看进度');
      } catch (e: any) {
        message.error('触发失败：' + (e?.message || '未知错误'));
      } finally {
        hide();
      }
      await loadAll();
    },
  });
}

async function handleResetToDesign() {
  Modal.confirm({
    title: '回到设计阶段',
    content: '将清理已有的方案、任务和工作空间，项目回到设计阶段重新拆分。此操作不可撤销，确定继续？',
    okText: '确定',
    okType: 'danger',
    async onOk() {
      const hide = message.loading('正在重置...', 0);
      try {
        const res = await resetToDesign(projectId.value);
        message.success(res.message || '已回到设计阶段');
      } catch (e: any) {
        message.error('重置失败：' + (e?.message || '未知错误'));
      } finally {
        hide();
      }
      await loadAll();
    },
  });
}

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

function resetDashboardState() {
  statusData.value = null;
  stages.value = [];
  summary.value = null;
  failedTasks.value = [];
}

async function loadAll() {
  const currentProjectId = projectId.value;
  if (!currentProjectId) {
    resetDashboardState();
    return;
  }
  const requestVersion = ++dashboardLoadVersion;
  loading.value = true;
  try {
    const [statusRes, stageRes] = await Promise.all([
      getProjectStatus(currentProjectId),
      getStageHistory(currentProjectId),
    ]);
    if (requestVersion !== dashboardLoadVersion || currentProjectId !== projectId.value) return;
    statusData.value = statusRes;
    stages.value = stageRes.stages || [];

    // Load failed tasks if in relevant state
    if (['executing', 'reworking', 'paused'].includes(statusRes.workflowStatus || statusRes.status || '')) {
      await loadFailedTasks(requestVersion);
      if (requestVersion !== dashboardLoadVersion || currentProjectId !== projectId.value) return;
    } else {
      failedTasks.value = [];
    }

    // 已完成时加载总结
    if (statusRes.workflowStatus === 'completed' || statusRes.status === 'completed') {
      try {
        const summaryRes = await getCompletionSummary(currentProjectId);
        if (requestVersion !== dashboardLoadVersion || currentProjectId !== projectId.value) return;
        summary.value = summaryRes;
      } catch {
        // 总结可能尚未生成
      }
    } else {
      summary.value = null;
    }
  } finally {
    if (requestVersion === dashboardLoadVersion && currentProjectId === projectId.value) {
      loading.value = false;
    }
  }
}

watch(
  () => route.query.projectId,
  async (value) => {
    projectId.value = (value as string) ?? '';
    resetDashboardState();
    await loadAll();
  },
  { immediate: true },
);

function goToTimeline() {
  router.push({ path: '/mvp/workflow/timeline', query: { projectId: projectId.value } });
}
function goToFeishu() {
  router.push({ path: '/mvp/workflow/feishu' });
}
function goToChat() {
  router.push({ path: '/mvp/chat', query: { projectId: projectId.value } });
}
function goToObjective() {
  router.push({ path: '/mvp/workflow/objective', query: { projectId: projectId.value } });
}
function goToSituation() {
  const workflowRunID = statusData.value?.workflowRunID;
  if (!workflowRunID) return;
  router.push({
    path: '/mvp/workflow/situation',
    query: { projectId: projectId.value, workflowRunId: workflowRunID },
  });
}
function goToMetaCognition() {
  router.push({ path: '/mvp/workflow/meta-cognition', query: { projectID: projectId.value } });
}

// Tab container
const activeTab = ref('execution');

const ExecutionPanel = defineAsyncComponent(() => import('./execution.vue'));
const ReviewPanel = defineAsyncComponent(() => import('./review.vue'));
const ReworkPanel = defineAsyncComponent(() => import('./rework.vue'));
const AcceptPanel = defineAsyncComponent(() => import('./accept.vue'));
const VerificationPanel = defineAsyncComponent(() => import('./verification.vue'));
const AutonomyPanel = defineAsyncComponent(() => import('./autonomy.vue'));
const TimelinePanel = defineAsyncComponent(() => import('./timeline.vue'));
const RegressionPanel = defineAsyncComponent(() => import('./regression.vue'));

// Auto-select tab based on workflow status
watch(currentStatus, (status) => {
  const tabMap: Record<string, string> = {
    reviewing: 'review',
    executing: 'execution',
    reworking: 'rework',
    accepting: 'accept',
    completed: 'verification',
  };
  if (tabMap[status]) {
    activeTab.value = tabMap[status];
  }
}, { immediate: true });
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
            <span>工作流仪表板</span>
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
                    :value-style="{ color: '#3f8600' }"
                  />
                </Col>
                <Col :span="6">
                  <Statistic
                    title="运行中"
                    :value="statusData.statusCounts?.['domain_running'] ?? statusData.activeRunningTasks ?? 0"
                    :value-style="{ color: '#1890ff' }"
                  />
                </Col>
                <Col :span="6">
                  <Statistic
                    title="失败"
                    :value="statusData.statusCounts?.['domain_failed'] ?? statusData.statusCounts?.['failed'] ?? 0"
                    :value-style="{ color: '#cf1322' }"
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
        <Card v-if="stages.length > 0" title="阶段流程" class="mb-4">
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

        <!-- 快速导航 -->
        <Card class="mb-4" size="small">
          <Space wrap>
            <Button @click="goToChat">进入对话</Button>
            <Button @click="goToObjective">目标约束</Button>
            <Button :disabled="!statusData?.workflowRunID" @click="goToSituation">态势仪表板</Button>
            <Button @click="goToMetaCognition">元认知</Button>
            <Button @click="goToFeishu">飞书协作</Button>
          </Space>
        </Card>

        <!-- 人工介入 -->
        <Card title="人工介入" class="mb-4" size="small">
          <template #extra>
            <Tag color="gold">最高权限</Tag>
          </template>

          <!-- 状态引导 -->
          <div class="bg-blue-50 border border-blue-200 rounded-lg p-3 mb-4">
            <p class="text-sm text-blue-700 whitespace-pre-line">{{ statusGuide }}</p>
            <p v-if="statusData?.pauseReason" class="text-sm text-orange-600 mt-2 font-medium">
              暂停原因：{{ statusData.pauseReason }}
            </p>
          </div>

          <!-- 操作按钮组 -->
          <Space wrap class="mb-3">
            <!-- 设计阶段 -->
            <Button v-if="currentStatus === 'designing'" type="primary" @click="handleConfirmPlan">
              确认方案
            </Button>

            <!-- 审核阶段 -->
            <Button v-if="currentStatus === 'reviewing'" type="primary" @click="handleManualApprove">
              手动通过审核
            </Button>
            <Button v-if="currentStatus === 'reviewing'" danger @click="handleManualReject">
              驳回方案
            </Button>

            <!-- 执行/返工阶段 -->
            <Button v-if="['executing', 'reworking'].includes(currentStatus)" danger @click="handlePause">
              暂停项目
            </Button>

            <!-- 暂停状态 -->
            <Button v-if="currentStatus === 'paused'" type="primary" @click="handleResume">
              恢复执行
            </Button>
            <Button v-if="currentStatus === 'paused'" danger @click="handleResetToDesign">
              回到设计
            </Button>

            <!-- 验收阶段 -->
            <Button v-if="currentStatus === 'accepting'" type="primary" @click="handleAcceptApprove">
              手动放行
            </Button>
            <Button v-if="currentStatus === 'accepting'" danger @click="handleAcceptRework">
              驳回返工
            </Button>

            <!-- 通用操作 -->
            <Button v-if="['executing', 'reworking', 'paused'].includes(currentStatus)" @click="handleTriggerReplan">
              触发重规划
            </Button>
          </Space>

          <!-- 失败任务列表（执行/返工/暂停时显示） -->
          <div v-if="failedTasks.length > 0">
            <Divider style="margin: 12px 0" />
            <div class="flex items-center justify-between mb-2">
              <span class="text-sm font-medium text-gray-700">失败/升级任务（{{ failedTasks.length }}）</span>
              <Space size="small">
                <Button size="small" type="primary" @click="handleRetryAll">全部重试</Button>
                <Button size="small" danger @click="handleSkipAll">全部跳过</Button>
              </Space>
            </div>
            <div class="max-h-60 overflow-y-auto">
              <div
                v-for="task in failedTasks"
                :key="task.id"
                class="flex items-center justify-between py-2 px-3 bg-red-50 rounded mb-1"
              >
                <div class="flex-1 min-w-0">
                  <span class="text-sm font-medium text-gray-800 truncate block">{{ task.name }}</span>
                  <span class="text-xs text-red-500">{{ task.result || task.errorMessage || task.status }}</span>
                </div>
                <Space size="small" class="ml-2 flex-shrink-0">
                  <Button size="small" @click="handleRetryTask(task)">重试</Button>
                  <Button size="small" danger @click="handleSkipTask(task)">跳过</Button>
                </Space>
              </div>
            </div>
          </div>
        </Card>

        <!-- 功能面板 -->
        <Tabs v-model:activeKey="activeTab" class="mb-4" type="card">
          <TabPane key="execution" tab="执行控制台">
            <ExecutionPanel :project-id="projectId" @changed="loadAll" />
          </TabPane>
          <TabPane key="review" tab="审核工作台">
            <ReviewPanel :project-id="projectId" @changed="loadAll" />
          </TabPane>
          <TabPane key="rework" tab="返工记录">
            <ReworkPanel :project-id="projectId" />
          </TabPane>
          <TabPane key="accept" tab="验收控制台">
            <AcceptPanel :project-id="projectId" @changed="loadAll" />
          </TabPane>
          <TabPane key="verification" tab="验证控制台">
            <VerificationPanel :project-id="projectId" @changed="loadAll" />
          </TabPane>
          <TabPane key="autonomy" tab="自治控制台">
            <AutonomyPanel :project-id="projectId" @changed="loadAll" />
          </TabPane>
          <TabPane key="timeline" tab="事件时间线">
            <TimelinePanel :project-id="projectId" />
          </TabPane>
          <TabPane key="regression" tab="评测样例">
            <RegressionPanel />
          </TabPane>
        </Tabs>

        <!-- 完成总结 -->
        <Card v-if="isCompleted && summary" title="完成总结" class="mb-4">
          <Row :gutter="16" class="mb-4">
            <Col :span="4">
              <Statistic title="总任务" :value="summary.totalTasks" />
            </Col>
            <Col :span="4">
              <Statistic title="成功" :value="summary.completedTasks" :value-style="{ color: '#3f8600' }" />
            </Col>
            <Col :span="4">
              <Statistic title="失败" :value="summary.failedTasks" :value-style="{ color: '#cf1322' }" />
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

      </template>
    </Spin>
  </Page>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { useRoute } from 'vue-router';

import { Page } from '@vben/common-ui';
import { Card, Empty, Spin, Tag, Timeline, TimelineItem, Row, Col, Statistic, Space } from 'ant-design-vue';
import {
  CheckCircleOutlined,
  CloseCircleOutlined,
  ClockCircleOutlined,
  ExclamationCircleOutlined,
  ReloadOutlined,
  RocketOutlined,
} from '@ant-design/icons-vue';

import {
  getTimeline,
  getStageHistory,
  getProjectTrace,
  getDeliveryReviews,
  type TimelineEvent,
  type StageHistoryItem,
  type ProjectTraceResult,
  type DeliveryReviewItem,
} from '#/api/mvp/workflow';
import { stageTypeMap } from '../consts';

defineOptions({ name: 'WorkflowTimeline' });

const props = defineProps<{ projectId?: string }>();

const route = useRoute();
const resolvedProjectId = computed(() => props.projectId || (route.query.projectId as string) || '');
const loading = ref(false);
const events = ref<TimelineEvent[]>([]);
const stages = ref<StageHistoryItem[]>([]);
const trace = ref<ProjectTraceResult | null>(null);
const deliveryReviews = ref<DeliveryReviewItem[]>([]);

// 当 projectId 变化时（包括 Tab 切换激活时 prop 传入）重新加载
watch(resolvedProjectId, (val) => {
  if (!val) {
    events.value = [];
    stages.value = [];
    trace.value = null;
    deliveryReviews.value = [];
    return;
  }
  loadData();
}, { immediate: true });

/** 事件类型 → 颜色/图标 */
function eventColor(eventType: string): string {
  if (eventType.includes('completed') || eventType.includes('approved')) return 'green';
  if (eventType.includes('failed') || eventType.includes('rejected')) return 'red';
  if (eventType.includes('delivery_prepared')) return 'blue';
  if (eventType.includes('review_required')) return 'orange';
  if (eventType.includes('sync_applied')) return 'green';
  if (eventType.includes('paused') || eventType.includes('escalated')) return 'orange';
  if (eventType.includes('created') || eventType.includes('started')) return 'blue';
  return 'gray';
}

function eventIcon(eventType: string) {
  if (eventType.includes('completed') || eventType.includes('approved')) return CheckCircleOutlined;
  if (eventType.includes('failed') || eventType.includes('rejected')) return CloseCircleOutlined;
  if (eventType.includes('delivery_prepared')) return RocketOutlined;
  if (eventType.includes('review_required')) return ExclamationCircleOutlined;
  if (eventType.includes('sync_applied')) return CheckCircleOutlined;
  if (eventType.includes('paused') || eventType.includes('escalated')) return ExclamationCircleOutlined;
  if (eventType.includes('started') || eventType.includes('created')) return RocketOutlined;
  return ClockCircleOutlined;
}

/** 格式化时间 */
function formatTime(t?: string) {
  if (!t) return '';
  const d = new Date(t);
  return `${d.getMonth() + 1}/${d.getDate()} ${d.getHours().toString().padStart(2, '0')}:${d.getMinutes().toString().padStart(2, '0')}:${d.getSeconds().toString().padStart(2, '0')}`;
}

const stageStatusColor: Record<string, string> = {
  running: 'processing',
  completed: 'success',
  failed: 'error',
  pending: 'default',
};

async function loadData() {
  if (!resolvedProjectId.value) return;
  loading.value = true;
  try {
    const [timelineRes, stageRes, traceRes, deliveryReviewRes] = await Promise.all([
      getTimeline(resolvedProjectId.value, 100),
      getStageHistory(resolvedProjectId.value),
      getProjectTrace(resolvedProjectId.value),
      getDeliveryReviews(resolvedProjectId.value),
    ]);
    events.value = timelineRes?.events ?? [];
    trace.value = traceRes ?? null;
    stages.value = traceRes?.stages ?? stageRes?.stages ?? [];
    deliveryReviews.value = deliveryReviewRes?.items ?? [];
  } finally {
    loading.value = false;
  }
}

function formatReasonColor(reason: string) {
  if (reason.includes('高风险')) return 'red';
  if (reason.includes('PR')) return 'magenta';
  if (reason.includes('回写')) return 'geekblue';
  return 'orange';
}

function countMapEntries(data?: Record<string, number>) {
  return Object.entries(data || {}).filter(([, count]) => Number(count) > 0);
}
</script>

<template>
  <Page auto-content-height>
    <Spin :spinning="loading">
      <Empty
        v-if="!resolvedProjectId"
        description="请从项目列表进入查看工作流时间线"
        class="mt-20"
      />
      <template v-else>
        <Card title="轨迹总览" class="mb-4">
          <Row :gutter="16">
            <Col :span="4">
              <Statistic title="事件数" :value="trace?.totalEvents || 0" />
            </Col>
            <Col :span="4">
              <Statistic title="阶段数" :value="trace?.totalStages || 0" />
            </Col>
            <Col :span="4">
              <Statistic title="任务数" :value="trace?.totalTasks || 0" />
            </Col>
            <Col :span="4">
              <Statistic title="返工轮次" :value="trace?.reworkRounds || 0" />
            </Col>
            <Col :span="4">
              <Statistic title="待处理人工节点" :value="trace?.pendingCheckpoints || 0" />
            </Col>
            <Col :span="4">
              <Statistic title="待处理决策" :value="trace?.pendingActions || 0" />
            </Col>
          </Row>
          <div class="mt-3 flex flex-wrap items-center gap-2 text-xs text-gray-500">
            <Tag v-if="trace?.workflowStatus" color="blue">工作流: {{ trace.workflowStatus }}</Tag>
            <Tag v-if="trace?.currentStage" color="cyan">当前阶段: {{ stageTypeMap[trace.currentStage]?.label ?? trace.currentStage }}</Tag>
            <Tag v-if="trace?.acceptDecision" :color="trace.acceptDecision === 'passed' ? 'green' : trace.acceptDecision === 'manual_review' ? 'orange' : 'red'">
              验收: {{ trace.acceptDecision }} / {{ Number(trace?.acceptScore || 0).toFixed(1) }}
            </Tag>
            <Tag color="volcano">审核问题: {{ trace?.openReviewIssues || 0 }}</Tag>
            <Tag color="gold">验收问题: {{ trace?.openAcceptIssues || 0 }}</Tag>
            <Tag color="orange">待人工交付: {{ trace?.pendingDeliveryReviews || 0 }}</Tag>
            <Tag color="magenta">PR 草稿: {{ trace?.prDraftTasks || 0 }}</Tag>
            <Tag color="red">高风险任务: {{ trace?.highRiskTasks || 0 }}</Tag>
            <Tag color="geekblue">待回写: {{ trace?.manualSyncTasks || 0 }}</Tag>
          </div>
          <div v-if="countMapEntries(trace?.deliveryModes).length" class="mt-3">
            <div class="mb-1 text-xs text-gray-400">交付形态</div>
            <Space wrap>
              <Tag v-for="[key, count] in countMapEntries(trace?.deliveryModes)" :key="`delivery-${key}`" color="purple">
                {{ key }}: {{ count }}
              </Tag>
            </Space>
          </div>
          <div v-if="countMapEntries(trace?.syncStatuses).length" class="mt-2">
            <div class="mb-1 text-xs text-gray-400">回写状态</div>
            <Space wrap>
              <Tag v-for="[key, count] in countMapEntries(trace?.syncStatuses)" :key="`sync-${key}`" color="geekblue">
                {{ key }}: {{ count }}
              </Tag>
            </Space>
          </div>
        </Card>

        <Card title="交付闸门" class="mb-4">
          <Empty v-if="deliveryReviews.length === 0" description="当前没有待人工处理的交付项" />
          <div v-else class="space-y-3">
            <div
              v-for="item in deliveryReviews"
              :key="`${item.workspaceID}-${item.taskID}`"
              class="rounded-lg border p-3"
            >
              <div class="flex flex-wrap items-center justify-between gap-2">
                <div class="flex flex-wrap items-center gap-2">
                  <span class="font-medium">{{ item.taskName }}</span>
                  <Tag color="blue">任务 #{{ item.taskID }}</Tag>
                  <Tag v-if="item.batchNo > 0" color="cyan">批次 {{ item.batchNo }}</Tag>
                  <Tag v-if="item.deliveryMode" color="purple">{{ item.deliveryMode }}</Tag>
                  <Tag v-if="item.syncStatus" color="geekblue">{{ item.syncStatus }}</Tag>
                  <Tag v-if="item.riskLevel" :color="item.riskLevel === 'high' ? 'red' : item.riskLevel === 'medium' ? 'orange' : 'green'">
                    {{ item.riskLevel }}
                  </Tag>
                </div>
                <span class="text-xs text-gray-400">{{ formatTime(item.updatedAt) }}</span>
              </div>
              <div class="mt-2 flex flex-wrap gap-2">
                <Tag
                  v-for="reason in item.reasons"
                  :key="`${item.taskID}-${reason}`"
                  :color="formatReasonColor(reason)"
                >
                  {{ reason }}
                </Tag>
              </div>
              <div class="mt-2 text-xs text-gray-500">
                {{ item.roleType }} / {{ item.executionMode }} / 状态 {{ item.taskStatus }}
              </div>
              <div v-if="item.deliveryTitle || item.deliveryRef || item.patchRef || item.diffSummary" class="mt-2 space-y-1 text-xs text-gray-600">
                <div v-if="item.deliveryTitle"><span class="text-gray-400">交付标题：</span>{{ item.deliveryTitle }}</div>
                <div v-if="item.deliveryRef" class="break-all"><span class="text-gray-400">交付引用：</span>{{ item.deliveryRef }}</div>
                <div v-if="item.patchRef" class="break-all"><span class="text-gray-400">Patch：</span>{{ item.patchRef }}</div>
                <div v-if="item.diffSummary" class="whitespace-pre-wrap"><span class="text-gray-400">Diff 摘要：</span>{{ item.diffSummary }}</div>
              </div>
            </div>
          </div>
        </Card>

        <div class="grid grid-cols-1 gap-4 xl:grid-cols-3">
          <!-- 阶段历程 -->
          <Card title="阶段历程" class="xl:col-span-1">
            <template #extra>
              <a @click="loadData"><ReloadOutlined /> 刷新</a>
            </template>
            <Empty v-if="stages.length === 0" description="暂无阶段记录" />
            <div v-else class="space-y-3">
              <div
                v-for="stage in stages"
                :key="stage.id"
                class="flex items-center justify-between rounded-lg border p-3"
              >
                <div>
                  <Tag :color="stageTypeMap[stage.stageType]?.color ?? 'default'">
                    {{ stageTypeMap[stage.stageType]?.label ?? stage.stageType }}
                  </Tag>
                  <span class="ml-2 text-xs text-gray-400">#{{ stage.stageNo }}</span>
                </div>
                <div class="flex items-center gap-2">
                  <Tag :color="stageStatusColor[stage.status] ?? 'default'" size="small">
                    {{ stage.status }}
                  </Tag>
                  <span class="text-xs text-gray-400">{{ formatTime(stage.startedAt) }}</span>
                </div>
              </div>
            </div>
          </Card>

          <!-- 事件时间线 -->
          <Card title="事件流" class="xl:col-span-2">
            <template #extra>
              <a @click="loadData"><ReloadOutlined /> 刷新</a>
            </template>
            <Empty v-if="events.length === 0" description="暂无事件记录" />
            <Timeline v-else>
              <TimelineItem
                v-for="event in events"
                :key="event.id"
                :color="eventColor(event.eventType)"
              >
                <template #dot>
                  <component :is="eventIcon(event.eventType)" />
                </template>
                <div class="flex items-start justify-between">
                  <div>
                    <span class="font-medium">{{ event.label }}</span>
                    <div class="mt-1 text-xs text-gray-400">
                      {{ event.entityType }} · {{ event.eventType }}
                    </div>
                  </div>
                  <span class="flex-shrink-0 text-xs text-gray-400">
                    {{ formatTime(event.createdAt) }}
                  </span>
                </div>
              </TimelineItem>
            </Timeline>
          </Card>
        </div>
      </template>
    </Spin>
  </Page>
</template>

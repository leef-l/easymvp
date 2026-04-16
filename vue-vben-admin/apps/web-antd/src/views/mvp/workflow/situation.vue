<script setup lang="ts">
import { computed, h, ref, watch } from 'vue';
import { useRoute } from 'vue-router';

import { Page } from '@vben/common-ui';

import { ReloadOutlined } from '@ant-design/icons-vue';
import {
  Alert,
  Button,
  Card,
  Col,
  Divider,
  Empty,
  Progress,
  Row,
  Space,
  Spin,
  Statistic,
  Table,
  Tooltip,
} from 'ant-design-vue';

import {
  getSituation,
  getSituationHistory,
  type SituationData,
  type SituationSnapshot,
} from '../../../api/mvp/workflow';

const route = useRoute();
const loading = ref(false);
const histLoading = ref(false);

const workflowRunId = ref<string>('');
const projectId = ref<string>('');
const taskId = ref<string>('');

const situation = ref<null | SituationData>(null);
const history = ref<SituationSnapshot[]>([]);
let situationLoadVersion = 0;
let historyLoadVersion = 0;

// ========== 衍生计算 ==========

const healthColor = computed(() => {
  const h = situation.value?.health;
  if (!h) return 'default';
  if (h.consecutiveFailures >= 3 || h.recentFailureRate >= 0.3) return 'error';
  if (h.consecutiveFailures >= 1 || h.recentFailureRate >= 0.1) return 'warning';
  return 'success';
});

const completionPercent = computed(() =>
  Math.round((situation.value?.progress?.completionRate ?? 0) * 100),
);

const batchPercent = computed(() =>
  Math.round((situation.value?.progress?.batchProgress ?? 0) * 100),
);

const utilizationPercent = computed(() =>
  Math.round((situation.value?.resource?.resourceUtilization ?? 0) * 100),
);

const anomalies = computed(() => situation.value?.anomalySignals ?? []);
const hasTaskFocus = computed(() => Boolean(situation.value?.health?.focusedTaskId || taskId.value));
const focusedTaskId = computed(() => situation.value?.health?.focusedTaskId || taskId.value || '');
const workflowRetryTitle = computed(() => hasTaskFocus.value ? '工作流重试' : '重试次数');
const workflowReworkTitle = computed(() => hasTaskFocus.value ? '工作流返工' : '返工轮次');
const canLoadHistory = computed(() => Boolean(projectId.value && workflowRunId.value));
const refreshLoading = computed(() => loading.value || histLoading.value);
const historyEmptyText = computed(() => {
  if (!workflowRunId.value) {
    return '缺少 workflowRunId，未加载快照历史';
  }
  if (!projectId.value) {
    return '缺少 projectId，未加载快照历史';
  }
  return hasTaskFocus.value ? '暂无该任务焦点快照' : '暂无快照历史';
});
const historyTitle = computed(() =>
  hasTaskFocus.value
    ? `快照历史（最近 20 条，任务 ${focusedTaskId.value} 焦点）`
    : '快照历史（最近 20 条）',
);

function trendIcon(trend: string) {
  if (trend === 'rising') return '↑';
  if (trend === 'falling') return '↓';
  return '→';
}

function trendColor(trend: string, inverse = false) {
  // inverse=true: 上升是好的（吞吐量）；默认：上升是坏的（失败率/时长）
  if (trend === 'rising') return inverse ? 'green' : 'red';
  if (trend === 'falling') return inverse ? 'red' : 'green';
  return '#888';
}

const historyColumns = [
  { title: '快照时间', dataIndex: 'snapshotAt', key: 'snapshotAt', width: 180 },
  {
    title: '完成率', key: 'completionRate',
    customRender: ({ record }: any) =>
      `${Math.round((record.progress?.completionRate ?? 0) * 100)}%`,
  },
  {
    title: '近期失败率', key: 'failureRate',
    customRender: ({ record }: any) =>
      `${Math.round((record.health?.recentFailureRate ?? 0) * 100)}%`,
  },
  {
    title: '连续失败', key: 'consecutiveFailures',
    customRender: ({ record }: any) => record.health?.consecutiveFailures ?? 0,
  },
  {
    title: '资源利用率', key: 'utilization',
    customRender: ({ record }: any) =>
      `${Math.round((record.resource?.resourceUtilization ?? 0) * 100)}%`,
  },
];

// ========== 数据加载 ==========

async function loadSituation() {
  const currentWorkflowRunId = workflowRunId.value;
  const currentTaskId = taskId.value;
  if (!currentWorkflowRunId) {
    situation.value = null;
    return;
  }
  const requestVersion = ++situationLoadVersion;
  loading.value = true;
  try {
    const res = await getSituation(currentWorkflowRunId, currentTaskId || undefined);
    if (
      requestVersion !== situationLoadVersion
      || currentWorkflowRunId !== workflowRunId.value
      || currentTaskId !== taskId.value
    ) return;
    situation.value = res.situation ?? null;
  } catch (error) {
    console.warn('[situation] loadSituation 失败:', error);
  } finally {
    if (
      requestVersion === situationLoadVersion
      && currentWorkflowRunId === workflowRunId.value
      && currentTaskId === taskId.value
    ) {
      loading.value = false;
    }
  }
}

async function loadHistory() {
  if (!canLoadHistory.value) {
    history.value = [];
    return;
  }
  const currentProjectId = projectId.value;
  const currentWorkflowRunId = workflowRunId.value;
  const currentTaskId = taskId.value;
  const requestVersion = ++historyLoadVersion;
  histLoading.value = true;
  try {
    const res = await getSituationHistory({
      projectID: currentProjectId,
      taskID: currentTaskId || undefined,
      workflowRunID: currentWorkflowRunId || undefined,
      limit: 20,
    });
    if (
      requestVersion !== historyLoadVersion
      || currentProjectId !== projectId.value
      || currentWorkflowRunId !== workflowRunId.value
      || currentTaskId !== taskId.value
    ) return;
    history.value = res.snapshots ?? [];
  } catch (error) {
    console.warn('[situation] loadHistory 失败:', error);
  } finally {
    if (
      requestVersion === historyLoadVersion
      && currentProjectId === projectId.value
      && currentWorkflowRunId === workflowRunId.value
      && currentTaskId === taskId.value
    ) {
      histLoading.value = false;
    }
  }
}

async function refreshDashboard() {
  await Promise.all([loadSituation(), loadHistory()]);
}

watch(
  [() => route.query.workflowRunId, () => route.query.projectId, () => route.query.taskId],
  async ([nextWorkflowRunId, nextProjectId, nextTaskId]) => {
    try {
      workflowRunId.value = (nextWorkflowRunId as string) ?? '';
      projectId.value = (nextProjectId as string) ?? '';
      taskId.value = (nextTaskId as string) ?? '';
      situation.value = null;
      history.value = [];
      await refreshDashboard();
    } catch (error) {
      console.warn('[situation] route 变更加载失败:', error);
    }
  },
  { immediate: true },
);
</script>

<template>
  <Page title="态势感知仪表板">
    <Space style="margin-bottom: 16px">
      <Button type="primary" :icon="h(ReloadOutlined)" :loading="refreshLoading" @click="refreshDashboard">
        刷新态势
      </Button>
    </Space>

    <Spin :spinning="refreshLoading">
      <Empty v-if="!situation && !refreshLoading" description="暂无态势数据，请确认 workflowRunId 参数" />

      <template v-if="situation">
        <Alert
          v-if="hasTaskFocus"
          type="info"
          show-icon
          style="margin-bottom: 16px"
          :message="`当前按任务 ${focusedTaskId} 聚焦查看预算`"
          description="当前态势会同时展示 workflow 总量与该任务自己的 retry / rework 预算；下方快照历史仅展示带同任务焦点的记录。"
        />

        <!-- 顶部状态概览 -->
        <Row :gutter="16" style="margin-bottom: 16px">
          <Col :span="6">
            <Card>
              <Statistic
                title="工作流状态"
                :value="situation.workflowStatus"
                :value-style="{ color: situation.workflowStatus === 'running' ? '#52c41a' : '#888' }"
              />
            </Card>
          </Col>
          <Col :span="6">
            <Card>
              <Statistic
                title="当前阶段"
                :value="situation.activeStage || '-'"
              />
            </Card>
          </Col>
          <Col :span="6">
            <Card>
              <Statistic
                title="整体完成率"
                :value="completionPercent"
                suffix="%"
                :value-style="{ color: completionPercent >= 80 ? '#52c41a' : '#1890ff' }"
              />
            </Card>
          </Col>
          <Col :span="6">
            <Card>
              <Statistic
                title="健康状态"
                :value="healthColor === 'success' ? '良好' : healthColor === 'warning' ? '注意' : '异常'"
                :value-style="{ color: healthColor === 'success' ? '#52c41a' : healthColor === 'warning' ? '#faad14' : '#ff4d4f' }"
              />
            </Card>
          </Col>
        </Row>

        <!-- 异常信号 -->
        <template v-if="anomalies.length > 0">
          <Alert
            v-for="(sig, i) in anomalies"
            :key="i"
            :type="sig.severity === 'critical' ? 'error' : sig.severity === 'warning' ? 'warning' : 'info'"
            :message="sig.message"
            :description="`类型: ${sig.type}  置信度: ${Math.round(sig.confidence * 100)}%`"
            show-icon
            style="margin-bottom: 8px"
            banner
          />
          <Divider />
        </template>

        <Row :gutter="16">
          <!-- 进度指标 -->
          <Col :span="8">
            <Card title="进度指标" style="margin-bottom: 16px">
              <div style="margin-bottom: 12px">
                <div style="margin-bottom: 4px; color: #888; font-size: 12px">整体完成率</div>
                <Progress :percent="completionPercent" status="active" />
              </div>
              <div style="margin-bottom: 12px">
                <div style="margin-bottom: 4px; color: #888; font-size: 12px">
                  当前批次进度（批次 {{ situation.progress?.currentBatchNo }} / {{ situation.progress?.totalBatches }}）
                </div>
                <Progress :percent="batchPercent" stroke-color="#722ed1" />
              </div>
              <Row :gutter="8">
                <Col :span="8">
                  <Statistic title="总任务" :value="situation.progress?.totalTasks ?? 0" />
                </Col>
                <Col :span="8">
                  <Statistic title="已完成" :value="situation.progress?.completedTasks ?? 0" :value-style="{ color: '#52c41a' }" />
                </Col>
                <Col :span="8">
                  <Statistic title="失败" :value="situation.progress?.failedTasks ?? 0" :value-style="{ color: situation.progress?.failedTasks ? '#ff4d4f' : '#888' }" />
                </Col>
              </Row>
            </Card>
          </Col>

          <!-- 健康指标 -->
          <Col :span="8">
            <Card title="健康指标" style="margin-bottom: 16px">
              <Row :gutter="8">
                <Col :span="12">
                  <Statistic
                    title="近期失败率"
                    :value="Math.round((situation.health?.recentFailureRate ?? 0) * 100)"
                    suffix="%"
                    :value-style="{ color: (situation.health?.recentFailureRate ?? 0) >= 0.2 ? '#ff4d4f' : '#888' }"
                  />
                </Col>
                <Col :span="12">
                  <Statistic
                    title="连续失败"
                    :value="situation.health?.consecutiveFailures ?? 0"
                    :value-style="{ color: (situation.health?.consecutiveFailures ?? 0) >= 3 ? '#ff4d4f' : '#888' }"
                  />
                </Col>
                <Col :span="12" style="margin-top: 12px">
                  <Statistic :title="workflowRetryTitle" :value="situation.health?.retryCount ?? 0" />
                </Col>
                <Col :span="12" style="margin-top: 12px">
                  <Statistic :title="workflowReworkTitle" :value="situation.health?.reworkRounds ?? 0" :value-style="{ color: (situation.health?.reworkRounds ?? 0) >= 2 ? '#faad14' : '#888' }" />
                </Col>
                <Col v-if="hasTaskFocus" :span="12" style="margin-top: 12px">
                  <Statistic title="当前任务重试" :value="situation.health?.taskRetryCount ?? 0" :value-style="{ color: (situation.health?.taskRetryCount ?? 0) >= 2 ? '#faad14' : '#888' }" />
                </Col>
                <Col v-if="hasTaskFocus" :span="12" style="margin-top: 12px">
                  <Statistic title="当前任务返工" :value="situation.health?.taskReworkRounds ?? 0" :value-style="{ color: (situation.health?.taskReworkRounds ?? 0) >= 2 ? '#ff4d4f' : '#888' }" />
                </Col>
                <Col :span="12" style="margin-top: 12px">
                  <Statistic title="升级次数" :value="situation.health?.escalationCount ?? 0" />
                </Col>
                <Col :span="12" style="margin-top: 12px">
                  <Statistic title="重规划次数" :value="situation.health?.replanCount ?? 0" />
                </Col>
              </Row>
            </Card>
          </Col>

          <!-- 资源指标 -->
          <Col :span="8">
            <Card title="资源指标" style="margin-bottom: 16px">
              <div style="margin-bottom: 12px">
                <div style="margin-bottom: 4px; color: #888; font-size: 12px">
                  资源利用率（{{ situation.resource?.runningConcurrency ?? 0 }} / {{ situation.resource?.maxConcurrency ?? 0 }} 并发）
                </div>
                <Progress
                  :percent="utilizationPercent"
                  :stroke-color="utilizationPercent >= 85 ? '#ff4d4f' : utilizationPercent >= 60 ? '#faad14' : '#52c41a'"
                />
              </div>
              <Row :gutter="8">
                <Col :span="12">
                  <Statistic title="锁定资源" :value="situation.resource?.lockedResourceCount ?? 0" />
                </Col>
                <Col :span="12">
                  <Statistic title="冲突次数" :value="situation.resource?.conflictCount ?? 0" :value-style="{ color: (situation.resource?.conflictCount ?? 0) > 0 ? '#faad14' : '#888' }" />
                </Col>
                <Col :span="12" style="margin-top: 12px">
                  <Tooltip :title="`剩余 Token: ${situation.resource?.estimatedTokensLeft ?? 0}`">
                    <Statistic title="已消耗 Token" :value="situation.resource?.tokensConsumed ?? 0" />
                  </Tooltip>
                </Col>
              </Row>
            </Card>
          </Col>
        </Row>

        <!-- 趋势指标 -->
        <Card title="趋势信号" style="margin-bottom: 16px">
          <Row :gutter="24">
            <Col :span="8">
              <div>失败率趋势</div>
              <span :style="{ fontSize: '24px', color: trendColor(situation.trend?.failureRateTrend ?? 'stable') }">
                {{ trendIcon(situation.trend?.failureRateTrend ?? 'stable') }}
                {{ situation.trend?.failureRateTrend ?? '-' }}
              </span>
            </Col>
            <Col :span="8">
              <div>任务时长趋势</div>
              <span :style="{ fontSize: '24px', color: trendColor(situation.trend?.durationTrend ?? 'stable') }">
                {{ trendIcon(situation.trend?.durationTrend ?? 'stable') }}
                {{ situation.trend?.durationTrend ?? '-' }}
              </span>
            </Col>
            <Col :span="8">
              <div>吞吐量趋势</div>
              <span :style="{ fontSize: '24px', color: trendColor(situation.trend?.throughputTrend ?? 'stable', true) }">
                {{ trendIcon(situation.trend?.throughputTrend ?? 'stable') }}
                {{ situation.trend?.throughputTrend ?? '-' }}
              </span>
            </Col>
          </Row>
        </Card>
      </template>
    </Spin>

    <!-- 快照历史 -->
    <Divider>{{ historyTitle }}</Divider>
    <Table
      :loading="histLoading"
      :data-source="history"
      :columns="historyColumns"
      :pagination="false"
      :locale="{ emptyText: historyEmptyText }"
      size="small"
      row-key="id"
    />
  </Page>
</template>

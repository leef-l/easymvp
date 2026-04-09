<script setup lang="ts">
import { ref, computed, onUnmounted, watch } from 'vue';
import { useRoute } from 'vue-router';

import { Page } from '@vben/common-ui';
import {
  Card,
  Tag,
  Empty,
  Spin,
  Progress,
  Statistic,
  Row,
  Col,
  Table,
  Button,
  Space,
  Badge,
  Tooltip,
  Modal,
  Drawer,
  Descriptions,
  DescriptionsItem,
  Tabs,
  TabPane,
  message,
} from 'ant-design-vue';

import {
  getExecutionStatus,
  getTaskReplay,
  retryTask,
  skipTask,
  type ExecutionStatusResult,
  type DomainTaskItem,
  type TaskReplayResult,
} from '../../../api/mvp/workflow';
import { stageRunStatusMap, taskStatusMap } from '../consts';

const props = defineProps<{ projectId?: string }>();

const route = useRoute();
const loading = ref(false);
const resolvedProjectId = computed(() => props.projectId || (route.query.projectId as string) || '');
const data = ref<ExecutionStatusResult | null>(null);
const replayVisible = ref(false);
const replayLoading = ref(false);
const selectedTask = ref<DomainTaskItem | null>(null);
const replayData = ref<TaskReplayResult | null>(null);
let pollTimer: ReturnType<typeof setInterval> | null = null;

type BadgeStatus = 'default' | 'processing' | 'success' | 'error' | 'warning';

onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer);
});

function startPoll() {
  if (pollTimer) clearInterval(pollTimer);
  pollTimer = setInterval(loadData, 5000);
}

function resetExecutionState() {
  data.value = null;
  replayVisible.value = false;
  replayLoading.value = false;
  selectedTask.value = null;
  replayData.value = null;
}

watch(
  resolvedProjectId,
  (value) => {
    if (pollTimer) {
      clearInterval(pollTimer);
      pollTimer = null;
    }
    resetExecutionState();
    if (!value) return;
    loadData();
    startPoll();
  },
  { immediate: true },
);

async function loadData() {
  loading.value = true;
  try {
    const res = await getExecutionStatus(resolvedProjectId.value);
    data.value = res;
  } catch {
    // silent
  } finally {
    loading.value = false;
  }
}

const progressPercent = computed(() => {
  if (!data.value || data.value.totalTasks === 0) return 0;
  return Math.round((data.value.completedTasks / data.value.totalTasks) * 100);
});

const progressStatus = computed(() => {
  if (!data.value) return 'active';
  if (data.value.failedTasks > 0) return 'exception';
  if (data.value.completedTasks === data.value.totalTasks) return 'success';
  return 'active';
});

// 按批次分组
const batchGroups = computed(() => {
  if (!data.value) return [];
  const groups = new Map<number, DomainTaskItem[]>();
  for (const task of data.value.tasks) {
    const batch = task.batchNo || 0;
    if (!groups.has(batch)) groups.set(batch, []);
    groups.get(batch)!.push(task);
  }
  return Array.from(groups.entries())
    .sort(([a], [b]) => a - b)
    .map(([batchNo, tasks]) => ({
      batchNo,
      label: batchNo === 0 ? '紧急任务' : `批次 ${batchNo}`,
      tasks,
      completed: tasks.filter((t) => t.status === 'completed').length,
      total: tasks.length,
    }));
});

const statusColorMap: Record<string, BadgeStatus> = {
  pending: 'default',
  running: 'processing',
  completed: 'success',
  failed: 'error',
  escalated: 'warning',
};

const deliveryColorMap: Record<string, string> = {
  patch: 'blue',
  pr: 'purple',
  manual: 'orange',
};

const syncColorMap: Record<string, string> = {
  pending: 'gold',
  applied: 'green',
  skipped: 'default',
  failed: 'red',
};

const columns = [
  { title: '任务名称', dataIndex: 'name', key: 'name', width: 180, ellipsis: true },
  { title: '状态', dataIndex: 'status', key: 'status', width: 90 },
  { title: '角色', dataIndex: 'roleType', key: 'roleType', width: 80 },
  { title: '执行方式', dataIndex: 'executionMode', key: 'executionMode', width: 80 },
  { title: '交付', key: 'delivery', width: 150 },
  { title: '结果/错误', key: 'resultInfo', width: 200, ellipsis: true },
  { title: '重试', dataIndex: 'retryCount', key: 'retryCount', width: 50 },
  { title: '操作', key: 'action', width: 140 },
];

async function showDetail(task: DomainTaskItem) {
  selectedTask.value = task;
  replayVisible.value = true;
  replayLoading.value = true;
  replayData.value = null;
  try {
    replayData.value = await getTaskReplay(resolvedProjectId.value, task.id);
  } catch {
    message.warning('任务回放数据加载失败');
  } finally {
    replayLoading.value = false;
  }
}

function formatTime(value?: string) {
  if (!value) return '-';
  return value.replace('T', ' ').replace('Z', '');
}

async function handleRetry(task: DomainTaskItem) {
  Modal.confirm({
    title: '确认重试',
    content: `确定要重试任务「${task.name}」吗？`,
    onOk: async () => {
      await retryTask({ projectID: resolvedProjectId.value, taskID: task.id });
      message.success('已提交重试');
      loadData();
    },
  });
}

async function handleSkip(task: DomainTaskItem) {
  Modal.confirm({
    title: '确认跳过',
    content: `跳过任务「${task.name}」将解除后续批次阻塞，确定跳过吗？`,
    onOk: async () => {
      await skipTask({ projectID: resolvedProjectId.value, taskID: task.id, reason: '手动跳过' });
      message.success('已跳过');
      loadData();
    },
  });
}

function showDetailRecord(record: Record<string, any>) {
  showDetail(record as DomainTaskItem);
}

function handleRetryRecord(record: Record<string, any>) {
  handleRetry(record as DomainTaskItem);
}

function handleSkipRecord(record: Record<string, any>) {
  handleSkip(record as DomainTaskItem);
}
</script>

<template>
  <Page auto-content-height>
    <Spin :spinning="loading">
      <Empty
        v-if="!data"
        description="请从项目列表进入查看执行状态"
        class="mt-20"
      />

      <template v-else>
        <!-- 执行概览 -->
        <Card title="执行控制台" class="mb-4">
          <Row :gutter="16" class="mb-4">
            <Col :span="4">
              <Statistic title="总任务" :value="data.totalTasks" />
            </Col>
            <Col :span="4">
              <Statistic title="已完成" :value="data.completedTasks" :value-style="{ color: '#3f8600' }" />
            </Col>
            <Col :span="4">
              <Statistic title="运行中" :value="data.runningTasks" :value-style="{ color: '#1890ff' }" />
            </Col>
            <Col :span="4">
              <Statistic title="待执行" :value="data.pendingTasks" />
            </Col>
            <Col :span="4">
              <Statistic title="失败" :value="data.failedTasks" :value-style="{ color: '#cf1322' }" />
            </Col>
            <Col :span="4">
              <Statistic title="已升级" :value="data.escalatedTasks" :value-style="{ color: '#fa8c16' }" />
            </Col>
          </Row>

          <Progress :percent="progressPercent" :status="progressStatus" />

          <div class="mt-2 text-gray-500 text-sm">
            <span v-if="data.activeBatch > 0">活跃批次: {{ data.activeBatch }}</span>
            <span v-if="data.stageStatus"> | 阶段状态:
              <Tag :color="stageRunStatusMap[data.stageStatus]?.color ?? 'default'" size="small">
                {{ stageRunStatusMap[data.stageStatus]?.label ?? data.stageStatus }}
              </Tag>
            </span>
          </div>
        </Card>

        <!-- 资源锁 -->
        <Card
          v-if="data.resourceLocks.length > 0"
          title="资源锁"
          class="mb-4"
          size="small"
        >
          <Space wrap>
            <Tag v-for="lock in data.resourceLocks" :key="lock.resource" color="orange">
              <Tooltip :title="`被任务「${lock.taskName}」锁定`">
                {{ lock.resource }}
              </Tooltip>
            </Tag>
          </Space>
        </Card>

        <!-- 按批次展示任务 -->
        <Card
          v-for="group in batchGroups"
          :key="group.batchNo"
          class="mb-4"
          size="small"
        >
          <template #title>
            <Space>
              <span>{{ group.label }}</span>
              <Badge
                :count="`${group.completed}/${group.total}`"
                :number-style="{ backgroundColor: group.completed === group.total ? '#52c41a' : '#1890ff' }"
              />
            </Space>
          </template>

          <Table
            :data-source="group.tasks"
            :columns="columns"
            :pagination="false"
            size="small"
            row-key="id"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'status'">
                <Badge :status="statusColorMap[record.status] ?? 'default'" :text="taskStatusMap[record.status]?.label ?? record.status" />
              </template>

              <template v-else-if="column.key === 'resultInfo'">
                <span v-if="record.errorMessage" class="text-red-500 text-xs cursor-pointer" @click="showDetailRecord(record)">
                  {{ record.errorMessage.length > 50 ? record.errorMessage.slice(0, 50) + '...' : record.errorMessage }}
                </span>
                <span v-else-if="record.result" class="text-green-600 text-xs cursor-pointer" @click="showDetailRecord(record)">
                  {{ record.result.length > 50 ? record.result.slice(0, 50) + '...' : record.result }}
                </span>
                <span v-else class="text-gray-400 text-xs">-</span>
              </template>

              <template v-else-if="column.key === 'delivery'">
                <Space size="small" wrap>
                  <Tag v-if="record.deliveryMode" :color="deliveryColorMap[record.deliveryMode] ?? 'default'">
                    {{ record.deliveryMode }}
                  </Tag>
                  <Tag v-if="record.deliveryTitle" color="geekblue">
                    {{ record.deliveryTitle }}
                  </Tag>
                  <Tag v-if="record.syncStatus" :color="syncColorMap[record.syncStatus] ?? 'default'">
                    {{ record.syncStatus }}
                  </Tag>
                </Space>
              </template>

              <template v-else-if="column.key === 'action'">
                <Space>
                  <Button type="link" size="small" @click="showDetailRecord(record)">详情</Button>
                  <Button
                    v-if="record.status === 'failed' || record.status === 'escalated'"
                    type="link"
                    size="small"
                    @click="handleRetryRecord(record)"
                  >
                    重试
                  </Button>
                  <Button
                    v-if="['failed', 'escalated', 'pending'].includes(record.status)"
                    type="link"
                    size="small"
                    danger
                    @click="handleSkipRecord(record)"
                  >
                    跳过
                  </Button>
                </Space>
              </template>
            </template>
          </Table>
        </Card>
      </template>
    </Spin>

    <Drawer
      v-model:open="replayVisible"
      :title="selectedTask?.name || '任务回放'"
      width="720"
      destroy-on-close
    >
      <Spin :spinning="replayLoading">
        <Empty v-if="!selectedTask" description="暂无任务数据" />
        <template v-else>
          <Descriptions :column="2" bordered size="small" class="mb-4">
            <DescriptionsItem label="状态">
              <Badge :status="statusColorMap[(replayData?.task?.status || selectedTask.status)] ?? 'default'" :text="taskStatusMap[replayData?.task?.status || selectedTask.status]?.label ?? replayData?.task?.status ?? selectedTask.status" />
            </DescriptionsItem>
            <DescriptionsItem label="阶段">{{ replayData?.stageType || '-' }}</DescriptionsItem>
            <DescriptionsItem label="角色">{{ replayData?.task?.roleType || selectedTask.roleType }}/{{ replayData?.task?.roleLevel || selectedTask.roleLevel }}</DescriptionsItem>
            <DescriptionsItem label="执行方式">{{ replayData?.task?.executionMode || selectedTask.executionMode || '-' }}</DescriptionsItem>
            <DescriptionsItem label="交付">{{ replayData?.task?.deliveryMode || selectedTask.deliveryMode || '-' }}</DescriptionsItem>
            <DescriptionsItem label="交付标题">{{ replayData?.task?.deliveryTitle || selectedTask.deliveryTitle || '-' }}</DescriptionsItem>
            <DescriptionsItem label="回写">{{ replayData?.task?.syncStatus || selectedTask.syncStatus || '-' }}</DescriptionsItem>
            <DescriptionsItem label="交付引用">{{ replayData?.task?.deliveryRef || selectedTask.deliveryRef || '-' }}</DescriptionsItem>
            <DescriptionsItem label="风险">{{ replayData?.task?.riskLevel || selectedTask.riskLevel || '-' }}</DescriptionsItem>
            <DescriptionsItem label="批次 / 重试">{{ replayData?.task?.batchNo || selectedTask.batchNo }} / {{ replayData?.task?.retryCount || selectedTask.retryCount }}</DescriptionsItem>
            <DescriptionsItem label="开始时间">{{ formatTime(replayData?.task?.startedAt || selectedTask.startedAt) }}</DescriptionsItem>
            <DescriptionsItem label="完成时间">{{ formatTime(replayData?.task?.completedAt || selectedTask.completedAt) }}</DescriptionsItem>
          </Descriptions>

          <div v-if="(replayData?.task?.description || selectedTask.description)" class="mb-4 rounded border bg-gray-50 p-3 text-xs leading-6 text-gray-700 whitespace-pre-wrap">
            {{ replayData?.task?.description || selectedTask.description }}
          </div>

          <Tabs>
            <TabPane key="logs" tab="日志">
              <Empty v-if="!replayData?.logs?.length" description="暂无任务日志" />
              <div v-else class="space-y-2">
                <div v-for="log in replayData.logs" :key="log.id" class="rounded border p-3 text-xs">
                  <div class="flex items-center justify-between">
                    <Space size="small">
                      <Tag color="blue">{{ log.action }}</Tag>
                      <Tag v-if="log.fromStatus || log.toStatus">{{ log.fromStatus || '-' }} -> {{ log.toStatus || '-' }}</Tag>
                      <Tag v-if="log.operator">{{ log.operator }}</Tag>
                    </Space>
                    <span class="text-gray-400">{{ formatTime(log.createdAt) }}</span>
                  </div>
                  <div v-if="log.message" class="mt-2 whitespace-pre-wrap text-gray-600">{{ log.message }}</div>
                </div>
              </div>
            </TabPane>

            <TabPane key="events" tab="事件">
              <Empty v-if="!replayData?.events?.length" description="暂无事件轨迹" />
              <div v-else class="space-y-2">
                <div v-for="event in replayData.events" :key="event.id" class="rounded border p-3 text-xs">
                  <div class="flex items-center justify-between">
                    <Space size="small">
                      <Tag color="cyan">{{ event.eventType }}</Tag>
                      <span class="font-medium">{{ event.label }}</span>
                    </Space>
                    <span class="text-gray-400">{{ formatTime(event.createdAt) }}</span>
                  </div>
                  <div v-if="event.payload" class="mt-2 whitespace-pre-wrap text-gray-500 break-all">{{ event.payload }}</div>
                </div>
              </div>
            </TabPane>

            <TabPane key="issues" tab="问题">
              <Empty v-if="!replayData?.issues?.length" description="暂无关联问题" />
              <div v-else class="space-y-2">
                <div v-for="issue in replayData.issues" :key="issue.id" class="rounded border p-3 text-xs">
                  <div class="flex items-center justify-between">
                    <Space size="small">
                      <Tag :color="issue.severity === 'blocker' ? 'red' : issue.severity === 'error' ? 'volcano' : issue.severity === 'warn' ? 'gold' : 'blue'">
                        {{ issue.severity }}
                      </Tag>
                      <Tag>{{ issue.ruleCode }}</Tag>
                      <span class="font-medium">{{ issue.title }}</span>
                    </Space>
                    <span class="text-gray-400">{{ formatTime(issue.createdAt) }}</span>
                  </div>
                  <div class="mt-2 whitespace-pre-wrap text-gray-600">{{ issue.detail }}</div>
                  <div v-if="issue.suggestedAction" class="mt-1 text-gray-500">建议：{{ issue.suggestedAction }}</div>
                </div>
              </div>
            </TabPane>

            <TabPane key="evidence" tab="证据">
              <Empty v-if="!replayData?.evidence?.length" description="暂无关联证据" />
              <div v-else class="space-y-2">
                <div v-for="item in replayData.evidence" :key="item.id" class="rounded border p-3 text-xs">
                  <div class="flex items-center justify-between">
                    <Space size="small">
                      <Tag color="purple">{{ item.evidenceType }}</Tag>
                      <Tag>{{ item.sourceType }}</Tag>
                    </Space>
                    <span class="text-gray-400">{{ formatTime(item.createdAt) }}</span>
                  </div>
                  <div class="mt-2 text-gray-700">{{ item.summary }}</div>
                  <div v-if="item.contentRef" class="mt-1 break-all text-gray-500">{{ item.contentRef }}</div>
                </div>
              </div>
            </TabPane>

            <TabPane key="handoffs" tab="交接">
              <Empty v-if="!replayData?.handoffs?.length" description="暂无交接记录" />
              <div v-else class="space-y-2">
                <div v-for="handoff in replayData.handoffs" :key="handoff.id" class="rounded border p-3 text-xs">
                  <div class="flex items-center justify-between">
                    <Space size="small">
                      <Tag color="orange">{{ handoff.handoffType }}</Tag>
                      <span>{{ handoff.fromTaskID || '-' }} -> {{ handoff.toTaskID || '-' }}</span>
                    </Space>
                    <span class="text-gray-400">{{ formatTime(handoff.createdAt) }}</span>
                  </div>
                  <div v-if="handoff.reason" class="mt-2 text-gray-700">{{ handoff.reason }}</div>
                  <div v-if="handoff.payload" class="mt-1 break-all whitespace-pre-wrap text-gray-500">{{ handoff.payload }}</div>
                </div>
              </div>
            </TabPane>

            <TabPane key="actions" tab="决策">
              <Empty v-if="!replayData?.actions?.length" description="暂无关联决策" />
              <div v-else class="space-y-2">
                <div v-for="action in replayData.actions" :key="action.id" class="rounded border p-3 text-xs">
                  <div class="flex items-center justify-between">
                    <Space size="small">
                      <Tag color="geekblue">{{ action.decisionType }}</Tag>
                      <Tag>{{ action.actionType }}</Tag>
                      <Tag :color="action.actionStatus === 'completed' ? 'green' : action.actionStatus === 'rejected' ? 'red' : 'orange'">
                        {{ action.actionStatus }}
                      </Tag>
                    </Space>
                    <span class="text-gray-400">{{ formatTime(action.createdAt) }}</span>
                  </div>
                  <div v-if="action.recommendation" class="mt-2 break-all whitespace-pre-wrap text-gray-600">{{ action.recommendation }}</div>
                </div>
              </div>
            </TabPane>
          </Tabs>
        </template>
      </Spin>
    </Drawer>
  </Page>
</template>

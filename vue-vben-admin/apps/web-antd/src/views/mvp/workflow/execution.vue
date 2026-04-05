<script setup lang="ts">
import { ref, onMounted, computed, onUnmounted } from 'vue';
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
  message,
} from 'ant-design-vue';

import {
  getExecutionStatus,
  retryTask,
  skipTask,
  type ExecutionStatusResult,
  type DomainTaskItem,
} from '../../../api/mvp/workflow';
import { stageTypeMap } from '../consts';

const route = useRoute();
const loading = ref(false);
const projectId = ref<string>('');
const data = ref<ExecutionStatusResult | null>(null);
let pollTimer: ReturnType<typeof setInterval> | null = null;

onMounted(() => {
  projectId.value = (route.query.projectId as string) ?? '';
  if (projectId.value) {
    loadData();
    pollTimer = setInterval(loadData, 5000);
  }
});

onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer);
});

async function loadData() {
  try {
    const res = await getExecutionStatus(projectId.value);
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

const statusColorMap: Record<string, string> = {
  pending: 'default',
  running: 'processing',
  completed: 'success',
  failed: 'error',
  escalated: 'warning',
};

const columns = [
  { title: '任务名称', dataIndex: 'name', key: 'name', width: 200 },
  { title: '状态', dataIndex: 'status', key: 'status', width: 100 },
  { title: '角色', dataIndex: 'roleType', key: 'roleType', width: 100 },
  { title: '执行方式', dataIndex: 'executionMode', key: 'executionMode', width: 100 },
  { title: '影响资源', dataIndex: 'affectedResources', key: 'affectedResources', width: 200 },
  { title: '重试次数', dataIndex: 'retryCount', key: 'retryCount', width: 80 },
  { title: '操作', key: 'action', width: 120 },
];

async function handleRetry(task: DomainTaskItem) {
  Modal.confirm({
    title: '确认重试',
    content: `确定要重试任务「${task.name}」吗？`,
    onOk: async () => {
      await retryTask({ projectID: projectId.value, taskID: task.id });
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
      await skipTask({ projectID: projectId.value, taskID: task.id, reason: '手动跳过' });
      message.success('已跳过');
      loadData();
    },
  });
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
              <Statistic title="已完成" :value="data.completedTasks" value-style="color: #3f8600" />
            </Col>
            <Col :span="4">
              <Statistic title="运行中" :value="data.runningTasks" value-style="color: #1890ff" />
            </Col>
            <Col :span="4">
              <Statistic title="待执行" :value="data.pendingTasks" />
            </Col>
            <Col :span="4">
              <Statistic title="失败" :value="data.failedTasks" value-style="color: #cf1322" />
            </Col>
            <Col :span="4">
              <Statistic title="已升级" :value="data.escalatedTasks" value-style="color: #fa8c16" />
            </Col>
          </Row>

          <Progress :percent="progressPercent" :status="progressStatus" />

          <div class="mt-2 text-gray-500 text-sm">
            <span v-if="data.activeBatch > 0">活跃批次: {{ data.activeBatch }}</span>
            <span v-if="data.stageStatus"> | 阶段状态:
              <Tag :color="stageTypeMap[data.stageStatus]?.color ?? 'default'" size="small">
                {{ data.stageStatus }}
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
                <Badge :status="statusColorMap[record.status] ?? 'default'" :text="record.status" />
              </template>

              <template v-else-if="column.key === 'affectedResources'">
                <Space v-if="record.affectedResources?.length" wrap size="small">
                  <Tag v-for="res in record.affectedResources" :key="res" size="small">
                    {{ res }}
                  </Tag>
                </Space>
                <span v-else class="text-gray-400">-</span>
              </template>

              <template v-else-if="column.key === 'action'">
                <Space>
                  <Button
                    v-if="record.status === 'failed'"
                    type="link"
                    size="small"
                    @click="handleRetry(record)"
                  >
                    重试
                  </Button>
                  <Button
                    v-if="record.status === 'failed' || record.status === 'pending'"
                    type="link"
                    size="small"
                    danger
                    @click="handleSkip(record)"
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
  </Page>
</template>

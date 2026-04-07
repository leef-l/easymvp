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
import { stageTypeMap, taskStatusMap } from '../consts';

const props = defineProps<{ projectId?: string }>();

const route = useRoute();
const loading = ref(false);
const resolvedProjectId = computed(() => props.projectId || (route.query.projectId as string) || '');
const data = ref<ExecutionStatusResult | null>(null);
let pollTimer: ReturnType<typeof setInterval> | null = null;

onMounted(() => {
  if (resolvedProjectId.value) {
    loadData();
    pollTimer = setInterval(loadData, 5000);
  }
});

onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer);
});

async function loadData() {
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

const statusColorMap: Record<string, string> = {
  pending: 'default',
  running: 'processing',
  completed: 'success',
  failed: 'error',
  escalated: 'warning',
};

const columns = [
  { title: '任务名称', dataIndex: 'name', key: 'name', width: 180, ellipsis: true },
  { title: '状态', dataIndex: 'status', key: 'status', width: 90 },
  { title: '角色', dataIndex: 'roleType', key: 'roleType', width: 80 },
  { title: '执行方式', dataIndex: 'executionMode', key: 'executionMode', width: 80 },
  { title: '结果/错误', key: 'resultInfo', width: 200, ellipsis: true },
  { title: '重试', dataIndex: 'retryCount', key: 'retryCount', width: 50 },
  { title: '操作', key: 'action', width: 140 },
];

function showDetail(task: DomainTaskItem) {
  Modal.info({
    title: task.name,
    width: 700,
    content: () => {
      const h = (window as any).__vue_create_element__ || null;
      // 用纯 HTML 字符串渲染
      return undefined;
    },
    okText: '关闭',
  });
  // 用 setTimeout 替换 content（Modal.info 不支持复杂 VNode）
  setTimeout(() => {
    const modalBody = document.querySelector('.ant-modal-confirm-content');
    if (modalBody) {
      modalBody.innerHTML = `
        <div style="max-height: 500px; overflow-y: auto; font-size: 13px; line-height: 1.6;">
          <p><b>描述：</b></p>
          <pre style="white-space: pre-wrap; background: #f5f5f5; padding: 8px; border-radius: 4px; max-height: 200px; overflow-y: auto;">${escapeHtml(task.description || '无')}</pre>
          <p style="margin-top: 12px;"><b>状态：</b> ${task.status} &nbsp; <b>角色：</b> ${task.roleType}/${task.roleLevel} &nbsp; <b>执行方式：</b> ${task.executionMode || '-'}</p>
          <p><b>批次：</b> ${task.batchNo} &nbsp; <b>重试次数：</b> ${task.retryCount}</p>
          ${task.affectedResources?.length ? `<p><b>影响资源：</b> ${task.affectedResources.join(', ')}</p>` : ''}
          ${task.result ? `<p style="margin-top: 12px;"><b>执行结果：</b></p><pre style="white-space: pre-wrap; background: #f0fff0; padding: 8px; border-radius: 4px; max-height: 200px; overflow-y: auto;">${escapeHtml(task.result)}</pre>` : ''}
          ${task.errorMessage ? `<p style="margin-top: 12px;"><b>错误信息：</b></p><pre style="white-space: pre-wrap; background: #fff0f0; padding: 8px; border-radius: 4px; max-height: 200px; overflow-y: auto; color: #cf1322;">${escapeHtml(task.errorMessage)}</pre>` : ''}
          ${task.startedAt ? `<p style="margin-top: 8px; color: #999;"><b>开始：</b> ${task.startedAt}</p>` : ''}
          ${task.completedAt ? `<p style="color: #999;"><b>完成：</b> ${task.completedAt}</p>` : ''}
        </div>
      `;
    }
  }, 100);
}

function escapeHtml(str: string): string {
  return str.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
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
                <Badge :status="statusColorMap[record.status] ?? 'default'" :text="taskStatusMap[record.status]?.label ?? record.status" />
              </template>

              <template v-else-if="column.key === 'resultInfo'">
                <span v-if="record.errorMessage" class="text-red-500 text-xs cursor-pointer" @click="showDetail(record)">
                  {{ record.errorMessage.length > 50 ? record.errorMessage.slice(0, 50) + '...' : record.errorMessage }}
                </span>
                <span v-else-if="record.result" class="text-green-600 text-xs cursor-pointer" @click="showDetail(record)">
                  {{ record.result.length > 50 ? record.result.slice(0, 50) + '...' : record.result }}
                </span>
                <span v-else class="text-gray-400 text-xs">-</span>
              </template>

              <template v-else-if="column.key === 'action'">
                <Space>
                  <Button type="link" size="small" @click="showDetail(record)">详情</Button>
                  <Button
                    v-if="record.status === 'failed' || record.status === 'escalated'"
                    type="link"
                    size="small"
                    @click="handleRetry(record)"
                  >
                    重试
                  </Button>
                  <Button
                    v-if="['failed', 'escalated', 'pending'].includes(record.status)"
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

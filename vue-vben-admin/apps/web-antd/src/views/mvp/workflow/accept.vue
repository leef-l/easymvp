<script setup lang="ts">
// @ts-nocheck
import { ref, onUnmounted, computed, watch } from 'vue';
import { useRoute } from 'vue-router';
import {
  Card,
  Descriptions,
  DescriptionsItem,
  Table,
  Tag,
  Button,
  Space,
  Spin,
  Empty,
  Modal,
  Input,
  Progress,
  Badge,
  Select,
  message,
} from 'ant-design-vue';
import {
  getAcceptStatus,
  getAcceptIssues,
  getAcceptEvidence,
  acceptApprove,
  acceptReject,
  acceptRerun,
  acceptRework,
  acceptIssueRework,
  type AcceptStatusResult,
  type AcceptIssueItem,
  type AcceptEvidenceItem,
} from '#/api/mvp/workflow';

const props = defineProps<{ projectId?: string }>();
const emit = defineEmits<{
  changed: [];
}>();

const route = useRoute();
const projectID = computed(() => props.projectId || (route.query.projectId as string) || '');

const loading = ref(false);
const status = ref<AcceptStatusResult | null>(null);
const issues = ref<AcceptIssueItem[]>([]);
const evidence = ref<AcceptEvidenceItem[]>([]);
const severityFilter = ref<string | undefined>(undefined);
const rejectReason = ref('');
const reworkReason = ref('');
const rejectModalVisible = ref(false);
const reworkModalVisible = ref(false);
let pollTimer: ReturnType<typeof setInterval> | null = null;
let loadRequestVersion = 0;

const formatScore = (percent?: number) => `${(percent ?? 0).toFixed(1)}`;

/** 加载全部数据 */
async function loadData() {
  const currentProjectId = projectID.value;
  if (!currentProjectId) {
    resetAcceptState();
    return;
  }
  const requestVersion = ++loadRequestVersion;
  loading.value = true;
  try {
    const [statusRes, issuesRes, evidenceRes] = await Promise.all([
      getAcceptStatus(currentProjectId),
      getAcceptIssues(currentProjectId, severityFilter.value),
      getAcceptEvidence(currentProjectId),
    ]);
    if (requestVersion !== loadRequestVersion || currentProjectId !== projectID.value) return;
    status.value = statusRes;
    issues.value = issuesRes?.issues ?? [];
    evidence.value = evidenceRes?.evidence ?? [];
  } catch {
    /* ignore */
  } finally {
    if (requestVersion === loadRequestVersion && currentProjectId === projectID.value) {
      loading.value = false;
    }
  }
}

/** 决策颜色 */
function decisionColor(d: string) {
  switch (d) {
    case 'passed':
      return 'green';
    case 'failed':
      return 'red';
    case 'manual_review':
      return 'orange';
    default:
      return 'blue';
  }
}

/** 决策文案 */
function decisionLabel(d: string) {
  switch (d) {
    case 'passed':
      return '通过';
    case 'failed':
      return '不通过';
    case 'manual_review':
      return '待人工审核';
    default:
      return d || '-';
  }
}

/** 状态颜色 */
function statusColor(s: string) {
  switch (s) {
    case 'completed':
      return 'green';
    case 'running':
      return 'processing';
    case 'failed':
      return 'red';
    default:
      return 'default';
  }
}

/** 严重级别颜色 */
const severityColor: Record<string, string> = {
  blocker: '#f5222d',
  error: '#fa541c',
  warn: '#faad14',
  info: '#1890ff',
};

/** 是否需要人工操作 */
const needsAction = computed(() => {
  if (!status.value) return false;
  const d = status.value.decision;
  const s = status.value.status;
  return d === 'manual_review' || d === 'failed' || (s === 'completed' && d !== 'passed');
});

/** 评分颜色 */
function scoreColor(score: number) {
  if (score >= 90) return '#52c41a';
  if (score >= 70) return '#1890ff';
  if (score >= 60) return '#faad14';
  return '#f5222d';
}

function evidenceTypeColor(type: string) {
  switch (type) {
    case 'ci':
      return 'geekblue';
    case 'delivery':
      return 'magenta';
    case 'diff':
      return 'purple';
    case 'task_output':
      return 'blue';
    case 'stage_output':
      return 'cyan';
    case 'handoff':
      return 'orange';
    default:
      return 'default';
  }
}

function sourceTypeColor(type: string) {
  switch (type) {
    case 'project_repo':
      return 'geekblue';
    case 'task_log':
      return 'gold';
    case 'workspace':
      return 'purple';
    case 'domain_task':
      return 'blue';
    case 'stage_run':
      return 'cyan';
    case 'handoff_record':
      return 'orange';
    default:
      return 'default';
  }
}

// 问题列表列
const issueColumns = [
  { title: '级别', dataIndex: 'severity', key: 'severity', width: 80 },
  { title: '规则', dataIndex: 'ruleCode', key: 'ruleCode', width: 140 },
  { title: '标题', dataIndex: 'title', key: 'title', width: 200 },
  { title: '详情', dataIndex: 'detail', key: 'detail' },
  { title: '期望', dataIndex: 'expectedValue', key: 'expectedValue', width: 120 },
  { title: '实际', dataIndex: 'actualValue', key: 'actualValue', width: 120 },
  { title: '建议', dataIndex: 'suggestedAction', key: 'suggestedAction', width: 160 },
  { title: '状态', dataIndex: 'status', key: 'status', width: 80 },
  { title: '操作', key: 'action', width: 110 },
];

// 证据列表列
const evidenceColumns = [
  { title: '类型', dataIndex: 'evidenceType', key: 'evidenceType', width: 120 },
  { title: '来源', dataIndex: 'sourceType', key: 'sourceType', width: 120 },
  { title: '摘要', dataIndex: 'summary', key: 'summary' },
  { title: '引用', dataIndex: 'contentRef', key: 'contentRef', width: 220, ellipsis: true },
  { title: '时间', dataIndex: 'createdAt', key: 'createdAt', width: 160 },
];

/** 人工放行 */
async function handleApprove() {
  Modal.confirm({
    title: '确认人工放行？',
    content: '放行后项目将推进到完成阶段',
    onOk: async () => {
      await acceptApprove(projectID.value, '管理员手动放行');
      message.success('已放行');
      await loadData();
      emit('changed');
    },
  });
}

/** 驳回 */
async function handleReject() {
  if (!rejectReason.value.trim()) {
    message.warning('请输入驳回原因');
    return;
  }
  await acceptReject(projectID.value, rejectReason.value);
  message.success('已驳回');
  rejectModalVisible.value = false;
  rejectReason.value = '';
  await loadData();
  emit('changed');
}

/** 重新验收 */
async function handleRerun() {
  Modal.confirm({
    title: '确认重新验收？',
    content: '将创建新一轮验收运行',
    onOk: async () => {
      await acceptRerun(projectID.value);
      message.success('已启动重新验收');
      await loadData();
      emit('changed');
    },
  });
}

/** 驳回并返工 */
async function handleRework() {
  if (!reworkReason.value.trim()) {
    message.warning('请输入返工原因');
    return;
  }
  await acceptRework(projectID.value, reworkReason.value);
  message.success('已触发返工');
  reworkModalVisible.value = false;
  reworkReason.value = '';
  await loadData();
  emit('changed');
}

function handleIssueRework(issue: AcceptIssueItem) {
  Modal.confirm({
    title: '将问题转为返工',
    content: `问题「${issue.title}」将作为返工原因回流到执行链路。`,
    onOk: async () => {
      await acceptIssueRework(projectID.value, [issue.id]);
      message.success('已触发返工');
      await loadData();
      emit('changed');
    },
  });
}

function handleIssueReworkRecord(record: Record<string, any>) {
  handleIssueRework(record as AcceptIssueItem);
}

/** 启动/停止轮询 */
function startPoll() {
  stopPoll();
  pollTimer = setInterval(() => {
    if (status.value?.status === 'running') {
      loadData();
    }
  }, 5000);
}
function stopPoll() {
  if (pollTimer) {
    clearInterval(pollTimer);
    pollTimer = null;
  }
}

function resetAcceptState() {
  status.value = null;
  issues.value = [];
  evidence.value = [];
  severityFilter.value = undefined;
  rejectReason.value = '';
  reworkReason.value = '';
  rejectModalVisible.value = false;
  reworkModalVisible.value = false;
}

watch(
  projectID,
  (value) => {
    stopPoll();
    resetAcceptState();
    if (!value) return;
    loadData();
    startPoll();
  },
  { immediate: true },
);
onUnmounted(stopPoll);
</script>

<template>
  <div class="p-4">
    <Spin :spinning="loading">
      <Empty v-if="!status || status.status === 'none'" description="暂无验收数据" />
      <template v-else>
        <!-- 验收概览 -->
        <Card title="验收概览" class="mb-4">
          <Descriptions :column="3" bordered size="small">
            <DescriptionsItem label="验收轮次">
              第 {{ status.acceptRound }} 轮
            </DescriptionsItem>
            <DescriptionsItem label="状态">
              <Tag :color="statusColor(status.status)">
                {{ status.status }}
              </Tag>
            </DescriptionsItem>
            <DescriptionsItem label="决策">
              <Tag :color="decisionColor(status.decision)">
                {{ decisionLabel(status.decision) }}
              </Tag>
            </DescriptionsItem>
            <DescriptionsItem label="评分">
              <Progress
                :percent="status.score"
                :stroke-color="scoreColor(status.score)"
                :format="formatScore"
                :size="[200, 14]"
              />
            </DescriptionsItem>
            <DescriptionsItem label="开始时间">
              {{ status.startedAt || '-' }}
            </DescriptionsItem>
            <DescriptionsItem label="完成时间">
              {{ status.finishedAt || '-' }}
            </DescriptionsItem>
            <DescriptionsItem label="阻塞">
              <Badge :count="status.blockerCount" :number-style="{ backgroundColor: '#f5222d' }" show-zero />
            </DescriptionsItem>
            <DescriptionsItem label="错误">
              <Badge :count="status.errorCount" :number-style="{ backgroundColor: '#fa541c' }" show-zero />
            </DescriptionsItem>
            <DescriptionsItem label="警告">
              <Badge :count="status.warnCount" :number-style="{ backgroundColor: '#faad14' }" show-zero />
            </DescriptionsItem>
          </Descriptions>
          <div v-if="status.summary" class="mt-2 text-gray-500">
            {{ status.summary }}
          </div>

          <!-- 操作按钮 -->
          <div v-if="needsAction" class="mt-4 text-right">
            <Space>
              <Button type="primary" @click="handleApprove">人工放行</Button>
              <Button @click="handleRerun">重新验收</Button>
              <Button danger @click="rejectModalVisible = true">驳回</Button>
              <Button danger type="dashed" @click="reworkModalVisible = true">驳回并返工</Button>
            </Space>
          </div>
        </Card>

        <!-- 问题列表 -->
        <Card class="mb-4">
          <template #title>
            <Space>
              <span>验收问题</span>
              <Select
                v-model:value="severityFilter"
                placeholder="按级别过滤"
                allow-clear
                style="width: 140px"
                @change="loadData"
              >
                <Select.Option value="blocker">阻塞</Select.Option>
                <Select.Option value="error">错误</Select.Option>
                <Select.Option value="warn">警告</Select.Option>
                <Select.Option value="info">提示</Select.Option>
              </Select>
            </Space>
          </template>
          <Table
            :columns="issueColumns"
            :data-source="issues"
            :pagination="false"
            row-key="id"
            size="small"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'severity'">
                <Tag :color="severityColor[record.severity] || 'default'">
                  {{ record.severity }}
                </Tag>
              </template>
              <template v-if="column.key === 'status'">
                <Tag :color="record.status === 'open' ? 'red' : 'green'">
                  {{ record.status === 'open' ? '待处理' : '已解决' }}
                </Tag>
              </template>
              <template v-else-if="column.key === 'action'">
                <Button
                  v-if="record.status === 'open'"
                  type="link"
                  size="small"
                  @click="handleIssueReworkRecord(record)"
                >
                  转返工
                </Button>
                <span v-else class="text-xs text-gray-400">-</span>
              </template>
            </template>
          </Table>
        </Card>

        <!-- 证据列表 -->
        <Card title="验收证据">
          <Table
            :columns="evidenceColumns"
            :data-source="evidence"
            :pagination="false"
            row-key="id"
            size="small"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'evidenceType'">
                <Tag :color="evidenceTypeColor(record.evidenceType)">{{ record.evidenceType }}</Tag>
              </template>
              <template v-else-if="column.key === 'sourceType'">
                <Tag :color="sourceTypeColor(record.sourceType)">{{ record.sourceType }}</Tag>
              </template>
              <template v-else-if="column.key === 'contentRef'">
                <span v-if="record.contentRef" class="text-xs text-gray-500 break-all">
                  {{ record.contentRef }}
                </span>
                <span v-else class="text-xs text-gray-400">-</span>
              </template>
            </template>
          </Table>
        </Card>
      </template>
    </Spin>

    <!-- 驳回弹窗 -->
    <Modal v-model:open="rejectModalVisible" title="驳回验收" @ok="handleReject">
      <Input.TextArea v-model:value="rejectReason" :rows="4" placeholder="请输入驳回原因" />
    </Modal>

    <!-- 返工弹窗 -->
    <Modal v-model:open="reworkModalVisible" title="驳回并返工" @ok="handleRework">
      <Input.TextArea v-model:value="reworkReason" :rows="4" placeholder="请输入返工原因" />
    </Modal>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { useRoute } from 'vue-router';
import {
  Card,
  Table,
  Tabs,
  TabPane,
  Tag,
  Button,
  Space,
  Spin,
  Empty,
  Modal,
  Select,
  Input,
  Typography,
  message,
} from 'ant-design-vue';
import {
  getAutonomyCheckpoints,
  autonomyApprove,
  autonomyReject,
  triggerReplan,
  getProjectReports,
  triggerReport,
  getAutonomyMode,
  setAutonomyMode,
  type AutonomyCheckpointItem,
  type AutonomyActionItem,
  type ProjectReportItem,
} from '#/api/mvp/workflow';

const route = useRoute();
const projectID = computed(() => (route.query.projectId as string) || '');

const loading = ref(false);
const activeTab = ref('decisions');
const checkpoints = ref<AutonomyCheckpointItem[]>([]);
const actions = ref<AutonomyActionItem[]>([]);
const reports = ref<ProjectReportItem[]>([]);
const reportTypeFilter = ref<string | undefined>(undefined);
const selectedReport = ref<ProjectReportItem | null>(null);
const reportDetailVisible = ref(false);
const autonomyMode = ref<string>('suggest');
const modeLoading = ref(false);
const rejectReason = ref('');

/** actionID → action 的快查表 */
const actionMap = computed(() => {
  const map = new Map<string, AutonomyActionItem>();
  for (const a of actions.value) {
    map.set(a.id, a);
  }
  return map;
});

/** 加载数据 */
async function loadData() {
  if (!projectID.value) return;
  loading.value = true;
  try {
    const [checkpointsRes, reportsRes] = await Promise.all([
      getAutonomyCheckpoints(projectID.value),
      getProjectReports(projectID.value, reportTypeFilter.value),
    ]);
    checkpoints.value = checkpointsRes?.checkpoints ?? [];
    actions.value = checkpointsRes?.actions ?? [];
    reports.value = reportsRes?.reports ?? [];
  } catch {
    /* ignore */
  } finally {
    loading.value = false;
  }
}

/** 待审批数 */
const pendingCount = computed(() =>
  checkpoints.value.filter((c) => c.status === 'open').length,
);

/** 决策等级标签 */
const levelMap: Record<string, { label: string; color: string }> = {
  A: { label: 'A 自动', color: 'green' },
  B: { label: 'B 建议', color: 'orange' },
  C: { label: 'C 人工', color: 'red' },
};

/** 动作类型标签 */
const actionTypeMap: Record<string, string> = {
  retry_task: '重试任务',
  trigger_rework: '触发返工',
  rerun_accept: '重新验收',
  pause_workflow: '暂停工作流',
  switch_executor: '切换执行器',
  replan_workflow: '重规划',
  notify_human: '通知人工',
  approve_complete: '批准完成',
};

/** 触发源标签 */
const triggerSourceMap: Record<string, string> = {
  'task.failed': '任务失败',
  'task.timeout': '任务超时',
  'task.retry_exhausted': '重试耗尽',
  'accept.passed': '验收通过',
  'accept.failed': '验收失败',
  'accept.manual_review': '人工审查',
  'workflow.circuit_break': '熔断',
  'rework.completed': '返工完成',
  'replan.suggested': '建议重规划',
  'human.override': '人工覆盖',
};

/** 节点状态 */
const checkpointStatusMap: Record<string, { label: string; color: string }> = {
  open: { label: '待处理', color: 'orange' },
  handled: { label: '已处理', color: 'green' },
  expired: { label: '已过期', color: 'default' },
  canceled: { label: '已取消', color: 'default' },
};

/** 报告类型标签 */
const reportTypeMap: Record<string, string> = {
  stage: '阶段报告',
  daily: '日报',
  weekly: '周报',
  summary: '总结',
};

// 决策列表列（基于 checkpoints + actions 联合展示）
const decisionColumns = [
  { title: '标题', dataIndex: 'title', key: 'title', ellipsis: true },
  { title: '类型', key: 'checkpointType', width: 100 },
  { title: '等级', key: 'decisionLevel', width: 80 },
  { title: '动作', key: 'actionType', width: 120 },
  { title: '触发源', key: 'triggerSource', width: 120 },
  { title: '建议', key: 'recommendation', ellipsis: true },
  { title: '状态', key: 'status', width: 100 },
  { title: '时间', dataIndex: 'createdAt', key: 'createdAt', width: 160 },
  { title: '操作', key: 'action', width: 200 },
];

// 报告列表列
const reportColumns = [
  { title: '类型', dataIndex: 'reportType', key: 'reportType', width: 100 },
  { title: '阶段', dataIndex: 'stageType', key: 'stageType', width: 80 },
  { title: '标题', dataIndex: 'title', key: 'title' },
  { title: '时间', dataIndex: 'createdAt', key: 'createdAt', width: 160 },
  { title: '操作', key: 'action', width: 80 },
];

/** 获取 checkpoint 关联的 action */
function getAction(record: AutonomyCheckpointItem): AutonomyActionItem | undefined {
  return actionMap.value.get(record.decisionActionId);
}

/** 解析建议 JSON */
function parseRecommendation(json?: string): string {
  if (!json) return '-';
  try {
    const obj = JSON.parse(json);
    if (obj.action) return `${obj.action}: ${obj.reasoning || obj.reason || ''}`;
    if (obj.original_action) return `原定: ${obj.original_action}`;
    return json;
  } catch {
    return json;
  }
}

/** 批准决策 */
async function handleApprove(record: AutonomyCheckpointItem) {
  const action = getAction(record);
  if (!action) {
    message.error('未找到关联的决策动作');
    return;
  }
  Modal.confirm({
    title: '确认批准该决策？',
    content: `${record.title}\n动作: ${actionTypeMap[action.actionType] || action.actionType}`,
    onOk: async () => {
      await autonomyApprove(action.id);
      message.success('已批准');
      loadData();
    },
  });
}

/** 拒绝决策弹窗 */
const rejectModalVisible = ref(false);
const rejectTarget = ref<AutonomyCheckpointItem | null>(null);

function handleReject(record: AutonomyCheckpointItem) {
  const action = getAction(record);
  if (!action) {
    message.error('未找到关联的决策动作');
    return;
  }
  rejectReason.value = '';
  rejectTarget.value = record;
  rejectModalVisible.value = true;
}

async function confirmReject() {
  if (!rejectTarget.value) return;
  const action = getAction(rejectTarget.value);
  if (!action) return;
  if (!rejectReason.value.trim()) {
    message.warning('请输入驳回理由');
    return;
  }
  await autonomyReject(action.id, rejectReason.value);
  message.success('已驳回');
  rejectModalVisible.value = false;
  rejectTarget.value = null;
  loadData();
}

/** 手动触发重规划 */
async function handleTriggerReplan() {
  Modal.confirm({
    title: '确认触发重规划？',
    content: '系统将分析当前失败任务并生成重规划建议',
    onOk: async () => {
      await triggerReplan(projectID.value);
      message.success('已触发重规划分析');
      loadData();
    },
  });
}

/** 手动生成报告 */
async function handleTriggerReport() {
  await triggerReport(projectID.value);
  message.success('报告已生成');
  loadData();
}

/** 查看报告详情 */
function viewReport(record: ProjectReportItem) {
  selectedReport.value = record;
  reportDetailVisible.value = true;
}

/** 加载自治模式 */
async function loadMode() {
  try {
    const res = await getAutonomyMode();
    autonomyMode.value = res.mode || 'suggest';
  } catch {
    /* ignore */
  }
}

/** 切换自治模式 */
async function handleModeChange(mode: string) {
  modeLoading.value = true;
  try {
    await setAutonomyMode(mode as 'auto' | 'suggest');
    autonomyMode.value = mode;
    message.success(`已切换为${mode === 'auto' ? '全自动' : '建议'}模式`);
  } catch {
    message.error('模式切换失败');
  } finally {
    modeLoading.value = false;
  }
}

onMounted(() => {
  loadData();
  loadMode();
});
</script>

<template>
  <div class="p-4">
    <!-- 自治模式切换 -->
    <Card size="small" class="mb-4">
      <Space align="center">
        <span class="font-medium">自治模式：</span>
        <Select
          :value="autonomyMode"
          :loading="modeLoading"
          style="width: 160px"
          @change="handleModeChange"
        >
          <Select.Option value="suggest">
            <Tag color="orange">建议型</Tag>
            决策需人工确认
          </Select.Option>
          <Select.Option value="auto">
            <Tag color="blue">全自动</Tag>
            系统自动执行
          </Select.Option>
        </Select>
        <Tag v-if="autonomyMode === 'auto'" color="red">注意：全自动模式下系统将自动执行重规划等操作</Tag>
      </Space>
    </Card>

    <Spin :spinning="loading">
      <Tabs v-model:activeKey="activeTab">
        <!-- 决策中心 -->
        <TabPane key="decisions">
          <template #tab>
            <span>
              决策中心
              <Tag v-if="pendingCount > 0" color="red" class="ml-1">{{ pendingCount }}</Tag>
            </span>
          </template>

          <Card size="small" class="mb-4">
            <Space>
              <Button type="primary" @click="handleTriggerReplan">手动触发重规划</Button>
            </Space>
          </Card>

          <Empty v-if="checkpoints.length === 0" description="暂无待处理决策" />
          <Table
            v-else
            :columns="decisionColumns"
            :data-source="checkpoints"
            :pagination="false"
            row-key="id"
            size="small"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'checkpointType'">
                <Tag color="blue">
                  {{ record.checkpointType === 'manual_review' ? '人工审查' : record.checkpointType === 'approval' ? '审批' : '升级' }}
                </Tag>
              </template>
              <template v-if="column.key === 'decisionLevel'">
                <Tag :color="levelMap[getAction(record)?.decisionLevel]?.color || 'default'">
                  {{ levelMap[getAction(record)?.decisionLevel]?.label || getAction(record)?.decisionLevel || '-' }}
                </Tag>
              </template>
              <template v-if="column.key === 'actionType'">
                <Tag color="purple">
                  {{ actionTypeMap[getAction(record)?.actionType] || getAction(record)?.actionType || '-' }}
                </Tag>
              </template>
              <template v-if="column.key === 'triggerSource'">
                {{ triggerSourceMap[getAction(record)?.triggerSource] || getAction(record)?.triggerSource || '-' }}
              </template>
              <template v-if="column.key === 'recommendation'">
                <Typography.Text :ellipsis="{ tooltip: true }" style="max-width: 300px">
                  {{ parseRecommendation(getAction(record)?.recommendation) }}
                </Typography.Text>
              </template>
              <template v-if="column.key === 'status'">
                <Tag :color="checkpointStatusMap[record.status]?.color || 'default'">
                  {{ checkpointStatusMap[record.status]?.label || record.status }}
                </Tag>
              </template>
              <template v-if="column.key === 'action'">
                <Space v-if="record.status === 'open'">
                  <Button size="small" type="primary" @click="handleApprove(record)">批准</Button>
                  <Button size="small" danger @click="handleReject(record)">驳回</Button>
                </Space>
                <span v-else class="text-gray-400">
                  {{ record.handleAction ? (record.handleAction === 'approve' ? '已批准' : '已驳回') : '-' }}
                </span>
              </template>
            </template>
          </Table>
        </TabPane>

        <!-- 报告中心 -->
        <TabPane key="reports" tab="报告中心">
          <Card size="small" class="mb-4">
            <Space>
              <Select
                v-model:value="reportTypeFilter"
                placeholder="按类型过滤"
                allow-clear
                style="width: 140px"
                @change="loadData"
              >
                <Select.Option value="stage">阶段报告</Select.Option>
                <Select.Option value="daily">日报</Select.Option>
                <Select.Option value="weekly">周报</Select.Option>
                <Select.Option value="summary">总结</Select.Option>
              </Select>
              <Button type="primary" @click="handleTriggerReport">生成报告</Button>
            </Space>
          </Card>

          <Empty v-if="reports.length === 0" description="暂无报告" />
          <Table
            v-else
            :columns="reportColumns"
            :data-source="reports"
            :pagination="false"
            row-key="id"
            size="small"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'reportType'">
                <Tag color="blue">{{ reportTypeMap[record.reportType] || record.reportType }}</Tag>
              </template>
              <template v-if="column.key === 'action'">
                <Button size="small" type="link" @click="viewReport(record)">查看</Button>
              </template>
            </template>
          </Table>
        </TabPane>
      </Tabs>
    </Spin>

    <!-- 报告详情弹窗 -->
    <Modal
      v-model:open="reportDetailVisible"
      :title="selectedReport?.title || '报告详情'"
      width="700px"
      :footer="null"
    >
      <div v-if="selectedReport" v-html="selectedReport.content" class="prose max-w-none" />
    </Modal>

    <!-- 驳回理由弹窗 -->
    <Modal
      v-model:open="rejectModalVisible"
      title="驳回决策"
      @ok="confirmReject"
    >
      <p v-if="rejectTarget">{{ rejectTarget.title }}</p>
      <Input.TextArea
        v-model:value="rejectReason"
        placeholder="请输入驳回理由"
        :rows="3"
      />
    </Modal>
  </div>
</template>

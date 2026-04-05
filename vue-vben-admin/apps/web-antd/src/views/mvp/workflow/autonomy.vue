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
  getAutonomyActions,
  getAutonomyGateRules,
  getAutonomyPolicyRules,
  autonomyApprove,
  autonomyReject,
  triggerReplan,
  getProjectReports,
  triggerReport,
  getAutonomyMode,
  setAutonomyMode,
  type AutonomyCheckpointItem,
  type AutonomyActionItem,
  type RiskGateRuleItem,
  type PolicyRuleItem,
  type ProjectReportItem,
} from '#/api/mvp/workflow';

const route = useRoute();
const projectID = computed(() => (route.query.projectId as string) || '');

const loading = ref(false);
const activeTab = ref('decisions');

// 决策中心
const checkpoints = ref<AutonomyCheckpointItem[]>([]);
const pendingActions = ref<AutonomyActionItem[]>([]);

// 审计历史
const allActions = ref<AutonomyActionItem[]>([]);
const actionStatusFilter = ref<string | undefined>(undefined);
const decisionTypeFilter = ref<string | undefined>(undefined);
const historyLoading = ref(false);

// 风险闸门 & 策略规则
const gateRules = ref<RiskGateRuleItem[]>([]);
const policyRules = ref<PolicyRuleItem[]>([]);

// 报告
const reports = ref<ProjectReportItem[]>([]);
const reportTypeFilter = ref<string | undefined>(undefined);
const selectedReport = ref<ProjectReportItem | null>(null);
const reportDetailVisible = ref(false);

// 自治模式
const autonomyMode = ref<string>('suggest');
const modeLoading = ref(false);

// 驳回弹窗
const rejectReason = ref('');
const rejectModalVisible = ref(false);
const rejectTarget = ref<AutonomyCheckpointItem | null>(null);

/** actionID → action 的快查表 */
const actionMap = computed(() => {
  const map = new Map<string, AutonomyActionItem>();
  for (const a of pendingActions.value) {
    map.set(a.id, a);
  }
  return map;
});

/** 加载决策中心 + 报告 */
async function loadData() {
  if (!projectID.value) return;
  loading.value = true;
  try {
    const [checkpointsRes, reportsRes] = await Promise.all([
      getAutonomyCheckpoints(projectID.value),
      getProjectReports(projectID.value, reportTypeFilter.value),
    ]);
    checkpoints.value = checkpointsRes?.checkpoints ?? [];
    pendingActions.value = checkpointsRes?.actions ?? [];
    reports.value = reportsRes?.reports ?? [];
  } catch {
    /* ignore */
  } finally {
    loading.value = false;
  }
}

/** 加载审计历史（独立刷新） */
async function loadHistory() {
  if (!projectID.value) return;
  historyLoading.value = true;
  try {
    const res = await getAutonomyActions(projectID.value, actionStatusFilter.value, decisionTypeFilter.value);
    allActions.value = res?.actions ?? [];
  } catch {
    /* ignore */
  } finally {
    historyLoading.value = false;
  }
}

/** 加载风险闸门 & 策略规则 */
async function loadRules() {
  if (!projectID.value) return;
  try {
    const [gateRes, policyRes] = await Promise.all([
      getAutonomyGateRules(projectID.value),
      getAutonomyPolicyRules(projectID.value),
    ]);
    gateRules.value = gateRes?.rules ?? [];
    policyRules.value = policyRes?.rules ?? [];
  } catch {
    /* ignore */
  }
}

/** 待审批数 */
const pendingCount = computed(() =>
  checkpoints.value.filter((c) => c.status === 'open').length,
);

// ==================== 映射表 ====================

const levelMap: Record<string, { label: string; color: string }> = {
  A: { label: 'A 自动', color: 'green' },
  B: { label: 'B 建议', color: 'orange' },
  C: { label: 'C 人工', color: 'red' },
};

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

/** 决策类型（decision_type）与动作类型（action_type）是不同语义 */
const decisionTypeMap: Record<string, string> = {
  task_recovery: '任务恢复',
  rework_flow: '返工流程',
  accept_flow: '验收流程',
  circuit_break: '熔断',
  replan: '重规划',
};

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

const checkpointStatusMap: Record<string, { label: string; color: string }> = {
  open: { label: '待处理', color: 'orange' },
  handled: { label: '已处理', color: 'green' },
  expired: { label: '已过期', color: 'default' },
  canceled: { label: '已取消', color: 'default' },
};

const actionStatusMap: Record<string, { label: string; color: string }> = {
  pending: { label: '待处理', color: 'blue' },
  waiting_human: { label: '等待人工', color: 'orange' },
  auto_executed: { label: '自动执行', color: 'green' },
  approved: { label: '已批准', color: 'green' },
  rejected: { label: '已驳回', color: 'red' },
  failed: { label: '执行失败', color: 'red' },
};

const gateTypeMap: Record<string, { label: string; color: string }> = {
  permission: { label: '权限', color: 'red' },
  quality: { label: '质量', color: 'orange' },
  cost: { label: '成本', color: 'gold' },
  runtime: { label: '运行时', color: 'blue' },
};

const reportTypeMap: Record<string, string> = {
  stage: '阶段报告',
  daily: '日报',
  weekly: '周报',
  summary: '总结',
};

// ==================== 列定义 ====================

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

const historyColumns = [
  { title: '时间', dataIndex: 'createdAt', key: 'createdAt', width: 160 },
  { title: '等级', key: 'decisionLevel', width: 80 },
  { title: '类型', key: 'decisionType', width: 100 },
  { title: '动作', key: 'actionType', width: 120 },
  { title: '触发源', key: 'triggerSource', width: 120 },
  { title: '状态', key: 'actionStatus', width: 100 },
  { title: '建议', key: 'recommendation', ellipsis: true },
  { title: '结果', key: 'result', width: 150, ellipsis: true },
];

const gateRuleColumns = [
  { title: '编码', dataIndex: 'gateCode', key: 'gateCode', width: 150 },
  { title: '名称', dataIndex: 'gateName', key: 'gateName' },
  { title: '类型', key: 'gateType', width: 100 },
  { title: '阻断动作', key: 'blockAction', width: 120 },
  { title: '降级动作', key: 'fallbackAction', width: 120 },
  { title: '优先级', dataIndex: 'priority', key: 'priority', width: 80 },
  { title: '状态', key: 'enabled', width: 80 },
];

const policyRuleColumns = [
  { title: '编码', dataIndex: 'ruleCode', key: 'ruleCode', width: 150 },
  { title: '名称', dataIndex: 'ruleName', key: 'ruleName' },
  { title: '决策类型', key: 'decisionType', width: 100 },
  { title: '等级', key: 'decisionLevel', width: 80 },
  { title: '触发源', key: 'triggerSource', width: 120 },
  { title: '优先级', dataIndex: 'priority', key: 'priority', width: 80 },
  { title: '状态', key: 'enabled', width: 80 },
];

const reportColumns = [
  { title: '类型', dataIndex: 'reportType', key: 'reportType', width: 100 },
  { title: '阶段', dataIndex: 'stageType', key: 'stageType', width: 80 },
  { title: '标题', dataIndex: 'title', key: 'title' },
  { title: '时间', dataIndex: 'createdAt', key: 'createdAt', width: 160 },
  { title: '操作', key: 'action', width: 80 },
];

// ==================== 辅助函数 ====================

function getAction(record: AutonomyCheckpointItem): AutonomyActionItem | undefined {
  return actionMap.value.get(record.decisionActionId);
}

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

function parseResult(json?: string): string {
  if (!json) return '-';
  try {
    const obj = JSON.parse(json);
    if (typeof obj === 'string') return obj;
    return JSON.stringify(obj);
  } catch {
    return json;
  }
}

// ==================== 操作 ====================

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

async function handleTriggerReport() {
  await triggerReport(projectID.value);
  message.success('报告已生成');
  loadData();
}

function viewReport(record: ProjectReportItem) {
  selectedReport.value = record;
  reportDetailVisible.value = true;
}

async function loadMode() {
  try {
    const res = await getAutonomyMode();
    autonomyMode.value = res.mode || 'suggest';
  } catch {
    /* ignore */
  }
}

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

/** Tab 切换时按需加载 */
function onTabChange(key: string) {
  if (key === 'history' && allActions.value.length === 0) {
    loadHistory();
  }
  if (key === 'gates' && gateRules.value.length === 0) {
    loadRules();
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
      <Tabs v-model:activeKey="activeTab" @change="onTabChange">
        <!-- Tab 1: 决策中心 -->
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

        <!-- Tab 2: 审计历史 -->
        <TabPane key="history" tab="审计历史">
          <Card size="small" class="mb-4">
            <Space>
              <Select
                v-model:value="actionStatusFilter"
                placeholder="按状态过滤"
                allow-clear
                style="width: 140px"
                @change="loadHistory"
              >
                <Select.Option value="pending">待处理</Select.Option>
                <Select.Option value="waiting_human">等待人工</Select.Option>
                <Select.Option value="auto_executed">自动执行</Select.Option>
                <Select.Option value="approved">已批准</Select.Option>
                <Select.Option value="rejected">已驳回</Select.Option>
                <Select.Option value="failed">执行失败</Select.Option>
              </Select>
              <Select
                v-model:value="decisionTypeFilter"
                placeholder="按决策类型过滤"
                allow-clear
                style="width: 140px"
                @change="loadHistory"
              >
                <Select.Option value="task_recovery">任务恢复</Select.Option>
                <Select.Option value="rework_flow">返工流程</Select.Option>
                <Select.Option value="accept_flow">验收流程</Select.Option>
                <Select.Option value="circuit_break">熔断</Select.Option>
                <Select.Option value="replan">重规划</Select.Option>
              </Select>
              <Button @click="loadHistory">刷新</Button>
            </Space>
          </Card>

          <Spin :spinning="historyLoading">
            <Empty v-if="allActions.length === 0" description="暂无决策记录" />
            <Table
              v-else
              :columns="historyColumns"
              :data-source="allActions"
              :pagination="{ pageSize: 20 }"
              row-key="id"
              size="small"
            >
              <template #bodyCell="{ column, record }">
                <template v-if="column.key === 'decisionLevel'">
                  <Tag :color="levelMap[record.decisionLevel]?.color || 'default'">
                    {{ levelMap[record.decisionLevel]?.label || record.decisionLevel }}
                  </Tag>
                </template>
                <template v-if="column.key === 'decisionType'">
                  {{ decisionTypeMap[record.decisionType] || record.decisionType }}
                </template>
                <template v-if="column.key === 'actionType'">
                  <Tag color="purple">
                    {{ actionTypeMap[record.actionType] || record.actionType }}
                  </Tag>
                </template>
                <template v-if="column.key === 'triggerSource'">
                  {{ triggerSourceMap[record.triggerSource] || record.triggerSource }}
                </template>
                <template v-if="column.key === 'actionStatus'">
                  <Tag :color="actionStatusMap[record.actionStatus]?.color || 'default'">
                    {{ actionStatusMap[record.actionStatus]?.label || record.actionStatus }}
                  </Tag>
                </template>
                <template v-if="column.key === 'recommendation'">
                  <Typography.Text :ellipsis="{ tooltip: true }" style="max-width: 250px">
                    {{ parseRecommendation(record.recommendation) }}
                  </Typography.Text>
                </template>
                <template v-if="column.key === 'result'">
                  <Typography.Text :ellipsis="{ tooltip: true }" style="max-width: 150px">
                    {{ parseResult(record.result) }}
                  </Typography.Text>
                </template>
              </template>
            </Table>
          </Spin>
        </TabPane>

        <!-- Tab 3: 风险闸门 -->
        <TabPane key="gates" tab="风险闸门">
          <!-- 闸门规则 -->
          <Card size="small" class="mb-4" title="闸门规则">
            <Empty v-if="gateRules.length === 0" description="暂无闸门规则" />
            <Table
              v-else
              :columns="gateRuleColumns"
              :data-source="gateRules"
              :pagination="false"
              row-key="id"
              size="small"
            >
              <template #bodyCell="{ column, record }">
                <template v-if="column.key === 'gateType'">
                  <Tag :color="gateTypeMap[record.gateType]?.color || 'default'">
                    {{ gateTypeMap[record.gateType]?.label || record.gateType }}
                  </Tag>
                </template>
                <template v-if="column.key === 'blockAction'">
                  <Tag color="red">{{ actionTypeMap[record.blockAction] || record.blockAction }}</Tag>
                </template>
                <template v-if="column.key === 'fallbackAction'">
                  <Tag v-if="record.fallbackAction" color="blue">
                    {{ actionTypeMap[record.fallbackAction] || record.fallbackAction }}
                  </Tag>
                  <span v-else class="text-gray-400">-</span>
                </template>
                <template v-if="column.key === 'enabled'">
                  <Tag :color="record.enabled ? 'green' : 'default'">
                    {{ record.enabled ? '启用' : '禁用' }}
                  </Tag>
                </template>
              </template>
            </Table>
          </Card>

          <!-- 策略规则 -->
          <Card size="small" title="策略规则">
            <Empty v-if="policyRules.length === 0" description="暂无策略规则" />
            <Table
              v-else
              :columns="policyRuleColumns"
              :data-source="policyRules"
              :pagination="false"
              row-key="id"
              size="small"
            >
              <template #bodyCell="{ column, record }">
                <template v-if="column.key === 'decisionType'">
                  {{ decisionTypeMap[record.decisionType] || record.decisionType }}
                </template>
                <template v-if="column.key === 'decisionLevel'">
                  <Tag :color="levelMap[record.decisionLevel]?.color || 'default'">
                    {{ levelMap[record.decisionLevel]?.label || record.decisionLevel }}
                  </Tag>
                </template>
                <template v-if="column.key === 'triggerSource'">
                  {{ triggerSourceMap[record.triggerSource] || record.triggerSource }}
                </template>
                <template v-if="column.key === 'enabled'">
                  <Tag :color="record.enabled ? 'green' : 'default'">
                    {{ record.enabled ? '启用' : '禁用' }}
                  </Tag>
                </template>
              </template>
            </Table>
          </Card>
        </TabPane>

        <!-- Tab 4: 报告中心 -->
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

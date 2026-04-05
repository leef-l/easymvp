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
  Typography,
  message,
} from 'ant-design-vue';
import {
  getAutonomyDecisions,
  approveDecision,
  rejectDecision,
  triggerReplan,
  getProjectReports,
  triggerReport,
  getAutonomyMode,
  setAutonomyMode,
  type AutonomyDecisionItem,
  type ProjectReportItem,
} from '#/api/mvp/workflow';

const route = useRoute();
const projectID = computed(() => (route.query.projectId as string) || '');

const loading = ref(false);
const activeTab = ref('decisions');
const decisions = ref<AutonomyDecisionItem[]>([]);
const reports = ref<ProjectReportItem[]>([]);
const decisionTypeFilter = ref<string | undefined>(undefined);
const reportTypeFilter = ref<string | undefined>(undefined);
const selectedReport = ref<ProjectReportItem | null>(null);
const reportDetailVisible = ref(false);
const autonomyMode = ref<string>('suggest');
const modeLoading = ref(false);

/** 加载数据 */
async function loadData() {
  if (!projectID.value) return;
  loading.value = true;
  try {
    const [decisionsRes, reportsRes] = await Promise.all([
      getAutonomyDecisions(projectID.value, decisionTypeFilter.value),
      getProjectReports(projectID.value, reportTypeFilter.value),
    ]);
    decisions.value = decisionsRes?.decisions ?? [];
    reports.value = reportsRes?.reports ?? [];
  } catch {
    /* ignore */
  } finally {
    loading.value = false;
  }
}

/** 待审批数 */
const pendingCount = computed(() =>
  decisions.value.filter((d) => d.humanAction === 'pending').length,
);

/** 决策类型标签 */
const decisionTypeMap: Record<string, { label: string; color: string }> = {
  replan: { label: '重规划', color: 'purple' },
  risk_escalate: { label: '风险升级', color: 'orange' },
  engine_switch: { label: '引擎切换', color: 'blue' },
  circuit_break: { label: '熔断', color: 'red' },
  report: { label: '报告', color: 'green' },
};

/** 报告类型标签 */
const reportTypeMap: Record<string, string> = {
  stage: '阶段报告',
  daily: '日报',
  weekly: '周报',
  summary: '总结',
};

// 决策列表列
const decisionColumns = [
  { title: '类型', dataIndex: 'decisionType', key: 'decisionType', width: 100 },
  { title: '触发源', dataIndex: 'triggerSource', key: 'triggerSource', width: 100 },
  { title: '建议', dataIndex: 'recommendation', key: 'recommendation' },
  { title: '模式', dataIndex: 'decisionMode', key: 'decisionMode', width: 80 },
  { title: '状态', dataIndex: 'humanAction', key: 'humanAction', width: 100 },
  { title: '时间', dataIndex: 'createdAt', key: 'createdAt', width: 160 },
  { title: '操作', key: 'action', width: 160 },
];

// 报告列表列
const reportColumns = [
  { title: '类型', dataIndex: 'reportType', key: 'reportType', width: 100 },
  { title: '阶段', dataIndex: 'stageType', key: 'stageType', width: 80 },
  { title: '标题', dataIndex: 'title', key: 'title' },
  { title: '时间', dataIndex: 'createdAt', key: 'createdAt', width: 160 },
  { title: '操作', key: 'action', width: 80 },
];

/** 解析建议 JSON */
function parseRecommendation(json: string): string {
  try {
    const obj = JSON.parse(json);
    if (obj.action) return `${obj.action}: ${obj.reasoning || obj.reason || ''}`;
    return json;
  } catch {
    return json;
  }
}

/** 批准决策 */
async function handleApprove(record: AutonomyDecisionItem) {
  Modal.confirm({
    title: '确认批准该决策？',
    content: parseRecommendation(record.recommendation),
    onOk: async () => {
      await approveDecision(projectID.value, record.id);
      message.success('已批准');
      loadData();
    },
  });
}

/** 拒绝决策 */
async function handleReject(record: AutonomyDecisionItem) {
  Modal.confirm({
    title: '确认拒绝该决策？',
    onOk: async () => {
      await rejectDecision(projectID.value, record.id);
      message.success('已拒绝');
      loadData();
    },
  });
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
              <Select
                v-model:value="decisionTypeFilter"
                placeholder="按类型过滤"
                allow-clear
                style="width: 140px"
                @change="loadData"
              >
                <Select.Option value="replan">重规划</Select.Option>
                <Select.Option value="circuit_break">熔断</Select.Option>
                <Select.Option value="risk_escalate">风险升级</Select.Option>
              </Select>
              <Button type="primary" @click="handleTriggerReplan">手动触发重规划</Button>
            </Space>
          </Card>

          <Table
            :columns="decisionColumns"
            :data-source="decisions"
            :pagination="false"
            row-key="id"
            size="small"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'decisionType'">
                <Tag :color="decisionTypeMap[record.decisionType]?.color || 'default'">
                  {{ decisionTypeMap[record.decisionType]?.label || record.decisionType }}
                </Tag>
              </template>
              <template v-if="column.key === 'recommendation'">
                <Typography.Text :ellipsis="{ tooltip: true }" style="max-width: 300px">
                  {{ parseRecommendation(record.recommendation) }}
                </Typography.Text>
              </template>
              <template v-if="column.key === 'decisionMode'">
                <Tag :color="record.decisionMode === 'auto' ? 'blue' : 'orange'">
                  {{ record.decisionMode === 'auto' ? '自动' : '建议' }}
                </Tag>
              </template>
              <template v-if="column.key === 'humanAction'">
                <Tag
                  :color="
                    record.humanAction === 'approved'
                      ? 'green'
                      : record.humanAction === 'rejected'
                        ? 'red'
                        : 'orange'
                  "
                >
                  {{
                    record.humanAction === 'approved'
                      ? '已批准'
                      : record.humanAction === 'rejected'
                        ? '已拒绝'
                        : '待审批'
                  }}
                </Tag>
              </template>
              <template v-if="column.key === 'action'">
                <Space v-if="record.humanAction === 'pending'">
                  <Button size="small" type="primary" @click="handleApprove(record)">批准</Button>
                  <Button size="small" danger @click="handleReject(record)">拒绝</Button>
                </Space>
                <span v-else class="text-gray-400">-</span>
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
  </div>
</template>

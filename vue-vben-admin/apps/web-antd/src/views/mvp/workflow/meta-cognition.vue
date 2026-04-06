<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue';
import {
  Card, Row, Col, Statistic, Table, Tag, Button, Space, Tabs, Progress,
  Descriptions, Popconfirm, message, Empty, Spin, Alert, Select,
} from 'ant-design-vue';
import { useRoute, useRouter } from 'vue-router';
import { getProjectList } from '#/api/mvp/project';
import type { ProjectItem } from '#/api/mvp/project/types';
import {
  getMetaObservationStats,
  getMetaObservations,
  getMetaAssessment,
  getMetaAssessmentHistory,
  runMetaAssessment,
  getMetaRecommendations,
  applyMetaRecommendation,
  rejectMetaRecommendation,
  getMetaLearning,
  type ObservationStats,
  type ObservationRecord,
  type AssessmentData,
  type TuneRecommendationItem,
  type LearningRecordItem,
} from '#/api/mvp/workflow';

const route = useRoute();
const router = useRouter();

// URL 参数优先；没有则由下拉选择
const projectID = computed(() => (route.query.projectID as string) || selectedProjectID.value);

// 下拉框选择的 projectID（仅在无 URL 参数时使用）
const selectedProjectID = ref('');
const projectOptions = ref<{ label: string; value: string }[]>([]);
const projectsLoading = ref(false);

async function loadProjects() {
  projectsLoading.value = true;
  try {
    const res = await getProjectList({ pageNum: 1, pageSize: 50 } as any);
    projectOptions.value = (res?.list ?? []).map((p: ProjectItem) => ({
      label: `${p.name}（${p.status ?? ''}）`,
      value: String(p.id),
    }));
  } finally {
    projectsLoading.value = false;
  }
}

function onSelectProject(val: string) {
  selectedProjectID.value = val;
  // 同步到 URL，方便分享和刷新
  router.replace({ query: { ...route.query, projectID: val } });
}

const loading = ref(false);
const activeTab = ref('overview');

// ==================== 数据 ====================
const stats = ref<ObservationStats | null>(null);
const observations = ref<ObservationRecord[]>([]);
const assessment = ref<AssessmentData | null>(null);
const assessmentHistory = ref<AssessmentData[]>([]);
const recommendations = ref<TuneRecommendationItem[]>([]);
const learningRecords = ref<LearningRecordItem[]>([]);

// ==================== 加载 ====================
async function loadAll() {
  if (!projectID.value) return;
  loading.value = true;
  try {
    const [statsRes, obsRes, assessRes, histRes, recRes, learnRes] = await Promise.allSettled([
      getMetaObservationStats(projectID.value),
      getMetaObservations(projectID.value, 50),
      getMetaAssessment(projectID.value),
      getMetaAssessmentHistory(projectID.value, 10),
      getMetaRecommendations(projectID.value),
      getMetaLearning(projectID.value),
    ]);
    if (statsRes.status === 'fulfilled') stats.value = statsRes.value.stats;
    if (obsRes.status === 'fulfilled') observations.value = obsRes.value.observations || [];
    if (assessRes.status === 'fulfilled') assessment.value = assessRes.value.assessment;
    if (histRes.status === 'fulfilled') assessmentHistory.value = histRes.value.assessments || [];
    if (recRes.status === 'fulfilled') recommendations.value = recRes.value.recommendations || [];
    if (learnRes.status === 'fulfilled') learningRecords.value = learnRes.value.records || [];
  } finally {
    loading.value = false;
  }
}

onMounted(async () => {
  // 没有 URL 参数时才加载项目列表供选择
  if (!route.query.projectID) {
    await loadProjects();
  }
  await loadAll();
});

// 选中项目后自动加载数据
watch(selectedProjectID, (val) => {
  if (val) loadAll();
});

// ==================== 操作 ====================
const assessmentRunning = ref(false);
async function handleRunAssessment() {
  assessmentRunning.value = true;
  try {
    const res = await runMetaAssessment(projectID.value, 7);
    assessment.value = res.assessment;
    message.success('评估完成');
    loadAll();
  } catch (e: any) {
    message.error(e.message || '评估失败');
  } finally {
    assessmentRunning.value = false;
  }
}

async function handleApply(id: string) {
  try {
    await applyMetaRecommendation(id);
    message.success('已应用');
    loadAll();
  } catch (e: any) {
    message.error(e.message || '应用失败');
  }
}

async function handleReject(id: string) {
  try {
    await rejectMetaRecommendation(id);
    message.success('已驳回');
    loadAll();
  } catch (e: any) {
    message.error(e.message || '驳回失败');
  }
}

// ==================== 表格列 ====================
const observationColumns = [
  { title: '时间', dataIndex: 'createdAt', width: 160, ellipsis: true },
  { title: '决策类型', dataIndex: 'decisionType', width: 140, ellipsis: true },
  { title: '触发源', dataIndex: 'triggerSource', width: 120 },
  { title: '级别', dataIndex: 'decisionLevel', width: 60 },
  { title: '动作', dataIndex: 'actionType', width: 120 },
  { title: '结果', dataIndex: 'outcome', width: 80 },
  { title: '效果', dataIndex: 'effectScore', width: 80 },
  { title: '人工', dataIndex: 'humanOverride', width: 60 },
  { title: '权重', dataIndex: 'signalWeight', width: 60 },
];

const recommendationColumns = [
  { title: '参数', dataIndex: 'parameter', width: 200, ellipsis: true },
  { title: '当前值', dataIndex: 'currentValue', width: 100 },
  { title: '建议值', dataIndex: 'suggestedValue', width: 100 },
  { title: '方向', dataIndex: 'direction', width: 80 },
  { title: '置信度', dataIndex: 'confidence', width: 80 },
  { title: '风险', dataIndex: 'riskLevel', width: 60 },
  { title: '状态', dataIndex: 'status', width: 80 },
  { title: '操作', key: 'action', width: 140 },
];

const learningColumns = [
  { title: '指标', dataIndex: 'metricKey', width: 250, ellipsis: true },
  { title: 'EMA值', dataIndex: 'emaValue', width: 100 },
  { title: '最新值', dataIndex: 'rawValue', width: 100 },
  { title: '样本数', dataIndex: 'sampleCount', width: 80 },
  { title: '最后更新', dataIndex: 'lastUpdated', width: 160, ellipsis: true },
];

// ==================== 辅助 ====================
function outcomeColor(outcome: string) {
  const map: Record<string, string> = { success: 'green', failure: 'red', neutral: 'blue', pending: 'default' };
  return map[outcome] || 'default';
}

function levelColor(level: string) {
  const map: Record<string, string> = { A: 'green', B: 'orange', C: 'red' };
  return map[level] || 'default';
}

function directionTag(dir: string) {
  return dir === 'conservative' ? '保守' : '激进';
}

function directionColor(dir: string) {
  return dir === 'conservative' ? 'green' : 'orange';
}

function statusColor(status: string) {
  const map: Record<string, string> = { pending: 'blue', applied: 'green', rejected: 'red', expired: 'default' };
  return map[status] || 'default';
}

function pct(v: number | undefined) {
  return v !== undefined ? `${(v * 100).toFixed(1)}%` : '-';
}
</script>

<template>
  <div style="padding: 16px">
    <!-- 无 URL 参数时展示项目选择器 -->
    <div v-if="!route.query.projectID" style="margin-bottom: 16px">
      <Card size="small">
        <Space>
          <span style="font-size: 13px; color: #595959">选择项目：</span>
          <Select
            v-model:value="selectedProjectID"
            :loading="projectsLoading"
            :options="projectOptions"
            placeholder="请选择要查看元认知数据的项目"
            style="width: 360px"
            show-search
            option-filter-prop="label"
            @change="onSelectProject"
          />
          <Button v-if="!selectedProjectID" size="small" @click="loadProjects">刷新列表</Button>
        </Space>
      </Card>
      <Alert
        v-if="!selectedProjectID"
        type="info"
        message="请先选择项目，或通过项目列表点击"元认知"按钮直接跳转"
        show-icon
        style="margin-top: 8px"
      />
    </div>

    <Spin :spinning="loading">
      <Tabs v-model:activeKey="activeTab">
        <!-- ==================== 总览 ==================== -->
        <Tabs.TabPane key="overview" tab="总览">
          <Row :gutter="16" style="margin-bottom: 16px">
            <Col :span="6">
              <Card>
                <Statistic title="观测总数" :value="stats?.total || 0" />
              </Card>
            </Col>
            <Col :span="6">
              <Card>
                <Statistic
                  title="人工干预率"
                  :value="stats ? (stats.humanOverrideRate * 100).toFixed(1) : '0'"
                  suffix="%"
                />
              </Card>
            </Col>
            <Col :span="6">
              <Card>
                <Statistic
                  title="策略准确率"
                  :value="assessment ? (assessment.policyAccuracy * 100).toFixed(1) : '-'"
                  suffix="%"
                />
              </Card>
            </Col>
            <Col :span="6">
              <Card>
                <Statistic
                  title="成本效率"
                  :value="assessment ? (assessment.costEfficiency * 100).toFixed(1) : '-'"
                  suffix="%"
                />
              </Card>
            </Col>
          </Row>

          <!-- 最新评估 -->
          <Card
            title="最新评估"
            :extra="undefined"
            style="margin-bottom: 16px"
          >
            <template #extra>
              <Button
                type="primary"
                size="small"
                :loading="assessmentRunning"
                @click="handleRunAssessment"
              >
                手动评估
              </Button>
            </template>
            <template v-if="assessment && assessment.sampleCount > 0">
              <Descriptions :column="3" bordered size="small">
                <Descriptions.Item label="评估周期">
                  {{ assessment.periodStart }} ~ {{ assessment.periodEnd }}
                </Descriptions.Item>
                <Descriptions.Item label="样本数">{{ assessment.sampleCount }}</Descriptions.Item>
                <Descriptions.Item label="策略准确率">
                  <Progress :percent="Number((assessment.policyAccuracy * 100).toFixed(1))" size="small" />
                </Descriptions.Item>
                <Descriptions.Item label="闸门误报率">{{ pct(assessment.gateFalsePositive) }}</Descriptions.Item>
                <Descriptions.Item label="闸门漏报率">{{ pct(assessment.gateFalseNegative) }}</Descriptions.Item>
                <Descriptions.Item label="人工干预率">{{ pct(assessment.humanOverrideRate) }}</Descriptions.Item>
                <Descriptions.Item label="摘要" :span="3">{{ assessment.summary }}</Descriptions.Item>
              </Descriptions>
              <div v-if="assessment.drifts && assessment.drifts.length" style="margin-top: 12px">
                <strong>参数偏差：</strong>
                <Tag v-for="d in assessment.drifts" :key="d.parameter" color="orange" style="margin: 4px">
                  {{ d.parameter }}：{{ d.currentValue.toFixed(3) }} → {{ d.optimalValue.toFixed(3) }}
                  (置信度 {{ (d.confidence * 100).toFixed(0) }}%)
                </Tag>
              </div>
            </template>
            <Empty v-else description="暂无评估数据，点击「手动评估」开始" />
          </Card>

          <!-- 待处理建议 -->
          <Card title="调参建议" style="margin-bottom: 16px">
            <Table
              :columns="recommendationColumns"
              :data-source="recommendations"
              :pagination="false"
              row-key="id"
              size="small"
            >
              <template #bodyCell="{ column, record }">
                <template v-if="column.dataIndex === 'direction'">
                  <Tag :color="directionColor(record.direction)">{{ directionTag(record.direction) }}</Tag>
                </template>
                <template v-else-if="column.dataIndex === 'confidence'">
                  {{ (record.confidence * 100).toFixed(0) }}%
                </template>
                <template v-else-if="column.dataIndex === 'riskLevel'">
                  <Tag :color="record.riskLevel === 'low' ? 'green' : record.riskLevel === 'medium' ? 'orange' : 'red'">
                    {{ record.riskLevel }}
                  </Tag>
                </template>
                <template v-else-if="column.dataIndex === 'status'">
                  <Tag :color="statusColor(record.status)">{{ record.status }}</Tag>
                </template>
                <template v-else-if="column.key === 'action'">
                  <Space v-if="record.status === 'pending'">
                    <Popconfirm title="确认应用此建议？" @confirm="handleApply(record.id)">
                      <Button type="primary" size="small">应用</Button>
                    </Popconfirm>
                    <Popconfirm title="确认驳回此建议？" @confirm="handleReject(record.id)">
                      <Button size="small" danger>驳回</Button>
                    </Popconfirm>
                  </Space>
                  <span v-else>-</span>
                </template>
              </template>
            </Table>
          </Card>
        </Tabs.TabPane>

        <!-- ==================== 观测记录 ==================== -->
        <Tabs.TabPane key="observations" tab="观测记录">
          <Table
            :columns="observationColumns"
            :data-source="observations"
            :pagination="{ pageSize: 20 }"
            row-key="id"
            size="small"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.dataIndex === 'decisionLevel'">
                <Tag :color="levelColor(record.decisionLevel)">{{ record.decisionLevel }}</Tag>
              </template>
              <template v-else-if="column.dataIndex === 'outcome'">
                <Tag :color="outcomeColor(record.outcome)">{{ record.outcome }}</Tag>
              </template>
              <template v-else-if="column.dataIndex === 'humanOverride'">
                <Tag v-if="record.humanOverride" color="red">是</Tag>
                <span v-else>-</span>
              </template>
              <template v-else-if="column.dataIndex === 'effectScore'">
                {{ record.effectScore?.toFixed(3) || '-' }}
              </template>
            </template>
          </Table>
        </Tabs.TabPane>

        <!-- ==================== 学习记录 ==================== -->
        <Tabs.TabPane key="learning" tab="学习记录">
          <Table
            :columns="learningColumns"
            :data-source="learningRecords"
            :pagination="false"
            row-key="id"
            size="small"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.dataIndex === 'emaValue'">
                <Progress
                  :percent="Number((record.emaValue * 100).toFixed(1))"
                  size="small"
                  :status="record.emaValue < 0.4 ? 'exception' : record.emaValue > 0.7 ? 'success' : 'active'"
                  style="width: 80px"
                />
              </template>
              <template v-else-if="column.dataIndex === 'rawValue'">
                {{ record.rawValue?.toFixed(4) }}
              </template>
            </template>
          </Table>
        </Tabs.TabPane>

        <!-- ==================== 评估历史 ==================== -->
        <Tabs.TabPane key="history" tab="评估历史">
          <Card
            v-for="item in assessmentHistory"
            :key="item.id"
            size="small"
            style="margin-bottom: 8px"
          >
            <Descriptions :column="4" size="small">
              <Descriptions.Item label="周期">{{ item.periodStart }} ~ {{ item.periodEnd }}</Descriptions.Item>
              <Descriptions.Item label="样本">{{ item.sampleCount }}</Descriptions.Item>
              <Descriptions.Item label="准确率">{{ pct(item.policyAccuracy) }}</Descriptions.Item>
              <Descriptions.Item label="干预率">{{ pct(item.humanOverrideRate) }}</Descriptions.Item>
            </Descriptions>
            <div style="margin-top: 4px; color: #888; font-size: 12px">{{ item.summary }}</div>
          </Card>
          <Empty v-if="!assessmentHistory.length" description="暂无评估历史" />
        </Tabs.TabPane>
      </Tabs>
    </Spin>
  </div>
</template>

<script setup lang="ts">
import { computed, onUnmounted, ref, watch } from 'vue';
import { useRoute } from 'vue-router';

import {
  Badge,
  Button,
  Card,
  Descriptions,
  DescriptionsItem,
  Empty,
  Input,
  message,
  Modal,
  Progress,
  Select,
  Space,
  Spin,
  Table,
  Tag,
} from 'ant-design-vue';

import {
  getVerificationEvidence,
  getVerificationIssues,
  getVerificationStatus,
  startVerification,
  type VerificationEvidenceItem,
  type VerificationIssueItem,
  verificationRepair,
  type VerificationStatusResult,
} from '#/api/mvp/workflow';

const props = defineProps<{ projectId?: string }>();

const route = useRoute();
const projectID = computed(() => props.projectId || (route.query.projectId as string) || '');

const loading = ref(false);
const status = ref<null | VerificationStatusResult>(null);
const issues = ref<VerificationIssueItem[]>([]);
const evidence = ref<VerificationEvidenceItem[]>([]);
const severityFilter = ref<string | undefined>(undefined);
const selectedIssueIDs = ref<string[]>([]);
const startReason = ref('');
const repairReason = ref('');
const startModalVisible = ref(false);
const repairModalVisible = ref(false);
let pollTimer: null | ReturnType<typeof setInterval> = null;

const needsPolling = computed(
  () => status.value?.status === 'running',
);

const openIssues = computed(() =>
  issues.value.filter((item) => item.status === 'open'),
);

const canRepair = computed(
  () => selectedIssueIDs.value.length > 0
    && selectedIssueIDs.value.every((id) =>
      openIssues.value.some((item) => item.id === id && !!item.domainTaskID),
    ),
);

function handleSelectionChange(keys: Array<number | string>) {
  selectedIssueIDs.value = keys.map(String);
}

async function loadData() {
  if (!projectID.value) return;
  loading.value = true;
  try {
    const [statusRes, issuesRes, evidenceRes] = await Promise.all([
      getVerificationStatus(projectID.value),
      getVerificationIssues(projectID.value, severityFilter.value),
      getVerificationEvidence(projectID.value),
    ]);
    status.value = statusRes?.status === 'none' ? null : statusRes;
    issues.value = issuesRes?.issues ?? [];
    evidence.value = evidenceRes?.evidence ?? [];

    const validSelection = new Set(openIssues.value.map((item) => item.id));
    selectedIssueIDs.value = selectedIssueIDs.value.filter((id) => validSelection.has(id));
  } finally {
    loading.value = false;
  }
}

function startPolling() {
  stopPolling();
  if (!needsPolling.value) return;
  pollTimer = setInterval(() => {
    loadData();
  }, 5000);
}

function stopPolling() {
  if (pollTimer) {
    clearInterval(pollTimer);
    pollTimer = null;
  }
}

watch(
  () => projectID.value,
  async () => {
    selectedIssueIDs.value = [];
    await loadData();
  },
  { immediate: true },
);

watch(
  () => severityFilter.value,
  async () => {
    await loadData();
  },
);

watch(needsPolling, (value) => {
  if (value) {
    startPolling();
  } else {
    stopPolling();
  }
}, { immediate: true });

onUnmounted(() => {
  stopPolling();
});

function statusColor(value?: string) {
  switch (value) {
    case 'completed': {
      return 'green';
    }
    case 'failed': {
      return 'red';
    }
    case 'running': {
      return 'processing';
    }
    default: {
      return 'default';
    }
  }
}

function decisionColor(value?: string) {
  switch (value) {
    case 'failed': {
      return 'red';
    }
    case 'manual_review': {
      return 'orange';
    }
    case 'passed': {
      return 'green';
    }
    default: {
      return 'default';
    }
  }
}

function severityColor(value: string) {
  switch (value) {
    case 'blocker': {
      return 'red';
    }
    case 'error': {
      return 'volcano';
    }
    case 'info': {
      return 'blue';
    }
    case 'warn': {
      return 'gold';
    }
    default: {
      return 'default';
    }
  }
}

async function handleStartVerification() {
  try {
    const res = await startVerification(projectID.value, startReason.value || '控制台手动触发验证');
    message.success(res.message || '验证已启动');
    startModalVisible.value = false;
    startReason.value = '';
    await loadData();
  } catch (error: any) {
    message.error(error?.message || '启动验证失败');
  }
}

async function handleRepairSelected() {
  if (!canRepair.value) {
    message.warning('请选择至少一条带任务关联的开放问题');
    return;
  }
  try {
    const res = await verificationRepair(
      projectID.value,
      selectedIssueIDs.value,
      repairReason.value || '控制台手动触发验证返工',
    );
    message.success(res.message || '已触发返工');
    repairModalVisible.value = false;
    repairReason.value = '';
    selectedIssueIDs.value = [];
    await loadData();
  } catch (error: any) {
    message.error(error?.message || '触发返工失败');
  }
}

function handleIssueRepair(issue: Record<string, any> | VerificationIssueItem) {
  const currentIssue = issue as VerificationIssueItem;
  if (!currentIssue.domainTaskID) {
    message.warning('该问题未关联任务，暂时无法直接返工');
    return;
  }
  Modal.confirm({
    title: '将验证问题转为返工',
    content: `问题「${currentIssue.title}」将作为返工原因回流到执行链路。`,
    onOk: async () => {
      const res = await verificationRepair(projectID.value, [currentIssue.id]);
      message.success(res.message || '已触发返工');
      await loadData();
    },
  });
}

const issueColumns = [
  { title: '级别', dataIndex: 'severity', key: 'severity', width: 90 },
  { title: '类型', dataIndex: 'issueType', key: 'issueType', width: 120 },
  { title: '标题', dataIndex: 'title', key: 'title', width: 220 },
  { title: '详情', dataIndex: 'detail', key: 'detail' },
  { title: '资源', dataIndex: 'resourceRef', key: 'resourceRef', width: 150 },
  { title: '状态', dataIndex: 'status', key: 'status', width: 110 },
  { title: '操作', key: 'action', width: 110, fixed: 'right' as const },
];

const evidenceColumns = [
  { title: '类型', dataIndex: 'evidenceType', key: 'evidenceType', width: 120 },
  { title: '来源', dataIndex: 'sourceType', key: 'sourceType', width: 120 },
  { title: '摘要', dataIndex: 'summary', key: 'summary' },
  { title: '时间', dataIndex: 'createdAt', key: 'createdAt', width: 170 },
];
</script>

<template>
  <Spin :spinning="loading">
    <Card class="mb-4">
      <template #title>验证控制台</template>
      <template #extra>
        <Space>
          <Button @click="loadData">刷新</Button>
          <Button type="primary" @click="startModalVisible = true">
            {{ status ? '重新验证' : '启动验证' }}
          </Button>
        </Space>
      </template>

      <Empty
        v-if="!projectID"
        description="缺少 projectId"
      />
      <Empty
        v-else-if="!status"
        description="暂无验证记录，可手动启动 Docker-first 验证"
      />
      <template v-else>
        <Descriptions :column="{ xs: 1, sm: 2, md: 3 }" bordered size="small">
          <DescriptionsItem label="运行状态">
            <Tag :color="statusColor(status.status)">
              {{ status.status }}
            </Tag>
          </DescriptionsItem>
          <DescriptionsItem label="验证结论">
            <Tag :color="decisionColor(status.decision)">
              {{ status.decision || '-' }}
            </Tag>
          </DescriptionsItem>
          <DescriptionsItem label="执行器">
            {{ status.runnerType || '-' }}
          </DescriptionsItem>
          <DescriptionsItem label="触发来源">
            {{ status.triggerSource || '-' }}
          </DescriptionsItem>
          <DescriptionsItem label="验证轮次">
            {{ status.verificationRound || '-' }}
          </DescriptionsItem>
          <DescriptionsItem label="证据数">
            {{ status.evidenceCount }}
          </DescriptionsItem>
          <DescriptionsItem label="开始时间">
            {{ status.startedAt || '-' }}
          </DescriptionsItem>
          <DescriptionsItem label="结束时间">
            {{ status.finishedAt || '-' }}
          </DescriptionsItem>
          <DescriptionsItem label="摘要" :span="3">
            {{ status.summary || '-' }}
          </DescriptionsItem>
        </Descriptions>

        <div class="mt-4">
          <Progress
            :percent="status.status === 'running' ? 65 : 100"
            :status="status.status === 'failed' ? 'exception' : status.status === 'running' ? 'active' : 'success'"
          />
        </div>

        <div class="mt-3 flex flex-wrap gap-4">
          <Badge status="error" :text="`blocker ${status.blockerCount}`" />
          <Badge status="error" :text="`error ${status.errorCount}`" />
          <Badge status="warning" :text="`warn ${status.warnCount}`" />
          <Badge status="processing" :text="`info ${status.infoCount}`" />
        </div>
      </template>
    </Card>

    <Card class="mb-4">
      <template #title>验证问题</template>
      <template #extra>
        <Space>
          <Select
            v-model:value="severityFilter"
            allow-clear
            placeholder="级别筛选"
            style="width: 140px"
            :options="[
              { label: 'Blocker', value: 'blocker' },
              { label: 'Error', value: 'error' },
              { label: 'Warn', value: 'warn' },
              { label: 'Info', value: 'info' },
            ]"
          />
          <Button
            :disabled="!canRepair"
            type="primary"
            @click="repairModalVisible = true"
          >
            返工所选问题
          </Button>
        </Space>
      </template>

      <Table
        row-key="id"
        :columns="issueColumns"
        :data-source="issues"
        :pagination="{ pageSize: 8 }"
        :row-selection="{
          selectedRowKeys: selectedIssueIDs,
          onChange: handleSelectionChange,
          getCheckboxProps: (record: VerificationIssueItem) => ({
            disabled: record.status !== 'open' || !record.domainTaskID,
          }),
        }"
        :scroll="{ x: 1200 }"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'severity'">
            <Tag :color="severityColor(record.severity)">
              {{ record.severity }}
            </Tag>
          </template>
          <template v-else-if="column.key === 'status'">
            <Tag :color="record.status === 'open' ? 'orange' : 'default'">
              {{ record.status }}
            </Tag>
          </template>
          <template v-else-if="column.key === 'action'">
            <Button
              size="small"
              type="link"
              :disabled="record.status !== 'open' || !record.domainTaskID"
              @click="handleIssueRepair(record)"
            >
              触发返工
            </Button>
          </template>
        </template>
      </Table>
    </Card>

    <Card>
      <template #title>验证证据</template>
      <Table
        row-key="id"
        :columns="evidenceColumns"
        :data-source="evidence"
        :pagination="{ pageSize: 8 }"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'evidenceType'">
            <Tag color="blue">{{ record.evidenceType }}</Tag>
          </template>
          <template v-else-if="column.key === 'sourceType'">
            <Tag color="purple">{{ record.sourceType }}</Tag>
          </template>
        </template>
      </Table>
    </Card>

    <Modal
      v-model:open="startModalVisible"
      title="启动 Docker-first 验证"
      ok-text="启动"
      cancel-text="取消"
      @ok="handleStartVerification"
    >
      <Input.TextArea
        v-model:value="startReason"
        :rows="4"
        placeholder="可填写本次验证原因，例如：准备发版前复检"
      />
    </Modal>

    <Modal
      v-model:open="repairModalVisible"
      title="将验证问题转为返工"
      ok-text="触发返工"
      cancel-text="取消"
      @ok="handleRepairSelected"
    >
      <p class="mb-3 text-gray-500">
        当前已选择 {{ selectedIssueIDs.length }} 条问题，只有已关联任务的开放问题可以直接回流返工。
      </p>
      <Input.TextArea
        v-model:value="repairReason"
        :rows="4"
        placeholder="可填写补充返工原因"
      />
    </Modal>
  </Spin>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
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
  Steps,
  Step,
  Badge,
  message,
} from 'ant-design-vue';
import {
  CheckCircleOutlined,
  CloseCircleOutlined,
  LoadingOutlined,
  ExclamationCircleOutlined,
} from '@ant-design/icons-vue';
import {
  getReviewStatus,
  getReviewIssues,
  manualApprove,
  manualReject,
  type ReviewStageTask,
  type ReviewIssueItem,
} from '#/api/mvp/workflow';

const route = useRoute();
const projectID = computed(() => (route.query.projectId as string) || '');

const loading = ref(false);
const reviewStatus = ref<any>(null);
const issues = ref<ReviewIssueItem[]>([]);
const rejectReason = ref('');
const rejectModalVisible = ref(false);

/** 加载审核数据 */
async function loadData() {
  if (!projectID.value) return;
  loading.value = true;
  try {
    const [statusRes, issuesRes] = await Promise.all([
      getReviewStatus(projectID.value),
      getReviewIssues(projectID.value),
    ]);
    reviewStatus.value = statusRes;
    issues.value = issuesRes?.issues ?? [];
  } catch {
    /* ignore */
  } finally {
    loading.value = false;
  }
}

/** 阶段任务状态图标 */
function taskStatusIcon(status: string) {
  switch (status) {
    case 'completed':
      return CheckCircleOutlined;
    case 'failed':
      return CloseCircleOutlined;
    case 'running':
      return LoadingOutlined;
    default:
      return ExclamationCircleOutlined;
  }
}

/** 阶段任务类型映射 */
const taskTypeMap: Record<string, string> = {
  precheck: '系统预检',
  auditor_review: '审计员审核',
  coordinator_optimize: '协调员优化',
  review_summary: '审核汇总',
};

/** 严重级别颜色 */
const severityColor: Record<string, string> = {
  error: 'red',
  warning: 'orange',
  info: 'blue',
};

/** 问题列表列 */
const issueColumns = [
  { title: '级别', dataIndex: 'severity', key: 'severity', width: 80 },
  { title: '来源', dataIndex: 'sourceRole', key: 'sourceRole', width: 100 },
  { title: '关联蓝图', dataIndex: 'taskName', key: 'taskName', width: 150 },
  { title: '问题描述', dataIndex: 'message', key: 'message' },
  { title: '状态', dataIndex: 'status', key: 'status', width: 80 },
];

/** 手动通过 */
async function handleApprove() {
  Modal.confirm({
    title: '确认手动通过审核？',
    content: '通过后项目将进入执行阶段',
    onOk: async () => {
      await manualApprove(projectID.value);
      message.success('审核已通过');
      loadData();
    },
  });
}

/** 手动驳回 */
async function handleReject() {
  if (!rejectReason.value.trim()) {
    message.warning('请输入驳回原因');
    return;
  }
  await manualReject(projectID.value, rejectReason.value);
  message.success('已驳回');
  rejectModalVisible.value = false;
  rejectReason.value = '';
  loadData();
}

onMounted(loadData);
</script>

<template>
  <div class="p-4">
    <Spin :spinning="loading">
      <Empty v-if="!reviewStatus" description="暂无审核数据" />
      <template v-else>
        <!-- 审核概览 -->
        <Card title="审核概览" class="mb-4">
          <Descriptions :column="3" bordered size="small">
            <DescriptionsItem label="方案版本">
              {{ reviewStatus.planVersionID || '-' }}
            </DescriptionsItem>
            <DescriptionsItem label="审核状态">
              <Tag
                :color="
                  reviewStatus.reviewStatus === 'approved'
                    ? 'green'
                    : reviewStatus.reviewStatus === 'rejected'
                      ? 'red'
                      : 'blue'
                "
              >
                {{
                  reviewStatus.reviewStatus === 'approved'
                    ? '已通过'
                    : reviewStatus.reviewStatus === 'rejected'
                      ? '已驳回'
                      : '审核中'
                }}
              </Tag>
            </DescriptionsItem>
            <DescriptionsItem label="蓝图数">
              {{ reviewStatus.blueprintCount }}
            </DescriptionsItem>
            <DescriptionsItem label="错误">
              <Badge
                :count="reviewStatus.errorCount"
                :number-style="{ backgroundColor: '#f5222d' }"
                show-zero
              />
            </DescriptionsItem>
            <DescriptionsItem label="警告">
              <Badge
                :count="reviewStatus.warningCount"
                :number-style="{ backgroundColor: '#faad14' }"
                show-zero
              />
            </DescriptionsItem>
            <DescriptionsItem label="阶段状态">
              <Tag
                :color="
                  reviewStatus.stageStatus === 'completed'
                    ? 'green'
                    : reviewStatus.stageStatus === 'failed'
                      ? 'red'
                      : 'processing'
                "
              >
                {{ reviewStatus.stageStatus || '-' }}
              </Tag>
            </DescriptionsItem>
          </Descriptions>

          <!-- 手动操作 -->
          <div
            v-if="reviewStatus.reviewStatus === 'pending'"
            class="mt-4 text-right"
          >
            <Space>
              <Button type="primary" @click="handleApprove">手动通过</Button>
              <Button danger @click="rejectModalVisible = true">
                手动驳回
              </Button>
            </Space>
          </div>
        </Card>

        <!-- 审核步骤 -->
        <Card
          v-if="reviewStatus.stageTasks?.length"
          title="审核步骤"
          class="mb-4"
        >
          <Steps direction="vertical" size="small">
            <Step
              v-for="task in reviewStatus.stageTasks"
              :key="task.id"
              :status="
                task.status === 'completed'
                  ? 'finish'
                  : task.status === 'failed'
                    ? 'error'
                    : task.status === 'running'
                      ? 'process'
                      : 'wait'
              "
            >
              <template #title>
                <Space>
                  <component :is="taskStatusIcon(task.status)" />
                  <span>{{ taskTypeMap[task.taskType] || task.taskType }}</span>
                  <Tag>{{ task.roleType }}</Tag>
                </Space>
              </template>
              <template #description>
                <span v-if="task.errorMessage" class="text-red-500">
                  {{ task.errorMessage }}
                </span>
              </template>
            </Step>
          </Steps>
        </Card>

        <!-- 审核问题列表 -->
        <Card title="审核问题">
          <Table
            :columns="issueColumns"
            :data-source="issues"
            :pagination="false"
            row-key="id"
            size="small"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'severity'">
                <Tag :color="severityColor[record.severity]">
                  {{ record.severity }}
                </Tag>
              </template>
              <template v-if="column.key === 'status'">
                <Tag
                  :color="record.status === 'open' ? 'red' : 'green'"
                >
                  {{ record.status === 'open' ? '待处理' : '已解决' }}
                </Tag>
              </template>
            </template>
          </Table>
        </Card>
      </template>
    </Spin>

    <!-- 驳回弹窗 -->
    <Modal
      v-model:open="rejectModalVisible"
      title="驳回审核"
      @ok="handleReject"
    >
      <Input.TextArea
        v-model:value="rejectReason"
        :rows="4"
        placeholder="请输入驳回原因"
      />
    </Modal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import { useVbenModal } from '@vben/common-ui';
import {
  Alert,
  Button,
  Descriptions,
  DescriptionsItem,
  Divider,
  message,
  Spin,
  Tag,
  Typography,
  TypographyParagraph,
  TypographyText,
} from 'ant-design-vue';
import { getTaskDetail } from '#/api/mvp/task';
import { retryTask } from '#/api/mvp/workflow';
import type { TaskItem } from '#/api/mvp/task/types';

// ===== 状态/颜色映射 =====
const STATUS_MAP: Record<string, { color: string; text: string }> = {
  pending: { color: 'default', text: '待执行' },
  running: { color: 'processing', text: '执行中' },
  completed: { color: 'success', text: '已完成' },
  failed: { color: 'error', text: '失败' },
  auditing: { color: 'warning', text: '审计中' },
  bug_found: { color: 'error', text: '发现Bug' },
  bug_dispatched: { color: 'purple', text: '修复中' },
  submit_error: { color: 'error', text: '提交错误' },
};

const ROLE_MAP: Record<string, { color: string; text: string }> = {
  architect: { color: 'purple', text: '架构师' },
  implementer: { color: 'blue', text: '实施员' },
  auditor: { color: 'orange', text: '审计员' },
};

const LEVEL_MAP: Record<string, { color: string; text: string }> = {
  lite: { color: 'default', text: 'Lite' },
  pro: { color: 'blue', text: 'Pro' },
  max: { color: 'purple', text: 'Max' },
};

// ===== 状态 =====
const detail = ref<TaskItem | null>(null);
const loading = ref(false);
const retrying = ref(false);

// 当前 projectId（从 setData 传入）
const currentProjectId = ref('');

/** 是否失败状态 */
const isFailed = computed(
  () => detail.value?.status === 'failed' || detail.value?.status === 'submit_error',
);

/** 是否已完成 */
const isCompleted = computed(() => detail.value?.status === 'completed');

// ===== Modal 配置 =====
const [Modal, modalApi] = useVbenModal({
  fullscreenButton: true,
  footer: false,
  async onOpenChange(isOpen: boolean) {
    if (isOpen) {
      const data = modalApi.getData<{ id: string; projectId?: string }>();
      if (data?.id) {
        currentProjectId.value = data.projectId || '';
        modalApi.setState({ title: '任务详情' });
        loading.value = true;
        try {
          detail.value = await getTaskDetail(data.id);
          // 更新标题
          if (detail.value?.name) {
            modalApi.setState({ title: `任务详情 · ${detail.value.name}` });
          }
        } catch {
          detail.value = null;
        } finally {
          loading.value = false;
        }
      }
    } else {
      detail.value = null;
      loading.value = false;
    }
  },
});

/** 重试 */
async function handleRetry() {
  if (!detail.value || !currentProjectId.value) return;
  retrying.value = true;
  try {
    await retryTask({ projectID: currentProjectId.value, taskID: detail.value.id });
    message.success('已提交重试请求');
    modalApi.close();
  } catch {
    message.error('重试失败');
  } finally {
    retrying.value = false;
  }
}
</script>

<template>
  <Modal class="w-[720px]">
    <Spin :spinning="loading">
      <div v-if="detail" class="task-detail">

        <!-- 失败提示 -->
        <Alert
          v-if="isFailed"
          type="error"
          class="mb-4"
          show-icon
          :message="'任务执行失败'"
          :description="detail.errorMessage || '未知错误'"
        >
          <template #action>
            <Button
              size="small"
              danger
              :loading="retrying"
              @click="handleRetry"
            >
              重试
            </Button>
          </template>
        </Alert>

        <!-- 基本信息 -->
        <Descriptions bordered :column="2" size="small">
          <DescriptionsItem label="任务名称" :span="2">
            <Typography>
              <TypographyText strong>{{ detail.name || '-' }}</TypographyText>
            </Typography>
          </DescriptionsItem>
          <DescriptionsItem label="状态">
            <Tag
              v-if="detail.status"
              :color="STATUS_MAP[detail.status]?.color || 'default'"
            >
              {{ STATUS_MAP[detail.status]?.text || detail.status }}
            </Tag>
            <span v-else>-</span>
          </DescriptionsItem>
          <DescriptionsItem label="批次号">
            {{ detail.batchNo ?? '-' }}
          </DescriptionsItem>
          <DescriptionsItem label="角色类型">
            <Tag
              v-if="detail.roleType"
              :color="ROLE_MAP[detail.roleType]?.color || 'default'"
            >
              {{ ROLE_MAP[detail.roleType]?.text || detail.roleType }}
            </Tag>
            <span v-else>-</span>
          </DescriptionsItem>
          <DescriptionsItem label="角色等级">
            <Tag
              v-if="detail.roleLevel"
              :color="LEVEL_MAP[detail.roleLevel]?.color || 'default'"
            >
              {{ LEVEL_MAP[detail.roleLevel]?.text || detail.roleLevel }}
            </Tag>
            <span v-else>-</span>
          </DescriptionsItem>
          <DescriptionsItem label="AI模型">
            {{ detail.modelID || '-' }}
          </DescriptionsItem>
          <DescriptionsItem label="排序">
            {{ detail.sort ?? '-' }}
          </DescriptionsItem>
          <DescriptionsItem label="开始时间">
            {{ detail.startedAt || '-' }}
          </DescriptionsItem>
          <DescriptionsItem label="完成时间">
            {{ detail.completedAt || '-' }}
          </DescriptionsItem>
          <DescriptionsItem label="创建时间">
            {{ detail.createdAt || '-' }}
          </DescriptionsItem>
          <DescriptionsItem label="更新时间">
            {{ detail.updatedAt || '-' }}
          </DescriptionsItem>
        </Descriptions>

        <!-- 任务描述 -->
        <Divider orientation="left" class="mt-4">任务描述</Divider>
        <div class="bg-gray-50 rounded p-3 text-sm text-gray-700 whitespace-pre-wrap min-h-[48px]">
          {{ detail.description || '（无描述）' }}
        </div>

        <!-- 涉及资源 -->
        <template v-if="detail.affectedResources">
          <Divider orientation="left">涉及资源</Divider>
          <div class="bg-gray-50 rounded p-3 text-sm text-gray-600 whitespace-pre-wrap">
            {{ detail.affectedResources }}
          </div>
        </template>

        <!-- 依赖任务 -->
        <template v-if="detail.dependsOn">
          <Divider orientation="left">依赖任务</Divider>
          <div class="bg-gray-50 rounded p-3 text-sm text-gray-600">
            {{ detail.dependsOn }}
          </div>
        </template>

        <!-- 执行结果（已完成时显示） -->
        <template v-if="isCompleted && detail.result">
          <Divider orientation="left">执行结果</Divider>
          <div class="result-block bg-gray-50 rounded p-3 text-sm text-gray-800 whitespace-pre-wrap max-h-[320px] overflow-y-auto font-mono">
            {{ detail.result }}
          </div>
        </template>

        <!-- 上下文摘要 -->
        <template v-if="detail.contextSummary">
          <Divider orientation="left">上下文摘要</Divider>
          <div class="bg-gray-50 rounded p-3 text-sm text-gray-600 whitespace-pre-wrap max-h-[200px] overflow-y-auto">
            {{ detail.contextSummary }}
          </div>
        </template>

        <!-- 错误信息（非失败状态也可能有历史错误） -->
        <template v-if="detail.errorMessage && !isFailed">
          <Divider orientation="left">错误信息</Divider>
          <div class="bg-red-50 rounded p-3 text-sm text-red-600 whitespace-pre-wrap">
            {{ detail.errorMessage }}
          </div>
        </template>

        <!-- 底部重试按钮 -->
        <div v-if="isFailed" class="mt-4 flex justify-end">
          <Button
            type="primary"
            danger
            :loading="retrying"
            @click="handleRetry"
          >
            重试此任务
          </Button>
        </div>
      </div>

      <!-- 空状态 -->
      <div v-else-if="!loading" class="text-center text-gray-400 py-12">
        暂无数据
      </div>
    </Spin>
  </Modal>
</template>

<style scoped>
.task-detail {
  padding: 0 2px;
}

.result-block {
  line-height: 1.6;
}
</style>

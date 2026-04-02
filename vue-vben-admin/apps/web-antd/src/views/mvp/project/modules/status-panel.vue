<script setup lang="ts">
import { ref, computed } from 'vue';
import { useVbenModal } from '@vben/common-ui';
import {
  message,
  Modal,
  Progress,
  Statistic,
  Tag,
} from 'ant-design-vue';
import {
  getProjectStatus,
  confirmPlan,
  resumeProject,
} from '#/api/mvp/workflow';
import type { ProjectItem } from '#/api/mvp/project/types';

const emit = defineEmits<{ refresh: [] }>();

/** 当前展示的项目基本信息 */
const project = ref<ProjectItem | null>(null);

/** 通过 workflow API 获取的实时状态数据 */
const statusData = ref<{
  status: string;
  pauseReason?: string;
  totalTasks: number;
  statusCounts: Record<string, number>;
} | null>(null);

/** 加载中 */
const loading = ref(false);

/** 状态对应的中文名称 */
const STATUS_LABELS: Record<string, string> = {
  designing: '设计中',
  running: '执行中',
  paused: '已暂停',
  completed: '已完成',
};

/** 状态对应的 Tag 颜色 */
const STATUS_COLORS: Record<string, string> = {
  designing: 'blue',
  running: 'green',
  paused: 'orange',
  completed: 'default',
};

/** 当前状态标签文字 */
const statusLabel = computed(() =>
  STATUS_LABELS[statusData.value?.status ?? ''] ?? statusData.value?.status ?? '-',
);

/** 当前状态颜色 */
const statusColor = computed(() =>
  STATUS_COLORS[statusData.value?.status ?? ''] ?? 'default',
);

/** 已完成任务数 */
const completedCount = computed(
  () => statusData.value?.statusCounts?.['completed'] ?? 0,
);

/** 运行中任务数 */
const runningCount = computed(
  () => statusData.value?.statusCounts?.['running'] ?? 0,
);

/** 失败任务数 */
const failedCount = computed(
  () => statusData.value?.statusCounts?.['failed'] ?? 0,
);

/** 等待中任务数 */
const pendingCount = computed(
  () => statusData.value?.statusCounts?.['pending'] ?? 0,
);

/** 总任务数 */
const totalTasks = computed(() => statusData.value?.totalTasks ?? 0);

/** 进度百分比（保留整数） */
const progressPercent = computed(() => {
  if (!totalTasks.value) return 0;
  return Math.floor((completedCount.value / totalTasks.value) * 100);
});

/** 进度条状态：有失败任务时显示异常色 */
const progressStatus = computed(() => {
  if (failedCount.value > 0) return 'exception';
  if (statusData.value?.status === 'completed') return 'success';
  return 'active';
});

/** 加载状态数据 */
async function loadStatus() {
  if (!project.value?.id) return;
  loading.value = true;
  try {
    statusData.value = await getProjectStatus(project.value.id);
  } catch {
    message.error('获取项目状态失败');
  } finally {
    loading.value = false;
  }
}

/** 确认方案操作 */
async function handleConfirmPlan() {
  if (!project.value?.id) return;
  Modal.confirm({
    title: '确认实施方案',
    content: '确认后系统将自动拆解并调度所有任务开始执行，是否继续？',
    okText: '确认执行',
    async onOk() {
      await confirmPlan(project.value!.id);
      message.success('方案已确认，任务开始调度执行');
      emit('refresh');
      await loadStatus();
    },
  });
}

/** 恢复项目操作 */
async function handleResume() {
  if (!project.value?.id) return;
  Modal.confirm({
    title: '恢复项目',
    content: '确定要恢复该项目的执行吗？',
    async onOk() {
      await resumeProject(project.value!.id);
      message.success('项目已恢复执行');
      emit('refresh');
      await loadStatus();
    },
  });
}

/** 弹窗（抽屉模式）配置 */
const [StatusDrawer, modalApi] = useVbenModal({
  fullscreenButton: false,
  title: '项目状态详情',
  footer: false,
  async onOpenChange(isOpen: boolean) {
    if (isOpen) {
      const data = modalApi.getData<{ project: ProjectItem } | null>();
      if (data?.project) {
        project.value = data.project;
        await loadStatus();
      }
    } else {
      project.value = null;
      statusData.value = null;
    }
  },
});
</script>

<template>
  <StatusDrawer class="w-[560px]">
    <div v-if="loading" class="flex justify-center py-8 text-gray-400">
      加载中...
    </div>
    <div v-else-if="project" class="flex flex-col gap-5">
      <!-- 项目基本信息 -->
      <div class="bg-gray-50 rounded-lg p-4">
        <h3 class="text-base font-semibold text-gray-800 mb-1">{{ project.name }}</h3>
        <p v-if="project.description" class="text-sm text-gray-500">{{ project.description }}</p>
      </div>

      <!-- 状态大标签 -->
      <div class="flex items-center gap-3">
        <span class="text-sm text-gray-500">当前状态：</span>
        <Tag :color="statusColor" class="text-sm px-3 py-0.5">
          {{ statusLabel }}
        </Tag>
      </div>

      <!-- 暂停原因（仅暂停时显示） -->
      <div
        v-if="statusData?.status === 'paused' && statusData?.pauseReason"
        class="bg-orange-50 border border-orange-200 rounded-lg p-3"
      >
        <p class="text-sm text-orange-600">
          <span class="font-medium">暂停原因：</span>{{ statusData.pauseReason }}
        </p>
      </div>

      <!-- 进度条 -->
      <div v-if="totalTasks > 0">
        <div class="flex justify-between text-sm text-gray-500 mb-1">
          <span>任务进度</span>
          <span>{{ completedCount }} / {{ totalTasks }}</span>
        </div>
        <Progress
          :percent="progressPercent"
          :status="progressStatus"
          :stroke-width="10"
        />
      </div>

      <!-- 任务统计卡片 -->
      <div v-if="totalTasks > 0" class="grid grid-cols-5 gap-3">
        <div class="bg-gray-50 rounded-lg p-3 text-center">
          <Statistic
            title="总任务"
            :value="totalTasks"
            :value-style="{ fontSize: '20px', color: '#595959' }"
          />
        </div>
        <div class="bg-blue-50 rounded-lg p-3 text-center">
          <Statistic
            title="执行中"
            :value="runningCount"
            :value-style="{ fontSize: '20px', color: '#1677ff' }"
          />
        </div>
        <div class="bg-green-50 rounded-lg p-3 text-center">
          <Statistic
            title="已完成"
            :value="completedCount"
            :value-style="{ fontSize: '20px', color: '#52c41a' }"
          />
        </div>
        <div class="bg-red-50 rounded-lg p-3 text-center">
          <Statistic
            title="失败"
            :value="failedCount"
            :value-style="{ fontSize: '20px', color: '#ff4d4f' }"
          />
        </div>
        <div class="bg-gray-50 rounded-lg p-3 text-center">
          <Statistic
            title="待执行"
            :value="pendingCount"
            :value-style="{ fontSize: '20px', color: '#8c8c8c' }"
          />
        </div>
      </div>

      <!-- 无任务时的提示 -->
      <div
        v-else-if="statusData?.status === 'designing'"
        class="text-center text-gray-400 text-sm py-4"
      >
        方案确认后将自动生成并调度 AI 任务
      </div>

      <!-- 操作按钮区 -->
      <div class="flex gap-2 pt-2 border-t border-gray-100">
        <!-- 设计中：确认方案 -->
        <button
          v-if="statusData?.status === 'designing'"
          class="px-4 py-1.5 bg-blue-600 text-white text-sm rounded-md hover:bg-blue-700 transition"
          @click="handleConfirmPlan"
        >
          确认方案开始执行
        </button>
        <!-- 暂停中：恢复执行 -->
        <button
          v-if="statusData?.status === 'paused'"
          class="px-4 py-1.5 bg-green-600 text-white text-sm rounded-md hover:bg-green-700 transition"
          @click="handleResume"
        >
          恢复执行
        </button>
        <!-- 刷新按钮 -->
        <button
          class="px-4 py-1.5 bg-gray-100 text-gray-600 text-sm rounded-md hover:bg-gray-200 transition"
          @click="loadStatus"
        >
          刷新状态
        </button>
      </div>
    </div>
  </StatusDrawer>
</template>

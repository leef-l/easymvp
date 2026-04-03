<script setup lang="ts">
import { h, ref, computed, onMounted, onUnmounted } from 'vue';
import type { VxeGridProps } from '#/adapter/vxe-table';
import { useRoute, useRouter } from 'vue-router';
import dayjs from 'dayjs';

import { useVbenModal } from '@vben/common-ui';
import {
  Button,
  Card,
  Col,
  message,
  Progress,
  Row,
  Select,
  SelectOption,
  Statistic,
  Tag,
  Input,
} from 'ant-design-vue';
import {
  ArrowLeftOutlined,
  ReloadOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  LoadingOutlined,
  UnorderedListOutlined,
} from '@ant-design/icons-vue';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import { getTaskTree, getTaskDetail } from '#/api/mvp/task';
import { getProjectList } from '#/api/mvp/project';
import { getProjectStatus, retryTask, skipTask } from '#/api/mvp/workflow';
import type { TaskItem } from '#/api/mvp/task/types';
import DetailDrawer from './modules/detail-drawer.vue';

// ===== 路由 =====
const route = useRoute();
const router = useRouter();
const projectId = computed(() => (route.query.projectId as string) || '');

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

// ===== 项目状态统计 =====
interface ProjectStatusData {
  status: string;
  pauseReason?: string;
  activeBatch: number;
  totalTasks: number;
  statusCounts: Record<string, number>;
}

// ===== 项目列表（供筛选下拉框使用） =====
interface ProjectOption {
  id: string;
  name: string;
}
const projectOptions = ref<ProjectOption[]>([]);

const projectStatus = ref<ProjectStatusData>({
  status: '',
  activeBatch: 0,
  totalTasks: 0,
  statusCounts: {},
});
const statusLoading = ref(false);

/** 计算进度百分比 */
const progressPercent = computed(() => {
  const total = projectStatus.value.totalTasks;
  if (!total) return 0;
  const completed = projectStatus.value.statusCounts?.completed || 0;
  return Math.round((completed / total) * 100);
});

/** 加载项目状态 */
async function loadProjectStatus() {
  if (!projectId.value) return;
  statusLoading.value = true;
  try {
    const res = await getProjectStatus(projectId.value);
    if (res) {
      projectStatus.value = res;
    }
  } catch {
    // 静默失败，不打断界面
  } finally {
    statusLoading.value = false;
  }
}

// 自动刷新定时器
let statusTimer: ReturnType<typeof setInterval> | null = null;

onMounted(() => {
  loadProjectStatus();
  loadProjectOptions();
  statusTimer = setInterval(loadProjectStatus, 5000);
});

onUnmounted(() => {
  if (statusTimer) clearInterval(statusTimer);
});

// ===== 搜索筛选 =====
const searchName = ref('');
const searchStatus = ref<string | undefined>(undefined);
const searchBatchNo = ref<number | undefined>(undefined);
const searchRole = ref<string | undefined>(undefined);
const searchProjectId = ref<string | undefined>(projectId.value || undefined);

/** 加载项目列表（供搜索下拉框使用） */
async function loadProjectOptions() {
  try {
    const res = await getProjectList({ pageNum: 1, pageSize: 200 });
    projectOptions.value = (res?.list ?? []).map((p: any) => ({
      id: String(p.id),
      name: p.name,
    }));
  } catch {
    // 静默
  }
}

// ===== 详情抽屉 =====
const [DetailDrawerComp, detailDrawerApi] = useVbenModal({
  connectedComponent: DetailDrawer,
  destroyOnClose: true,
});

// ===== 表格列配置 =====
const gridOptions: VxeGridProps<TaskItem> = {
  columns: [
    {
      field: 'projectName',
      title: '所属项目',
      width: 150,
      visible: !projectId.value, // 项目维度进入时隐藏
    },
    {
      field: 'name',
      title: '任务名称',
      treeNode: true,
      minWidth: 200,
    },
    {
      field: 'roleType',
      title: '角色类型',
      width: 100,
      slots: {
        default: ({ row }: { row: TaskItem }) => {
          const m = ROLE_MAP[row.roleType] ?? { color: 'default', text: row.roleType };
          return [h(Tag, { color: m.color }, () => m.text)];
        },
      },
    },
    {
      field: 'roleLevel',
      title: '角色等级',
      width: 90,
      slots: {
        default: ({ row }: { row: TaskItem }) => {
          if (!row.roleLevel) return ['-'];
          const m = LEVEL_MAP[row.roleLevel] ?? { color: 'default', text: row.roleLevel };
          return [h(Tag, { color: m.color }, () => m.text)];
        },
      },
    },
    {
      field: 'batchNo',
      title: '批次',
      width: 70,
    },
    {
      field: 'status',
      title: '状态',
      width: 110,
      slots: {
        default: ({ row }: { row: TaskItem }) => {
          if (!row.status) return ['-'];
          const m = STATUS_MAP[row.status] ?? { color: 'default', text: row.status };
          return [h(Tag, { color: m.color }, () => m.text)];
        },
      },
    },
    {
      field: 'startedAt',
      title: '开始时间',
      width: 160,
      formatter({ cellValue }) {
        return cellValue ? dayjs(cellValue).format('YYYY-MM-DD HH:mm:ss') : '-';
      },
    },
    {
      field: 'completedAt',
      title: '完成时间',
      width: 160,
      formatter({ cellValue }) {
        return cellValue ? dayjs(cellValue).format('YYYY-MM-DD HH:mm:ss') : '-';
      },
    },
    {
      title: '操作',
      width: 200,
      fixed: 'right',
      slots: { default: 'action' },
    },
  ],
  pagerConfig: { enabled: false },
  treeConfig: {
    childrenField: 'children',
    expandAll: true,
  },
  proxyConfig: {
    ajax: {
      query: async () => {
        const params: Record<string, any> = {};
        // 优先用 URL 参数中的项目 ID，否则用搜索栏选择的项目
        const pid = projectId.value || searchProjectId.value;
        if (pid) params.projectID = pid;
        if (searchName.value) params.name = searchName.value;
        if (searchStatus.value) params.status = searchStatus.value;
        if (searchBatchNo.value !== undefined && searchBatchNo.value !== null) {
          params.batchNo = searchBatchNo.value;
        }
        if (searchRole.value) params.roleType = searchRole.value;
        const list = (await getTaskTree(params as any)) ?? [];
        return { items: list, total: list.length };
      },
    },
  },
  toolbarConfig: {
    custom: true,
    refresh: false, // 用自定义刷新按钮
  },
};

const [Grid, gridApi] = useVbenVxeGrid({ gridOptions });

// ===== 操作函数 =====

/** 查看详情 */
function handleView(row: TaskItem) {
  detailDrawerApi.setData({ id: row.id, projectId: row.projectID }).open();
}

/** 重试失败任务 */
async function handleRetry(row: TaskItem) {
  const pid = projectId.value || row.projectID;
  if (!pid) return;
  try {
    await retryTask({ projectID: pid, taskID: row.id });
    message.success('已提交重试请求');
    gridApi.reload();
    loadProjectStatus();
  } catch {
    message.error('重试失败');
  }
}

/** 跳过失败任务 */
async function handleSkip(row: TaskItem) {
  const pid = projectId.value || row.projectID;
  if (!pid) return;
  const reason = window.prompt('请输入跳过原因：');
  if (!reason) return;
  try {
    await skipTask({ projectID: pid, taskID: row.id, reason });
    message.success('已跳过该任务');
    gridApi.reload();
    loadProjectStatus();
  } catch {
    message.error('跳过失败');
  }
}

/** 跳转到对话页 */
function handleViewConversation(row: TaskItem) {
  // 直接用 task 的 conversationID 跳转对话页
  router.push({
    path: '/mvp/chat',
    query: { projectId: projectId.value, conversationId: row.conversationID || '' },
  });
}

/** 手动刷新 */
function handleRefresh() {
  gridApi.reload();
  loadProjectStatus();
}

/** 搜索 */
function handleSearch() {
  gridApi.reload();
}

/** 返回项目列表 */
function handleBack() {
  router.push('/mvp/project');
}

// ===== 批次选项（动态从数据中收集） =====
const batchOptions = ref<number[]>([]);

// 监听表格数据变化，收集批次号
async function refreshBatchOptions() {
  try {
    const params: Record<string, any> = {};
    if (projectId.value) params.projectID = projectId.value;
    const list = await getTaskTree(params as any);
    const batchSet = new Set<number>();
    function collectBatch(items: TaskItem[]) {
      for (const item of items) {
        if (item.batchNo !== undefined && item.batchNo !== null) {
          batchSet.add(item.batchNo);
        }
        if (item.children?.length) collectBatch(item.children);
      }
    }
    collectBatch(list ?? []);
    batchOptions.value = [...batchSet].sort((a, b) => a - b);
  } catch {
    // 忽略
  }
}

onMounted(() => {
  refreshBatchOptions();
});
</script>

<template>
  <div class="task-monitor-page flex flex-col h-full bg-gray-50 p-4 overflow-y-auto">
    <DetailDrawerComp />

    <!-- 页面头部 -->
    <div class="flex items-center justify-between mb-4">
      <div class="flex items-center gap-3">
        <Button type="text" :icon="h(ArrowLeftOutlined)" @click="handleBack">
          返回项目列表
        </Button>
        <span class="text-lg font-semibold text-gray-800">项目任务监控</span>
        <Tag v-if="projectStatus.status" :color="STATUS_MAP[projectStatus.status]?.color || 'default'">
          {{ STATUS_MAP[projectStatus.status]?.text || projectStatus.status }}
        </Tag>
      </div>
      <Button :icon="h(ReloadOutlined)" @click="handleRefresh">刷新</Button>
    </div>

    <!-- 统计卡片（仅项目维度时显示） -->
    <template v-if="projectId">
      <Row :gutter="16" class="mb-4">
        <Col :span="6">
          <Card size="small" class="stat-card">
            <Statistic
              title="总任务"
              :value="projectStatus.totalTasks"
              :prefix="h(UnorderedListOutlined, { style: { color: '#8c8c8c' } })"
            />
          </Card>
        </Col>
        <Col :span="6">
          <Card size="small" class="stat-card">
            <Statistic
              title="执行中"
              :value="projectStatus.statusCounts?.running || 0"
              :value-style="{ color: '#1677ff' }"
              :prefix="h(LoadingOutlined, { style: { color: '#1677ff' } })"
            />
          </Card>
        </Col>
        <Col :span="6">
          <Card size="small" class="stat-card">
            <Statistic
              title="已完成"
              :value="projectStatus.statusCounts?.completed || 0"
              :value-style="{ color: '#52c41a' }"
              :prefix="h(CheckCircleOutlined, { style: { color: '#52c41a' } })"
            />
          </Card>
        </Col>
        <Col :span="6">
          <Card size="small" class="stat-card">
            <Statistic
              title="失败"
              :value="(projectStatus.statusCounts?.failed || 0) + (projectStatus.statusCounts?.submit_error || 0)"
              :value-style="{ color: '#ff4d4f' }"
              :prefix="h(CloseCircleOutlined, { style: { color: '#ff4d4f' } })"
            />
          </Card>
        </Col>
      </Row>

      <!-- 进度条 -->
      <Card size="small" class="mb-4">
        <div class="flex items-center gap-3">
          <span class="text-gray-600 text-sm whitespace-nowrap">执行进度</span>
          <Progress
            :percent="progressPercent"
            :stroke-color="{ '0%': '#108ee9', '100%': '#87d068' }"
            class="flex-1"
          />
          <span class="text-gray-500 text-sm whitespace-nowrap">
            {{ projectStatus.statusCounts?.completed || 0 }} / {{ projectStatus.totalTasks }}
          </span>
        </div>
        <div class="flex items-center gap-4 mt-2 text-sm">
          <span v-if="projectStatus.activeBatch > 0" class="text-blue-500">
            当前执行批次：第 {{ projectStatus.activeBatch }} 批
          </span>
          <span v-else-if="projectStatus.status === 'running'" class="text-green-500">
            所有批次已完成
          </span>
          <span v-if="projectStatus.pauseReason" class="text-orange-500">
            暂停原因：{{ projectStatus.pauseReason }}
          </span>
        </div>
      </Card>
    </template>

    <!-- 搜索筛选栏 -->
    <Card size="small" class="mb-3">
      <div class="flex flex-wrap gap-3 items-center">
        <Select
          v-if="!projectId"
          v-model:value="searchProjectId"
          placeholder="所属项目"
          allow-clear
          show-search
          option-filter-prop="label"
          style="width: 180px"
          @change="handleSearch"
        >
          <SelectOption
            v-for="p in projectOptions"
            :key="p.id"
            :value="p.id"
            :label="p.name"
          >
            {{ p.name }}
          </SelectOption>
        </Select>
        <Input
          v-model:value="searchName"
          placeholder="任务名称"
          allow-clear
          style="width: 180px"
          @press-enter="handleSearch"
          @change="(e: Event) => { if (!(e.target as HTMLInputElement).value) handleSearch(); }"
        />
        <Select
          v-model:value="searchStatus"
          placeholder="状态"
          allow-clear
          style="width: 130px"
          @change="handleSearch"
        >
          <SelectOption v-for="(v, k) in STATUS_MAP" :key="k" :value="k">
            <Tag :color="v.color" class="mr-1" />{{ v.text }}
          </SelectOption>
        </Select>
        <Select
          v-model:value="searchBatchNo"
          placeholder="批次"
          allow-clear
          style="width: 100px"
          @change="handleSearch"
        >
          <SelectOption v-for="b in batchOptions" :key="b" :value="b">
            批次 {{ b }}
          </SelectOption>
        </Select>
        <Select
          v-model:value="searchRole"
          placeholder="角色类型"
          allow-clear
          style="width: 120px"
          @change="handleSearch"
        >
          <SelectOption v-for="(v, k) in ROLE_MAP" :key="k" :value="k">
            <Tag :color="v.color" class="mr-1" />{{ v.text }}
          </SelectOption>
        </Select>
        <Button type="primary" @click="handleSearch">查询</Button>
        <Button
          @click="() => {
            searchName = '';
            searchStatus = undefined;
            searchBatchNo = undefined;
            searchRole = undefined;
            if (!projectId) searchProjectId = undefined;
            handleSearch();
          }"
        >
          重置
        </Button>
      </div>
    </Card>

    <!-- 任务树表格 -->
    <div class="flex-1 min-h-0">
      <Grid>
        <!-- 操作列 -->
        <template #action="{ row }">
          <!-- 查看详情（所有状态都显示） -->
          <Button type="link" size="small" @click="handleView(row)">详情</Button>

          <!-- 执行中：查看对话 -->
          <Button
            v-if="row.status === 'running'"
            type="link"
            size="small"
            @click="handleViewConversation(row)"
          >
            查看对话
          </Button>

          <!-- 已完成：查看结果 -->
          <Button
            v-if="row.status === 'completed'"
            type="link"
            size="small"
            @click="handleView(row)"
          >
            查看结果
          </Button>

          <!-- 失败：重试 + 跳过 -->
          <Button
            v-if="row.status === 'failed' || row.status === 'submit_error'"
            type="link"
            size="small"
            danger
            @click="handleRetry(row)"
          >
            重试
          </Button>
          <Button
            v-if="row.status === 'failed' || row.status === 'bug_found'"
            type="link"
            size="small"
            @click="handleSkip(row)"
          >
            跳过
          </Button>
        </template>
      </Grid>
    </div>
  </div>
</template>

<style scoped>
.task-monitor-page {
  min-height: 100%;
}

.stat-card {
  border-radius: 8px;
}

:deep(.vxe-table--body-wrapper) {
  overflow-y: auto;
}
</style>

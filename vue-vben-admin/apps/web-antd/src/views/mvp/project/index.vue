<script setup lang="ts">
import { useRouter } from 'vue-router';
import type { VbenFormProps } from '#/adapter/form';
import type { VxeGridProps } from '#/adapter/vxe-table';

import { Page, useVbenModal } from '@vben/common-ui';
import { Button, Dropdown, Menu, MenuItem, message, Modal, Progress, Tag } from 'ant-design-vue';
import { DownOutlined } from '@ant-design/icons-vue';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import { getProjectList, deleteProject } from '#/api/mvp/project';
import { confirmPlan, resumeProject, batchProjectStats } from '#/api/mvp/workflow';
import type { ProjectItem } from '#/api/mvp/project/types';
import { stageTypeMap } from '../consts';

import FormModal from './modules/form.vue';
import PauseModal from './modules/pause-modal.vue';
import StatusPanel from './modules/status-panel.vue';

const router = useRouter();

/** 状态中文名称映射 */
const STATUS_LABELS: Record<string, string> = {
  designing: '设计中',
  reviewing: '方案审核中',
  running: '执行中',
  executing: '执行中',
  reworking: '返工中',
  accepting: '验收中',
  paused: '已暂停',
  completed: '已完成',
  failed: '已失败',
  canceled: '已取消',
};

/** 状态 Tag 颜色映射 */
const STATUS_COLORS: Record<string, string> = {
  designing: 'blue',
  reviewing: 'cyan',
  running: 'green',
  executing: 'green',
  reworking: 'orange',
  accepting: 'purple',
  paused: 'orange',
  completed: 'default',
  failed: 'red',
  canceled: 'default',
};

/** 新建项目弹窗 */
const [FormModalComp, formModalApi] = useVbenModal({
  connectedComponent: FormModal,
  destroyOnClose: true,
});

/** 暂停弹窗 */
const [PauseModalComp, pauseModalApi] = useVbenModal({
  connectedComponent: PauseModal,
  destroyOnClose: true,
});

/** 项目状态详情抽屉 */
const [StatusPanelComp, statusPanelApi] = useVbenModal({
  connectedComponent: StatusPanel,
  destroyOnClose: true,
});

/** 搜索表单配置 */
const formOptions: VbenFormProps = {
  collapsed: false,
  showCollapseButton: false,
  submitOnEnter: true,
  schema: [
    {
      component: 'Input',
      componentProps: { placeholder: '请输入项目名称', allowClear: true },
      fieldName: 'name',
      label: '项目名称',
    },
    {
      component: 'Select',
      fieldName: 'status',
      label: '项目状态',
      componentProps: {
        placeholder: '全部状态',
        allowClear: true,
        options: [
          { label: '设计中', value: 'designing' },
          { label: '方案审核中', value: 'reviewing' },
          { label: '执行中', value: 'executing' },
          { label: '返工中', value: 'reworking' },
          { label: '验收中', value: 'accepting' },
          { label: '已暂停', value: 'paused' },
          { label: '已完成', value: 'completed' },
        ],
      },
    },
  ],
};

/** 表格列配置 */
const gridOptions: VxeGridProps<ProjectItem> = {
  columns: [
    { title: '序号', type: 'seq', width: 60 },
    {
      field: 'name',
      title: '项目名称',
      minWidth: 160,
      slots: { default: 'col-name' },
    },
    {
      field: 'projectCategory',
      title: '项目分类',
      width: 120,
    },
    {
      field: 'status',
      title: '状态',
      width: 100,
      slots: { default: 'col-status' },
    },
    {
      field: 'currentStage',
      title: '当前阶段',
      width: 100,
      slots: { default: 'col-stage' },
    },
    {
      field: 'taskProgress',
      title: '任务进度',
      width: 180,
      slots: { default: 'col-progress' },
    },
    {
      field: 'pauseReason',
      title: '暂停原因',
      minWidth: 160,
      showOverflow: 'tooltip',
    },
    {
      field: 'createdAt',
      title: '创建时间',
      width: 180,
      formatter: 'formatDateTime',
      sortable: true,
    },
    {
      title: '操作',
      width: 200,
      fixed: 'right',
      slots: { default: 'action' },
    },
  ],
  height: 'auto',
  pagerConfig: {},
  proxyConfig: {
    ajax: {
      query: async ({ page, sorts }, formValues) => {
        const params: Record<string, any> = {
          pageNum: page.currentPage,
          pageSize: page.pageSize,
          ...formValues,
        };
        if (sorts && sorts.length > 0) {
          const sort = sorts[0];
          if (sort?.field && sort?.order) {
            params.orderBy = sort.field;
            params.orderDir = sort.order;
          }
        }
        const res = await getProjectList(params as any);
        const items = res?.list ?? [];

        // 异步加载运行时统计（进度、阶段）
        if (items.length > 0) {
          const ids = items.map((p: ProjectItem) => p.id);
          batchProjectStats(ids).then((statsRes) => {
            const statsMap = new Map(
              (statsRes?.stats ?? []).map((s) => [s.projectID, s]),
            );
            for (const item of items) {
              const stat = statsMap.get(item.id);
              if (stat) {
                item.currentStage = stat.currentStage;
                item.totalTasks = stat.totalTasks;
                item.completedTasks = stat.completedTasks;
                item.failedTasks = stat.failedTasks;
                item.runningTasks = stat.runningTasks;
              }
            }
            // 触发表格刷新渲染
            gridApi.grid?.loadData(items);
          }).catch(() => { /* 静默失败 */ });
        }

        return { items, total: res?.total ?? 0 };
      },
    },
  },
  sortConfig: { remote: true, trigger: 'cell' },
  toolbarConfig: { custom: true, refresh: true, search: true },
};

const [Grid, gridApi] = useVbenVxeGrid({ formOptions, gridOptions });

/** 新建项目 */
function handleCreate() {
  formModalApi.setData(null).open();
}

/** 编辑项目 */
function handleEdit(row: ProjectItem) {
  formModalApi.setData({ id: row.id }).open();
}

/** 进入对话（与AI架构师对话） */
function handleChat(row: ProjectItem) {
  router.push({ path: '/mvp/chat', query: { projectId: row.id } });
}

/** 查看详情（打开状态面板） */
function handleViewStatus(row: ProjectItem) {
  statusPanelApi.setData({ project: row }).open();
}

/** 打开执行控制台 */
function handleExecution(row: ProjectItem) {
  router.push({ path: '/mvp/workflow/execution', query: { projectId: row.id } });
}

/** 打开仪表板 */
function handleDashboard(row: ProjectItem) {
  router.push({ path: '/mvp/workflow/dashboard', query: { projectId: row.id } });
}

/** 打开元认知页面 */
function handleMetaCognition(row: ProjectItem) {
  router.push({ path: '/mvp/workflow/meta-cognition', query: { projectID: row.id } });
}

/** 确认方案（设计阶段 -> 执行阶段） */
function handleConfirmPlan(row: ProjectItem) {
  Modal.confirm({
    title: '确认实施方案',
    content: '确认后系统将先执行方案审核（系统预检 + AI审核 + 调度优化），审核通过后自动开始执行。',
    okText: '确认方案',
    async onOk() {
      const res = await confirmPlan(row.id);
      if (res?.reviewPassed) {
        message.success('方案审核通过，已进入执行阶段');
      } else {
        const errorList = (res?.issues || [])
          .filter((i: any) => i.severity === 'error')
          .map((i: any) => `• ${i.taskName ? `[${i.taskName}] ` : ''}${i.message}`)
          .slice(0, 20);
        Modal.error({
          title: `方案审核未通过（${res?.errorCount || 0} 个错误，${res?.warningCount || 0} 个警告）`,
          content: errorList.length > 0
            ? errorList.join('\n')
            : (res?.rejectReason || '审核未通过，请检查方案后重试'),
          width: 600,
        });
      }
      gridApi.reload();
    },
  });
}

/** 暂停项目 */
function handlePause(row: ProjectItem) {
  pauseModalApi.setData({ projectID: row.id }).open();
}

/** 恢复项目 */
function handleResume(row: ProjectItem) {
  Modal.confirm({
    title: '恢复项目',
    content: '确定要恢复该项目的执行吗？',
    async onOk() {
      await resumeProject(row.id);
      message.success('项目已恢复执行');
      gridApi.reload();
    },
  });
}

/** 删除项目 */
function handleDelete(row: ProjectItem) {
  Modal.confirm({
    title: '确认删除',
    content: `确定要删除项目「${row.name}」吗？此操作不可恢复。`,
    okType: 'danger',
    async onOk() {
      await deleteProject(row.id);
      message.success('删除成功');
      gridApi.reload();
    },
  });
}
</script>

<template>
  <Page auto-content-height>
    <!-- 弹窗组件挂载 -->
    <FormModalComp @success="() => gridApi.reload()" />
    <PauseModalComp @success="() => gridApi.reload()" />
    <StatusPanelComp @refresh="() => gridApi.reload()" />

    <Grid>
      <!-- 工具栏操作区 -->
      <template #toolbar-actions>
        <Button type="primary" @click="handleCreate">新建项目</Button>
      </template>

      <!-- 项目名称列 -->
      <template #col-name="{ row }">
        <span class="font-medium text-gray-800">{{ row.name }}</span>
        <p v-if="row.description" class="text-xs text-gray-400 mt-0.5 truncate max-w-xs">
          {{ row.description }}
        </p>
      </template>

      <!-- 状态 Tag 列 -->
      <template #col-status="{ row }">
        <Tag :color="STATUS_COLORS[row.status] ?? 'default'">
          {{ STATUS_LABELS[row.status] ?? row.status ?? '-' }}
        </Tag>
      </template>

      <!-- 当前阶段列 -->
      <template #col-stage="{ row }">
        <Tag v-if="row.currentStage" :color="stageTypeMap[row.currentStage]?.color ?? 'default'" size="small">
          {{ stageTypeMap[row.currentStage]?.label ?? row.currentStage }}
        </Tag>
        <span v-else class="text-xs text-gray-300">-</span>
      </template>

      <!-- 任务进度列 -->
      <template #col-progress="{ row }">
        <div v-if="row.totalTasks > 0" class="flex items-center gap-2">
          <Progress
            :percent="
              row.totalTasks
                ? Math.floor(((row.completedTasks ?? 0) / row.totalTasks) * 100)
                : 0
            "
            size="small"
            class="flex-1"
          />
          <span class="text-xs text-gray-400 whitespace-nowrap">
            {{ row.completedTasks ?? 0 }}/{{ row.totalTasks }}
          </span>
        </div>
        <span v-else class="text-xs text-gray-300">-</span>
      </template>

      <!-- 操作列：2 个主按钮 + 更多下拉 -->
      <template #action="{ row }">
        <!-- 主按钮1：根据状态显示最关键操作 -->
        <Button v-if="row.status === 'designing'" type="link" size="small" @click="handleChat(row)">对话</Button>
        <Button v-else-if="row.status === 'paused'" type="link" size="small" class="text-green-600" @click="handleResume(row)">恢复</Button>
        <Button v-else type="link" size="small" @click="handleDashboard(row)">仪表板</Button>

        <!-- 主按钮2 -->
        <Button v-if="row.status === 'designing'" type="link" size="small" @click="handleDashboard(row)">仪表板</Button>
        <Button v-else-if="['running', 'executing', 'reworking'].includes(row.status)" type="link" size="small" class="text-orange-500" @click="handlePause(row)">暂停</Button>
        <Button v-else-if="row.status === 'paused'" type="link" size="small" @click="handleDashboard(row)">仪表板</Button>

        <!-- 更多操作下拉 -->
        <Dropdown>
          <Button type="link" size="small">
            更多 <DownOutlined />
          </Button>
          <template #overlay>
            <Menu>
              <MenuItem v-if="row.status === 'designing'" @click="handleConfirmPlan(row)">确认方案</MenuItem>
              <MenuItem v-if="row.status === 'designing'" @click="handleDashboard(row)">查看详情</MenuItem>
              <MenuItem v-if="row.status !== 'designing'" @click="handleChat(row)">进入对话</MenuItem>
              <MenuItem v-if="row.status !== 'designing'" @click="handleDashboard(row)">仪表板</MenuItem>
              <MenuItem @click="handleEdit(row)">编辑</MenuItem>
              <MenuItem danger @click="handleDelete(row)">删除</MenuItem>
            </Menu>
          </template>
        </Dropdown>
      </template>
    </Grid>
  </Page>
</template>

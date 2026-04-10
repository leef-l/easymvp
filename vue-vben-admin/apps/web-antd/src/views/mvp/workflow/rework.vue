<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useRoute } from 'vue-router';

import { Page } from '@vben/common-ui';

import {
  CheckCircleOutlined,
  ReloadOutlined,
  SyncOutlined,
  ToolOutlined,
} from '@ant-design/icons-vue';
import {
  Card,
  Collapse,
  CollapsePanel,
  Descriptions,
  DescriptionsItem,
  Empty,
  Spin,
  Tag,
  Timeline,
  TimelineItem,
} from 'ant-design-vue';

import { getReworkStatus, type ReworkRoundInfo, type ReworkStageInfo } from '#/api/mvp/workflow';

defineOptions({ name: 'WorkflowRework' });

const props = defineProps<{ projectId?: string }>();

const route = useRoute();
const resolvedProjectId = computed(() => props.projectId || (route.query.projectId as string) || '');
const loading = ref(false);

const hasRework = ref(false);
const reworkRounds = ref(0);
const currentStage = ref<null | ReworkStageInfo>(null);
const history = ref<ReworkRoundInfo[]>([]);

const stageStatusConfig: Record<string, { color: string; label: string; }> = {
  running: { label: '返工中', color: 'processing' },
  completed: { label: '已完成', color: 'success' },
  failed: { label: '失败', color: 'error' },
  pending: { label: '待启动', color: 'default' },
};

function formatTime(t?: string) {
  if (!t) return '-';
  const d = new Date(t);
  return `${d.getMonth() + 1}/${d.getDate()} ${d.getHours().toString().padStart(2, '0')}:${d.getMinutes().toString().padStart(2, '0')}`;
}

/** 截断过长文本 */
function truncate(text: string, max = 200) {
  if (!text) return '-';
  return text.length > max ? `${text.slice(0, max)}...` : text;
}

async function loadData() {
  if (!resolvedProjectId.value) return;
  loading.value = true;
  try {
    const res = await getReworkStatus(resolvedProjectId.value);
    hasRework.value = res?.hasRework ?? false;
    reworkRounds.value = res?.reworkRounds ?? 0;
    currentStage.value = res?.currentStage ?? null;
    history.value = res?.history ?? [];
  } finally {
    loading.value = false;
  }
}

function resetReworkState() {
  hasRework.value = false;
  reworkRounds.value = 0;
  currentStage.value = null;
  history.value = [];
}

watch(
  resolvedProjectId,
  (value) => {
    resetReworkState();
    if (!value) return;
    loadData();
  },
  { immediate: true },
);
</script>

<template>
  <Page auto-content-height>
    <Spin :spinning="loading">
      <Empty
        v-if="!resolvedProjectId"
        description="请从项目列表进入查看返工状态"
        class="mt-20"
      />

      <template v-else-if="!hasRework && !loading">
        <Card>
          <Empty description="当前工作流无返工记录">
            <template #image>
              <CheckCircleOutlined style="font-size: 48px; color: #52c41a" />
            </template>
          </Empty>
        </Card>
      </template>

      <template v-else>
        <!-- 返工概览 -->
        <Card title="返工概览" class="mb-4">
          <template #extra>
            <a @click="loadData"><ReloadOutlined /> 刷新</a>
          </template>
          <Descriptions :column="3" bordered size="small">
            <DescriptionsItem label="返工轮次">
              <Tag color="orange">{{ reworkRounds }} 轮</Tag>
            </DescriptionsItem>
            <DescriptionsItem label="当前状态">
              <Tag
                v-if="currentStage"
                :color="stageStatusConfig[currentStage.status]?.color ?? 'default'"
              >
                <template #icon>
                  <SyncOutlined v-if="currentStage.status === 'running'" spin />
                </template>
                {{ stageStatusConfig[currentStage.status]?.label ?? currentStage.status }}
              </Tag>
              <span v-else class="text-gray-400">无活跃返工</span>
            </DescriptionsItem>
            <DescriptionsItem v-if="currentStage" label="开始时间">
              {{ formatTime(currentStage.startedAt) }}
            </DescriptionsItem>
          </Descriptions>
        </Card>

        <!-- 返工历史 -->
        <Card title="返工历史" class="mb-4">
          <Timeline>
            <TimelineItem
              v-for="round in history"
              :key="round.round"
              color="orange"
            >
              <template #dot>
                <ToolOutlined />
              </template>

              <div class="mb-2 font-medium">
                第 {{ round.round }} 轮返工
                <span class="ml-2 text-xs text-gray-400">{{ formatTime(round.createdAt) }}</span>
              </div>

              <Collapse :bordered="false" class="bg-transparent">
                <CollapsePanel :key="`round-${round.round}`" header="查看详情">
                  <Descriptions :column="1" size="small" bordered>
                    <DescriptionsItem label="失败任务">
                      <div>
                        <Tag color="red">{{ round.failedTaskName || round.failedTaskID }}</Tag>
                      </div>
                    </DescriptionsItem>
                    <DescriptionsItem label="失败原因">
                      <div class="max-w-lg whitespace-pre-wrap text-xs text-gray-600">
                        {{ truncate(round.failedReason, 500) }}
                      </div>
                    </DescriptionsItem>
                    <DescriptionsItem v-if="round.analysisTaskID" label="分析任务 ID">
                      {{ round.analysisTaskID }}
                    </DescriptionsItem>
                    <DescriptionsItem v-if="round.analysisResult" label="分析结果">
                      <div class="max-w-lg whitespace-pre-wrap text-xs text-gray-600">
                        {{ truncate(round.analysisResult, 500) }}
                      </div>
                    </DescriptionsItem>
                    <DescriptionsItem label="交接类型">
                      <Tag>{{ round.handoffType }}</Tag>
                    </DescriptionsItem>
                  </Descriptions>
                </CollapsePanel>
              </Collapse>
            </TimelineItem>
          </Timeline>
        </Card>
      </template>
    </Spin>
  </Page>
</template>

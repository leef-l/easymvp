<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { useRoute } from 'vue-router';

import { Page } from '@vben/common-ui';
import { Card, Empty, Spin, Tag, Timeline, TimelineItem } from 'ant-design-vue';
import {
  CheckCircleOutlined,
  CloseCircleOutlined,
  ClockCircleOutlined,
  ExclamationCircleOutlined,
  ReloadOutlined,
  RocketOutlined,
} from '@ant-design/icons-vue';

import { getTimeline, getStageHistory, type TimelineEvent, type StageHistoryItem } from '#/api/mvp/workflow';
import { stageTypeMap } from '../consts';

defineOptions({ name: 'WorkflowTimeline' });

const route = useRoute();
const projectId = ref((route.query.projectId as string) ?? '');
const loading = ref(false);
const events = ref<TimelineEvent[]>([]);
const stages = ref<StageHistoryItem[]>([]);

/** 事件类型 → 颜色/图标 */
function eventColor(eventType: string): string {
  if (eventType.includes('completed') || eventType.includes('approved')) return 'green';
  if (eventType.includes('failed') || eventType.includes('rejected')) return 'red';
  if (eventType.includes('paused') || eventType.includes('escalated')) return 'orange';
  if (eventType.includes('created') || eventType.includes('started')) return 'blue';
  return 'gray';
}

function eventIcon(eventType: string) {
  if (eventType.includes('completed') || eventType.includes('approved')) return CheckCircleOutlined;
  if (eventType.includes('failed') || eventType.includes('rejected')) return CloseCircleOutlined;
  if (eventType.includes('paused') || eventType.includes('escalated')) return ExclamationCircleOutlined;
  if (eventType.includes('started') || eventType.includes('created')) return RocketOutlined;
  return ClockCircleOutlined;
}

/** 格式化时间 */
function formatTime(t?: string) {
  if (!t) return '';
  const d = new Date(t);
  return `${d.getMonth() + 1}/${d.getDate()} ${d.getHours().toString().padStart(2, '0')}:${d.getMinutes().toString().padStart(2, '0')}:${d.getSeconds().toString().padStart(2, '0')}`;
}

const stageStatusColor: Record<string, string> = {
  running: 'processing',
  completed: 'success',
  failed: 'error',
  pending: 'default',
};

async function loadData() {
  if (!projectId.value) return;
  loading.value = true;
  try {
    const [timelineRes, stageRes] = await Promise.all([
      getTimeline(projectId.value, 100),
      getStageHistory(projectId.value),
    ]);
    events.value = timelineRes?.events ?? [];
    stages.value = stageRes?.stages ?? [];
  } finally {
    loading.value = false;
  }
}

onMounted(loadData);
</script>

<template>
  <Page auto-content-height>
    <Spin :spinning="loading">
      <Empty
        v-if="!projectId"
        description="请从项目列表进入查看工作流时间线"
        class="mt-20"
      />
      <template v-else>
        <div class="grid grid-cols-1 gap-4 xl:grid-cols-3">
          <!-- 阶段历程 -->
          <Card title="阶段历程" class="xl:col-span-1">
            <template #extra>
              <a @click="loadData"><ReloadOutlined /> 刷新</a>
            </template>
            <Empty v-if="stages.length === 0" description="暂无阶段记录" />
            <div v-else class="space-y-3">
              <div
                v-for="stage in stages"
                :key="stage.id"
                class="flex items-center justify-between rounded-lg border p-3"
              >
                <div>
                  <Tag :color="stageTypeMap[stage.stageType]?.color ?? 'default'">
                    {{ stageTypeMap[stage.stageType]?.label ?? stage.stageType }}
                  </Tag>
                  <span class="ml-2 text-xs text-gray-400">#{{ stage.stageNo }}</span>
                </div>
                <div class="flex items-center gap-2">
                  <Tag :color="stageStatusColor[stage.status] ?? 'default'" size="small">
                    {{ stage.status }}
                  </Tag>
                  <span class="text-xs text-gray-400">{{ formatTime(stage.startedAt) }}</span>
                </div>
              </div>
            </div>
          </Card>

          <!-- 事件时间线 -->
          <Card title="事件流" class="xl:col-span-2">
            <Empty v-if="events.length === 0" description="暂无事件记录" />
            <Timeline v-else>
              <TimelineItem
                v-for="event in events"
                :key="event.id"
                :color="eventColor(event.eventType)"
              >
                <template #dot>
                  <component :is="eventIcon(event.eventType)" />
                </template>
                <div class="flex items-start justify-between">
                  <div>
                    <span class="font-medium">{{ event.label }}</span>
                    <div class="mt-1 text-xs text-gray-400">
                      {{ event.entityType }} · {{ event.eventType }}
                    </div>
                  </div>
                  <span class="flex-shrink-0 text-xs text-gray-400">
                    {{ formatTime(event.createdAt) }}
                  </span>
                </div>
              </TimelineItem>
            </Timeline>
          </Card>
        </div>
      </template>
    </Spin>
  </Page>
</template>

<script setup lang="ts">
import type { Component } from 'vue';

import { computed, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';

import { Page } from '@vben/common-ui';
import { Alert, Button, Card, Empty, Progress, Spin, Tag } from 'ant-design-vue';
import {
  CheckCircleOutlined,
  CloseCircleOutlined,
  ExclamationCircleOutlined,
  RightOutlined,
  ReloadOutlined,
  SettingOutlined,
} from '@ant-design/icons-vue';

import { getSystemCheck } from '#/api/mvp/workflow';
import type { SystemCheckItem } from '#/api/mvp/workflow';

defineOptions({ name: 'MvpDashboard' });

const router = useRouter();
const loading = ref(false);
const allPass = ref(false);
const items = ref<SystemCheckItem[]>([]);

async function fetchCheck() {
  loading.value = true;
  try {
    const res = await getSystemCheck();
    items.value = res.items ?? [];
    allPass.value = res.allPass ?? false;
  } finally {
    loading.value = false;
  }
}

onMounted(fetchCheck);

function goToLink(link: string) {
  if (link) router.push(link);
}

type StatusType = 'error' | 'ok' | 'warning';

const STATUS_CONFIG: Record<
  StatusType,
  {
    color: string;
    icon: Component;
    text: string;
    cardClass: string;
    iconClass: string;
    progressStroke: string;
  }
> = {
  ok: {
    color: 'success',
    icon: CheckCircleOutlined,
    text: '正常',
    cardClass: 'border-green-200 bg-green-50/80',
    iconClass: 'text-green-500',
    progressStroke: '#16a34a',
  },
  warning: {
    color: 'warning',
    icon: ExclamationCircleOutlined,
    text: '警告',
    cardClass: 'border-yellow-200 bg-yellow-50/80',
    iconClass: 'text-yellow-500',
    progressStroke: '#d97706',
  },
  error: {
    color: 'error',
    icon: CloseCircleOutlined,
    text: '异常',
    cardClass: 'border-red-200 bg-red-50/80',
    iconClass: 'text-red-500',
    progressStroke: '#dc2626',
  },
};

const statusCounts = computed(() => {
  const counts: Record<StatusType, number> = { error: 0, ok: 0, warning: 0 };
  for (const item of items.value) {
    const status = (item.status || 'warning') as StatusType;
    if (counts[status] !== undefined) {
      counts[status] += 1;
    }
  }
  return counts;
});

const completionRate = computed(() => {
  if (items.value.length === 0) return 0;
  return Math.round((statusCounts.value.ok / items.value.length) * 100);
});

const nextActions = computed(() =>
  items.value.filter((item) => item.link && item.status !== 'ok').slice(0, 4),
);
</script>

<template>
  <Page auto-content-height>
    <div class="space-y-4 p-1 md:p-2">
      <Card :bordered="false" class="overflow-hidden">
        <div
          class="rounded-2xl bg-gradient-to-r from-slate-900 via-blue-900 to-cyan-700 p-6 text-white"
        >
          <div
            class="flex flex-col gap-5 lg:flex-row lg:items-center lg:justify-between"
          >
            <div class="max-w-2xl">
              <div class="mb-2 text-sm text-white/70">MVP 系统概览</div>
              <div class="text-2xl font-semibold md:text-3xl">
                {{ allPass ? '当前系统状态良好，可直接进入项目协作。' : '仍有配置项待完善，建议先处理风险项。' }}
              </div>
              <div class="mt-3 text-sm leading-6 text-white/75">
                这里会集中展示 MVP 运行所依赖的引擎、任务流与核心配置检查结果，方便你在创建项目或排障前快速确认系统状态。
              </div>
            </div>

            <div class="grid min-w-[280px] grid-cols-2 gap-3 md:grid-cols-4">
              <div class="rounded-2xl bg-white/10 p-4 backdrop-blur">
                <div class="text-xs text-white/60">检测项总数</div>
                <div class="mt-2 text-2xl font-semibold">{{ items.length }}</div>
              </div>
              <div class="rounded-2xl bg-white/10 p-4 backdrop-blur">
                <div class="text-xs text-white/60">通过率</div>
                <div class="mt-2 text-2xl font-semibold">{{ completionRate }}%</div>
              </div>
              <div class="rounded-2xl bg-white/10 p-4 backdrop-blur">
                <div class="text-xs text-white/60">待处理项</div>
                <div class="mt-2 text-2xl font-semibold">
                  {{ statusCounts.warning + statusCounts.error }}
                </div>
              </div>
              <div class="rounded-2xl bg-white/10 p-4 backdrop-blur">
                <div class="text-xs text-white/60">系统结论</div>
                <div class="mt-2 text-lg font-semibold">
                  {{ allPass ? '可运行' : '需处理' }}
                </div>
              </div>
            </div>
          </div>
        </div>
      </Card>

      <div class="grid grid-cols-1 gap-4 xl:grid-cols-3">
        <Card :bordered="false" class="xl:col-span-2">
          <template #title>健康检测</template>
          <template #extra>
            <Button :loading="loading" @click="fetchCheck">
              <template #icon><ReloadOutlined /></template>
              刷新检测
            </Button>
          </template>

          <Alert
            v-if="!loading && allPass"
            type="success"
            message="所有配置项检测通过，系统可正常运行"
            show-icon
            class="mb-4"
          />
          <Alert
            v-else-if="!loading && !allPass && items.length > 0"
            type="warning"
            message="部分配置项未就绪，请优先处理警告与异常项"
            show-icon
            class="mb-4"
          />

          <Spin :spinning="loading">
            <div class="grid grid-cols-1 gap-3 md:grid-cols-3">
              <div
                class="rounded-2xl border p-4"
                :class="STATUS_CONFIG.ok.cardClass"
              >
                <div class="text-xs text-slate-500">正常项</div>
                <div class="mt-2 text-3xl font-semibold text-slate-900">
                  {{ statusCounts.ok }}
                </div>
                <Progress
                  :percent="items.length ? Math.round((statusCounts.ok / items.length) * 100) : 0"
                  :show-info="false"
                  :stroke-color="STATUS_CONFIG.ok.progressStroke"
                  class="mt-3"
                />
              </div>
              <div
                class="rounded-2xl border p-4"
                :class="STATUS_CONFIG.warning.cardClass"
              >
                <div class="text-xs text-slate-500">警告项</div>
                <div class="mt-2 text-3xl font-semibold text-slate-900">
                  {{ statusCounts.warning }}
                </div>
                <Progress
                  :percent="items.length ? Math.round((statusCounts.warning / items.length) * 100) : 0"
                  :show-info="false"
                  :stroke-color="STATUS_CONFIG.warning.progressStroke"
                  class="mt-3"
                />
              </div>
              <div
                class="rounded-2xl border p-4"
                :class="STATUS_CONFIG.error.cardClass"
              >
                <div class="text-xs text-slate-500">异常项</div>
                <div class="mt-2 text-3xl font-semibold text-slate-900">
                  {{ statusCounts.error }}
                </div>
                <Progress
                  :percent="items.length ? Math.round((statusCounts.error / items.length) * 100) : 0"
                  :show-info="false"
                  :stroke-color="STATUS_CONFIG.error.progressStroke"
                  class="mt-3"
                />
              </div>
            </div>

            <div class="mt-4 space-y-3">
              <div
                v-for="item in items"
                :key="item.key"
                class="flex flex-col gap-3 rounded-2xl border p-4 transition-colors md:flex-row md:items-center"
                :class="STATUS_CONFIG[(item.status || 'warning') as StatusType].cardClass"
              >
                <div
                  class="flex h-11 w-11 items-center justify-center rounded-full bg-white shadow-sm"
                >
                  <component
                    :is="STATUS_CONFIG[(item.status || 'warning') as StatusType].icon"
                    class="text-lg"
                    :class="STATUS_CONFIG[(item.status || 'warning') as StatusType].iconClass"
                  />
                </div>
                <div class="min-w-0 flex-1">
                  <div class="flex flex-wrap items-center gap-2">
                    <span class="font-medium text-slate-900">{{ item.name }}</span>
                    <Tag
                      :color="STATUS_CONFIG[(item.status || 'warning') as StatusType].color"
                      size="small"
                    >
                      {{ STATUS_CONFIG[(item.status || 'warning') as StatusType].text }}
                    </Tag>
                  </div>
                  <div class="mt-1 text-sm text-slate-500">{{ item.message }}</div>
                </div>
                <Button
                  v-if="item.link"
                  type="primary"
                  ghost
                  @click="goToLink(item.link)"
                >
                  去配置
                </Button>
              </div>
            </div>
          </Spin>
        </Card>

        <div class="space-y-4">
          <Card :bordered="false">
            <template #title>
              <div class="flex items-center gap-2">
                <SettingOutlined />
                <span>推荐操作</span>
              </div>
            </template>

            <div v-if="nextActions.length > 0" class="space-y-3">
              <button
                v-for="item in nextActions"
                :key="item.key"
                class="flex w-full items-center justify-between rounded-2xl border border-slate-200 px-4 py-3 text-left transition hover:border-blue-300 hover:bg-blue-50"
                @click="goToLink(item.link!)"
              >
                <div class="min-w-0">
                  <div class="font-medium text-slate-900">{{ item.name }}</div>
                  <div class="mt-1 text-xs text-slate-500">{{ item.message }}</div>
                </div>
                <RightOutlined class="ml-3 flex-shrink-0 text-slate-400" />
              </button>
            </div>
            <Empty
              v-else
              description="当前没有待处理配置项"
              :image="Empty.PRESENTED_IMAGE_SIMPLE"
            />
          </Card>

          <Card :bordered="false" title="运行建议">
            <div class="space-y-3 text-sm text-slate-600">
              <div class="rounded-2xl bg-slate-50 p-4">
                <div class="font-medium text-slate-900">创建项目前</div>
                <div class="mt-1">
                  先确保执行引擎、模型和回调配置通过检测，避免任务进入执行阶段后中途失败。
                </div>
              </div>
              <div class="rounded-2xl bg-slate-50 p-4">
                <div class="font-medium text-slate-900">排障时</div>
                <div class="mt-1">
                  优先处理异常项，其次处理警告项；异常通常会直接阻塞项目工作流。
                </div>
              </div>
              <div class="rounded-2xl bg-slate-50 p-4">
                <div class="font-medium text-slate-900">建议目标</div>
                <div class="mt-1">
                  保持通过率在 100%，这样架构师拆解、调度执行和审核流都能稳定运行。
                </div>
              </div>
            </div>
          </Card>
        </div>
      </div>
    </div>
  </Page>
</template>

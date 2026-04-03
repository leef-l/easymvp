<script setup lang="ts">
import { h, ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { Page } from '@vben/common-ui';
import { Alert, Button, Card, Spin, Tag } from 'ant-design-vue';
import {
  CheckCircleOutlined,
  CloseCircleOutlined,
  ExclamationCircleOutlined,
  ReloadOutlined,
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

const STATUS_CONFIG: Record<
  string,
  { color: string; icon: any; text: string }
> = {
  ok: { color: 'success', icon: CheckCircleOutlined, text: '正常' },
  warning: { color: 'warning', icon: ExclamationCircleOutlined, text: '警告' },
  error: { color: 'error', icon: CloseCircleOutlined, text: '异常' },
};
</script>

<template>
  <Page auto-content-height>
    <Card title="配置检测" :bordered="false">
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
        message="部分配置项未就绪，请按提示完成配置后再创建项目"
        show-icon
        class="mb-4"
      />

      <Spin :spinning="loading">
        <div class="space-y-3">
          <div
            v-for="item in items"
            :key="item.key"
            class="flex items-center gap-3 rounded-lg border p-3 transition-colors"
            :class="{
              'border-green-200 bg-green-50': item.status === 'ok',
              'border-yellow-200 bg-yellow-50': item.status === 'warning',
              'border-red-200 bg-red-50': item.status === 'error',
            }"
          >
            <component
              :is="STATUS_CONFIG[item.status]?.icon"
              :class="{
                'text-green-500': item.status === 'ok',
                'text-yellow-500': item.status === 'warning',
                'text-red-500': item.status === 'error',
              }"
              class="flex-shrink-0 text-lg"
            />
            <div class="min-w-0 flex-1">
              <div class="flex items-center gap-2">
                <span class="font-medium">{{ item.name }}</span>
                <Tag :color="STATUS_CONFIG[item.status]?.color" size="small">
                  {{ STATUS_CONFIG[item.status]?.text }}
                </Tag>
              </div>
              <div class="mt-1 text-sm text-gray-500">{{ item.message }}</div>
            </div>
            <Button
              v-if="item.link"
              size="small"
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
  </Page>
</template>

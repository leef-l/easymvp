<script setup lang="ts">
import { onMounted, ref } from 'vue';

import { Alert, Card, Empty, Space, Spin, Table, Tag } from 'ant-design-vue';

import {
  getRegressionScenarios,
  type RegressionScenariosResult,
  type RegressionScenarioItem,
} from '#/api/mvp/workflow';

defineOptions({ name: 'WorkflowRegressionPanel' });

const loading = ref(false);
const manifest = ref<RegressionScenariosResult | null>(null);

const columns = [
  { title: '场景', dataIndex: 'name', key: 'name', width: 180 },
  { title: '编码', dataIndex: 'scenarioCode', key: 'scenarioCode', width: 180 },
  { title: '状态', dataIndex: 'status', key: 'status', width: 100 },
  { title: '目录', dataIndex: 'workspaceDir', key: 'workspaceDir', width: 220 },
  { title: '目标', dataIndex: 'goal', key: 'goal' },
];

async function loadData() {
  loading.value = true;
  try {
    manifest.value = await getRegressionScenarios();
  } catch {
    manifest.value = null;
  } finally {
    loading.value = false;
  }
}

function statusColor(status: string) {
  switch (status) {
    case 'ready':
      return 'green';
    case 'planned':
      return 'gold';
    default:
      return 'default';
  }
}

onMounted(loadData);
</script>

<template>
  <Spin :spinning="loading">
    <Empty v-if="!manifest || !(manifest.scenarios || []).length" description="暂无回归样例" />
    <template v-else>
      <Card size="small" class="mb-4">
        <Space wrap>
          <Tag color="blue">版本 {{ manifest.version }}</Tag>
          <Tag color="cyan">更新时间 {{ manifest.updatedAt }}</Tag>
          <Tag :color="manifest.valid ? 'green' : 'red'">
            {{ manifest.valid ? '校验通过' : '校验失败' }}
          </Tag>
          <Tag color="green">ready {{ manifest.readyCount || 0 }}</Tag>
          <Tag color="gold">planned {{ manifest.plannedCount || 0 }}</Tag>
          <Tag v-if="manifest.manifestPath" color="default">{{ manifest.manifestPath }}</Tag>
        </Space>
      </Card>

      <Alert
        class="mb-4"
        :message="manifest.message || '暂无校验摘要'"
        :type="manifest.valid ? 'success' : 'warning'"
        show-icon
      />

      <Table
        :columns="columns"
        :data-source="manifest.scenarios"
        :pagination="false"
        row-key="scenarioCode"
        size="small"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'status'">
            <Tag :color="statusColor((record as RegressionScenarioItem).status)">
              {{ (record as RegressionScenarioItem).status }}
            </Tag>
          </template>
          <template v-else-if="column.key === 'goal'">
            <div class="space-y-1">
              <div>{{ (record as RegressionScenarioItem).goal }}</div>
              <div v-if="(record as RegressionScenarioItem).checkpoints?.length" class="text-xs text-gray-500">
                观察点：{{ (record as RegressionScenarioItem).checkpoints.join('；') }}
              </div>
            </div>
          </template>
        </template>
      </Table>
    </template>
  </Spin>
</template>

<script setup lang="ts">
import { h, ref, watch } from 'vue';
import { useRoute } from 'vue-router';

import { Page } from '@vben/common-ui';

import { ReloadOutlined, SaveOutlined } from '@ant-design/icons-vue';
import { message } from 'ant-design-vue';
import {
  Alert,
  Button,
  Card,
  Col,
  Descriptions,
  DescriptionsItem,
  Form,
  FormItem,
  Input,
  InputNumber,
  Row,
  Select,
  SelectOption,
  Space,
  Spin,
} from 'ant-design-vue';

import {
  getObjective,
  type ObjectiveData,
  saveObjective,
} from '../../../api/mvp/workflow';

const route = useRoute();
const projectId = ref<string>('');
const loading = ref(false);
const saving = ref(false);
const isEdit = ref(false);

const defaultForm: Partial<ObjectiveData> = {
  deliveryGoal: '',
  qualityFloor: 0.8,
  tokenBudget: 0,
  timeBudgetHours: 0,
  costBudgetCents: 0,
  riskTolerance: 'medium',
  maxAutoRetries: 3,
  maxAutoReworks: 3,
  maxAutoReplans: 2,
  maxStallMinutes: 30,
  autonomyLevel: 'supervised',
  maxSideEffectLevel: 'reversible',
};

const form = ref<Partial<ObjectiveData>>({ ...defaultForm });

const riskOptions = [
  { label: '低（保守）', value: 'low' },
  { label: '中（均衡）', value: 'medium' },
  { label: '高（激进）', value: 'high' },
];

const autonomyOptions = [
  { label: '监督模式（所有关键动作需人工确认）', value: 'supervised' },
  { label: '辅助模式（低风险自动，高风险人工）', value: 'assisted' },
  { label: '全自动（所有动作自动执行）', value: 'full_auto' },
];

const sideEffectOptions = [
  { label: '无副作用', value: 'none' },
  { label: '可逆副作用', value: 'reversible' },
  { label: '不可逆副作用', value: 'irreversible' },
];

async function loadObjective() {
  if (!projectId.value) return;
  loading.value = true;
  try {
    const res = await getObjective(projectId.value);
    form.value = { ...defaultForm };
    if (res.objective) {
      form.value = { ...form.value, ...res.objective };
    }
  } finally {
    loading.value = false;
  }
}

async function handleSave() {
  if (!projectId.value) {
    message.error('缺少项目 ID');
    return;
  }
  saving.value = true;
  try {
    await saveObjective({ projectID: projectId.value, ...form.value });
    message.success('目标约束已保存');
    isEdit.value = false;
  } catch (error: any) {
    message.error(error?.message || '保存失败');
  } finally {
    saving.value = false;
  }
}

watch(
  () => route.query.projectId,
  async (value) => {
    projectId.value = (value as string) ?? '';
    form.value = { ...defaultForm };
    if (!projectId.value) return;
    await loadObjective();
  },
  { immediate: true },
);
</script>

<template>
  <Page title="目标层管理">
    <Space style="margin-bottom: 16px">
      <Button
        :icon="h(ReloadOutlined)"
        :loading="loading"
        @click="loadObjective"
      >
        重新加载
      </Button>
      <Button
        v-if="!isEdit"
        type="primary"
        @click="isEdit = true"
      >
        编辑
      </Button>
      <template v-if="isEdit">
        <Button
          type="primary"
          :icon="h(SaveOutlined)"
          :loading="saving"
          @click="handleSave"
        >
          保存
        </Button>
        <Button @click="isEdit = false; loadObjective()">取消</Button>
      </template>
    </Space>

    <Alert
      type="info"
      message="目标层约束用于驾驭自治系统的行为边界：Token 预算耗尽、时间超限、风险超标时，系统将自动降级或暂停并请求人工介入。"
      show-icon
      style="margin-bottom: 16px"
    />

    <Spin :spinning="loading">
      <Row :gutter="16">
        <!-- 交付目标 -->
        <Col :span="24">
          <Card title="交付目标" style="margin-bottom: 16px">
            <FormItem label="交付目标描述" :label-col="{ span: 4 }">
              <Input
                v-if="isEdit"
                v-model:value="form.deliveryGoal"
                placeholder="描述本次项目期望达到的交付目标"
                :maxlength="500"
                show-count
              />
              <span v-else>{{ form.deliveryGoal || '（未设置）' }}</span>
            </FormItem>
          </Card>
        </Col>

        <!-- 质量与预算约束 -->
        <Col :span="12">
          <Card title="质量与预算约束" style="margin-bottom: 16px">
            <Form :label-col="{ span: 10 }" :wrapper-col="{ span: 14 }">
              <FormItem label="质量下限（0-1）">
                <InputNumber
                  v-if="isEdit"
                  v-model:value="form.qualityFloor"
                  :min="0" :max="1" :step="0.05"
                  style="width: 100%"
                />
                <span v-else>{{ form.qualityFloor }}</span>
              </FormItem>

              <FormItem label="Token 预算">
                <InputNumber
                  v-if="isEdit"
                  v-model:value="form.tokenBudget"
                  :min="0" :step="10000"
                  style="width: 100%"
                  addon-after="tokens"
                />
                <span v-else>{{ form.tokenBudget || '不限' }}</span>
              </FormItem>

              <FormItem label="时间预算">
                <InputNumber
                  v-if="isEdit"
                  v-model:value="form.timeBudgetHours"
                  :min="0" :step="1"
                  style="width: 100%"
                  addon-after="小时"
                />
                <span v-else>{{ form.timeBudgetHours || '不限' }}</span>
              </FormItem>

              <FormItem label="成本预算">
                <InputNumber
                  v-if="isEdit"
                  v-model:value="form.costBudgetCents"
                  :min="0" :step="100"
                  style="width: 100%"
                  addon-after="分（人民币）"
                />
                <span v-else>{{ form.costBudgetCents ? `¥${(form.costBudgetCents / 100).toFixed(2)}` : '不限' }}</span>
              </FormItem>

              <FormItem label="截止日期">
                <Input
                  v-if="isEdit"
                  v-model:value="form.deadlineAt"
                  placeholder="格式：2026-06-01T18:00:00+08:00"
                />
                <span v-else>{{ form.deadlineAt || '不限' }}</span>
              </FormItem>
            </Form>
          </Card>
        </Col>

        <!-- 风险与自治约束 -->
        <Col :span="12">
          <Card title="风险与自治约束" style="margin-bottom: 16px">
            <Form :label-col="{ span: 10 }" :wrapper-col="{ span: 14 }">
              <FormItem label="风险容忍度">
                <Select
                  v-if="isEdit"
                  v-model:value="form.riskTolerance"
                  style="width: 100%"
                >
                  <SelectOption v-for="opt in riskOptions" :key="opt.value" :value="opt.value">
                    {{ opt.label }}
                  </SelectOption>
                </Select>
                <span v-else>{{ riskOptions.find(o => o.value === form.riskTolerance)?.label || form.riskTolerance }}</span>
              </FormItem>

              <FormItem label="自治级别">
                <Select
                  v-if="isEdit"
                  v-model:value="form.autonomyLevel"
                  style="width: 100%"
                >
                  <SelectOption v-for="opt in autonomyOptions" :key="opt.value" :value="opt.value">
                    {{ opt.label }}
                  </SelectOption>
                </Select>
                <span v-else>{{ autonomyOptions.find(o => o.value === form.autonomyLevel)?.label || form.autonomyLevel }}</span>
              </FormItem>

              <FormItem label="最大副作用级别">
                <Select
                  v-if="isEdit"
                  v-model:value="form.maxSideEffectLevel"
                  style="width: 100%"
                >
                  <SelectOption v-for="opt in sideEffectOptions" :key="opt.value" :value="opt.value">
                    {{ opt.label }}
                  </SelectOption>
                </Select>
                <span v-else>{{ sideEffectOptions.find(o => o.value === form.maxSideEffectLevel)?.label || form.maxSideEffectLevel }}</span>
              </FormItem>

              <FormItem label="最大停滞时间">
                <InputNumber
                  v-if="isEdit"
                  v-model:value="form.maxStallMinutes"
                  :min="1" :step="5"
                  style="width: 100%"
                  addon-after="分钟"
                />
                <span v-else>{{ form.maxStallMinutes }} 分钟</span>
              </FormItem>
            </Form>
          </Card>
        </Col>

        <!-- 自动行为上限 -->
        <Col :span="24">
          <Card title="自动行为上限">
            <Descriptions :column="3" bordered size="small">
              <DescriptionsItem label="最大自动重试次数">
                <InputNumber
                  v-if="isEdit"
                  v-model:value="form.maxAutoRetries"
                  :min="0" :max="10"
                  style="width: 80px"
                />
                <span v-else>{{ form.maxAutoRetries }} 次</span>
              </DescriptionsItem>
              <DescriptionsItem label="最大自动返工轮次">
                <InputNumber
                  v-if="isEdit"
                  v-model:value="form.maxAutoReworks"
                  :min="0" :max="10"
                  style="width: 80px"
                />
                <span v-else>{{ form.maxAutoReworks }} 轮</span>
              </DescriptionsItem>
              <DescriptionsItem label="最大自动重规划次数">
                <InputNumber
                  v-if="isEdit"
                  v-model:value="form.maxAutoReplans"
                  :min="0" :max="5"
                  style="width: 80px"
                />
                <span v-else>{{ form.maxAutoReplans }} 次</span>
              </DescriptionsItem>
            </Descriptions>
          </Card>
        </Col>
      </Row>
    </Spin>
  </Page>
</template>

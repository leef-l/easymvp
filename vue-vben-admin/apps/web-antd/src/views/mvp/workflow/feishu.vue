<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue';

import { Page } from '@vben/common-ui';
import {
  Alert,
  Button,
  Card,
  Form,
  FormItem,
  Input,
  Modal,
  Select,
  Space,
  Switch,
  Table,
  Tabs,
  Tag,
  message,
} from 'ant-design-vue';

import {
  bindFeishuUser,
  getFeishuBindings,
  getFeishuConfig,
  saveFeishuConfig,
  testFeishuMessage,
  unbindFeishuUser,
  type FeishuBindingItem,
  type FeishuConfigItem,
} from '#/api/mvp/workflow';

defineOptions({ name: 'MvpWorkflowFeishu' });

const loading = ref(false);
const bindingsLoading = ref(false);
const saving = ref(false);
const bindingSubmitting = ref(false);
const testing = ref(false);

const config = reactive<FeishuConfigItem>({
  enabled: 0,
  appId: '',
  appSecret: '',
  verificationToken: '',
  encryptKey: '',
  defaultNotifyUserIds: '',
  connectionMode: 'webhook',
  callbackPath: '/api/mvp/collab/feishu/callback',
  wsRunning: false,
});

const bindForm = reactive({
  userId: '',
  platformUserId: '',
  platformName: '',
});

const bindings = ref<FeishuBindingItem[]>([]);

const enabledChecked = computed({
  get: () => config.enabled === 1,
  set: (checked: boolean) => {
    config.enabled = checked ? 1 : 0;
  },
});

const callbackUrl = computed(() => {
  if (typeof window === 'undefined') {
    return config.callbackPath;
  }
  return `${window.location.origin}${config.callbackPath}`;
});

async function loadConfig() {
  loading.value = true;
  try {
    const res = await getFeishuConfig();
    Object.assign(config, res.config ?? {});
  } finally {
    loading.value = false;
  }
}

async function loadBindings() {
  bindingsLoading.value = true;
  try {
    const res = await getFeishuBindings();
    bindings.value = res.bindings ?? [];
  } finally {
    bindingsLoading.value = false;
  }
}

async function saveConfigForm() {
  saving.value = true;
  try {
    await saveFeishuConfig({ ...config });
    message.success('飞书配置已保存');
    await loadConfig();
  } finally {
    saving.value = false;
  }
}

async function submitBinding() {
  if (!bindForm.userId || !bindForm.platformUserId) {
    message.warning('请先填写系统用户ID和飞书 open_id');
    return;
  }
  bindingSubmitting.value = true;
  try {
    await bindFeishuUser({
      userId: bindForm.userId,
      platformUserId: bindForm.platformUserId,
      platformName: bindForm.platformName || undefined,
    });
    message.success('绑定成功');
    bindForm.userId = '';
    bindForm.platformUserId = '';
    bindForm.platformName = '';
    await loadBindings();
  } finally {
    bindingSubmitting.value = false;
  }
}

function confirmUnbind(row: FeishuBindingItem) {
  Modal.confirm({
    title: '确认解绑',
    content: `确认解绑系统用户 ${row.userId} 与飞书用户 ${row.platformUserId} 吗？`,
    okType: 'danger',
    async onOk() {
      await unbindFeishuUser(row.id);
      message.success('解绑成功');
      await loadBindings();
    },
  });
}

async function sendTest(row: FeishuBindingItem) {
  await testFeishuMessage({
    bindingId: row.id,
    content: `EasyMVP 飞书测试消息：系统用户 ${row.userId} 绑定正常。`,
  });
  message.success('测试消息已发送');
}

async function testConnection() {
  if (bindings.value.length === 0) {
    message.warning('请先绑定飞书用户再测试');
    return;
  }
  testing.value = true;
  try {
    const first = bindings.value[0];
    await testFeishuMessage({
      bindingId: first.id,
      content: `EasyMVP 连通测试：飞书配置正常，系统用户 ${first.userId} 绑定有效。`,
    });
    message.success('测试消息已发送，请在飞书查收');
  } finally {
    testing.value = false;
  }
}

async function copyCallbackUrl() {
  await navigator.clipboard.writeText(callbackUrl.value);
  message.success('回调地址已复制');
}

onMounted(async () => {
  await Promise.all([loadConfig(), loadBindings()]);
});

const columns = [
  { title: '系统用户ID', dataIndex: 'userId', key: 'userId', width: 140 },
  { title: '飞书 OpenID', dataIndex: 'platformUserId', key: 'platformUserId', ellipsis: true },
  { title: '飞书名称', dataIndex: 'platformName', key: 'platformName', width: 180 },
  { title: '创建人', dataIndex: 'createdBy', key: 'createdBy', width: 140 },
  { title: '部门', dataIndex: 'deptId', key: 'deptId', width: 120 },
  { title: '更新时间', dataIndex: 'updatedAt', key: 'updatedAt', width: 180 },
  { title: '操作', key: 'action', width: 220, fixed: 'right' as const },
];
</script>

<template>
  <Page auto-content-height>
    <div class="space-y-4">
      <Alert
        type="info"
        show-icon
        message="飞书一期管理面"
        description="这里用于保存飞书应用配置、维护系统用户与飞书 open_id 的映射，并直接做联通测试。审批回调仍走 EasyMVP 原服务链。"
      />

      <Tabs>
        <Tabs.TabPane key="config" tab="配置">
          <Card :loading="loading" title="飞书应用配置">
            <Form layout="vertical">
              <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
                <FormItem label="启用飞书通知">
                  <Switch v-model:checked="enabledChecked" />
                </FormItem>
                <FormItem label="默认通知系统用户ID">
                  <Input
                    v-model:value="config.defaultNotifyUserIds"
                    placeholder="多个用户ID用英文逗号分隔"
                  />
                </FormItem>
                <FormItem label="App ID">
                  <Input v-model:value="config.appId" placeholder="cli_xxx" />
                </FormItem>
                <FormItem label="App Secret">
                  <Input.Password v-model:value="config.appSecret" placeholder="飞书 App Secret" />
                </FormItem>
                <FormItem label="Verification Token">
                  <Input
                    v-model:value="config.verificationToken"
                    placeholder="飞书 Verification Token"
                  />
                </FormItem>
                <FormItem label="Encrypt Key">
                  <Input.Password
                    v-model:value="config.encryptKey"
                    placeholder="飞书 Encrypt Key"
                  />
                </FormItem>
              </div>

              <FormItem label="连接模式">
                <Select
                  v-model:value="config.connectionMode"
                  style="width: 240px"
                  :options="[
                    { label: '回调模式（Webhook）', value: 'webhook' },
                    { label: '长连接模式（WebSocket）', value: 'websocket' },
                  ]"
                />
                <div style="color: #999; font-size: 12px; margin-top: 4px">
                  Webhook：飞书主动推送事件到你的服务（需公网）；WebSocket：服务主动建立长连接接收事件（无需公网）
                </div>
              </FormItem>

              <FormItem label="回调地址">
                <Space class="w-full">
                  <Input :value="callbackUrl" readonly />
                  <Button @click="copyCallbackUrl">复制</Button>
                </Space>
              </FormItem>

              <div class="flex justify-end">
                <Space>
                  <Button @click="loadConfig">刷新</Button>
                  <Button :loading="testing" @click="testConnection">测试连通</Button>
                  <Button type="primary" :loading="saving" @click="saveConfigForm">
                    保存配置
                  </Button>
                </Space>
              </div>
            </Form>
          </Card>
        </Tabs.TabPane>

        <Tabs.TabPane key="binding" tab="绑定">
          <Card title="飞书用户绑定">
            <div class="mb-4 rounded-xl border border-dashed border-slate-200 bg-slate-50 p-4">
              <Form layout="vertical">
                <div class="grid grid-cols-1 gap-4 md:grid-cols-3">
                  <FormItem label="系统用户ID" required>
                    <Input v-model:value="bindForm.userId" placeholder="请输入系统用户ID" />
                  </FormItem>
                  <FormItem label="飞书 OpenID" required>
                    <Input
                      v-model:value="bindForm.platformUserId"
                      placeholder="请输入飞书 open_id"
                    />
                  </FormItem>
                  <FormItem label="飞书显示名">
                    <Input v-model:value="bindForm.platformName" placeholder="可选" />
                  </FormItem>
                </div>
                <div class="flex justify-end">
                  <Button type="primary" :loading="bindingSubmitting" @click="submitBinding">
                    新增 / 重绑
                  </Button>
                </div>
              </Form>
            </div>

            <Table
              row-key="id"
              :columns="columns"
              :data-source="bindings"
              :loading="bindingsLoading"
              :pagination="{ pageSize: 10 }"
              :scroll="{ x: 1100 }"
            >
              <template #bodyCell="{ column, record }">
                <template v-if="column.key === 'platformName'">
                  <span>{{ record.platformName || '-' }}</span>
                </template>
                <template v-else-if="column.key === 'updatedAt'">
                  <span>{{ record.updatedAt || '-' }}</span>
                </template>
                <template v-else-if="column.key === 'action'">
                  <Space>
                    <Button size="small" @click="sendTest(record)">测试消息</Button>
                    <Button size="small" danger @click="confirmUnbind(record)">解绑</Button>
                  </Space>
                </template>
              </template>
            </Table>
          </Card>
        </Tabs.TabPane>

        <Tabs.TabPane key="status" tab="状态">
          <Card title="当前接入状态">
            <div class="space-y-3">
              <div class="flex items-center justify-between rounded-lg border p-3">
                <span>飞书总开关</span>
                <Tag :color="config.enabled === 1 ? 'green' : 'default'">
                  {{ config.enabled === 1 ? '已开启' : '未开启' }}
                </Tag>
              </div>
              <div class="flex items-center justify-between rounded-lg border p-3">
                <span>App 凭证</span>
                <Tag :color="config.appId && config.appSecret ? 'green' : 'orange'">
                  {{ config.appId && config.appSecret ? '已配置' : '待补齐' }}
                </Tag>
              </div>
              <div class="flex items-center justify-between rounded-lg border p-3">
                <span>回调签名参数</span>
                <Tag :color="config.encryptKey ? 'green' : 'orange'">
                  {{ config.encryptKey ? '已配置' : '待补齐' }}
                </Tag>
              </div>
              <div class="flex items-center justify-between rounded-lg border p-3">
                <span>连接模式</span>
                <Tag :color="config.connectionMode === 'websocket' ? 'purple' : 'blue'">
                  {{ config.connectionMode === 'websocket' ? 'WebSocket 长连接' : 'Webhook 回调' }}
                </Tag>
              </div>
              <div
                v-if="config.connectionMode === 'websocket'"
                class="flex items-center justify-between rounded-lg border p-3"
              >
                <span>WebSocket 长连接状态</span>
                <Tag :color="config.wsRunning ? 'green' : 'orange'">
                  {{ config.wsRunning ? '在线' : '未连接' }}
                </Tag>
              </div>
              <div class="flex items-center justify-between rounded-lg border p-3">
                <span>已绑定用户数</span>
                <Tag color="blue">{{ bindings.length }}</Tag>
              </div>
            </div>
          </Card>
        </Tabs.TabPane>
      </Tabs>
    </div>
  </Page>
</template>

<script setup lang="ts">
// @ts-nocheck
import { onMounted, reactive, ref } from 'vue';

import { Page } from '@vben/common-ui';

import {
  Alert,
  Button,
  Card,
  Form,
  FormItem,
  Input,
  message,
  Modal,
  Space,
  Switch,
  Table,
  Tabs,
  Tag,
} from 'ant-design-vue';

import {
  bindTelegramUser,
  type FeishuBindingItem,
  getTelegramBindings,
  getTelegramConfig,
  saveTelegramConfig,
  setTelegramCommands,
  type TelegramCommandItem,
  type TelegramConfigItem,
  testTelegramMessage,
  unbindTelegramUser,
} from '#/api/mvp/workflow';

defineOptions({ name: 'MvpWorkflowTelegram' });

const loading = ref(false);
const saving = ref(false);
const bindingsLoading = ref(false);
const bindingSubmitting = ref(false);
const testing = ref(false);
const commandSaving = ref(false);

const config = reactive<TelegramConfigItem>({
  enabled: 0,
  botToken: '',
  botRunning: false,
});

const bindForm = reactive({
  userId: '',
  platformUserId: '',
  platformName: '',
});

const bindings = ref<FeishuBindingItem[]>([]);

// 命令菜单
const commands = ref<TelegramCommandItem[]>([]);
const commandTabLoaded = ref(false);

const defaultCommands: TelegramCommandItem[] = [
  { command: 'start', description: '开始使用 / 帮助' },
  { command: 'help', description: '查看所有功能' },
  { command: 'list', description: '我的项目列表' },
  { command: 'quit', description: '退出对话模式' },
];

async function loadConfig() {
  loading.value = true;
  try {
    const res = await getTelegramConfig();
    Object.assign(config, res.config ?? {});
  } finally {
    loading.value = false;
  }
}

async function loadBindings() {
  bindingsLoading.value = true;
  try {
    const res = await getTelegramBindings();
    bindings.value = res.bindings ?? [];
  } finally {
    bindingsLoading.value = false;
  }
}

async function saveConfigForm() {
  saving.value = true;
  try {
    await saveTelegramConfig({ enabled: config.enabled, botToken: config.botToken });
    message.success('Telegram 配置已保存');
    await loadConfig();
  } finally {
    saving.value = false;
  }
}

async function submitBinding() {
  if (!bindForm.userId || !bindForm.platformUserId) {
    message.warning('请填写系统用户ID和 Telegram chat_id');
    return;
  }
  bindingSubmitting.value = true;
  try {
    await bindTelegramUser({
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
    content: `确认解绑系统用户 ${row.userId} 与 Telegram 用户 ${row.platformUserId} 吗？`,
    okType: 'danger',
    async onOk() {
      await unbindTelegramUser(row.id);
      message.success('解绑成功');
      await loadBindings();
    },
  });
}

async function sendTest(row: FeishuBindingItem) {
  testing.value = true;
  try {
    await testTelegramMessage({
      bindingId: row.id,
      content: `EasyMVP Telegram 联通测试成功 🎉 系统用户 ${row.userId} 绑定有效。`,
    });
    message.success('测试消息已发送，请在 Telegram 查收');
  } catch (error: any) {
    message.error(error?.message || '发送失败');
  } finally {
    testing.value = false;
  }
}

async function deployCommands(useDefault = false) {
  commandSaving.value = true;
  try {
    const res = await setTelegramCommands(
      useDefault ? { useDefault: true } : { commands: commands.value },
    );
    message.success(res.message || '命令菜单已更新');
    commands.value = res.commands ?? defaultCommands;
  } catch (error: any) {
    message.error(error?.message || '设置失败');
  } finally {
    commandSaving.value = false;
  }
}

function addCommand() {
  commands.value.push({ command: '', description: '' });
}

function removeCommand(idx: number) {
  commands.value.splice(idx, 1);
}

function onTabChange(key: number | string) {
  if (String(key) === 'commands' && !commandTabLoaded.value) {
    commandTabLoaded.value = true;
    commands.value = [...defaultCommands];
  }
}

function handleEnabledChange(checked: boolean | number | string) {
  config.enabled = checked === true || checked === 1 || checked === '1' ? 1 : 0;
}

onMounted(async () => {
  await Promise.all([loadConfig(), loadBindings()]);
});

const columns = [
  { title: '系统用户ID', dataIndex: 'userId', key: 'userId', width: 140 },
  { title: 'Telegram chat_id', dataIndex: 'platformUserId', key: 'platformUserId', ellipsis: true },
  { title: 'Telegram 用户名', dataIndex: 'platformName', key: 'platformName', width: 180 },
  { title: '创建人', dataIndex: 'createdBy', key: 'createdBy', width: 140 },
  { title: '更新时间', dataIndex: 'updatedAt', key: 'updatedAt', width: 180 },
  { title: '操作', key: 'action', width: 200, fixed: 'right' as const },
];
</script>

<template>
  <Page auto-content-height>
    <Tabs default-active-key="config" @change="onTabChange">
      <!-- ── Bot 配置 ─────────────────────────────── -->
      <Tabs.TabPane key="config" tab="Bot 配置">
        <Card title="Telegram Bot 配置" :loading="loading">
          <Alert
            class="mb-4"
            type="info"
            show-icon
            message="如何获取 Bot Token"
            description="在 Telegram 搜索 @BotFather → 发送 /newbot → 按提示输入名称和用户名 → 获得 Token。"
          />

          <Form layout="vertical">
            <FormItem label="启用 Telegram Bot">
              <Switch
                :checked="config.enabled === 1"
                @change="handleEnabledChange"
              />
              <Tag v-if="config.botRunning" color="green" class="ml-3">Polling 运行中</Tag>
              <Tag v-else color="default" class="ml-3">未运行</Tag>
            </FormItem>
            <FormItem label="Bot Token">
              <Input
                v-model:value="config.botToken"
                placeholder="输入新 Token（留空则不更新）"
                allow-clear
              />
              <div class="text-gray-400 text-xs mt-1">格式：123456789:AAHxxxxxxxxxxxxxxxxxxxx</div>
            </FormItem>

            <FormItem>
              <Button type="primary" :loading="saving" @click="saveConfigForm">保存配置</Button>
            </FormItem>
          </Form>
        </Card>
      </Tabs.TabPane>

      <!-- ── 用户绑定 ─────────────────────────────── -->
      <Tabs.TabPane key="bindings" tab="用户绑定">
        <Card title="绑定 Telegram 用户" class="mb-4">
          <Alert
            class="mb-4"
            type="info"
            show-icon
            message="如何获取 chat_id"
            description="在 Telegram 向 Bot 发任意消息，然后访问 https://api.telegram.org/bot<Token>/getUpdates，在返回的 JSON 中找到 message.chat.id 字段即为 chat_id。"
          />
          <Form layout="inline">
            <FormItem label="系统用户ID">
              <Input v-model:value="bindForm.userId" placeholder="如：1000000000000000001" style="width: 200px" />
            </FormItem>
            <FormItem label="Telegram chat_id">
              <Input v-model:value="bindForm.platformUserId" placeholder="如：123456789" style="width: 160px" />
            </FormItem>
            <FormItem label="用户名（选填）">
              <Input v-model:value="bindForm.platformName" placeholder="@username" style="width: 140px" />
            </FormItem>
            <FormItem>
              <Button type="primary" :loading="bindingSubmitting" @click="submitBinding">绑定</Button>
            </FormItem>
          </Form>
        </Card>

        <Card title="已绑定用户" :loading="bindingsLoading">
          <Table
            :data-source="bindings"
            :columns="columns"
            row-key="id"
            size="small"
            :scroll="{ x: 900 }"
            :pagination="false"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'action'">
                <Space>
                  <Button size="small" :loading="testing" @click="sendTest(record as FeishuBindingItem)">
                    测试消息
                  </Button>
                  <Button size="small" danger @click="confirmUnbind(record as FeishuBindingItem)">
                    解绑
                  </Button>
                </Space>
              </template>
            </template>
          </Table>
        </Card>
      </Tabs.TabPane>

      <!-- ── 命令菜单 ─────────────────────────────── -->
      <Tabs.TabPane key="commands" tab="命令菜单">
        <Card title="Bot 命令菜单（/command）">
          <Alert
            class="mb-4"
            type="info"
            show-icon
            message="说明"
            description="Telegram Bot 命令菜单显示在输入框左侧的菜单按钮里，用户点击即可快速发送命令。点击「推送到 Telegram」后立即生效。"
          />

          <div v-for="(cmd, idx) in commands" :key="idx" class="flex gap-2 mb-2 items-center">
            <Input
              v-model:value="cmd.command"
              placeholder="命令（不含/，如 help）"
              style="width: 160px"
              addon-before="/"
            />
            <Input
              v-model:value="cmd.description"
              placeholder="描述（如：查看所有功能）"
              style="width: 260px"
            />
            <Button size="small" danger @click="removeCommand(idx)">删除</Button>
          </div>

          <div class="mt-3">
            <Space>
              <Button @click="addCommand">+ 添加命令</Button>
              <Button type="primary" :loading="commandSaving" @click="deployCommands(false)">
                推送到 Telegram
              </Button>
              <Button :loading="commandSaving" @click="deployCommands(true)">恢复默认</Button>
            </Space>
          </div>

          <div class="mt-4 text-gray-400 text-sm">
            默认命令：/start /help /list /quit
          </div>
        </Card>
      </Tabs.TabPane>
    </Tabs>
  </Page>
</template>

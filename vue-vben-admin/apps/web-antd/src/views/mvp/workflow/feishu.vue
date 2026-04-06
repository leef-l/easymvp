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
  createChatMenu,
  deleteChatMenu,
  getBotMenu,
  getChatMenu,
  getFeishuBindings,
  getFeishuConfig,
  saveFeishuConfig,
  setBotMenu,
  testFeishuMessage,
  unbindFeishuUser,
  type BotMenuItem,
  type ChatMenuItem,
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

// ─── 机器人菜单 ────────────────────────────────────────────────────────────────
const menuLoading = ref(false);
const menuSaving = ref(false);
const menuIsDefault = ref(true);
const menuItems = ref<BotMenuItem[]>([]);
const defaultMenuItems = ref<BotMenuItem[]>([]);

async function loadBotMenu() {
  menuLoading.value = true;
  try {
    const res = await getBotMenu();
    menuItems.value = JSON.parse(JSON.stringify(res.menuItems ?? []));
    menuIsDefault.value = res.isDefault ?? true;
    defaultMenuItems.value = res.defaultItems ?? [];
  } finally {
    menuLoading.value = false;
  }
}

async function deployMenu(useDefault = false) {
  menuSaving.value = true;
  try {
    const payload = useDefault
      ? { useDefault: true }
      : { menuItems: menuItems.value };
    const res = await setBotMenu(payload);
    message.success(res.message || '菜单已推送到飞书');
    await loadBotMenu();
  } finally {
    menuSaving.value = false;
  }
}

function addTopMenu() {
  if (menuItems.value.length >= 3) {
    message.warning('飞书机器人菜单最多支持 3 个一级菜单');
    return;
  }
  menuItems.value.push({ eventKey: `PARENT_${Date.now()}`, name: '新菜单', children: [] });
}

function addSubMenu(parentIdx: number) {
  const parent = menuItems.value[parentIdx];
  if (!parent.children) parent.children = [];
  if (parent.children.length >= 5) {
    message.warning('每个一级菜单最多支持 5 个子菜单');
    return;
  }
  parent.children.push({ eventKey: `item_${Date.now()}`, name: '新菜单项' });
}

function removeTopMenu(idx: number) {
  menuItems.value.splice(idx, 1);
}

function removeSubMenu(parentIdx: number, subIdx: number) {
  menuItems.value[parentIdx].children?.splice(subIdx, 1);
}

// 在 onMounted 中加载菜单（懒加载：切换到菜单 tab 时再加载）
const menuTabLoaded = ref(false);

// ─── 群菜单 ────────────────────────────────────────────────────────────────────
const chatMenuLoading = ref(false);
const chatMenuSaving = ref(false);
const chatMenuDeleting = ref(false);
const chatID = ref('');
const baseURL = ref(window.location.origin);
const chatMenuList = ref<any[]>([]);

const defaultChatMenuPreview: ChatMenuItem[] = [
  {
    name: '项目管理',
    children: [
      { name: '项目列表', url: '{baseURL}/mvp/project/index' },
      { name: '新建项目', url: '{baseURL}/mvp/workflow/create' },
      { name: '项目仪表盘', url: '{baseURL}/mvp/workflow/dashboard' },
    ],
  },
  {
    name: '任务管理',
    children: [
      { name: '任务列表', url: '{baseURL}/mvp/task/index' },
      { name: '工作流状态', url: '{baseURL}/mvp/workflow/situation' },
    ],
  },
  {
    name: '系统设置',
    children: [
      { name: '飞书配置', url: '{baseURL}/mvp/workflow/feishu' },
      { name: 'AI 配置', url: '{baseURL}/ai/model/index' },
    ],
  },
];

async function loadChatMenu() {
  if (!chatID.value) {
    message.warning('请先输入群 chat_id');
    return;
  }
  chatMenuLoading.value = true;
  try {
    const res = await getChatMenu(chatID.value);
    chatMenuList.value = res.menuItems ?? [];
  } catch (e: any) {
    message.error(e?.message || '获取群菜单失败');
  } finally {
    chatMenuLoading.value = false;
  }
}

async function handleCreateChatMenu() {
  if (!chatID.value) {
    message.warning('请先输入群 chat_id');
    return;
  }
  chatMenuSaving.value = true;
  try {
    const res = await createChatMenu({ chatId: chatID.value });
    message.success(res.message || '群菜单创建成功');
    await loadChatMenu();
  } catch (e: any) {
    message.error(e?.message || '创建群菜单失败');
  } finally {
    chatMenuSaving.value = false;
  }
}

async function handleDeleteAllChatMenu() {
  if (!chatID.value || chatMenuList.value.length === 0) return;
  chatMenuDeleting.value = true;
  try {
    const ids = chatMenuList.value.map((m: any) => m.chat_menu_top_level_id).filter(Boolean);
    await deleteChatMenu({ chatId: chatID.value, menuIds: ids });
    message.success('群菜单已清空');
    chatMenuList.value = [];
  } catch (e: any) {
    message.error(e?.message || '删除群菜单失败');
  } finally {
    chatMenuDeleting.value = false;
  }
}

function onTabChange(key: string) {
  if (key === 'menu' && !menuTabLoaded.value) {
    menuTabLoaded.value = true;
    loadBotMenu();
  }
}

// EventKey 说明文档
const menuEventKeyDocs = [
  { key: 'list_projects', desc: '列出我的项目列表' },
  { key: 'create_project_tip', desc: '引导用户创建新项目（Bot 会提示需要哪些信息）' },
  { key: 'help', desc: '显示帮助信息（所有可用指令）' },
  { key: 'project_status_tip', desc: '查询项目执行进度（Bot 会进一步询问项目名称）' },
  { key: 'list_tasks_tip', desc: '查看项目任务列表（Bot 会进一步询问项目名称）' },
  { key: 'retry_task_tip', desc: '重试项目失败任务（Bot 会进一步询问项目名称）' },
  { key: 'review_status_tip', desc: '查看人工审核状态（Bot 会进一步询问项目名称）' },
  { key: 'accept_status_tip', desc: '查看验收状态（Bot 会进一步询问项目名称）' },
  { key: 'autonomy_status_tip', desc: '查看自治检查点状态（Bot 会进一步询问项目名称）' },
  { key: '自定义key', desc: 'Bot 将 key 文本作为指令发送给 AI 处理（AI 会理解并响应）' },
];
</script>

<template>
  <Page auto-content-height>
    <div class="space-y-4">
      <Alert
        type="info"
        show-icon
        message="EasyMVP 飞书 Bot — 在飞书里完成项目管理全流程"
        description="接入后可在飞书单聊/群聊中：创建项目、与架构师AI对话、查看执行进度、处理审核/验收/自治检查点，并自动接收任务失败、项目完成等关键通知。查看「使用说明」标签了解详情。"
      />

      <Tabs @change="onTabChange">
        <Tabs.TabPane key="guide" tab="使用说明">
          <div class="space-y-4">
            <!-- 能力介绍 -->
            <Card title="🤖 飞书 Bot 能做什么">
              <div class="space-y-3 text-sm text-gray-700">
                <div class="rounded-lg bg-blue-50 p-3">
                  <div class="mb-2 font-semibold text-blue-700">📁 项目全生命周期管理</div>
                  <div class="space-y-1 text-gray-600">
                    <div>• 直接说需求创建项目，支持：软件开发、游戏开发、数据分析、内容创作、运营策划</div>
                    <div>• 查看项目列表、执行进度、暂停/恢复项目</div>
                    <div>• 说「确认方案」启动自动执行流水线</div>
                  </div>
                </div>
                <div class="rounded-lg bg-green-50 p-3">
                  <div class="mb-2 font-semibold text-green-700">💬 与 AI 角色直接对话</div>
                  <div class="space-y-1 text-gray-600">
                    <div>• 创建项目后，直接在飞书里和<strong>架构师 AI</strong> 对话，描述需求、拆解任务</div>
                    <div>• AI 回复完成后自动推送到飞书，无需刷新后台</div>
                    <div>• 说「退出对话」结束当前对话上下文</div>
                  </div>
                </div>
                <div class="rounded-lg bg-orange-50 p-3">
                  <div class="mb-2 font-semibold text-orange-700">📋 任务管理</div>
                  <div class="space-y-1 text-gray-600">
                    <div>• 查看项目任务列表和状态</div>
                    <div>• 重试所有失败任务（或指定单个任务 ID）</div>
                    <div>• 跳过阻塞任务解除卡点</div>
                  </div>
                </div>
                <div class="rounded-lg bg-purple-50 p-3">
                  <div class="mb-2 font-semibold text-purple-700">🔔 主动推送通知（无需查后台）</div>
                  <div class="space-y-1 text-gray-600">
                    <div>• ✅ 项目全部执行完成 → 飞书通知</div>
                    <div>• ❌ 任务失败 → 飞书通知（含失败原因）</div>
                    <div>• 🔍 需要人工审核 → 飞书通知（可直接回复审核指令）</div>
                    <div>• 🎯 需要人工验收 → 飞书通知（可直接回复验收指令）</div>
                    <div>• 🤖 自治检查点需确认 → 飞书通知（可直接回复批准/拒绝）</div>
                  </div>
                </div>
                <div class="rounded-lg bg-gray-50 p-3">
                  <div class="mb-2 font-semibold text-gray-700">📌 使用方式</div>
                  <div class="space-y-1 text-gray-600">
                    <div>• <strong>单聊</strong>：直接给 Bot 发消息，无需 @</div>
                    <div>• <strong>群聊</strong>：@EasyMVP 后跟随指令</div>
                    <div>• <strong>自然语言</strong>：不用记命令，直接描述意图，AI 自动理解</div>
                  </div>
                </div>
              </div>
            </Card>

            <!-- 事件订阅说明 -->
            <Card title="📡 飞书开放平台事件订阅配置">
              <Alert
                type="warning"
                show-icon
                message="必须在飞书开放平台配置事件订阅，Bot 才能收到消息"
                class="mb-4"
              />
              <div class="space-y-3">
                <div class="rounded-lg border border-red-200 bg-red-50 p-3">
                  <div class="mb-2 font-semibold text-red-700">🔴 必须订阅的事件</div>
                  <div class="space-y-2">
                    <div class="flex items-start gap-3">
                      <Tag color="red" class="mt-0.5 shrink-0">必须</Tag>
                      <div>
                        <div class="font-mono font-semibold">im.message.receive_v1</div>
                        <div class="mt-0.5 text-xs text-gray-500">接收用户发给 Bot 的消息（单聊 + 群聊 @Bot），所有飞书功能的基础</div>
                      </div>
                    </div>
                    <div class="flex items-start gap-3">
                      <Tag color="orange" class="mt-0.5 shrink-0">菜单需要</Tag>
                      <div>
                        <div class="font-mono font-semibold">application.bot.menu_v6</div>
                        <div class="mt-0.5 text-xs text-gray-500">接收用户点击机器人菜单的事件，启用「机器人菜单」功能时必须订阅</div>
                      </div>
                    </div>
                  </div>
                </div>
                <div class="rounded-lg border border-gray-200 bg-gray-50 p-3">
                  <div class="mb-2 font-semibold text-gray-700">⚪ 不需要订阅的事件</div>
                  <div class="space-y-1 text-xs text-gray-500">
                    <div>• 审批事件（approval.*）：EasyMVP 有自己的审核流，不走飞书审批</div>
                    <div>• 机器人进出群（im.chat.member.*）：当前版本不需要</div>
                    <div>• 消息已读（im.message.message_read_v1）：可选，不影响核心功能</div>
                  </div>
                </div>
              </div>

              <div class="mt-4 space-y-2 text-sm">
                <div class="font-semibold text-gray-700">配置步骤</div>
                <div class="rounded bg-gray-100 p-3 font-mono text-xs leading-6 text-gray-700">
                  1. 飞书开发者后台 → 选择你的应用<br>
                  2. 左侧「事件订阅」<br>
                  3. 添加事件：<br>
                  &nbsp;&nbsp;&nbsp;→ 搜索 im.message.receive_v1 → 添加（必须）<br>
                  &nbsp;&nbsp;&nbsp;→ 搜索 application.bot.menu_v6 → 添加（使用菜单功能时必须）<br>
                  4. Webhook 模式：填写「请求地址」= 本页「回调地址」中的 URL<br>
                  &nbsp;&nbsp;&nbsp;WebSocket 模式：选择「使用长连接接收事件」，无需填 URL
                </div>
              </div>
            </Card>

            <!-- 权限配置说明 -->
            <Card title="🔑 飞书应用权限配置">
              <div class="space-y-2 text-sm">
                <div class="text-gray-500">在飞书开发者后台 → 权限管理 中开启以下权限：</div>
                <div class="overflow-x-auto">
                  <table class="w-full text-xs">
                    <thead>
                      <tr class="bg-gray-50 text-left">
                        <th class="px-3 py-2 font-semibold">权限标识</th>
                        <th class="px-3 py-2 font-semibold">用途</th>
                        <th class="px-3 py-2 font-semibold">是否必须</th>
                      </tr>
                    </thead>
                    <tbody class="divide-y divide-gray-100">
                      <tr>
                        <td class="px-3 py-2 font-mono">im:message</td>
                        <td class="px-3 py-2 text-gray-600">读取消息内容（接收 im.message.receive_v1）</td>
                        <td class="px-3 py-2"><Tag color="red" class="text-xs">必须</Tag></td>
                      </tr>
                      <tr>
                        <td class="px-3 py-2 font-mono">im:message:send_as_bot</td>
                        <td class="px-3 py-2 text-gray-600">以 Bot 身份发送消息（回复用户、主动推送）</td>
                        <td class="px-3 py-2"><Tag color="red" class="text-xs">必须</Tag></td>
                      </tr>
                      <tr>
                        <td class="px-3 py-2 font-mono">im:message.group_at_msg:readonly</td>
                        <td class="px-3 py-2 text-gray-600">接收群聊中 @Bot 的消息</td>
                        <td class="px-3 py-2"><Tag color="orange" class="text-xs">群聊需要</Tag></td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>
            </Card>

            <!-- 接入流程 -->
            <Card title="🚀 快速接入流程">
              <div class="space-y-2 text-sm text-gray-700">
                <div class="flex items-start gap-2">
                  <div class="mt-0.5 flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-blue-100 text-xs font-bold text-blue-600">1</div>
                  <div>在飞书开放平台创建企业自建应用，获取 App ID 和 App Secret</div>
                </div>
                <div class="flex items-start gap-2">
                  <div class="mt-0.5 flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-blue-100 text-xs font-bold text-blue-600">2</div>
                  <div>切换到「配置」标签，填入 App ID / App Secret，选择连接模式，保存</div>
                </div>
                <div class="flex items-start gap-2">
                  <div class="mt-0.5 flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-blue-100 text-xs font-bold text-blue-600">3</div>
                  <div>在飞书开发者后台订阅 <span class="rounded bg-gray-100 px-1 font-mono text-xs">im.message.receive_v1</span> 事件，开启必要权限</div>
                </div>
                <div class="flex items-start gap-2">
                  <div class="mt-0.5 flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-blue-100 text-xs font-bold text-blue-600">4</div>
                  <div>切换到「绑定」标签，将系统用户与飞书 open_id 绑定（每人绑定一次即可）</div>
                </div>
                <div class="flex items-start gap-2">
                  <div class="mt-0.5 flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-blue-100 text-xs font-bold text-blue-600">5</div>
                  <div>在「绑定」页点击「测试消息」，确认收到飞书消息</div>
                </div>
                <div class="flex items-start gap-2">
                  <div class="mt-0.5 flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-blue-100 text-xs font-bold text-blue-600">6</div>
                  <div>（可选）切换到「机器人菜单」标签，点击「一键推送菜单」为用户设置快捷菜单 → 接入完成 🎉</div>
                </div>
              </div>
            </Card>
          </div>
        </Tabs.TabPane>

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

        <Tabs.TabPane key="menu" tab="机器人菜单">
          <div class="space-y-4">
            <Alert
              type="info"
              show-icon
              message="飞书机器人菜单"
              description="飞书机器人菜单需在「飞书开发者后台 → 应用功能 → 机器人 → 自定义菜单」中手动配置，此处管理 EventKey 与指令的映射关系，保存后后端会识别对应的菜单点击事件。"
            />
            <Alert
              type="warning"
              show-icon
              message="配置步骤：① 在此保存 EventKey 配置 → ② 去飞书开发者后台创建菜单，响应动作选「推送事件」并填写 EventKey → ③ 订阅 application.bot.menu_v6 事件"
            />

            <Card :loading="menuLoading" title="当前菜单配置">
              <template #extra>
                <Space>
                  <Tag :color="menuIsDefault ? 'blue' : 'green'">
                    {{ menuIsDefault ? '使用默认菜单' : '自定义菜单' }}
                  </Tag>
                  <Button size="small" @click="loadBotMenu">刷新</Button>
                </Space>
              </template>

              <div class="space-y-3">
                <div
                  v-for="(topItem, topIdx) in menuItems"
                  :key="topIdx"
                  class="rounded-lg border border-gray-200 bg-gray-50 p-3"
                >
                  <div class="mb-2 flex items-center gap-2">
                    <span class="text-xs font-semibold text-gray-500">一级菜单 {{ topIdx + 1 }}</span>
                    <Input
                      v-model:value="topItem.name"
                      size="small"
                      style="width: 140px"
                      placeholder="菜单名称"
                    />
                    <Input
                      v-model:value="topItem.eventKey"
                      size="small"
                      style="width: 180px"
                      placeholder="EventKey（唯一标识）"
                    />
                    <Button size="small" danger @click="removeTopMenu(topIdx)">删除</Button>
                  </div>
                  <div class="ml-4 space-y-1">
                    <div
                      v-for="(subItem, subIdx) in topItem.children"
                      :key="subIdx"
                      class="flex items-center gap-2"
                    >
                      <span class="text-xs text-gray-400">└</span>
                      <Input
                        v-model:value="subItem.name"
                        size="small"
                        style="width: 130px"
                        placeholder="子菜单名称"
                      />
                      <Input
                        v-model:value="subItem.eventKey"
                        size="small"
                        style="width: 180px"
                        placeholder="EventKey（唯一标识）"
                      />
                      <Button size="small" danger @click="removeSubMenu(topIdx, subIdx)">删除</Button>
                    </div>
                    <Button
                      size="small"
                      type="dashed"
                      style="margin-top: 4px"
                      @click="addSubMenu(topIdx)"
                    >
                      + 添加子菜单
                    </Button>
                  </div>
                </div>

                <Button
                  v-if="menuItems.length < 3"
                  type="dashed"
                  block
                  @click="addTopMenu"
                >
                  + 添加一级菜单
                </Button>
              </div>

              <div class="mt-4 flex justify-end gap-2">
                <Button :loading="menuSaving" @click="deployMenu(true)">
                  恢复默认
                </Button>
                <Button type="primary" :loading="menuSaving" @click="deployMenu(false)">
                  保存配置
                </Button>
              </div>
            </Card>

            <!-- 菜单 EventKey 说明 -->
            <Card title="EventKey 说明（点击菜单项时触发的指令）">
              <div class="overflow-x-auto">
                <table class="w-full text-xs">
                  <thead>
                    <tr class="bg-gray-50 text-left">
                      <th class="px-3 py-2 font-semibold">EventKey</th>
                      <th class="px-3 py-2 font-semibold">触发动作</th>
                    </tr>
                  </thead>
                  <tbody class="divide-y divide-gray-100">
                    <tr v-for="(row, i) in menuEventKeyDocs" :key="i">
                      <td class="px-3 py-1.5 font-mono text-blue-600">{{ row.key }}</td>
                      <td class="px-3 py-1.5 text-gray-600">{{ row.desc }}</td>
                    </tr>
                  </tbody>
                </table>
              </div>
              <div class="mt-2 text-xs text-gray-400">
                自定义 EventKey 时，Bot 会将其作为文本指令发送给 AI 处理（支持自然语言）。
              </div>
            </Card>
          </div>
        </Tabs.TabPane>

        <Tabs.TabPane key="chat-menu" tab="群菜单">
          <div class="space-y-4">
            <Alert
              type="info"
              show-icon
              message="飞书群菜单（快捷跳转）"
              description="在指定飞书群中添加快捷跳转菜单，用户点击菜单项可直接跳转到 EasyMVP 后台对应页面。需要 Bot 已加入目标群，并开启 im:chat 权限。"
            />

            <Card title="群菜单管理">
              <div class="space-y-4">
                <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
                  <FormItem label="群 chat_id">
                    <Input
                      v-model:value="chatID"
                      placeholder="oc_xxxxxxxxxxxxxxxxxx"
                      allow-clear
                    />
                  </FormItem>
                  <FormItem label="后台访问地址">
                    <Input
                      v-model:value="baseURL"
                      placeholder="https://easymvp.example.com"
                    />
                  </FormItem>
                </div>

                <div class="flex gap-2">
                  <Button :loading="chatMenuLoading" @click="loadChatMenu">查看当前菜单</Button>
                  <Button type="primary" :loading="chatMenuSaving" @click="handleCreateChatMenu">
                    一键创建默认菜单
                  </Button>
                  <Button
                    danger
                    :loading="chatMenuDeleting"
                    :disabled="chatMenuList.length === 0"
                    @click="handleDeleteAllChatMenu"
                  >
                    清空群菜单
                  </Button>
                </div>

                <!-- 当前群菜单展示 -->
                <div v-if="chatMenuList.length > 0" class="rounded-lg border border-gray-200 bg-gray-50 p-3">
                  <div class="mb-2 text-sm font-semibold text-gray-600">当前群菜单（{{ chatMenuList.length }} 个一级菜单）</div>
                  <div class="space-y-2">
                    <div v-for="(top, i) in chatMenuList" :key="i" class="rounded border border-gray-100 bg-white p-2">
                      <div class="font-medium text-gray-700">{{ top.chat_menu_item?.name ?? '-' }}</div>
                      <div v-if="top.children?.length" class="ml-3 mt-1 space-y-0.5">
                        <div v-for="(sub, j) in top.children" :key="j" class="text-xs text-gray-500">
                          └ {{ sub.chat_menu_item?.name }}
                          <span v-if="sub.chat_menu_item?.redirect_link?.common_url" class="text-blue-400">
                            → {{ sub.chat_menu_item.redirect_link.common_url }}
                          </span>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </Card>

            <!-- 默认菜单预览 -->
            <Card title="默认菜单结构预览">
              <div class="space-y-2 text-sm text-gray-600">
                <div v-for="(top, i) in defaultChatMenuPreview" :key="i" class="rounded border border-gray-100 p-2">
                  <div class="font-medium text-gray-700">{{ top.name }}</div>
                  <div class="ml-3 mt-1 space-y-0.5">
                    <div v-for="(sub, j) in top.children" :key="j" class="text-xs text-gray-500">
                      └ {{ sub.name }} → <span class="text-blue-400">{{ sub.url?.replace('{baseURL}', baseURL) }}</span>
                    </div>
                  </div>
                </div>
              </div>
              <div class="mt-3 text-xs text-gray-400">
                链接地址会自动替换为「后台访问地址」字段的值。
              </div>
            </Card>

            <!-- 如何获取 chat_id -->
            <Card title="如何获取群 chat_id">
              <div class="space-y-1 text-sm text-gray-600">
                <div>方法一：飞书开发者后台 → 群管理 → 找到目标群 → 复制 chat_id</div>
                <div>方法二：调用飞书 API <span class="rounded bg-gray-100 px-1 font-mono text-xs">GET /open-apis/im/v1/chats</span> 列出 Bot 所在的群</div>
                <div>方法三：在目标群发送消息，从飞书 Webhook 日志中获取 <span class="font-mono text-xs">chat_id</span> 字段</div>
              </div>
            </Card>
          </div>
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

<script setup lang="ts">
import type { ChatMessage } from '#/api/mvp/chat';

import { computed, nextTick, onMounted, onUnmounted, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import { useAppConfig } from '@vben/hooks';
import { useAccessStore } from '@vben/stores';

import {
  ArrowLeftOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined,
  LoadingOutlined,
  RobotOutlined,
  SearchOutlined,
  SyncOutlined,
} from '@ant-design/icons-vue';
import { Page } from '@vben/common-ui';
import { Avatar, Badge, Button, message, Modal, Spin, Tag, Tooltip } from 'ant-design-vue';

import { getChatHistory, sendMessage } from '#/api/mvp/chat';
import { confirmPlan, getProjectStatus, parseTasks } from '#/api/mvp/workflow';
import { getProjectDetail } from '#/api/mvp/project';
import { getConversationList } from '#/api/mvp/conversation';

import ChatInput from './components/ChatInput.vue';
import ChatMessageComp from './components/ChatMessage.vue';

/** 扩展消息类型，包含流式输出缓冲 */
interface MessageWithStream extends ChatMessage {
  streamingContent?: string;
}

// ======================== 路由参数 ========================

const route = useRoute();
const router = useRouter();

/** 当前项目 ID */
const projectId = ref<string>((route.query.projectId as string) || '');
/** 当前对话 ID */
const conversationId = ref<string>((route.query.conversationId as string) || '');
const needsTaskEntry = computed(
  () => !projectId.value && !conversationId.value,
);

// ======================== 项目状态 ========================

/** 项目名称 */
const projectName = ref<string>('加载中...');
/** 项目状态 */
const projectStatus = ref<string>('');
/** 草稿任务数（架构师解析出的待确认任务） */
const draftTaskCount = ref(0);
/** 是否正在确认方案 */
const confirmingPlan = ref(false);
/** 是否正在检查拆分 */
const parsingTasks = ref(false);

/** 项目状态配置 */
const STATUS_CONFIG: Record<
  string,
  { color: string; text: string; icon: any }
> = {
  designing: {
    color: 'processing',
    text: '设计中',
    icon: SyncOutlined,
  },
  pending: {
    color: 'default',
    text: '待开始',
    icon: LoadingOutlined,
  },
  running: {
    color: 'success',
    text: '执行中',
    icon: SyncOutlined,
  },
  completed: {
    color: 'success',
    text: '已完成',
    icon: CheckCircleOutlined,
  },
  failed: {
    color: 'error',
    text: '失败',
    icon: ExclamationCircleOutlined,
  },
  paused: {
    color: 'warning',
    text: '已暂停',
    icon: ExclamationCircleOutlined,
  },
};

/** 获取状态配置 */
function getStatusConfig(status: string) {
  return (
    STATUS_CONFIG[status] || { color: 'default', text: status, icon: null }
  );
}

// ======================== 消息列表 ========================

/** 消息列表 */
const messages = ref<MessageWithStream[]>([]);
/** 是否正在加载历史 */
const loadingHistory = ref(false);
/** 消息区域 DOM 引用 */
const messagesContainerRef = ref<HTMLElement | null>(null);

/** 滚动到底部 */
async function scrollToBottom(smooth = false) {
  await nextTick();
  if (messagesContainerRef.value) {
    messagesContainerRef.value.scrollTo({
      top: messagesContainerRef.value.scrollHeight,
      behavior: smooth ? 'smooth' : 'auto',
    });
  }
}

// ======================== SSE 流式输出 ========================

const { apiURL } = useAppConfig(import.meta.env, import.meta.env.PROD);
const accessStore = useAccessStore();

/** 当前 SSE AbortController，用于中断连接 */
let sseAbortController: AbortController | null = null;

/**
 * 使用 fetch + ReadableStream 实现 SSE
 * 因为 EventSource 不支持自定义 header（无法传 JWT token）
 */
async function connectSSE(replyID: string, targetMessageIndex: number) {
  // 中断上一次连接
  if (sseAbortController) {
    sseAbortController.abort();
  }
  sseAbortController = new AbortController();

  const token = accessStore.accessToken;
  const url = `${apiURL}/mvp/chat/sse?messageID=${replyID}`;

  try {
    const response = await fetch(url, {
      headers: {
        Authorization: `Bearer ${token}`,
        Accept: 'text/event-stream',
      },
      signal: sseAbortController.signal,
    });

    if (!response.ok) {
      throw new Error(`SSE 连接失败: ${response.status}`);
    }

    const reader = response.body?.getReader();
    if (!reader) {
      throw new Error('无法获取响应流');
    }

    const decoder = new TextDecoder('utf-8');
    let buffer = '';
    let scrollTimer: ReturnType<typeof setTimeout> | null = null;

    // 防抖滚动：流式输出中最多每 100ms 滚动一次
    const debouncedScroll = () => {
      if (scrollTimer) return;
      scrollTimer = setTimeout(() => {
        scrollTimer = null;
        scrollToBottom();
      }, 100);
    };

    // 跟踪已收到的最大 chunk index（用于重连去重）
    let lastChunkIndex = -1;

    // 跟踪是否收到过失败信号，用于 done 时判断最终状态
    let hasFailed = false;

    // 逐块读取数据
    while (true) {
      const { done, value } = await reader.read();
      if (done) break;

      // 解码并追加到缓冲区
      buffer += decoder.decode(value, { stream: true });

      // 按行解析标准 SSE 格式（event: xxx\ndata: {...}\n\n）
      const lastNewline = buffer.lastIndexOf('\n');
      if (lastNewline === -1) continue;

      const complete = buffer.substring(0, lastNewline);
      buffer = buffer.substring(lastNewline + 1);
      const lines = complete.split('\n');

      // 当前事件的 event 类型（每对 event/data 行共享）
      let currentEvent = '';

      for (const line of lines) {
        const trimmed = line.trim();

        // 解析 event: 行
        if (trimmed.startsWith('event:')) {
          currentEvent = trimmed.slice(6).trim();
          continue;
        }

        if (!trimmed.startsWith('data:')) continue;

        const jsonStr = trimmed.slice(5).trim();
        if (!jsonStr) continue;

        // 使用当前 event 类型，然后重置
        const eventType = currentEvent || '';
        currentEvent = '';

        try {
          const data = JSON.parse(jsonStr);
          const msg = messages.value[targetMessageIndex];

          // 按事件类型分发处理
          switch (eventType) {
            case 'full': {
              // 完整消息（已完成或已失败的消息直接返回）
              if (msg) {
                msg.content = data.content || msg.content;
                msg.streamingContent = undefined;
                msg.status = data.status || 'completed';
                if (data.status === 'failed') hasFailed = true;
              }

              break;
            }
            case 'error': {
              // 错误事件
              if (msg) {
                msg.status = 'failed';
                msg.content = data.error || msg.streamingContent || '（响应出现错误）';
                msg.streamingContent = undefined;
              }
              hasFailed = true;

              break;
            }
            case 'done': {
              // 流结束信号
              if (msg) {
                if (!hasFailed && msg.status !== 'failed') {
                  // 没有收到过失败信号，标记完成
                  msg.status = 'completed';
                }
                msg.content = msg.streamingContent || msg.content;
                msg.streamingContent = undefined;
              }
              if (scrollTimer) { clearTimeout(scrollTimer); scrollTimer = null; }
              await scrollToBottom();
              return;
            }
            case 'chunk':
            default: {
              // 续写通知事件
              if (data.event === 'continue') {
                if (msg) {
                  msg.streamingContent = (msg.streamingContent || '') + '\n\n> _AI 正在继续生成..._\n\n';
                  debouncedScroll();
                }
                break;
              }
              // chunk 事件或无事件类型（兼容旧格式）
              if (data.done) {
                // 兼容：data.done 标记（无 event 类型时的结束信号）
                if (msg) {
                  if (!hasFailed && msg.status !== 'failed') {
                    msg.status = 'completed';
                  }
                  msg.content = msg.streamingContent || msg.content;
                  msg.streamingContent = undefined;
                }
                if (scrollTimer) { clearTimeout(scrollTimer); scrollTimer = null; }
                await scrollToBottom();
                return;
              }
              if (data.content !== undefined && msg) {
                // chunk 去重：跳过已收到的 chunk（重连时后端会重发已有 chunks）
                const chunkIdx = data.index ?? -1;
                if (chunkIdx >= 0 && chunkIdx <= lastChunkIndex) {
                  break; // 跳过重复 chunk
                }
                if (chunkIdx > lastChunkIndex) {
                  lastChunkIndex = chunkIdx;
                }
                msg.streamingContent = (msg.streamingContent || '') + data.content;
                debouncedScroll();
              }

              break;
            }
          }
        } catch {
          // 忽略 JSON 解析错误
        }
      }
    }
  } catch (err: any) {
    if (err.name === 'AbortError') {
      // 主动中断，不处理
      return;
    }
    console.error('SSE 连接异常:', err);
    // 标记消息为失败
    const msg = messages.value[targetMessageIndex];
    if (msg) {
      msg.status = 'failed';
      msg.content = msg.streamingContent || '（响应出现错误，请重试）';
      msg.streamingContent = undefined;
    }
  }
}

// ======================== 发送消息 ========================

/** 是否正在等待 AI 响应 */
const isSending = ref(false);

/** 发送消息处理 */
async function handleSend(content: string) {
  if (!content.trim() || isSending.value) return;
  if (needsTaskEntry.value) {
    message.warning('请从 MVP任务 页面进入任务对话');
    return;
  }
  if (!conversationId.value) {
    message.error('对话 ID 未设置');
    return;
  }

  isSending.value = true;

  // 添加用户消息到列表
  const now = new Date().toISOString();
  const userMsg: MessageWithStream = {
    id: `local-user-${Date.now()}`,
    role: 'user',
    content,
    status: 'completed',
    createdAt: now,
  };
  messages.value.push(userMsg);
  await scrollToBottom(true);

  // 创建 AI 消息占位（streaming 状态）
  const aiMsgIndex = messages.value.length;
  const aiMsg: MessageWithStream = {
    id: `local-ai-${Date.now()}`,
    role: 'assistant',
    content: '',
    status: 'streaming',
    streamingContent: '',
    createdAt: new Date().toISOString(),
  };
  messages.value.push(aiMsg);
  await scrollToBottom();

  try {
    // 调用发送 API
    const result = await sendMessage({
      conversationID: conversationId.value,
      content,
    });

    // 更新 AI 消息 ID（用真实 ID）
    messages.value[aiMsgIndex].id = result.replyID;

    // 开始 SSE 流式接收
    await connectSSE(result.replyID, aiMsgIndex);
  } catch (err: any) {
    console.error('发送消息失败:', err);
    message.error('发送失败，请重试');
    // 标记 AI 消息为失败
    const msg = messages.value[aiMsgIndex];
    if (msg) {
      msg.status = 'failed';
      msg.content = '（发送失败）';
      msg.streamingContent = undefined;
    }
  } finally {
    isSending.value = false;
    // AI 回复完成后刷新状态（架构师可能解析出了新的草稿任务）
    loadProjectStatus();
  }
}

// ======================== 确认方案 ========================

/** 检查拆分按钮点击：先 dryRun 检查，再弹窗确认创建 */
async function handleParseTasks() {
  if (!projectId.value) return;
  parsingTasks.value = true;
  try {
    // 第一步：dryRun 仅检查
    const check = await parseTasks(projectId.value, true);
    if (!check.hasTasks) {
      message.info('架构师回复中未检测到任务清单，请先让AI完成任务拆分');
      return;
    }
    // 第二步：弹窗确认
    Modal.confirm({
      title: '检测到任务清单',
      content: `架构师回复中解析出 ${check.taskCount} 个任务，是否创建为草案任务？`,
      okText: '确定创建',
      cancelText: '取消',
      async onOk() {
        // 第三步：实际创建
        const res = await parseTasks(projectId.value, false);
        await loadProjectStatus();
        if (res.message) {
          // 异步提取中，提示用户等待
          message.info(res.message);
        } else if (res.taskCount > 0) {
          message.success(`已创建 ${res.taskCount} 个草案任务`);
        } else {
          message.info('未提取到任务，请检查架构师回复格式');
        }
      },
    });
  } catch (err: any) {
    message.error(err?.message || '拆分方案失败');
  } finally {
    parsingTasks.value = false;
  }
}

/** 确认方案按钮点击 */
async function handleConfirmPlan() {
  if (!projectId.value) return;
  confirmingPlan.value = true;
  try {
    await confirmPlan(projectId.value);
    message.success('方案已提交审核，请等待审核结果');
    projectStatus.value = 'reviewing';
    await loadProjectStatus();
  } catch (err: any) {
    message.error(err?.message || '确认方案失败');
  } finally {
    confirmingPlan.value = false;
  }
}

// ======================== 数据加载 ========================

/** 加载项目信息和状态 */
async function loadProjectStatus() {
  if (!projectId.value) return;
  try {
    const [detail, status] = await Promise.all([
      getProjectDetail(projectId.value),
      getProjectStatus(projectId.value),
    ]);
    projectName.value = detail.name || '未命名项目';
    projectStatus.value = status.status || '';
    draftTaskCount.value = status.statusCounts?.draft || 0;
  } catch (err) {
    console.error('加载项目状态失败:', err);
    projectName.value = '项目';
  }
}

/** 加载对话历史 */
async function loadHistory() {
  if (!conversationId.value) return;
  loadingHistory.value = true;
  try {
    const result = await getChatHistory(conversationId.value);
    messages.value = (result.list || []).map((msg) => ({
      ...msg,
      streamingContent: msg.status === 'streaming' ? (msg.content || '') : undefined,
    }));
    await scrollToBottom();

    // 自动恢复 streaming 消息的 SSE 连接
    const streamingIdx = messages.value.findIndex((m) => m.status === 'streaming');
    if (streamingIdx >= 0) {
      const streamingMsg = messages.value[streamingIdx];
      isSending.value = true;
      connectSSE(streamingMsg.id, streamingIdx).finally(() => {
        isSending.value = false;
        loadProjectStatus();
      });
    }
  } catch (err) {
    console.error('加载历史消息失败:', err);
    message.error('加载历史消息失败');
  } finally {
    loadingHistory.value = false;
  }
}

/** 返回项目列表 */
function handleBack() {
  router.push({ path: '/mvp/project' });
}

// ======================== 生命周期 ========================

onMounted(async () => {
  if (needsTaskEntry.value) {
    projectName.value = 'MVP任务对话';
    return;
  }
  // 如果没有 conversationId 但有 projectId，查询架构师对话
  if (!conversationId.value && projectId.value) {
    try {
      const res = await getConversationList({
        pageNum: 1,
        pageSize: 1,
        projectID: projectId.value,
        roleType: 'architect',
        orderBy: 'created_at',
        orderDir: 'desc',
      } as any);
      if (res?.list?.length) {
        conversationId.value = res.list[0].id;
      }
    } catch {
      // 忽略
    }
  }
  // 并行加载项目状态和历史消息
  await Promise.all([loadProjectStatus(), loadHistory()]);
});

onUnmounted(() => {
  // 页面销毁时中断 SSE 连接
  if (sseAbortController) {
    sseAbortController.abort();
  }
});
</script>

<template>
  <Page auto-content-height :content-class="'h-full'">
  <div class="chat-page">
    <!-- ===== 顶部栏 ===== -->
    <div class="chat-header">
      <!-- 左侧：返回 + 项目信息 -->
      <div class="header-left">
        <Button
          type="text"
          class="back-btn"
          @click="handleBack"
        >
          <template #icon>
            <ArrowLeftOutlined />
          </template>
          返回
        </Button>

        <div class="project-info">
          <!-- 项目图标 -->
          <Avatar
            :size="32"
            style="background: linear-gradient(135deg, #667eea 0%, #764ba2 100%)"
          >
            <template #icon>
              <RobotOutlined />
            </template>
          </Avatar>
          <!-- 项目名称 -->
          <span class="project-name">{{ projectName }}</span>
          <!-- 状态 Tag -->
          <Tooltip
            v-if="projectStatus"
            :title="getStatusConfig(projectStatus).text"
          >
            <Tag :color="getStatusConfig(projectStatus).color" class="status-tag">
              <component
                :is="getStatusConfig(projectStatus).icon"
                v-if="getStatusConfig(projectStatus).icon"
                style="margin-right: 4px"
              />
              {{ getStatusConfig(projectStatus).text }}
            </Tag>
          </Tooltip>
        </div>
      </div>

      <!-- 右侧：操作按钮 -->
      <div class="header-right">
        <!-- 检查拆分按钮：设计中/暂停状态 且无草稿任务时显示 -->
        <Button
          v-if="(projectStatus === 'designing' || projectStatus === 'paused') && draftTaskCount === 0"
          :loading="parsingTasks"
          class="parse-btn"
          @click="handleParseTasks"
        >
          <template #icon>
            <SearchOutlined />
          </template>
          拆分方案
        </Button>
        <!-- 设计中/暂停状态 且有草稿任务时显示确认方案按钮 -->
        <Badge :count="draftTaskCount" :offset="[-6, 0]">
          <Button
            v-if="projectStatus === 'designing' || projectStatus === 'paused'"
            type="primary"
            :loading="confirmingPlan"
            :disabled="draftTaskCount === 0"
            class="confirm-btn"
            @click="handleConfirmPlan"
          >
            <template #icon>
              <CheckCircleOutlined />
            </template>
            确认方案
            <span v-if="draftTaskCount > 0" class="task-count">
              ({{ draftTaskCount }}个任务)
            </span>
          </Button>
        </Badge>
      </div>
    </div>

    <!-- ===== 消息区域 ===== -->
    <div class="chat-body">
      <div
        v-if="needsTaskEntry"
        class="entry-guide"
      >
        <div class="entry-guide-icon">
          <RobotOutlined />
        </div>
        <div class="entry-guide-title">请从 MVP任务 进入任务对话</div>
        <div class="entry-guide-desc">
          当前页面需要任务上下文才能加载对应对话。请先进入具体项目的
          <span class="entry-guide-highlight">MVP任务</span>
          页面，再点击任务行中的“查看对话”。
        </div>
        <Button type="primary" class="entry-guide-btn" @click="router.push('/mvp/project')">
          返回 MVP项目
        </Button>
      </div>

      <!-- 加载历史中的 Spin -->
      <Spin
        v-else-if="loadingHistory"
        class="loading-spin"
        size="large"
        tip="加载对话历史..."
      />

      <!-- 消息列表 -->
      <div
        v-else
        ref="messagesContainerRef"
        class="messages-container"
      >
        <!-- 空状态 -->
        <div
          v-if="messages.length === 0"
          class="empty-state"
        >
          <div class="empty-icon">
            <RobotOutlined />
          </div>
          <p class="empty-title">AI 架构师已就绪</p>
          <p class="empty-desc">请告诉我您想要构建什么，我将为您设计最优方案</p>
        </div>

        <!-- 消息列表 -->
        <template v-else>
          <!-- 历史提示 -->
          <div class="history-divider">
            <span>— 历史消息 —</span>
          </div>

          <!-- 消息气泡 -->
          <ChatMessageComp
            v-for="msg in messages"
            :key="msg.id"
            :message="msg"
          />
        </template>

        <!-- 底部留白，保证最后一条消息不被输入框遮挡 -->
        <div class="messages-bottom-padding" />
      </div>
    </div>

    <!-- ===== 输入区域 ===== -->
    <div v-if="!needsTaskEntry" class="chat-footer">
      <ChatInput
        :loading="isSending"
        @send="handleSend"
      />
    </div>
  </div>
  </Page>
</template>

<style scoped>
/* 整体布局：flex 纵向 */
.chat-page {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: #fff;
  overflow: hidden;
}

/* ===== 顶部栏 ===== */
.chat-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 20px;
  background: #fff;
  border-bottom: 1px solid #f0f0f0;
  box-shadow: 0 1px 4px rgb(0 0 0 / 6%);
  flex-shrink: 0;
  z-index: 10;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
}

.back-btn {
  color: #595959;
  flex-shrink: 0;
}

.project-info {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
}

.project-name {
  font-size: 16px;
  font-weight: 600;
  color: #1a1a1a;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 300px;
}

.status-tag {
  flex-shrink: 0;
  cursor: default;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.parse-btn {
  border-radius: 8px;
}

.confirm-btn {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: none;
  border-radius: 8px;
}

.confirm-btn:hover {
  opacity: 0.9;
}

.task-count {
  font-size: 12px;
  opacity: 0.85;
  margin-left: 2px;
}

/* ===== 消息区域 ===== */
.chat-body {
  flex: 1;
  overflow: hidden;
  position: relative;
  background: #fafafa;
}

.entry-guide {
  display: flex;
  height: 100%;
  min-height: 360px;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 24px;
  text-align: center;
}

.entry-guide-icon {
  display: flex;
  height: 88px;
  width: 88px;
  align-items: center;
  justify-content: center;
  border-radius: 9999px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #fff;
  font-size: 38px;
  box-shadow: 0 12px 32px rgb(102 126 234 / 28%);
}

.entry-guide-title {
  margin-top: 20px;
  font-size: 22px;
  font-weight: 600;
  color: #1f2937;
}

.entry-guide-desc {
  margin-top: 12px;
  max-width: 520px;
  line-height: 1.8;
  color: #6b7280;
}

.entry-guide-highlight {
  margin: 0 4px;
  font-weight: 600;
  color: #4f46e5;
}

.entry-guide-btn {
  margin-top: 24px;
  border-radius: 10px;
}

/* 加载 Spin */
.loading-spin {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
}

/* 消息滚动容器 */
.messages-container {
  height: 100%;
  overflow-y: auto;
  padding: 20px 20px 0;
  scroll-behavior: auto;
}

/* 滚动条样式 */
.messages-container::-webkit-scrollbar {
  width: 6px;
}

.messages-container::-webkit-scrollbar-track {
  background: transparent;
}

.messages-container::-webkit-scrollbar-thumb {
  background: #d9d9d9;
  border-radius: 3px;
}

.messages-container::-webkit-scrollbar-thumb:hover {
  background: #bfbfbf;
}

/* 历史消息分割线 */
.history-divider {
  text-align: center;
  margin-bottom: 20px;
  color: #bfbfbf;
  font-size: 12px;
  position: relative;
}

.history-divider::before,
.history-divider::after {
  content: '';
  display: inline-block;
  width: 60px;
  height: 1px;
  background: #e8e8e8;
  vertical-align: middle;
  margin: 0 8px;
}

/* 底部留白 */
.messages-bottom-padding {
  height: 20px;
}

/* ===== 空状态 ===== */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  min-height: 300px;
  text-align: center;
  padding: 40px 20px;
}

.empty-icon {
  width: 80px;
  height: 80px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 20px;
  font-size: 36px;
  color: #fff;
  box-shadow: 0 8px 24px rgb(102 126 234 / 30%);
}

.empty-title {
  font-size: 18px;
  font-weight: 600;
  color: #1a1a1a;
  margin: 0 0 8px;
}

.empty-desc {
  font-size: 14px;
  color: #8c8c8c;
  max-width: 320px;
  line-height: 1.6;
  margin: 0;
}

/* ===== 输入区域 ===== */
.chat-footer {
  flex-shrink: 0;
  z-index: 10;
}

/* ===== 响应式 ===== */
@media (max-width: 768px) {
  .chat-header {
    padding: 10px 12px;
  }

  .project-name {
    max-width: 150px;
    font-size: 14px;
  }

  .messages-container {
    padding: 12px 12px 0;
  }

  .confirm-btn span:not(.anticon) {
    display: none;
  }
}
</style>

<script setup lang="ts">
import type { ChatMessage } from '#/api/mvp/chat';

import { computed, ref } from 'vue';

import { FileTextOutlined } from '@ant-design/icons-vue';
import { Avatar } from 'ant-design-vue';

/** 组件属性 */
const props = defineProps<{
  message: ChatMessage & { streamingContent?: string };
}>();

/** 是否是 AI 消息 */
const isAssistant = computed(() => props.message.role === 'assistant');

/** 是否正在流式输出 */
const isStreaming = computed(() => props.message.status === 'streaming');

/** 显示的内容（流式输出时用 streamingContent，否则用 content） */
const displayContent = computed(() => {
  if (isStreaming.value && props.message.streamingContent !== undefined) {
    return props.message.streamingContent;
  }
  return props.message.content;
});

/** 文件内容分隔标记 */
const FILE_CONTENT_SEPARATOR = '\n\n---\n以下是读取的文件内容：\n';

/** 用户消息正文（不含文件内容部分） */
const userMainContent = computed(() => {
  const content = displayContent.value || '';
  const idx = content.indexOf(FILE_CONTENT_SEPARATOR);
  return idx >= 0 ? content.substring(0, idx) : content;
});

/** 用户消息中的文件内容（折叠部分） */
const userFileContent = computed(() => {
  const content = displayContent.value || '';
  const idx = content.indexOf(FILE_CONTENT_SEPARATOR);
  return idx >= 0 ? content.substring(idx + FILE_CONTENT_SEPARATOR.length) : '';
});

/** 是否有文件内容 */
const hasFileContent = computed(() => userFileContent.value.length > 0);

/** 是否展开文件内容 */
const fileExpanded = ref(false);

/** 格式化时间（从 ISO 字符串提取时分） */
function formatTime(dateStr: string): string {
  if (!dateStr) return '';
  try {
    const date = new Date(dateStr);
    const hours = date.getHours().toString().padStart(2, '0');
    const minutes = date.getMinutes().toString().padStart(2, '0');
    return `${hours}:${minutes}`;
  } catch {
    return '';
  }
}

/**
 * 简单的 Markdown 渲染函数
 * 处理常见格式：代码块、行内代码、粗体、斜体、链接、换行
 */
function renderMarkdown(text: string): string {
  if (!text) return '';

  let html = text;

  // 代码块（带语言标记）：```lang\ncode\n```
  html = html.replace(
    /```(\w*)\n?([\s\S]*?)```/g,
    (_, lang, code) => {
      const langClass = lang ? ` class="language-${lang}"` : '';
      const escapedCode = code
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;');
      return `<pre class="code-block"><div class="code-lang">${lang || 'code'}</div><code${langClass}>${escapedCode}</code></pre>`;
    },
  );

  // 行内代码：`code`
  html = html.replace(
    /`([^`]+)`/g,
    '<code class="inline-code">$1</code>',
  );

  // 粗体：**text** 或 __text__
  html = html.replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>');
  html = html.replace(/__(.+?)__/g, '<strong>$1</strong>');

  // 斜体：*text* 或 _text_
  html = html.replace(/\*(.+?)\*/g, '<em>$1</em>');
  html = html.replace(/_(.+?)_/g, '<em>$1</em>');

  // 标题：# ## ###
  html = html.replace(/^### (.+)$/gm, '<h3>$1</h3>');
  html = html.replace(/^## (.+)$/gm, '<h2>$1</h2>');
  html = html.replace(/^# (.+)$/gm, '<h1>$1</h1>');

  // 无序列表：- item 或 * item
  html = html.replace(/^[*-] (.+)$/gm, '<li>$1</li>');
  html = html.replace(/(<li>.*<\/li>\n?)+/g, '<ul>$&</ul>');

  // 有序列表：1. item
  html = html.replace(/^\d+\. (.+)$/gm, '<li>$1</li>');

  // 链接：[text](url)
  html = html.replace(
    /\[([^\]]+)\]\(([^)]+)\)/g,
    '<a href="$2" target="_blank" rel="noopener noreferrer">$1</a>',
  );

  // 段落换行（两个换行转 <p>，单个换行转 <br>）
  html = html.replace(/\n\n/g, '</p><p>');
  html = html.replace(/\n/g, '<br>');
  html = `<p>${html}</p>`;

  // 清理 pre 标签内部被错误处理的 p/br 标签
  html = html.replace(/<pre([^>]*)>([\s\S]*?)<\/pre>/g, (_, attrs, content) => {
    const cleaned = content.replace(/<\/?p>|<br>/g, '\n').replace(/^\n|\n$/g, '');
    return `<pre${attrs}>${cleaned}</pre>`;
  });

  return html;
}
</script>

<template>
  <div
    class="message-wrapper"
    :class="{ 'message-user': !isAssistant, 'message-assistant': isAssistant }"
  >
    <!-- AI 消息：左对齐布局 -->
    <template v-if="isAssistant">
      <!-- AI 头像 -->
      <Avatar
        class="message-avatar"
        :size="36"
        style="background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); flex-shrink: 0"
      >
        AI
      </Avatar>

      <!-- 消息内容区 -->
      <div class="message-content-wrapper">
        <!-- 角色标签 -->
        <div class="message-role-label">AI 架构师</div>

        <!-- 气泡 -->
        <div class="message-bubble bubble-assistant">
          <!-- 流式输出中显示动画光标 -->
          <div
            v-if="isStreaming && !displayContent"
            class="typing-indicator"
          >
            <span /><span /><span />
          </div>
          <!-- Markdown 渲染内容 -->
          <div
            v-else
            class="markdown-body"
            v-html="renderMarkdown(displayContent)"
          />
          <!-- 流式光标 -->
          <span v-if="isStreaming && displayContent" class="streaming-cursor">▊</span>
        </div>

        <!-- 底部元信息 -->
        <div class="message-meta">
          <span v-if="message.modelName" class="meta-model">{{ message.modelName }}</span>
          <span class="meta-time">{{ formatTime(message.createdAt) }}</span>
        </div>
      </div>
    </template>

    <!-- 用户消息：右对齐布局 -->
    <template v-else>
      <!-- 消息内容区 -->
      <div class="message-content-wrapper align-right">
        <!-- 气泡 -->
        <div class="message-bubble bubble-user">
          <div class="user-text">{{ userMainContent }}</div>
          <!-- 文件内容折叠区 -->
          <div v-if="hasFileContent" class="file-content-toggle" @click.stop="fileExpanded = !fileExpanded">
            <FileTextOutlined style="margin-right: 4px" />
            {{ fileExpanded ? '收起文件内容' : '查看读取的文件内容' }}
          </div>
          <div v-if="hasFileContent && fileExpanded" class="file-content-area">
            <div class="markdown-body" v-html="renderMarkdown(userFileContent)" />
          </div>
        </div>
        <!-- 底部元信息 -->
        <div class="message-meta justify-end">
          <span class="meta-time">{{ formatTime(message.createdAt) }}</span>
        </div>
      </div>

      <!-- 用户头像 -->
      <Avatar
        class="message-avatar"
        :size="36"
        style="background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%); flex-shrink: 0"
      >
        我
      </Avatar>
    </template>
  </div>
</template>

<style scoped>
/* 消息行容器 */
.message-wrapper {
  display: flex;
  gap: 12px;
  margin-bottom: 20px;
  align-items: flex-start;
}

/* AI 消息左对齐 */
.message-assistant {
  flex-direction: row;
  padding-right: 15%;
}

/* 用户消息右对齐 */
.message-user {
  flex-direction: row-reverse;
  padding-left: 15%;
}

/* 头像 */
.message-avatar {
  margin-top: 4px;
}

/* 内容区域 */
.message-content-wrapper {
  display: flex;
  flex-direction: column;
  gap: 4px;
  max-width: 100%;
  min-width: 0;
}

.align-right {
  align-items: flex-end;
}

/* 角色标签 */
.message-role-label {
  font-size: 12px;
  color: #8c8c8c;
  margin-bottom: 2px;
  padding-left: 4px;
}

/* 气泡通用样式 */
.message-bubble {
  padding: 12px 16px;
  border-radius: 12px;
  word-break: break-word;
  overflow-wrap: break-word;
  line-height: 1.6;
  font-size: 14px;
  position: relative;
}

/* AI 气泡 */
.bubble-assistant {
  background: #f5f5f5;
  color: #1a1a1a;
  border-top-left-radius: 4px;
  border: 1px solid #e8e8e8;
}

/* 用户气泡 */
.bubble-user {
  background: #1677ff;
  color: #fff;
  border-top-right-radius: 4px;
}

/* 用户消息文本（保留换行） */
.user-text {
  white-space: pre-wrap;
}

/* 文件内容折叠切换按钮 */
.file-content-toggle {
  margin-top: 8px;
  padding: 4px 10px;
  font-size: 12px;
  color: rgba(255, 255, 255, 0.75);
  background: rgba(255, 255, 255, 0.15);
  border-radius: 4px;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  user-select: none;
  transition: background 0.2s;
}

.file-content-toggle:hover {
  background: rgba(255, 255, 255, 0.25);
  color: #fff;
}

/* 文件内容展开区域 */
.file-content-area {
  margin-top: 8px;
  padding: 8px;
  background: rgba(0, 0, 0, 0.2);
  border-radius: 8px;
  max-height: 400px;
  overflow-y: auto;
  color: #f0f0f0;
}

.file-content-area .markdown-body :deep(.code-block) {
  margin: 6px 0;
}

/* 底部元信息 */
.message-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 4px;
}

.justify-end {
  justify-content: flex-end;
}

.meta-model {
  font-size: 11px;
  color: #9254de;
  background: #f9f0ff;
  border: 1px solid #d3adf7;
  padding: 0 6px;
  border-radius: 4px;
}

.meta-time {
  font-size: 11px;
  color: #bfbfbf;
}

/* 流式输出闪烁光标 */
.streaming-cursor {
  display: inline-block;
  animation: blink 1s step-end infinite;
  color: #1677ff;
  font-weight: bold;
  margin-left: 2px;
}

@keyframes blink {
  0%, 100% { opacity: 1; }
  50% { opacity: 0; }
}

/* 打字指示器（三个跳动的点） */
.typing-indicator {
  display: flex;
  gap: 4px;
  padding: 4px 0;
  align-items: center;
}

.typing-indicator span {
  display: inline-block;
  width: 8px;
  height: 8px;
  background: #1677ff;
  border-radius: 50%;
  animation: typing-bounce 1.2s infinite ease-in-out;
}

.typing-indicator span:nth-child(1) { animation-delay: 0s; }
.typing-indicator span:nth-child(2) { animation-delay: 0.2s; }
.typing-indicator span:nth-child(3) { animation-delay: 0.4s; }

@keyframes typing-bounce {
  0%, 80%, 100% { transform: scale(0.6); opacity: 0.4; }
  40% { transform: scale(1); opacity: 1; }
}

/* Markdown 渲染样式 */
.markdown-body :deep(p) {
  margin: 0 0 8px;
}

.markdown-body :deep(p:last-child) {
  margin-bottom: 0;
}

.markdown-body :deep(h1),
.markdown-body :deep(h2),
.markdown-body :deep(h3) {
  margin: 12px 0 6px;
  font-weight: 600;
  line-height: 1.4;
}

.markdown-body :deep(h1) { font-size: 18px; }
.markdown-body :deep(h2) { font-size: 16px; }
.markdown-body :deep(h3) { font-size: 14px; }

.markdown-body :deep(ul),
.markdown-body :deep(ol) {
  margin: 6px 0;
  padding-left: 20px;
}

.markdown-body :deep(li) {
  margin-bottom: 4px;
}

/* 行内代码 */
.markdown-body :deep(.inline-code) {
  background: #e8e8e8;
  color: #d4380d;
  padding: 1px 5px;
  border-radius: 3px;
  font-family: 'SFMono-Regular', Consolas, monospace;
  font-size: 13px;
}

/* 代码块 */
.markdown-body :deep(.code-block) {
  background: #1e1e1e;
  border-radius: 8px;
  margin: 10px 0;
  overflow: hidden;
  position: relative;
}

.markdown-body :deep(.code-lang) {
  background: #2d2d2d;
  color: #858585;
  font-size: 11px;
  padding: 4px 12px;
  font-family: 'SFMono-Regular', Consolas, monospace;
}

.markdown-body :deep(.code-block code) {
  display: block;
  padding: 12px 16px;
  color: #d4d4d4;
  font-family: 'SFMono-Regular', Consolas, monospace;
  font-size: 13px;
  line-height: 1.6;
  overflow-x: auto;
  white-space: pre;
}

.markdown-body :deep(strong) {
  font-weight: 600;
}

.markdown-body :deep(a) {
  color: #1677ff;
  text-decoration: none;
}

.markdown-body :deep(a:hover) {
  text-decoration: underline;
}

/* 响应式：移动端减小缩进 */
@media (max-width: 768px) {
  .message-assistant {
    padding-right: 5%;
  }

  .message-user {
    padding-left: 5%;
  }
}
</style>

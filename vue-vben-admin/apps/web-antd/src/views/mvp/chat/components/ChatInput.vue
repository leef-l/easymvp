<script setup lang="ts">
import { computed, ref } from 'vue';

import { Button, Tooltip } from 'ant-design-vue';
import {
  EnterOutlined,
  SendOutlined,
} from '@ant-design/icons-vue';

/** 组件属性 */
const props = defineProps<{
  /** 是否正在发送/等待响应 */
  loading: boolean;
  /** 是否禁用（如对话已结束） */
  disabled?: boolean;
}>();

/** 事件 */
const emit = defineEmits<{
  send: [content: string];
}>();

/** 输入内容 */
const inputValue = ref('');

/** 是否可以发送（非空、非加载中） */
const canSend = computed(() => {
  return inputValue.value.trim().length > 0 && !props.loading && !props.disabled;
});

/** 发送消息 */
function handleSend() {
  if (!canSend.value) return;
  const content = inputValue.value.trim();
  inputValue.value = '';
  emit('send', content);
}

/**
 * 键盘事件处理
 * Enter 发送，Shift+Enter 换行
 */
function handleKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault();
    handleSend();
  }
}
</script>

<template>
  <div class="chat-input-area">
    <div class="input-container">
      <!-- 文本输入框 -->
      <textarea
        v-model="inputValue"
        class="chat-textarea"
        placeholder="输入消息，Enter 发送，Shift+Enter 换行..."
        :disabled="loading || disabled"
        :rows="3"
        @keydown="handleKeydown"
      />

      <!-- 底部操作栏 -->
      <div class="input-footer">
        <!-- 提示文字 -->
        <div class="input-hint">
          <Tooltip title="Enter 发送，Shift+Enter 换行">
            <span class="hint-text">
              <EnterOutlined style="margin-right: 4px" />
              Enter 发送
            </span>
          </Tooltip>
        </div>

        <!-- 字数统计 -->
        <span class="char-count" :class="{ 'char-warn': inputValue.length > 4000 }">
          {{ inputValue.length }} 字
        </span>

        <!-- 发送按钮 -->
        <Button
          type="primary"
          :disabled="!canSend"
          :loading="loading"
          class="send-btn"
          @click="handleSend"
        >
          <template #icon>
            <SendOutlined />
          </template>
          发送
        </Button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.chat-input-area {
  padding: 12px 16px 16px;
  background: #fff;
  border-top: 1px solid #f0f0f0;
}

.input-container {
  background: #fafafa;
  border: 1px solid #e8e8e8;
  border-radius: 12px;
  overflow: hidden;
  transition: border-color 0.2s;
}

.input-container:focus-within {
  border-color: #1677ff;
  box-shadow: 0 0 0 2px rgb(22 119 255 / 10%);
}

/* 文本输入框 */
.chat-textarea {
  display: block;
  width: 100%;
  padding: 12px 16px 8px;
  border: none;
  outline: none;
  background: transparent;
  font-size: 14px;
  line-height: 1.6;
  color: #1a1a1a;
  resize: none;
  min-height: 72px;
  max-height: 200px;
  box-sizing: border-box;
  font-family: inherit;
}

.chat-textarea::placeholder {
  color: #bfbfbf;
}

.chat-textarea:disabled {
  cursor: not-allowed;
  color: #bfbfbf;
}

/* 底部操作栏 */
.input-footer {
  display: flex;
  align-items: center;
  padding: 6px 12px 10px;
  gap: 8px;
}

.input-hint {
  flex: 1;
}

.hint-text {
  font-size: 12px;
  color: #bfbfbf;
  cursor: default;
  user-select: none;
}

/* 字数统计 */
.char-count {
  font-size: 12px;
  color: #bfbfbf;
}

.char-warn {
  color: #ff4d4f;
}

/* 发送按钮 */
.send-btn {
  border-radius: 8px;
}

/* 响应式 */
@media (max-width: 768px) {
  .chat-input-area {
    padding: 8px 12px 12px;
  }

  .hint-text {
    display: none;
  }
}
</style>

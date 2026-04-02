<script setup lang="ts">
import { ref } from 'vue';
import { useVbenModal } from '@vben/common-ui';
import { message, Textarea } from 'ant-design-vue';
import { pauseProject } from '#/api/mvp/workflow';

const emit = defineEmits<{ success: [] }>();

/** 暂停原因输入值 */
const pauseReason = ref('');
/** 当前操作的项目ID */
const projectID = ref('');

/** 弹窗配置 */
const [Modal, modalApi] = useVbenModal({
  fullscreenButton: false,
  title: '暂停项目',
  onCancel() {
    pauseReason.value = '';
    modalApi.close();
  },
  onConfirm: async () => {
    if (!pauseReason.value.trim()) {
      message.warning('请输入暂停原因');
      return;
    }
    modalApi.lock();
    try {
      await pauseProject({
        projectID: projectID.value,
        pauseReason: pauseReason.value.trim(),
      });
      message.success('项目已暂停');
      emit('success');
      pauseReason.value = '';
      modalApi.close();
    } catch {
      message.error('暂停操作失败，请稍后重试');
    } finally {
      modalApi.lock(false);
    }
  },
  async onOpenChange(isOpen: boolean) {
    if (isOpen) {
      const data = modalApi.getData<{ projectID: string } | null>();
      if (data?.projectID) {
        projectID.value = data.projectID;
      }
      pauseReason.value = '';
    }
  },
});
</script>

<template>
  <Modal class="w-[480px]">
    <div class="flex flex-col gap-3">
      <p class="text-gray-600 text-sm">请说明暂停项目的原因，便于后续恢复时了解上下文。</p>
      <Textarea
        v-model:value="pauseReason"
        :rows="4"
        placeholder="请输入暂停原因（如：需要调整需求、等待资源等）"
        :maxlength="500"
        show-count
      />
    </div>
  </Modal>
</template>

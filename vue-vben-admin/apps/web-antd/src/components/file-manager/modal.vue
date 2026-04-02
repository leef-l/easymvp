<script setup lang="ts">
import { ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';
import { Upload, Button, message } from 'ant-design-vue';
import { InboxOutlined } from '@ant-design/icons-vue';

import { uploadFile } from '../../api/upload';
import type { FileManagerItem } from './index.vue';

interface ModalData {
  mode?: 'all' | 'image';
  multiple?: boolean;
  maxCount?: number;
  accept?: string;
  maxSize?: number;
}

const emit = defineEmits<{
  confirm: [files: FileManagerItem[]];
}>();

const uploadedFiles = ref<FileManagerItem[]>([]);
const modalData = ref<ModalData>({});

const [Modal, modalApi] = useVbenModal({
  onOpenChange(isOpen: boolean) {
    if (isOpen) {
      const data = modalApi.getData<ModalData>();
      if (data) {
        modalData.value = data;
      }
      uploadedFiles.value = [];
    }
  },
});

async function handleUpload(options: any) {
  const { file } = options;

  if (modalData.value.maxSize) {
    const sizeMB = file.size / 1024 / 1024;
    if (sizeMB > modalData.value.maxSize) {
      message.error(`文件大小不能超过 ${modalData.value.maxSize}MB`);
      return;
    }
  }

  try {
    const res = await uploadFile(file);
    const item: FileManagerItem = {
      uid: `${Date.now()}-${Math.random()}`,
      url: res.url,
      name: file.name,
    };
    uploadedFiles.value.push(item);
  } catch {
    message.error('上传失败');
  }
}

function handleConfirm() {
  if (uploadedFiles.value.length === 0) {
    message.warning('请先上传文件');
    return;
  }
  emit('confirm', uploadedFiles.value);
  modalApi.close();
}
</script>

<template>
  <Modal title="文件管理" class="w-[520px]">
    <Upload.Dragger
      :custom-request="handleUpload"
      :multiple="modalData.multiple ?? false"
      :accept="modalData.accept || undefined"
      :show-upload-list="false"
    >
      <p class="ant-upload-drag-icon">
        <InboxOutlined />
      </p>
      <p class="ant-upload-text">点击或拖拽文件到此区域上传</p>
    </Upload.Dragger>

    <div v-if="uploadedFiles.length > 0" style="margin-top: 12px">
      <div
        v-for="file in uploadedFiles"
        :key="file.uid"
        style="
          padding: 4px 8px;
          margin-bottom: 4px;
          border: 1px solid #f0f0f0;
          border-radius: 4px;
          font-size: 13px;
        "
      >
        {{ file.name }}
      </div>
    </div>

    <template #footer>
      <Button @click="modalApi.close()">取消</Button>
      <Button type="primary" @click="handleConfirm">确认选择</Button>
    </template>
  </Modal>
</template>

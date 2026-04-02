<script setup lang="ts">
import { computed } from 'vue';

import { useVbenModal } from '@vben/common-ui';
import { Button } from 'ant-design-vue';

import FileManagerModal from '#/components/file-manager/modal.vue';
import type { FileManagerItem } from '#/components/file-manager/index.vue';

interface Props {
  value?: string;
  maxCount?: number;
  accept?: string;
  maxSize?: number;
  disabled?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  value: '',
  maxCount: 1,
  accept: '',
  maxSize: 10,
  disabled: false,
});

const emit = defineEmits<{
  'update:value': [val: string];
}>();

/** 解析 value 为文件列表 */
const fileItems = computed(() => {
  if (!props.value) return [];
  return props.value
    .split(',')
    .filter(Boolean)
    .map((url, i) => ({ uid: `${i}`, url, name: url.split('/').pop() || '' }));
});

/** 文件管理器 Modal */
const [PickerModal, pickerApi] = useVbenModal({
  connectedComponent: FileManagerModal,
});

function openPicker() {
  if (props.disabled) return;
  const remaining = props.maxCount - fileItems.value.length;
  if (remaining <= 0) return;
  pickerApi.setData({
    mode: 'all',
    multiple: remaining > 1,
    maxCount: remaining,
    accept: props.accept,
    maxSize: props.maxSize,
  });
  pickerApi.open();
}

function onPickerConfirm(files: FileManagerItem[]) {
  const currentUrls = props.value ? props.value.split(',').filter(Boolean) : [];
  const newUrls = [...currentUrls, ...files.map((f) => f.url)];
  emit('update:value', newUrls.join(','));
}

function removeFile(index: number) {
  const urls = props.value ? props.value.split(',').filter(Boolean) : [];
  urls.splice(index, 1);
  emit('update:value', urls.join(','));
}
</script>

<template>
  <div class="file-upload">
    <div v-if="fileItems.length > 0" class="file-upload__list">
      <div v-for="(file, index) in fileItems" :key="file.uid" class="file-upload__item">
        <a :href="file.url" target="_blank" class="file-upload__link" :title="file.name">
          {{ file.name }}
        </a>
        <span
          v-if="!disabled"
          class="file-upload__remove"
          @click="removeFile(index)"
        >
          ×
        </span>
      </div>
    </div>
    <Button
      v-if="fileItems.length < maxCount && !disabled"
      @click="openPicker"
    >
      选择文件
    </Button>
    <PickerModal @confirm="onPickerConfirm" />
  </div>
</template>

<style scoped>
.file-upload__list {
  margin-bottom: 8px;
}

.file-upload__item {
  display: flex;
  align-items: center;
  padding: 4px 8px;
  border: 1px solid #f0f0f0;
  border-radius: 4px;
  margin-bottom: 4px;
}

.file-upload__link {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: #1677ff;
  font-size: 13px;
}

.file-upload__remove {
  margin-left: 8px;
  cursor: pointer;
  color: #999;
  font-size: 14px;
}

.file-upload__remove:hover {
  color: #ff4d4f;
}
</style>

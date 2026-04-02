<script setup lang="ts">
import { computed, ref, watch } from 'vue';

import { useVbenModal } from '@vben/common-ui';
import { PlusOutlined } from '@ant-design/icons-vue';
import { Image } from 'ant-design-vue';

import FileManagerModal from '#/components/file-manager/modal.vue';
import type { FileManagerItem } from '#/components/file-manager/index.vue';

interface Props {
  value?: string;
  maxCount?: number;
  disabled?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  value: '',
  maxCount: 1,
  disabled: false,
});

const emit = defineEmits<{
  'update:value': [val: string];
}>();

/** 解析 value 为图片列表 */
const imageList = computed(() => {
  if (!props.value) return [];
  return props.value
    .split(',')
    .filter(Boolean)
    .map((url, i) => ({ uid: `${i}`, url, name: url.split('/').pop() || '' }));
});

const previewVisible = ref(false);
const previewUrl = ref('');

/** 文件管理器 Modal */
const [PickerModal, pickerApi] = useVbenModal({
  connectedComponent: FileManagerModal,
});

function openPicker() {
  if (props.disabled) return;
  const remaining = props.maxCount - imageList.value.length;
  if (remaining <= 0) return;
  pickerApi.setData({
    mode: 'image',
    multiple: remaining > 1,
    maxCount: remaining,
    accept: 'image/*',
  });
  pickerApi.open();
}

function onPickerConfirm(files: FileManagerItem[]) {
  const currentUrls = props.value ? props.value.split(',').filter(Boolean) : [];
  const newUrls = [...currentUrls, ...files.map((f) => f.url)];
  emit('update:value', newUrls.join(','));
}

function removeImage(index: number) {
  const urls = props.value ? props.value.split(',').filter(Boolean) : [];
  urls.splice(index, 1);
  emit('update:value', urls.join(','));
}

function handlePreview(url: string) {
  previewUrl.value = url;
  previewVisible.value = true;
}
</script>

<template>
  <div class="image-upload">
    <div class="image-upload__list">
      <div
        v-for="(img, index) in imageList"
        :key="img.uid"
        class="image-upload__item"
      >
        <img
          :src="img.url"
          :alt="img.name"
          class="image-upload__thumb"
          @click="handlePreview(img.url)"
        />
        <div
          v-if="!disabled"
          class="image-upload__remove"
          @click.stop="removeImage(index)"
        >
          ×
        </div>
      </div>
      <div
        v-if="imageList.length < maxCount && !disabled"
        class="image-upload__add"
        @click="openPicker"
      >
        <PlusOutlined style="font-size: 20px; color: #999" />
        <span class="image-upload__add-text">选择图片</span>
      </div>
    </div>
    <PickerModal @confirm="onPickerConfirm" />
    <Image
      :preview="{
        visible: previewVisible,
        onVisibleChange: (val: boolean) => (previewVisible = val),
      }"
      :src="previewUrl"
      :style="{ display: 'none' }"
    />
  </div>
</template>

<style scoped>
.image-upload__list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.image-upload__item {
  position: relative;
  width: 104px;
  height: 104px;
  border: 1px solid #d9d9d9;
  border-radius: 8px;
  overflow: hidden;
  cursor: pointer;
}

.image-upload__thumb {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.image-upload__remove {
  position: absolute;
  top: 2px;
  right: 2px;
  width: 20px;
  height: 20px;
  background: rgba(0, 0, 0, 0.5);
  color: #fff;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  font-size: 14px;
  line-height: 1;
}

.image-upload__add {
  width: 104px;
  height: 104px;
  border: 1px dashed #d9d9d9;
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: border-color 0.3s;
  gap: 4px;
}

.image-upload__add:hover {
  border-color: #1677ff;
}

.image-upload__add-text {
  font-size: 12px;
  color: #999;
}
</style>

<script setup lang="ts">
import { ref, watch } from 'vue';

import { uploadFile } from '#/api/upload/uploader';

// TinyMCE v8 core (theme is built-in, no separate themes/silver)
import 'tinymce';
import 'tinymce/models/dom';
import 'tinymce/icons/default';

// Plugins
import 'tinymce/plugins/advlist';
import 'tinymce/plugins/autolink';
import 'tinymce/plugins/charmap';
import 'tinymce/plugins/code';
import 'tinymce/plugins/directionality';
import 'tinymce/plugins/emoticons';
import 'tinymce/plugins/fullscreen';
import 'tinymce/plugins/image';
import 'tinymce/plugins/insertdatetime';
import 'tinymce/plugins/link';
import 'tinymce/plugins/lists';
import 'tinymce/plugins/media';
import 'tinymce/plugins/nonbreaking';
import 'tinymce/plugins/preview';
import 'tinymce/plugins/searchreplace';
import 'tinymce/plugins/table';
import 'tinymce/plugins/visualblocks';
import 'tinymce/plugins/visualchars';
import 'tinymce/plugins/wordcount';

import Editor from '@tinymce/tinymce-vue';

interface Props {
  value?: string;
  disabled?: boolean;
  height?: number;
}

const props = withDefaults(defineProps<Props>(), {
  value: '',
  disabled: false,
  height: 400,
});

const emit = defineEmits<{
  'update:value': [val: string];
}>();

const content = ref('');

watch(
  () => props.value,
  (val) => {
    if (val !== content.value) {
      content.value = val || '';
    }
  },
  { immediate: true },
);

function handleUpdate(val: string) {
  content.value = val;
  emit('update:value', val);
}

const plugins = [
  'advlist', 'autolink', 'charmap', 'code', 'directionality',
  'emoticons', 'fullscreen', 'image', 'insertdatetime', 'link',
  'lists', 'media', 'nonbreaking', 'preview', 'searchreplace',
  'table', 'visualblocks', 'visualchars', 'wordcount',
];

const toolbar =
  'undo redo | blocks fontfamily fontsize | bold italic underline strikethrough | ' +
  'alignleft aligncenter alignright alignjustify | outdent indent | ' +
  'bullist numlist | forecolor backcolor removeformat | ' +
  'table image media link charmap emoticons | ' +
  'searchreplace fullscreen preview code';

async function imagesUploadHandler(blobInfo: any): Promise<string> {
  const file = blobInfo.blob() as File;
  const res = await uploadFile(file);
  return res.url;
}

const initOptions = {
  height: props.height,
  plugins,
  toolbar,
  branding: false,
  promotion: false,
  menubar: 'file edit view insert format tools table',
  images_upload_handler: imagesUploadHandler,
  skin_url: `${import.meta.env.BASE_URL}tinymce/skins/ui/oxide`,
  content_css: `${import.meta.env.BASE_URL}tinymce/skins/content/default/content.min.css`,
};
</script>

<template>
  <Editor
    :disabled="disabled"
    :init="initOptions"
    :model-value="content"
    @update:model-value="handleUpdate"
  />
</template>

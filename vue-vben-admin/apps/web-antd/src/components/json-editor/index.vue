<script setup lang="ts">
import { ref, watch } from 'vue';

import JsonEditorVue from 'json-editor-vue';

interface Props {
  value?: string;
  height?: number;
  disabled?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  value: '',
  height: 300,
  disabled: false,
});

const emit = defineEmits<{
  'update:value': [val: string];
}>();

const jsonValue = ref<any>(undefined);

/** Parse JSON string to object, fallback to undefined */
function parseJson(str: string): any {
  if (!str) return undefined;
  try {
    return JSON.parse(str);
  } catch {
    return undefined;
  }
}

// Sync prop -> internal
watch(
  () => props.value,
  (val) => {
    const parsed = parseJson(val);
    if (JSON.stringify(parsed) !== JSON.stringify(jsonValue.value)) {
      jsonValue.value = parsed;
    }
  },
  { immediate: true },
);

// Sync internal -> emit
function handleChange(val: any) {
  jsonValue.value = val;
  try {
    const str = typeof val === 'string' ? val : JSON.stringify(val, null, 2);
    emit('update:value', str);
  } catch {
    // ignore stringify errors
  }
}
</script>

<template>
  <JsonEditorVue
    :model-value="jsonValue"
    :main-menu-bar="true"
    :mode="'tree'"
    :navigation-bar="false"
    :read-only="disabled"
    :style="{ height: `${height}px`, width: '100%' }"
    @update:model-value="handleChange"
  />
</template>

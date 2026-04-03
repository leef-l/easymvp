<script setup lang="ts">
import { ref } from 'vue';
import { Checkbox, message, Modal } from 'ant-design-vue';

import { getRoleAiEngineCodes, grantRoleAiEngine } from '#/api/system/role';

const emit = defineEmits<{ success: [] }>();

const visible = ref(false);
const confirmLoading = ref(false);
const roleId = ref('');
const checkedValues = ref<string[]>([]);

const engineOptions = [
  { label: 'Aider', value: 'aider', description: '本地命令行代码执行引擎' },
  { label: 'OpenHands', value: 'openhands', description: '远程 Agent 执行引擎' },
];

async function open(id: string) {
  visible.value = true;
  roleId.value = id;
  checkedValues.value = [];

  try {
    checkedValues.value = await getRoleAiEngineCodes(id);
  } catch {
    message.error('加载 AI 引擎权限失败');
  }
}

async function handleOk() {
  confirmLoading.value = true;
  try {
    await grantRoleAiEngine({
      id: roleId.value,
      engineCodes: checkedValues.value,
    });
    message.success('AI 引擎权限保存成功');
    visible.value = false;
    emit('success');
  } finally {
    confirmLoading.value = false;
  }
}

defineExpose({ open });
</script>

<template>
  <Modal
    v-model:open="visible"
    title="AI引擎权限"
    :confirm-loading="confirmLoading"
    width="520px"
    @ok="handleOk"
  >
    <div class="flex flex-col gap-3 py-2">
      <Checkbox.Group v-model:value="checkedValues" class="w-full">
        <div
          v-for="option in engineOptions"
          :key="option.value"
          class="mb-3 rounded border border-gray-200 px-3 py-3"
        >
          <Checkbox :value="option.value">
            {{ option.label }}
          </Checkbox>
          <div class="mt-1 pl-6 text-sm text-gray-500">
            {{ option.description }}
          </div>
        </div>
      </Checkbox.Group>
    </div>
  </Modal>
</template>

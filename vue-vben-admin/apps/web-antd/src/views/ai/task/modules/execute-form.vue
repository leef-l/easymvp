<script setup lang="ts">
import { useVbenForm } from '#/adapter/form';
import { useVbenModal } from '@vben/common-ui';
import { message } from 'ant-design-vue';

import { executeTask } from '#/api/ai/task';

const emit = defineEmits<{ success: [] }>();

const engineOptions = [
  { label: 'Aider', value: 'aider' },
  { label: 'OpenHands', value: 'openhands' },
];

const [Form, formApi] = useVbenForm({
  showDefaultActions: false,
  schema: [
    {
      component: 'Input',
      fieldName: 'title',
      label: '任务标题',
      rules: 'required',
    },
    {
      component: 'Select',
      fieldName: 'engineCode',
      label: '执行引擎',
      rules: 'required',
      componentProps: { options: engineOptions, class: 'w-full' },
    },
    {
      component: 'Input',
      fieldName: 'repoPath',
      label: '仓库路径',
      rules: 'required',
    },
    {
      component: 'Input',
      fieldName: 'worktreePath',
      label: '工作目录',
    },
    {
      component: 'Input',
      fieldName: 'branchName',
      label: '分支名',
    },
    {
      component: 'Textarea',
      fieldName: 'instruction',
      label: '任务指令',
      rules: 'required',
      componentProps: { rows: 6 },
    },
  ],
});

const [Modal, modalApi] = useVbenModal({
  fullscreenButton: false,
  onCancel() {
    modalApi.close();
  },
  onConfirm: async () => {
    const values = await formApi.validateAndSubmitForm();
    if (!values) return;
    modalApi.lock();
    try {
      await executeTask(values);
      message.success('任务创建成功');
      emit('success');
      modalApi.close();
    } finally {
      modalApi.lock(false);
    }
  },
  async onOpenChange(isOpen: boolean) {
    if (!isOpen) return;
    formApi.resetForm();
    formApi.setValues({
      engineCode: 'openhands',
    });
  },
});
</script>

<template>
  <Modal class="w-[720px]">
    <Form />
  </Modal>
</template>

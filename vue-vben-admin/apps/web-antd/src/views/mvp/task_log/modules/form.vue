<script setup lang="ts">
import type { TaskTreeOption } from '../../task-tree-options';

import { ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import { message } from 'ant-design-vue';

import { useVbenForm } from '#/adapter/form';
import { getTaskTree } from '#/api/mvp/task';
import {
  createTaskLog,
  getTaskLogDetail,
  updateTaskLog,
} from '#/api/mvp/task_log';

import { toTaskTreeOptions } from '../../task-tree-options';

const emit = defineEmits<{ success: [] }>();

const taskIDOptions = ref<TaskTreeOption[]>([]);

const isEdit = ref(false);
const editId = ref('');

/** 表单配置 */
const [Form, formApi] = useVbenForm({
  showDefaultActions: false,
  schema: [
    {
      component: 'TreeSelect',
      fieldName: 'taskID',
      label: '任务ID',
      rules: 'selectRequired',
      componentProps: {
        treeData: taskIDOptions.value,
        fieldNames: { label: 'label', value: 'value', children: 'children' },
        placeholder: '请选择任务ID',
        allowClear: true,
        treeDefaultExpandAll: true,
        class: 'w-full',
      },
    },
    {
      component: 'Input',
      fieldName: 'action',
      label: '动作',
      rules: 'required',
      componentProps: { placeholder: '请输入动作', maxlength: 50 },
    },
    {
      component: 'Input',
      fieldName: 'fromStatus',
      label: '原状态',
      componentProps: { placeholder: '请输入原状态', maxlength: 20 },
    },
    {
      component: 'Input',
      fieldName: 'toStatus',
      label: '新状态',
      componentProps: { placeholder: '请输入新状态', maxlength: 20 },
    },
    {
      component: 'Textarea',
      fieldName: 'message',
      label: '日志内容',
      componentProps: { placeholder: '请输入日志内容', rows: 4, maxlength: 65_535 },
    },
    {
      component: 'Input',
      fieldName: 'operator',
      label: '操作者',
      componentProps: { placeholder: '请输入操作者', maxlength: 50 },
    },
  ],
});

/** Modal 配置 */
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
      if (isEdit.value) {
        await updateTaskLog({ id: editId.value, ...values });
        message.success('更新成功');
      } else {
        await createTaskLog(values);
        message.success('创建成功');
      }
      emit('success');
      modalApi.close();
    } finally {
      modalApi.lock(false);
    }
  },
  async onOpenChange(isOpen: boolean) {
    if (isOpen) {
      const data = modalApi.getData<null | { id?: string }>();
      // 加载任务ID树形数据
      try {
        const taskRes = await getTaskTree();
        taskIDOptions.value = toTaskTreeOptions(taskRes ?? []);
        formApi.updateSchema([
          {
            fieldName: 'taskID',
            componentProps: { treeData: taskIDOptions.value },
          },
        ]);
      } catch {
        // ignore
      }
      if (data?.id) {
        isEdit.value = true;
        editId.value = data.id;
        modalApi.setState({ title: '编辑任务日志表' });
        try {
          const detail = await getTaskLogDetail(data.id);
          if (detail) {
            formApi.setValues(detail);
          }
        } catch {
          message.error('获取详情失败');
        }
      } else {
        isEdit.value = false;
        editId.value = '';
        modalApi.setState({ title: '新建任务日志表' });
        formApi.resetForm();
      }
    }
  },
});
</script>

<template>
  <Modal class="w-[600px]">
    <Form />
  </Modal>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useVbenModal } from '@vben/common-ui';
import { useVbenForm } from '#/adapter/form';
import { message } from 'ant-design-vue';
import {
  getConversationDetail,
  createConversation,
  updateConversation,
} from '#/api/mvp/conversation';
import { getProjectList } from '#/api/mvp/project';
import { getTaskTree } from '#/api/mvp/task';

const projectIDOptions = ref<{ label: string; value: string }[]>([]);
const taskIDOptions = ref<{ label: string; value: string }[]>([]);

const emit = defineEmits<{ success: [] }>();
const isEdit = ref(false);
const editId = ref('');

/** 表单配置 */
const [Form, formApi] = useVbenForm({
  showDefaultActions: false,
  schema: [
    {
      component: 'Select',
      fieldName: 'projectID',
      label: '项目ID',
      rules: 'selectRequired',
      componentProps: { options: projectIDOptions, placeholder: '请选择项目ID', allowClear: true, class: 'w-full' },
    },
    {
      component: 'TreeSelect',
      fieldName: 'taskID',
      label: '关联任务ID，NULL=项目级对话',
      componentProps: {
        treeData: taskIDOptions.value,
        fieldNames: { label: 'name', value: 'id', children: 'children' },
        placeholder: '请选择关联任务ID，NULL=项目级对话',
        allowClear: true,
        treeDefaultExpandAll: true,
        class: 'w-full',
      },
    },
    {
      component: 'Input',
      fieldName: 'title',
      label: '对话标题',
      componentProps: { placeholder: '请输入对话标题', maxlength: 200 },
    },
    {
      component: 'Input',
      fieldName: 'roleType',
      label: '对话角色类型',
      rules: 'required',
      componentProps: { placeholder: '请输入对话角色类型', maxlength: 20 },
    },
    {
      component: 'Switch',
      fieldName: 'status',
      label: '状态',
      componentProps: { checkedValue: 1, unCheckedValue: 0 },
      defaultValue: active,
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
        await updateConversation({ id: editId.value, ...values });
        message.success('更新成功');
      } else {
        await createConversation(values);
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
      const data = modalApi.getData<{ id?: string } | null>();
      // 加载项目ID选项
      try {
        const projectRes = await getProjectList({ pageNum: 1, pageSize: 1000 });
        projectIDOptions.value = (projectRes?.list ?? []).map((item: any) => ({
          label: item.name || item.id,
          value: item.id,
        }));
      } catch {
        // ignore
      }
      // 加载关联任务ID，NULL=项目级对话树形数据
      try {
        const taskRes = await getTaskTree();
        taskIDOptions.value = taskRes ?? [];
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
        modalApi.setState({ title: '编辑MVP对话表' });
        try {
          const detail = await getConversationDetail(data.id);
          if (detail) {
            formApi.setValues(detail);
          }
        } catch {
          message.error('获取详情失败');
        }
      } else {
        isEdit.value = false;
        editId.value = '';
        modalApi.setState({ title: '新建MVP对话表' });
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

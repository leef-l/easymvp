<script setup lang="ts">
import { ref } from 'vue';
import { useVbenModal } from '@vben/common-ui';
import { useVbenForm } from '#/adapter/form';
import { message } from 'ant-design-vue';
import {
  getMessageDetail,
  createMessage,
  updateMessage,
} from '#/api/mvp/message';
import { getConversationList } from '#/api/mvp/conversation';

const conversationIDOptions = ref<{ label: string; value: string }[]>([]);

const emit = defineEmits<{ success: [] }>();
const isEdit = ref(false);
const editId = ref('');

/** 表单配置 */
const [Form, formApi] = useVbenForm({
  showDefaultActions: false,
  schema: [
    {
      component: 'Select',
      fieldName: 'conversationID',
      label: '对话ID',
      rules: 'selectRequired',
      componentProps: { options: conversationIDOptions, placeholder: '请选择对话ID', allowClear: true, class: 'w-full' },
    },
    {
      component: 'Input',
      fieldName: 'role',
      label: '消息角色',
      rules: [
        { required: true, message: '消息角色不能为空' },
      ],
      componentProps: { placeholder: '请输入消息角色', maxlength: 20 },
    },
    {
      component: 'Textarea',
      fieldName: 'content',
      label: '消息内容',
      rules: 'required',
      componentProps: { placeholder: '请输入消息内容', rows: 4, maxlength: 4294967295 },
    },
    {
      component: 'Input',
      fieldName: 'modelID',
      label: '使用的AI模型ID',
      componentProps: { placeholder: '请输入使用的AI模型ID' },
    },
    {
      component: 'Input',
      fieldName: 'tokenUsage',
      label: 'token消耗',
      componentProps: { placeholder: '请输入token消耗' },
    },
    {
      component: 'Switch',
      fieldName: 'status',
      label: '状态',
      componentProps: { checkedValue: 1, unCheckedValue: 0 },
      defaultValue: completed,
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
        await updateMessage({ id: editId.value, ...values });
        message.success('更新成功');
      } else {
        await createMessage(values);
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
      // 加载对话ID选项
      try {
        const conversationRes = await getConversationList({ pageNum: 1, pageSize: 1000 });
        conversationIDOptions.value = (conversationRes?.list ?? []).map((item: any) => ({
          label: item.title || item.id,
          value: item.id,
        }));
      } catch {
        // ignore
      }
      if (data?.id) {
        isEdit.value = true;
        editId.value = data.id;
        modalApi.setState({ title: '编辑MVP消息表' });
        try {
          const detail = await getMessageDetail(data.id);
          if (detail) {
            formApi.setValues(detail);
          }
        } catch {
          message.error('获取详情失败');
        }
      } else {
        isEdit.value = false;
        editId.value = '';
        modalApi.setState({ title: '新建MVP消息表' });
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

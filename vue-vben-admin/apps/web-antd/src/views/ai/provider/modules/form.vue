<script setup lang="ts">
import { ref } from 'vue';
import { useVbenModal } from '@vben/common-ui';
import { useVbenForm } from '#/adapter/form';
import { message } from 'ant-design-vue';
import {
  getProviderDetail,
  createProvider,
  updateProvider,
} from '#/api/ai/provider';

const emit = defineEmits<{ success: [] }>();
const isEdit = ref(false);
const editId = ref('');

/** 供应商类型选项 */
const providerTypeOptions = [
  { label: 'OpenAI 兼容', value: 'openai_compatible' },
  { label: 'Anthropic', value: 'anthropic' },
  { label: 'Google', value: 'google' },
];

/** 表单配置 */
const [Form, formApi] = useVbenForm({
  showDefaultActions: false,
  schema: [
    {
      component: 'Input',
      fieldName: 'name',
      label: '供应商名称',
      rules: [
        { required: true, message: '供应商名称不能为空' },
      ],
      componentProps: { placeholder: '请输入供应商名称，如：OpenAI', maxlength: 100 },
    },
    {
      component: 'Input',
      fieldName: 'code',
      label: '供应商代码',
      rules: [
        { required: true, message: '供应商代码不能为空' },
      ],
      componentProps: { placeholder: '请输入唯一标识代码，如：openai', maxlength: 50 },
    },
    {
      component: 'Select',
      fieldName: 'providerType',
      label: '供应商类型',
      rules: 'selectRequired',
      componentProps: {
        options: providerTypeOptions,
        placeholder: '请选择供应商类型',
        allowClear: true,
        class: 'w-full',
      },
    },
    {
      component: 'Input',
      fieldName: 'baseURL',
      label: 'Base URL',
      rules: [
        { required: true, message: 'Base URL 不能为空' },
      ],
      componentProps: { placeholder: '请输入 API 基础地址，如：https://api.openai.com/v1', maxlength: 500 },
    },
    {
      component: 'Switch',
      fieldName: 'status',
      label: '状态',
      componentProps: { checkedValue: 1, unCheckedValue: 0, checkedChildren: '启用', unCheckedChildren: '禁用' },
      defaultValue: 1,
    },
    {
      component: 'InputNumber',
      fieldName: 'sort',
      label: '排序',
      componentProps: { placeholder: '数字越小越靠前', class: 'w-full', min: 0 },
      defaultValue: 0,
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
        await updateProvider({ id: editId.value, ...values });
        message.success('更新成功');
      } else {
        await createProvider(values);
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
      if (data?.id) {
        isEdit.value = true;
        editId.value = data.id;
        modalApi.setState({ title: '编辑供应商' });
        try {
          const detail = await getProviderDetail(data.id);
          if (detail) {
            formApi.setValues(detail);
          }
        } catch {
          message.error('获取详情失败');
        }
      } else {
        isEdit.value = false;
        editId.value = '';
        modalApi.setState({ title: '新建供应商' });
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

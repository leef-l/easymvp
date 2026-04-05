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

/** 表单配置 */
const [Form, formApi] = useVbenForm({
  showDefaultActions: false,
  schema: [
    {
      component: 'Input',
      fieldName: 'name',
      label: '供应商名称',
      rules: 'required',
      componentProps: { placeholder: '请输入供应商名称', maxlength: 100 },
    },
    {
      component: 'Input',
      fieldName: 'code',
      label: '供应商代码',
      rules: 'required',
      componentProps: { placeholder: '请输入供应商代码', maxlength: 50 },
    },
    {
      component: 'Select',
      fieldName: 'providerType',
      label: 'Provider 类型',
      rules: 'required',
      componentProps: {
        showSearch: true,
        allowClear: true,
        mode: 'combobox',
        filterOption: false,
        placeholder: '请选择或输入 Provider 类型',
        options: [
          { label: 'OpenAI', value: 'openai' },
          { label: 'Anthropic', value: 'anthropic' },
          { label: 'Google Gemini', value: 'google' },
          { label: 'DeepSeek', value: 'deepseek' },
          { label: '阿里云百炼 (Qwen)', value: 'qwen' },
          { label: '智谱 GLM', value: 'zhipu' },
          { label: 'Moonshot (Kimi)', value: 'moonshot' },
          { label: 'MiniMax', value: 'minimax' },
          { label: '百度文心', value: 'baidu' },
          { label: '腾讯混元', value: 'tencent' },
          { label: '字节豆包', value: 'bytedance' },
          { label: '零一万物', value: '01ai' },
          { label: 'Mistral', value: 'mistral' },
          { label: 'Groq', value: 'groq' },
          { label: 'Cohere', value: 'cohere' },
          { label: '腾讯云 Coding Plan', value: 'tencent_coding' },
          { label: '阿里云 Coding Plan', value: 'aliyun_coding' },
          { label: '百度 Coding Plan', value: 'baidu_coding' },
        ],
      },
    },
    {
      component: 'Input',
      fieldName: 'baseURL',
      label: 'API基础地址',
      rules: 'required',
      componentProps: { placeholder: '请输入URL地址', maxlength: 500, addonBefore: 'https://' },
    },
    {
      component: 'IconPicker',
      fieldName: 'icon',
      label: '图标URL',
      componentProps: { placeholder: '请选择图标' },
    },
    {
      component: 'Switch',
      fieldName: 'status',
      label: '状态',
      componentProps: { checkedValue: 1, unCheckedValue: 0 },
      defaultValue: 1,
    },
    {
      component: 'InputNumber',
      fieldName: 'sort',
      label: '排序',
      componentProps: { placeholder: '请输入排序', class: 'w-full' },
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
        modalApi.setState({ title: '编辑AI供应商表' });
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
        modalApi.setState({ title: '新建AI供应商表' });
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

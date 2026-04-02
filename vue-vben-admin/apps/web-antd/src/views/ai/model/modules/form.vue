<script setup lang="ts">
import { ref } from 'vue';
import { useVbenModal } from '@vben/common-ui';
import { useVbenForm } from '#/adapter/form';
import { message } from 'ant-design-vue';
import {
  getModelDetail,
  createModel,
  updateModel,
} from '#/api/ai/model';
import { getPlanList } from '#/api/ai/plan';
import { getProviderList } from '#/api/ai/provider';

const emit = defineEmits<{ success: [] }>();
const isEdit = ref(false);
const editId = ref('');

const planOptions = ref<{ label: string; value: string }[]>([]);
const providerOptions = ref<{ label: string; value: string }[]>([]);

/** 表单配置 */
const [Form, formApi] = useVbenForm({
  showDefaultActions: false,
  schema: [
    {
      component: 'Input',
      fieldName: 'name',
      label: '模型名称',
      rules: [
        { required: true, message: '模型名称不能为空' },
      ],
      componentProps: { placeholder: '请输入模型显示名称，如：GPT-4o', maxlength: 100 },
    },
    {
      component: 'Input',
      fieldName: 'modelCode',
      label: '模型代码',
      rules: [
        { required: true, message: '模型代码不能为空' },
      ],
      componentProps: { placeholder: '请输入 API 调用代码，如：gpt-4o', maxlength: 100 },
    },
    {
      component: 'Select',
      fieldName: 'planID',
      label: '所属套餐',
      rules: 'selectRequired',
      componentProps: {
        options: planOptions,
        placeholder: '请选择套餐',
        allowClear: true,
        class: 'w-full',
        showSearch: true,
        optionFilterProp: 'label',
      },
    },
    {
      component: 'Select',
      fieldName: 'providerID',
      label: '所属供应商',
      rules: 'selectRequired',
      componentProps: {
        options: providerOptions,
        placeholder: '请选择供应商',
        allowClear: true,
        class: 'w-full',
        showSearch: true,
        optionFilterProp: 'label',
      },
    },
    {
      component: 'InputNumber',
      fieldName: 'maxTokens',
      label: '最大 Tokens',
      componentProps: {
        placeholder: '请输入最大输出 token 数',
        class: 'w-full',
        min: 1,
        max: 2000000,
      },
    },
    {
      component: 'InputNumber',
      fieldName: 'contextWindow',
      label: '上下文窗口',
      componentProps: {
        placeholder: '请输入上下文窗口大小',
        class: 'w-full',
        min: 1,
        max: 10000000,
      },
    },
    {
      component: 'Input',
      fieldName: 'capability',
      label: '模型能力',
      componentProps: { placeholder: '如：chat, vision, embedding', maxlength: 20 },
    },
    {
      component: 'Select',
      fieldName: 'supportsStream',
      label: '支持流式输出',
      componentProps: {
        options: [
          { label: '支持', value: 1 },
          { label: '不支持', value: 0 },
        ],
        placeholder: '请选择',
        allowClear: true,
        class: 'w-full',
      },
      defaultValue: 1,
    },
    {
      component: 'Textarea',
      fieldName: 'rolePrompt',
      label: '默认角色提示词',
      componentProps: {
        placeholder: '请输入默认的系统角色提示词（System Prompt）',
        rows: 4,
        maxlength: 65535,
        showCount: true,
      },
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
        await updateModel({ id: editId.value, ...values });
        message.success('更新成功');
      } else {
        await createModel(values);
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
      // 并行加载套餐和供应商选项
      await Promise.allSettled([
        getPlanList({ pageNum: 1, pageSize: 1000 }).then((res) => {
          planOptions.value = (res?.list ?? []).map((item: any) => ({
            label: item.name || item.id,
            value: item.id,
          }));
        }),
        getProviderList({ pageNum: 1, pageSize: 1000 }).then((res) => {
          providerOptions.value = (res?.list ?? []).map((item: any) => ({
            label: item.name || item.id,
            value: item.id,
          }));
        }),
      ]);
      if (data?.id) {
        isEdit.value = true;
        editId.value = data.id;
        modalApi.setState({ title: '编辑模型' });
        try {
          const detail = await getModelDetail(data.id);
          if (detail) {
            formApi.setValues(detail);
          }
        } catch {
          message.error('获取详情失败');
        }
      } else {
        isEdit.value = false;
        editId.value = '';
        modalApi.setState({ title: '新建模型' });
        formApi.resetForm();
      }
    }
  },
});
</script>

<template>
  <Modal class="w-[640px]">
    <Form />
  </Modal>
</template>

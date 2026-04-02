<script setup lang="ts">
import { ref } from 'vue';
import { useVbenModal } from '@vben/common-ui';
import { useVbenForm } from '#/adapter/form';
import { message } from 'ant-design-vue';
import {
  getPlanDetail,
  createPlan,
  updatePlan,
} from '#/api/ai/plan';
import { getProviderList } from '#/api/ai/provider';

const emit = defineEmits<{ success: [] }>();
const isEdit = ref(false);
const editId = ref('');

const providerOptions = ref<{ label: string; value: string }[]>([]);

/** 表单配置 */
const [Form, formApi] = useVbenForm({
  showDefaultActions: false,
  schema: [
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
      component: 'Input',
      fieldName: 'name',
      label: '套餐名称',
      rules: [
        { required: true, message: '套餐名称不能为空' },
      ],
      componentProps: { placeholder: '请输入套餐名称，如：GPT-4 官方', maxlength: 100 },
    },
    {
      component: 'Input',
      fieldName: 'code',
      label: '套餐代码',
      rules: [
        { required: true, message: '套餐代码不能为空' },
      ],
      componentProps: { placeholder: '请输入唯一标识代码', maxlength: 50 },
    },
    {
      component: 'InputPassword',
      fieldName: 'apiKey',
      label: 'API Key',
      componentProps: {
        placeholder: '请输入 API Key（加密存储）',
        maxlength: 500,
        autocomplete: 'new-password',
      },
    },
    {
      component: 'InputPassword',
      fieldName: 'apiSecret',
      label: 'API Secret',
      componentProps: {
        placeholder: '请输入 API Secret（部分供应商需要）',
        maxlength: 500,
        autocomplete: 'new-password',
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
        await updatePlan({ id: editId.value, ...values });
        message.success('更新成功');
      } else {
        await createPlan(values);
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
      // 加载供应商选项
      try {
        const providerRes = await getProviderList({ pageNum: 1, pageSize: 1000 });
        providerOptions.value = (providerRes?.list ?? []).map((item: any) => ({
          label: item.name || item.id,
          value: item.id,
        }));
      } catch {
        // ignore
      }
      if (data?.id) {
        isEdit.value = true;
        editId.value = data.id;
        modalApi.setState({ title: '编辑套餐' });
        try {
          const detail = await getPlanDetail(data.id);
          if (detail) {
            formApi.setValues(detail);
          }
        } catch {
          message.error('获取详情失败');
        }
      } else {
        isEdit.value = false;
        editId.value = '';
        modalApi.setState({ title: '新建套餐' });
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

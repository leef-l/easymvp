<script setup lang="ts">
import { h, ref } from 'vue';
import { useVbenModal } from '@vben/common-ui';
import { useVbenForm } from '#/adapter/form';
import { message, Tooltip } from 'ant-design-vue';
import { QuestionCircleOutlined } from '@ant-design/icons-vue';
import {
  getPlanDetail,
  createPlan,
  updatePlan,
} from '#/api/ai/plan';
import { getProviderList } from '#/api/ai/provider';

const providerIDOptions = ref<{ label: string; value: string }[]>([]);
/** 渲染带 Tooltip 的表单 label */
function tooltipLabel(label: string, tip: string) {
  return () => h('span', {}, [
    label + ' ',
    h(Tooltip, { title: tip }, {
      default: () => h(QuestionCircleOutlined, { style: { color: '#999', marginLeft: '4px' } }),
    }),
  ]);
}

const emit = defineEmits<{ success: [] }>();
const isEdit = ref(false);
const editId = ref('');

/** 表单配置 */
const [Form, formApi] = useVbenForm({
  showDefaultActions: false,
  schema: [
    {
      component: 'Select',
      fieldName: 'providerID',
      label: '供应商ID',
      rules: 'selectRequired',
      componentProps: { options: providerIDOptions, placeholder: '请选择供应商ID', allowClear: true, class: 'w-full' },
    },
    {
      component: 'Input',
      fieldName: 'name',
      label: '套餐名称',
      rules: 'required',
      componentProps: { placeholder: '请输入套餐名称', maxlength: 100 },
    },
    {
      component: 'Input',
      fieldName: 'code',
      label: '套餐代码',
      rules: 'required',
      componentProps: { placeholder: '请输入套餐代码', maxlength: 50 },
    },
    {
      component: 'Input',
      fieldName: 'apiKey',
      label: tooltipLabel('API Key', '加密存储'),
      componentProps: { placeholder: '请输入API Key（加密存储）', maxlength: 500 },
    },
    {
      component: 'Input',
      fieldName: 'apiSecret',
      label: tooltipLabel('API Secret', '部分供应商需要'),
      componentProps: { placeholder: '请输入API Secret（部分供应商需要）', maxlength: 500 },
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
        await updatePlan({ id: editId.value, ...values });
        message.success('更新成功');
      } else {
        const res = await createPlan(values);
        const initCount = (res as any)?.initCount ?? 0;
        message.success(initCount > 0 ? `套餐创建成功，已自动初始化 ${initCount} 个 AI 模型` : '创建成功');
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
      // 加载供应商ID选项
      try {
        const providerRes = await getProviderList({ pageNum: 1, pageSize: 1000 });
        providerIDOptions.value = (providerRes?.list ?? []).map((item: any) => ({
          label: item.name || item.id,
          value: item.id,
        }));
      } catch {
        // ignore
      }
      if (data?.id) {
        isEdit.value = true;
        editId.value = data.id;
        modalApi.setState({ title: '编辑AI套餐表' });
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
        modalApi.setState({ title: '新建AI套餐表' });
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

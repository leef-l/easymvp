<script setup lang="ts">
import { h, ref } from 'vue';
import { useVbenModal } from '@vben/common-ui';
import { useVbenForm } from '#/adapter/form';
import { message, Tooltip } from 'ant-design-vue';
import { QuestionCircleOutlined } from '@ant-design/icons-vue';
import {
  getModelDetail,
  createModel,
  updateModel,
} from '#/api/ai/model';
import { getPlanList } from '#/api/ai/plan';
import { getProviderList } from '#/api/ai/provider';

/** 是否支持流式输出选项 */
const supportsStreamOptions = [
  { label: '否', value: 0 },
  { label: '是', value: 1 },
];

const planIDOptions = ref<{ label: string; value: string }[]>([]);
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
      fieldName: 'planID',
      label: '套餐ID',
      rules: 'selectRequired',
      componentProps: { options: planIDOptions, placeholder: '请选择套餐ID', allowClear: true, class: 'w-full' },
    },
    {
      component: 'Select',
      fieldName: 'providerID',
      label: tooltipLabel('供应商ID', '冗余便于查询'),
      rules: 'selectRequired',
      componentProps: { options: providerIDOptions, placeholder: '请选择供应商ID（冗余便于查询）', allowClear: true, class: 'w-full' },
    },
    {
      component: 'Input',
      fieldName: 'name',
      label: '模型显示名称',
      rules: 'required',
      componentProps: { placeholder: '请输入模型显示名称', maxlength: 100 },
    },
    {
      component: 'Input',
      fieldName: 'modelCode',
      label: tooltipLabel('模型代码', 'API调用用'),
      rules: 'required',
      componentProps: { placeholder: '请输入模型代码（API调用用）', maxlength: 100 },
    },
    {
      component: 'Input',
      fieldName: 'capability',
      label: '能力',
      componentProps: { placeholder: '请输入能力', maxlength: 20 },
    },
    {
      component: 'Input',
      fieldName: 'maxTokens',
      label: '最大输出token',
      componentProps: { placeholder: '请输入最大输出token' },
    },
    {
      component: 'Input',
      fieldName: 'contextWindow',
      label: '上下文窗口大小',
      componentProps: { placeholder: '请输入上下文窗口大小' },
    },
    {
      component: 'Select',
      fieldName: 'supportsStream',
      label: '是否支持流式输出',
      componentProps: { options: supportsStreamOptions, placeholder: '请选择是否支持流式输出', allowClear: true, class: 'w-full' },
    },
    {
      component: 'Textarea',
      fieldName: 'rolePrompt',
      label: '默认角色提示词',
      componentProps: { placeholder: '请输入默认角色提示词', rows: 4, maxlength: 65535 },
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
      // 加载套餐ID选项
      try {
        const planRes = await getPlanList({ pageNum: 1, pageSize: 1000 });
        planIDOptions.value = (planRes?.list ?? []).map((item: any) => ({
          label: item.name || item.id,
          value: item.id,
        }));
      } catch {
        // ignore
      }
      // 加载供应商ID（冗余便于查询）选项
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
        modalApi.setState({ title: '编辑AI模型表' });
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
        modalApi.setState({ title: '新建AI模型表' });
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

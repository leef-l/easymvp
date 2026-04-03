<script setup lang="ts">
import { h, ref } from 'vue';
import { useVbenModal } from '@vben/common-ui';
import { useVbenForm } from '#/adapter/form';
import { message, Tooltip } from 'ant-design-vue';
import { QuestionCircleOutlined } from '@ant-design/icons-vue';
import {
  getConfigDetail,
  createConfig,
  updateConfig,
} from '#/api/mvp/config';

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
      component: 'Input',
      fieldName: 'configKey',
      label: tooltipLabel('配置键', '唯一'),
      rules: 'required',
      componentProps: { placeholder: '请输入配置键（唯一）', maxlength: 100 },
    },
    {
      component: 'Textarea',
      fieldName: 'configValue',
      label: '配置值',
      rules: 'required',
      componentProps: { placeholder: '请输入配置值', rows: 4, maxlength: 65535 },
    },
    {
      component: 'Input',
      fieldName: 'configType',
      label: '值类型',
      componentProps: { placeholder: '请输入值类型', maxlength: 20 },
    },
    {
      component: 'Input',
      fieldName: 'category',
      label: '分类',
      componentProps: { placeholder: '请输入分类', maxlength: 50 },
    },
    {
      component: 'Input',
      fieldName: 'description',
      label: '配置说明',
      componentProps: { placeholder: '请输入配置说明', maxlength: 255 },
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
        await updateConfig({ id: editId.value, ...values });
        message.success('更新成功');
      } else {
        await createConfig(values);
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
        modalApi.setState({ title: '编辑MVP配置表' });
        try {
          const detail = await getConfigDetail(data.id);
          if (detail) {
            formApi.setValues(detail);
          }
        } catch {
          message.error('获取详情失败');
        }
      } else {
        isEdit.value = false;
        editId.value = '';
        modalApi.setState({ title: '新建MVP配置表' });
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

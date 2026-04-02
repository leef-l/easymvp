<script setup lang="ts">
import { h, ref } from 'vue';
import { useVbenModal } from '@vben/common-ui';
import { useVbenForm } from '#/adapter/form';
import { message, Tooltip } from 'ant-design-vue';
import { QuestionCircleOutlined } from '@ant-design/icons-vue';
import {
  getProjectDetail,
  createProject,
  updateProject,
} from '#/api/mvp/project';

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
      fieldName: 'name',
      label: '项目名称',
      rules: 'required',
      componentProps: { placeholder: '请输入项目名称', maxlength: 200 },
    },
    {
      component: 'Textarea',
      fieldName: 'description',
      label: '项目简介',
      componentProps: { placeholder: '请输入项目简介', rows: 4, maxlength: 65535 },
    },
    {
      component: 'Switch',
      fieldName: 'status',
      label: '状态',
      componentProps: { checkedValue: 1, unCheckedValue: 0 },
      defaultValue: 'designing',
    },
    {
      component: 'Textarea',
      fieldName: 'pauseReason',
      label: '暂停原因',
      componentProps: { placeholder: '请输入暂停原因', rows: 4, maxlength: 65535 },
    },
    {
      component: 'Textarea',
      fieldName: 'globalContext',
      label: tooltipLabel('项目全局上下文', '架构师需求分析+方案设计的压缩摘要'),
      componentProps: { placeholder: '请输入项目全局上下文（架构师需求分析+方案设计的压缩摘要）', rows: 4, maxlength: 65535 },
    },
    {
      component: 'Input',
      fieldName: 'architectModelID',
      label: '架构师使用的AI模型ID',
      componentProps: { placeholder: '请输入架构师使用的AI模型ID' },
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
        await updateProject({ id: editId.value, ...values });
        message.success('更新成功');
      } else {
        await createProject(values);
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
        modalApi.setState({ title: '编辑MVP项目表' });
        try {
          const detail = await getProjectDetail(data.id);
          if (detail) {
            formApi.setValues(detail);
          }
        } catch {
          message.error('获取详情失败');
        }
      } else {
        isEdit.value = false;
        editId.value = '';
        modalApi.setState({ title: '新建MVP项目表' });
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

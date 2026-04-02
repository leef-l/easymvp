<script setup lang="ts">
import { h, ref } from 'vue';
import { useRouter } from 'vue-router';
import { useVbenModal } from '@vben/common-ui';
import { useVbenForm } from '#/adapter/form';
import { message, Tooltip } from 'ant-design-vue';
import { QuestionCircleOutlined } from '@ant-design/icons-vue';
import {
  getProjectDetail,
  updateProject,
} from '#/api/mvp/project';
import { createProject as workflowCreateProject } from '#/api/mvp/workflow';
import { getModelList } from '#/api/ai/model';

const router = useRouter();
const modelIDOptions = ref<{ label: string; value: string }[]>([]);

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

/** 新建时的表单字段 */
const createSchema = [
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
    component: 'Select',
    fieldName: 'architectModelID',
    label: '架构师AI模型',
    rules: 'selectRequired',
    componentProps: { options: modelIDOptions, placeholder: '请选择架构师使用的AI模型', allowClear: true, showSearch: true, optionFilterProp: 'label', class: 'w-full' },
  },
];

/** 编辑时的表单字段 */
const editSchema = [
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
    component: 'Select',
    fieldName: 'architectModelID',
    label: '架构师AI模型',
    componentProps: { options: modelIDOptions, placeholder: '请选择架构师使用的AI模型', allowClear: true, showSearch: true, optionFilterProp: 'label', class: 'w-full' },
  },
];

/** 表单配置 */
const [Form, formApi] = useVbenForm({
  showDefaultActions: false,
  schema: createSchema,
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
        emit('success');
        modalApi.close();
      } else {
        // 调用工作流 API 创建项目（会自动初始化架构师角色和对话）
        const res = await workflowCreateProject({
          name: values.name,
          description: values.description || '',
          architectModelID: values.architectModelID,
        });
        message.success('项目创建成功，正在跳转到对话页面...');
        emit('success');
        modalApi.close();
        // 跳转到对话页面
        router.push({
          path: '/mvp/chat',
          query: {
            projectId: res.projectID,
            conversationId: res.conversationID,
          },
        });
      }
    } finally {
      modalApi.lock(false);
    }
  },
  async onOpenChange(isOpen: boolean) {
    if (isOpen) {
      // 加载AI模型选项
      try {
        const modelRes = await getModelList({ pageNum: 1, pageSize: 1000 });
        modelIDOptions.value = (modelRes?.list ?? []).map((item: any) => ({
          label: `${item.name} (${item.modelCode})`,
          value: item.id,
        }));
      } catch {
        // ignore
      }
      const data = modalApi.getData<{ id?: string } | null>();
      if (data?.id) {
        isEdit.value = true;
        editId.value = data.id;
        modalApi.setState({ title: '编辑项目' });
        formApi.setState({ schema: editSchema });
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
        modalApi.setState({ title: '新建项目' });
        formApi.setState({ schema: createSchema });
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

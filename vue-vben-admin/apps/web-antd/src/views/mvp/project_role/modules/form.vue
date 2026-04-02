<script setup lang="ts">
import { h, ref } from 'vue';
import { useVbenModal } from '@vben/common-ui';
import { useVbenForm } from '#/adapter/form';
import { message, Tooltip } from 'ant-design-vue';
import { QuestionCircleOutlined } from '@ant-design/icons-vue';
import {
  getProjectRoleDetail,
  createProjectRole,
  updateProjectRole,
} from '#/api/mvp/project_role';
import { getProjectList } from '#/api/mvp/project';

const projectIDOptions = ref<{ label: string; value: string }[]>([]);
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
      fieldName: 'projectID',
      label: '项目ID',
      rules: 'selectRequired',
      componentProps: { options: projectIDOptions, placeholder: '请选择项目ID', allowClear: true, class: 'w-full' },
    },
    {
      component: 'Input',
      fieldName: 'roleType',
      label: '角色类型',
      rules: 'required',
      componentProps: { placeholder: '请输入角色类型', maxlength: 20 },
    },
    {
      component: 'Input',
      fieldName: 'roleLevel',
      label: '角色等级',
      componentProps: { placeholder: '请输入角色等级', maxlength: 10 },
    },
    {
      component: 'Input',
      fieldName: 'modelID',
      label: 'AI模型ID',
      rules: 'required',
      componentProps: { placeholder: '请输入AI模型ID' },
    },
    {
      component: 'Textarea',
      fieldName: 'systemPrompt',
      label: tooltipLabel('系统提示词', '角色设定'),
      componentProps: { placeholder: '请输入系统提示词（角色设定）', rows: 4, maxlength: 65535 },
    },
    {
      component: 'Switch',
      fieldName: 'status',
      label: '状态',
      componentProps: { checkedValue: 1, unCheckedValue: 0 },
      defaultValue: 1,
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
        await updateProjectRole({ id: editId.value, ...values });
        message.success('更新成功');
      } else {
        await createProjectRole(values);
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
      // 加载项目ID选项
      try {
        const projectRes = await getProjectList({ pageNum: 1, pageSize: 1000 });
        projectIDOptions.value = (projectRes?.list ?? []).map((item: any) => ({
          label: item.name || item.id,
          value: item.id,
        }));
      } catch {
        // ignore
      }
      if (data?.id) {
        isEdit.value = true;
        editId.value = data.id;
        modalApi.setState({ title: '编辑项目角色配置表' });
        try {
          const detail = await getProjectRoleDetail(data.id);
          if (detail) {
            formApi.setValues(detail);
          }
        } catch {
          message.error('获取详情失败');
        }
      } else {
        isEdit.value = false;
        editId.value = '';
        modalApi.setState({ title: '新建项目角色配置表' });
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

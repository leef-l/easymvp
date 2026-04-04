<script setup lang="ts">
import { h, ref } from 'vue';
import { useVbenModal } from '@vben/common-ui';
import { useVbenForm } from '#/adapter/form';
import { message, Tooltip } from 'ant-design-vue';
import { QuestionCircleOutlined } from '@ant-design/icons-vue';
import {
  getRolePresetDetail,
  createRolePreset,
  updateRolePreset,
} from '#/api/mvp/role_preset';
import { getModelList } from '#/api/ai/model';
import { roleTypeOptions, roleLevelOptions, projectCategoryOptions, executionModeOptions } from '../../consts';

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

/** 表单配置 */
const [Form, formApi] = useVbenForm({
  showDefaultActions: false,
  schema: [
    {
      component: 'Select',
      fieldName: 'projectCategory',
      label: '项目分类',
      rules: 'selectRequired',
      componentProps: { options: projectCategoryOptions, placeholder: '请选择项目分类', allowClear: true, class: 'w-full' },
      defaultValue: '软件开发',
    },
    {
      component: 'Select',
      fieldName: 'roleType',
      label: '角色类型',
      rules: 'selectRequired',
      componentProps: { options: roleTypeOptions, placeholder: '请选择角色类型', allowClear: true, class: 'w-full' },
    },
    {
      component: 'Select',
      fieldName: 'roleLevel',
      label: '角色等级',
      componentProps: { options: roleLevelOptions, placeholder: '请选择角色等级', allowClear: true, class: 'w-full' },
    },
    {
      component: 'Select',
      fieldName: 'modelID',
      label: 'AI模型',
      rules: 'selectRequired',
      componentProps: { options: modelIDOptions, placeholder: '请选择AI模型', allowClear: true, showSearch: true, optionFilterProp: 'label', class: 'w-full' },
    },
    {
      component: 'Textarea',
      fieldName: 'systemPrompt',
      label: tooltipLabel('默认系统提示词', '角色设定，创建项目时架构师提示词会根据项目名称动态生成'),
      componentProps: { placeholder: '请输入默认系统提示词（角色设定）', rows: 4, maxlength: 65535 },
    },
    {
      component: 'Select',
      fieldName: 'executionMode',
      label: tooltipLabel('执行方式', '决定该角色用什么方式执行任务：Chat对话、Aider代码编辑、OpenHands沙箱'),
      componentProps: { options: executionModeOptions, placeholder: '请选择执行方式', allowClear: true, class: 'w-full' },
      defaultValue: 'chat',
    },
    {
      component: 'Switch',
      fieldName: 'status',
      label: '状态',
      componentProps: { checkedValue: '1', unCheckedValue: '0' },
      defaultValue: '1',
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
        await updateRolePreset({ id: editId.value, ...values });
        message.success('更新成功');
      } else {
        await createRolePreset(values);
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
      // 加载AI模型选项
      try {
        const modelRes = await getModelList({ pageNum: 1, pageSize: 1000 });
        modelIDOptions.value = (modelRes?.list ?? []).map((item: any) => ({
          label: `${item.name} (${item.modelCode})`,
          value: item.id,
        }));
      } catch { /* ignore */ }
      const data = modalApi.getData<{ id?: string } | null>();
      if (data?.id) {
        isEdit.value = true;
        editId.value = data.id;
        modalApi.setState({ title: '编辑角色预设' });
        try {
          const detail = await getRolePresetDetail(data.id);
          if (detail) {
            formApi.setValues(detail);
          }
        } catch {
          message.error('获取详情失败');
        }
      } else {
        isEdit.value = false;
        editId.value = '';
        modalApi.setState({ title: '新建角色预设' });
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

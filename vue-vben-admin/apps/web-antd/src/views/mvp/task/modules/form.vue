<script setup lang="ts">
import { h, ref } from 'vue';
import { useVbenModal } from '@vben/common-ui';
import { useVbenForm } from '#/adapter/form';
import { message, Tooltip } from 'ant-design-vue';
import { QuestionCircleOutlined } from '@ant-design/icons-vue';
import {
  getTaskDetail,
  createTask,
  updateTask,
  getTaskTree,
} from '#/api/mvp/task';
import type { TaskItem } from '#/api/mvp/task/types';

const treeData = ref<TaskItem[]>([]);
import { getProjectList } from '#/api/mvp/project';
import { getConversationList } from '#/api/mvp/conversation';

const projectIDOptions = ref<{ label: string; value: string }[]>([]);
const conversationIDOptions = ref<{ label: string; value: string }[]>([]);
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
      component: 'TreeSelect',
      fieldName: 'parentID',
      label: '父任务ID，0=顶级',
      componentProps: {
        treeData: treeData.value,
        fieldNames: { label: 'name', value: 'id', children: 'children' },
        placeholder: '请选择父任务ID，0=顶级',
        allowClear: true,
        treeDefaultExpandAll: true,
        class: 'w-full',
      },
    },
    {
      component: 'Input',
      fieldName: 'name',
      label: '任务名称',
      rules: 'required',
      componentProps: { placeholder: '请输入任务名称', maxlength: 500 },
    },
    {
      component: 'Textarea',
      fieldName: 'description',
      label: '任务描述',
      componentProps: { placeholder: '请输入任务描述', rows: 4, maxlength: 65535 },
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
      label: '使用的AI模型ID',
      componentProps: { placeholder: '请输入使用的AI模型ID' },
    },
    {
      component: 'Select',
      fieldName: 'conversationID',
      label: '任务对话ID，用于检测任务状态',
      componentProps: { options: conversationIDOptions, placeholder: '请选择任务对话ID，用于检测任务状态', allowClear: true, class: 'w-full' },
    },
    {
      component: 'Switch',
      fieldName: 'status',
      label: '状态',
      componentProps: { checkedValue: 1, unCheckedValue: 0 },
      defaultValue: pending,
    },
    {
      component: 'InputNumber',
      fieldName: 'sort',
      label: '排序',
      componentProps: { placeholder: '请输入排序', class: 'w-full' },
    },
    {
      component: 'Input',
      fieldName: 'batchNo',
      label: '执行批次号，同批次内可并行，批次间串行',
      componentProps: { placeholder: '请输入执行批次号，同批次内可并行，批次间串行' },
    },
    {
      component: 'Input',
      fieldName: 'affectedResources',
      label: tooltipLabel('涉及的资源范围', '文件/模块'),
      componentProps: { placeholder: '请输入涉及的资源范围（文件/模块），用于并发冲突检测' },
    },
    {
      component: 'Input',
      fieldName: 'dependsOn',
      label: '依赖的任务ID列表',
      componentProps: { placeholder: '请输入依赖的任务ID列表' },
    },
    {
      component: 'Textarea',
      fieldName: 'result',
      label: '任务执行结果',
      componentProps: { placeholder: '请输入任务执行结果', rows: 4, maxlength: 4294967295 },
    },
    {
      component: 'Textarea',
      fieldName: 'contextSummary',
      label: '任务完成后的上下文压缩摘要，供后续AI读取',
      componentProps: { placeholder: '请输入任务完成后的上下文压缩摘要，供后续AI读取', rows: 4, maxlength: 65535 },
    },
    {
      component: 'Textarea',
      fieldName: 'errorMessage',
      label: '错误信息',
      componentProps: { placeholder: '请输入错误信息', rows: 4, maxlength: 65535 },
    },
    {
      component: 'DatePicker',
      fieldName: 'startedAt',
      label: '开始时间',
      componentProps: { showTime: true, placeholder: '请选择开始时间', class: 'w-full', valueFormat: 'YYYY-MM-DD HH:mm:ss' },
    },
    {
      component: 'DatePicker',
      fieldName: 'completedAt',
      label: '完成时间',
      componentProps: { showTime: true, placeholder: '请选择完成时间', class: 'w-full', valueFormat: 'YYYY-MM-DD HH:mm:ss' },
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
        await updateTask({ id: editId.value, ...values });
        message.success('更新成功');
      } else {
        await createTask(values);
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
      // 加载树形数据
      try {
        const res = await getTaskTree();
        treeData.value = [
          { id: '0', title: '顶级节点', children: res ?? [] } as TaskItem,
        ];
        formApi.updateSchema([
          {
            fieldName: 'parentID',
            componentProps: { treeData: treeData.value },
          },
        ]);
      } catch {
        // ignore
      }
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
      // 加载任务对话ID，用于检测任务状态选项
      try {
        const conversationRes = await getConversationList({ pageNum: 1, pageSize: 1000 });
        conversationIDOptions.value = (conversationRes?.list ?? []).map((item: any) => ({
          label: item.title || item.id,
          value: item.id,
        }));
      } catch {
        // ignore
      }
      if (data?.id) {
        isEdit.value = true;
        editId.value = data.id;
        modalApi.setState({ title: '编辑MVP任务表' });
        try {
          const detail = await getTaskDetail(data.id);
          if (detail) {
            formApi.setValues(detail);
          }
        } catch {
          message.error('获取详情失败');
        }
      } else {
        isEdit.value = false;
        editId.value = '';
        modalApi.setState({ title: '新建MVP任务表' });
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

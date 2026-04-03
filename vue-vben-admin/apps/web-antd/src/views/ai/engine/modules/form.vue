<script setup lang="ts">
import { ref } from 'vue';

import { useVbenForm } from '#/adapter/form';
import { useVbenModal } from '@vben/common-ui';
import { Alert, message } from 'ant-design-vue';

import { getEngineDetail, updateEngine } from '#/api/ai/engine';
import { getModelDetail, getModelList } from '#/api/ai/model';

const emit = defineEmits<{ success: [] }>();
const engineCode = ref('');
const modelOptions = ref<{ label: string; value: string }[]>([]);

async function syncModelPreview(modelID?: string) {
  const normalizedModelID = modelID ? String(modelID) : '';
  if (!normalizedModelID) {
    formApi.setValues({
      baseURLPreview: '',
      apiKeyPreview: '',
    });
    return;
  }
  const detail = await getModelDetail(normalizedModelID);
  formApi.setValues({
    baseURLPreview: detail.baseURL || '',
    apiKeyPreview: detail.apiKeyMasked || '',
  });
}

const [Form, formApi] = useVbenForm({
  showDefaultActions: false,
  schema: [
    {
      component: 'Input',
      fieldName: 'engineCode',
      label: '引擎编码',
      componentProps: { disabled: true },
    },
    {
      component: 'Select',
      fieldName: 'defaultModelID',
      label: '默认模型（执行任务）',
      rules: 'selectRequired',
      componentProps: {
        options: modelOptions,
        placeholder: '请选择默认模型',
        allowClear: true,
        showSearch: true,
        optionFilterProp: 'label',
        class: 'w-full',
        onChange: async (value: string) => {
          await syncModelPreview(value);
        },
      },
    },
    {
      component: 'Input',
      fieldName: 'baseURLPreview',
      label: 'Base URL',
      componentProps: {
        disabled: true,
        placeholder: '随所选 AI 模型自动带出',
      },
    },
    {
      component: 'Input',
      fieldName: 'apiKeyPreview',
      label: 'API Key',
      componentProps: {
        disabled: true,
        placeholder: '随所选 AI 模型自动带出',
      },
    },
    {
      component: 'InputNumber',
      fieldName: 'timeoutSeconds',
      label: '超时秒数（仅执行任务）',
      componentProps: { class: 'w-full', min: 1 },
      defaultValue: 600,
    },
    {
      component: 'InputNumber',
      fieldName: 'maxSteps',
      label: '最大步数（仅执行任务）',
      componentProps: { class: 'w-full', min: 1 },
      defaultValue: 20,
    },
    {
      component: 'Input',
      fieldName: 'workspaceRoot',
      label: '工作区根目录（仅执行任务）',
      componentProps: { placeholder: '仅 AI 执行任务生效' },
    },
    {
      component: 'Input',
      fieldName: 'commandTemplate',
      label: '命令模板（仅执行任务）',
      componentProps: { placeholder: '仅 AI 执行任务生效' },
    },
    {
      component: 'Textarea',
      fieldName: 'extraConfigText',
      label: '额外配置（仅执行任务 JSON）',
      componentProps: { rows: 5, placeholder: '仅 AI 执行任务生效' },
    },
    {
      component: 'Switch',
      fieldName: 'status',
      label: '启用状态',
      componentProps: { checkedValue: 1, unCheckedValue: 0 },
      defaultValue: 1,
    },
  ],
});

const [Modal, modalApi] = useVbenModal({
  fullscreenButton: false,
  onCancel() {
    modalApi.close();
  },
  onConfirm: async () => {
    const values = await formApi.validateAndSubmitForm();
    if (!values) return;
    let extraConfig: Record<string, any> | undefined;
    if (values.extraConfigText) {
      try {
        extraConfig = JSON.parse(values.extraConfigText);
      } catch {
        message.error('额外配置 JSON 格式不正确');
        return;
      }
    }
    modalApi.lock();
    try {
      await updateEngine({
        engineCode: engineCode.value,
        defaultModelID: values.defaultModelID,
        timeoutSeconds: values.timeoutSeconds,
        maxSteps: values.maxSteps,
        workspaceRoot: values.workspaceRoot,
        commandTemplate: values.commandTemplate,
        extraConfig,
        status: values.status,
      });
      message.success('保存成功');
      emit('success');
      modalApi.close();
    } finally {
      modalApi.lock(false);
    }
  },
  async onOpenChange(isOpen: boolean) {
    if (!isOpen) return;
    const data = modalApi.getData<{ engineCode: string; name: string }>();
    engineCode.value = data.engineCode;
    modalApi.setState({ title: `编辑引擎 - ${data.name}` });
    formApi.resetForm();
    try {
        const modelRes = await getModelList({ pageNum: 1, pageSize: 1000 });
        modelOptions.value = (modelRes?.list ?? []).map((item: any) => ({
          label: `${item.name} (${item.modelCode})`,
          value: String(item.id),
        }));
    } catch {
      message.error('加载模型列表失败');
      return;
    }
    const detail = await getEngineDetail(data.engineCode);
    formApi.setValues({
      engineCode: detail.engineCode,
      defaultModelID: detail.defaultModelID ? String(detail.defaultModelID) : undefined,
      baseURLPreview: detail.baseURL,
      apiKeyPreview: detail.apiKeyMasked || '',
      timeoutSeconds: detail.timeoutSeconds,
      maxSteps: detail.maxSteps,
      workspaceRoot: detail.workspaceRoot,
      commandTemplate: detail.commandTemplate,
      extraConfigText: detail.extraConfig || '',
      status: detail.configStatus || detail.status || 0,
    });
    await syncModelPreview(detail.defaultModelID ? String(detail.defaultModelID) : undefined);
  },
});
</script>

<template>
  <Modal class="w-[720px]">
    <Alert
      class="mb-4"
      type="info"
      show-icon
      message="这里的超时、步数、工作区、命令模板仅对 AI 模块里的“执行任务”生效。"
      description="MVP 项目任务请到“MVP 项目”的代码工作目录，以及“MVP -> 配置管理”里维护全局调度/看门狗配置。"
    />
    <Form />
  </Modal>
</template>

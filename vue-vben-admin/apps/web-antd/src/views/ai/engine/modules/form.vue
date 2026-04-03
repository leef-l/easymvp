<script setup lang="ts">
import { ref } from 'vue';

import { useVbenForm } from '#/adapter/form';
import { useVbenModal } from '@vben/common-ui';
import { message } from 'ant-design-vue';

import { getEngineDetail, updateEngine } from '#/api/ai/engine';
import { getModelList } from '#/api/ai/model';

const emit = defineEmits<{ success: [] }>();
const engineCode = ref('');
const modelOptions = ref<{ label: string; value: string }[]>([]);

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
      component: 'Input',
      fieldName: 'baseURL',
      label: 'Base URL',
      componentProps: { placeholder: '请输入服务地址' },
    },
    {
      component: 'InputPassword',
      fieldName: 'apiKey',
      label: 'API Key',
      componentProps: { placeholder: '如需更新请输入新 API Key' },
    },
    {
      component: 'Select',
      fieldName: 'defaultModelID',
      label: '默认模型',
      rules: 'selectRequired',
      componentProps: {
        options: modelOptions,
        placeholder: '请选择默认模型',
        allowClear: true,
        showSearch: true,
        optionFilterProp: 'label',
        class: 'w-full',
      },
    },
    {
      component: 'InputNumber',
      fieldName: 'timeoutSeconds',
      label: '超时秒数',
      componentProps: { class: 'w-full', min: 1 },
      defaultValue: 600,
    },
    {
      component: 'InputNumber',
      fieldName: 'maxSteps',
      label: '最大步数',
      componentProps: { class: 'w-full', min: 1 },
      defaultValue: 20,
    },
    {
      component: 'Input',
      fieldName: 'workspaceRoot',
      label: '工作区根目录',
      componentProps: { placeholder: '请输入工作区根目录' },
    },
    {
      component: 'Input',
      fieldName: 'commandTemplate',
      label: '命令模板',
      componentProps: { placeholder: 'Aider 可选命令模板' },
    },
    {
      component: 'Textarea',
      fieldName: 'extraConfigText',
      label: '额外配置(JSON)',
      componentProps: { rows: 5, placeholder: '{\n  "key": "value"\n}' },
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
        baseURL: values.baseURL,
        apiKey: values.apiKey,
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
        value: item.id,
      }));
    } catch {
      message.error('加载模型列表失败');
      return;
    }
    const detail = await getEngineDetail(data.engineCode);
    formApi.setValues({
      engineCode: detail.engineCode,
      baseURL: detail.baseURL,
      defaultModelID: detail.defaultModelID,
      timeoutSeconds: detail.timeoutSeconds,
      maxSteps: detail.maxSteps,
      workspaceRoot: detail.workspaceRoot,
      commandTemplate: detail.commandTemplate,
      extraConfigText: detail.extraConfig || '',
      status: detail.configStatus || detail.status || 0,
    });
  },
});
</script>

<template>
  <Modal class="w-[720px]">
    <Form />
  </Modal>
</template>

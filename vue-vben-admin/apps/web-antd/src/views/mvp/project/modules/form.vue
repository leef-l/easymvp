<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import { useVbenModal } from '@vben/common-ui';
import { useVbenForm } from '#/adapter/form';
import { message } from 'ant-design-vue';
import { getModelList } from '#/api/ai/model';
import { createProject } from '#/api/mvp/workflow';

const router = useRouter();

/** AI 模型选项列表 */
const modelOptions = ref<{ label: string; value: string }[]>([]);

/** 加载可用的架构师 AI 模型列表 */
async function loadModelOptions() {
  try {
    const res = await getModelList({ pageNum: 1, pageSize: 200 });
    modelOptions.value = (res?.list ?? []).map((item) => ({
      label: item.name,
      value: item.id,
    }));
  } catch {
    message.error('加载AI模型列表失败');
  }
}

onMounted(() => {
  loadModelOptions();
});

/** 表单配置 */
const [Form, formApi] = useVbenForm({
  showDefaultActions: false,
  schema: [
    {
      component: 'Input',
      fieldName: 'name',
      label: '项目名称',
      rules: [{ required: true, message: '项目名称不能为空' }],
      componentProps: {
        placeholder: '请输入项目名称，简洁清晰即可',
        maxlength: 200,
      },
    },
    {
      component: 'Textarea',
      fieldName: 'description',
      label: '项目简介',
      componentProps: {
        placeholder: '请描述您想要构建的产品或功能，AI架构师将根据此信息为您设计方案',
        rows: 5,
        maxlength: 5000,
      },
    },
    {
      component: 'Select',
      fieldName: 'architectModelID',
      label: '架构师AI模型',
      rules: [{ required: true, message: '请选择架构师AI模型' }],
      componentProps: {
        placeholder: '请选择用于架构设计的AI模型',
        options: modelOptions,
        showSearch: true,
        filterOption: (input: string, option: any) =>
          option.label.toLowerCase().includes(input.toLowerCase()),
      },
    },
  ],
});

/** 弹窗配置 */
const [Modal, modalApi] = useVbenModal({
  fullscreenButton: false,
  title: '新建项目',
  onCancel() {
    modalApi.close();
  },
  onConfirm: async () => {
    const values = await formApi.validateAndSubmitForm();
    if (!values) return;
    modalApi.lock();
    try {
      const res = await createProject(values as any);
      message.success('项目创建成功，正在进入对话...');
      modalApi.close();
      // 创建成功后跳转到与AI架构师的对话页面
      router.push({
        path: '/mvp/chat',
        query: {
          projectId: res?.projectID,
          conversationId: res?.conversationID,
        },
      });
    } catch {
      message.error('创建项目失败，请稍后重试');
    } finally {
      modalApi.lock(false);
    }
  },
  async onOpenChange(isOpen: boolean) {
    if (isOpen) {
      formApi.resetForm();
      // 每次打开时刷新模型列表，确保数据最新
      loadModelOptions();
    }
  },
});
</script>

<template>
  <Modal class="w-[560px]">
    <Form />
  </Modal>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useVbenModal } from '@vben/common-ui';
import { useVbenForm } from '#/adapter/form';
import { message } from 'ant-design-vue';
import {
  getProjectCategoryDetail,
  createProjectCategory,
  updateProjectCategory,
} from '#/api/mvp/project_category';

const emit = defineEmits<{ success: [] }>();
const isEdit = ref(false);
const editId = ref('');

/** 状态选项 */
const statusOptions = [
  { label: '禁用', value: '0' },
  { label: '启用', value: '1' },
];

/** 表单配置 */
const [Form, formApi] = useVbenForm({
  showDefaultActions: false,
  schema: [
    {
      component: 'Input',
      fieldName: 'categoryCode',
      label: '分类代码',
      rules: 'required',
      componentProps: { placeholder: '请输入分类代码（如 software_dev）', class: 'w-full' },
    },
    {
      component: 'Input',
      fieldName: 'displayName',
      label: '显示名称',
      rules: 'required',
      componentProps: { placeholder: '请输入显示名称（如 软件开发）', class: 'w-full' },
    },
    {
      component: 'Input',
      fieldName: 'familyCode',
      label: '所属系',
      componentProps: { placeholder: '请输入所属系（如 tech）', class: 'w-full' },
    },
    {
      component: 'Textarea',
      fieldName: 'description',
      label: '描述',
      componentProps: { placeholder: '请输入描述', rows: 3, maxlength: 500 },
    },
    {
      component: 'Select',
      fieldName: 'status',
      label: '状态',
      componentProps: { options: statusOptions, placeholder: '请选择状态', allowClear: true, class: 'w-full' },
      defaultValue: '1',
    },
    {
      component: 'InputNumber',
      fieldName: 'sort',
      label: '排序',
      componentProps: { placeholder: '请输入排序', class: 'w-full' },
      defaultValue: 0,
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
        await updateProjectCategory({ id: editId.value, ...values });
        message.success('更新成功');
      } else {
        await createProjectCategory(values);
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
        modalApi.setState({ title: '编辑项目分类' });
        try {
          const detail = await getProjectCategoryDetail(data.id);
          if (detail) {
            formApi.setValues(detail);
          }
        } catch {
          message.error('获取详情失败');
        }
      } else {
        isEdit.value = false;
        editId.value = '';
        modalApi.setState({ title: '新建项目分类' });
        formApi.resetForm();
      }
    }
  },
});
</script>

<template>
  <Modal class="w-[560px]">
    <Form />
  </Modal>
</template>

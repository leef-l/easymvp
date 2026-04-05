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

/** 表单配置 */
const [Form, formApi] = useVbenForm({
  showDefaultActions: false,
  schema: [
    {
      component: 'Input',
      fieldName: 'categoryCode',
      label: '稳定分类编码',
      rules: 'required',
      componentProps: { placeholder: '请输入稳定分类编码', maxlength: 64 },
    },
    {
      component: 'Input',
      fieldName: 'displayName',
      label: '展示名称',
      rules: 'required',
      componentProps: { placeholder: '请输入展示名称', maxlength: 64 },
    },
    {
      component: 'Input',
      fieldName: 'familyCode',
      label: '能力家族编码',
      rules: 'required',
      componentProps: { placeholder: '请输入能力家族编码', maxlength: 32 },
    },
    {
      component: 'Input',
      fieldName: 'description',
      label: '分类说明',
      componentProps: { placeholder: '请输入分类说明', maxlength: 255 },
    },
    {
      component: 'Switch',
      fieldName: 'status',
      label: '1启用 0停用',
      componentProps: { checkedValue: 1, unCheckedValue: 0 },
      defaultValue: 1,
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
        modalApi.setState({ title: '编辑项目分类配置表' });
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
        modalApi.setState({ title: '新建项目分类配置表' });
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

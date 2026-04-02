<script setup lang="ts">
import { ref } from 'vue';
import { useVbenModal } from '@vben/common-ui';
import { useVbenForm } from '#/adapter/form';
import { message } from 'ant-design-vue';
import {
  getDeptDetail,
  createDept,
  updateDept,
  getDeptTree,
} from '#/api/system/dept';
import type { DeptItem } from '#/api/system/dept/types';

const treeData = ref<DeptItem[]>([]);

const emit = defineEmits<{ success: [] }>();
const isEdit = ref(false);
const editId = ref('');

/** 表单配置 */
const [Form, formApi] = useVbenForm({
  showDefaultActions: false,
  schema: [
    {
      component: 'TreeSelect',
      fieldName: 'parentID',
      label: '上级部门',
      componentProps: {
        treeData: treeData.value,
        fieldNames: { label: 'title', value: 'id', children: 'children' },
        placeholder: '请选择上级部门',
        allowClear: true,
        treeDefaultExpandAll: true,
        class: 'w-full',
      },
    },
    {
      component: 'Input',
      fieldName: 'title',
      label: '部门名称',
      rules: 'required',
      componentProps: { placeholder: '请输入部门名称', maxlength: 50 },
    },
    {
      component: 'Input',
      fieldName: 'username',
      label: '部门负责人姓名',
      componentProps: { placeholder: '请输入部门负责人姓名', maxlength: 50 },
    },
    {
      component: 'Input',
      fieldName: 'email',
      label: '负责人邮箱',
      componentProps: { placeholder: '请输入负责人邮箱', maxlength: 100 },
    },
    {
      component: 'InputNumber',
      fieldName: 'sort',
      label: '排序（升序）',
      componentProps: { placeholder: '请输入排序（升序）', class: 'w-full' },
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
        await updateDept({ id: editId.value, ...values });
        message.success('更新成功');
      } else {
        await createDept(values);
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
        const res = await getDeptTree();
        treeData.value = [
          { id: '0', title: '顶级节点', children: res ?? [] } as DeptItem,
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
      if (data?.id) {
        isEdit.value = true;
        editId.value = data.id;
        modalApi.setState({ title: '编辑部门表' });
        try {
          const detail = await getDeptDetail(data.id);
          if (detail) {
            formApi.setValues(detail);
          }
        } catch {
          message.error('获取详情失败');
        }
      } else {
        isEdit.value = false;
        editId.value = '';
        modalApi.setState({ title: '新建部门表' });
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

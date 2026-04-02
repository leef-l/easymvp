<script setup lang="ts">
import { ref } from 'vue';
import { useVbenModal } from '@vben/common-ui';
import { useVbenForm } from '#/adapter/form';
import { message } from 'ant-design-vue';
import {
  getUsersDetail,
  createUsers,
  updateUsers,
} from '#/api/system/users';
import { getDeptTree } from '#/api/system/dept';
import { getRoleTree } from '#/api/system/role';
import type { DeptItem } from '#/api/system/dept/types';

const emit = defineEmits<{ success: [] }>();
const isEdit = ref(false);
const editId = ref('');
const deptTreeData = ref<DeptItem[]>([]);

/** 表单配置 */
const [Form, formApi] = useVbenForm({
  showDefaultActions: false,
  schema: [
    {
      component: 'Input',
      fieldName: 'username',
      label: '登录用户名',
      rules: 'required',
      componentProps: { placeholder: '请输入登录用户名', maxlength: 50 },
    },
    {
      component: 'InputPassword',
      fieldName: 'password',
      label: '密码',
      rules: 'required',
      componentProps: { placeholder: '请输入密码' },
      dependencies: {
        triggerFields: ['_mode'],
        if: () => !isEdit.value,
      },
    },
    {
      component: 'Input',
      fieldName: 'nickname',
      label: '昵称',
      componentProps: { placeholder: '请输入昵称', maxlength: 50 },
    },
    {
      component: 'Input',
      fieldName: 'email',
      label: '邮箱',
      componentProps: { placeholder: '请输入邮箱', maxlength: 100 },
    },
    {
      component: 'TreeSelect',
      fieldName: 'deptId',
      label: '所属部门',
      componentProps: {
        treeData: [],
        fieldNames: { label: 'title', value: 'id', children: 'children' },
        placeholder: '请选择所属部门',
        allowClear: true,
        treeDefaultExpandAll: true,
        class: 'w-full',
      },
    },
    {
      component: 'Cascader',
      fieldName: 'roleIds',
      label: '角色',
      componentProps: {
        options: [],
        fieldNames: { label: 'title', value: 'id', children: 'children' },
        placeholder: '请选择角色',
        multiple: true,
        allowClear: true,
        class: 'w-full',
      },
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
        await updateUsers({ id: editId.value, ...values });
        message.success('更新成功');
      } else {
        await createUsers(values);
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

      // 加载部门树
      try {
        const res = await getDeptTree();
        deptTreeData.value = [
          { id: '0', title: '顶级部门', children: res ?? [] } as any,
        ];
        formApi.updateSchema([
          {
            fieldName: 'deptId',
            componentProps: { treeData: deptTreeData.value },
          },
        ]);
      } catch { /* ignore */ }

      // 加载角色树
      try {
        const res = await getRoleTree();
        formApi.updateSchema([
          {
            fieldName: 'roleIds',
            componentProps: { options: res ?? [] },
          },
        ]);
      } catch { /* ignore */ }

      if (data?.id) {
        isEdit.value = true;
        editId.value = data.id;
        modalApi.setState({ title: '编辑用户' });
        try {
          const detail = await getUsersDetail(data.id);
          if (detail) {
            formApi.setValues(detail);
          }
        } catch {
          message.error('获取详情失败');
        }
      } else {
        isEdit.value = false;
        editId.value = '';
        modalApi.setState({ title: '新建用户' });
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

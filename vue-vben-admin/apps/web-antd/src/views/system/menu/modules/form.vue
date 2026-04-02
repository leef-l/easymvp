<script setup lang="ts">
import { ref } from 'vue';
import { useVbenModal } from '@vben/common-ui';
import { useVbenForm } from '#/adapter/form';
import { message } from 'ant-design-vue';
import {
  getMenuDetail,
  createMenu,
  updateMenu,
  getMenuTree,
} from '#/api/system/menu';
import type { MenuItem } from '#/api/system/menu/types';

const treeData = ref<MenuItem[]>([]);

/** 类型选项 */
const typeOptions = [
  { label: '目录', value: 1 },
  { label: '菜单', value: 2 },
  { label: '按钮', value: 3 },
  { label: '外链', value: 4 },
  { label: '内链', value: 5 },
];

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
      label: '上级菜单',
      componentProps: {
        treeData: treeData.value,
        fieldNames: { label: 'title', value: 'id', children: 'children' },
        placeholder: '请选择上级菜单',
        allowClear: true,
        treeDefaultExpandAll: true,
        class: 'w-full',
      },
    },
    {
      component: 'Input',
      fieldName: 'title',
      label: '菜单名称',
      rules: 'required',
      componentProps: { placeholder: '请输入菜单名称', maxlength: 50 },
    },
    {
      component: 'Select',
      fieldName: 'type',
      label: '类型',
      componentProps: { options: typeOptions, placeholder: '请选择类型', allowClear: true, class: 'w-full' },
    },
    {
      component: 'Input',
      fieldName: 'path',
      label: '前端路由路径',
      componentProps: { placeholder: '请输入前端路由路径', maxlength: 200 },
    },
    {
      component: 'Input',
      fieldName: 'component',
      label: '前端组件路径',
      componentProps: { placeholder: '请输入前端组件路径', maxlength: 200 },
    },
    {
      component: 'Input',
      fieldName: 'permission',
      label: '权限标识',
      componentProps: { placeholder: '请输入权限标识，如 system:dept:list', maxlength: 100 },
    },
    {
      component: 'Input',
      fieldName: 'icon',
      label: '菜单图标（图标名称）',
      componentProps: { placeholder: '请输入菜单图标（图标名称）', maxlength: 100 },
    },
    {
      component: 'InputNumber',
      fieldName: 'sort',
      label: '排序（升序）',
      componentProps: { placeholder: '请输入排序（升序）', class: 'w-full' },
    },
    {
      component: 'Switch',
      fieldName: 'isShow',
      label: '是否显示',
      componentProps: { checkedValue: 1, unCheckedValue: 0 },
      defaultValue: 1,
    },
    {
      component: 'Switch',
      fieldName: 'isCache',
      label: '是否缓存',
      componentProps: { checkedValue: 1, unCheckedValue: 0 },
      defaultValue: 0,
    },
    {
      component: 'Input',
      fieldName: 'linkURL',
      label: '外链/内链地址（type=4或5时有效）',
      componentProps: { placeholder: '请输入外链/内链地址（type=4或5时有效）', maxlength: 500 },
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
        await updateMenu({ id: editId.value, ...values });
        message.success('更新成功');
      } else {
        await createMenu(values);
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
        const res = await getMenuTree();
        treeData.value = [
          { id: '0', title: '顶级节点', children: res ?? [] } as MenuItem,
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
        modalApi.setState({ title: '编辑菜单表' });
        try {
          const detail = await getMenuDetail(data.id);
          if (detail) {
            formApi.setValues(detail);
          }
        } catch {
          message.error('获取详情失败');
        }
      } else {
        isEdit.value = false;
        editId.value = '';
        modalApi.setState({ title: '新建菜单表' });
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

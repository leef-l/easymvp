<script setup lang="ts">
import { ref } from 'vue';
import { message, Modal, Tree } from 'ant-design-vue';
import { getMenuTree } from '#/api/system/menu';
import { getRoleMenuIds, grantRoleMenu } from '#/api/system/role';
import type { MenuItem } from '#/api/system/menu/types';

const emit = defineEmits<{ success: [] }>();

const visible = ref(false);
const confirmLoading = ref(false);
const roleId = ref('');
const checkedKeys = ref<string[]>([]);
const halfCheckedKeys = ref<string[]>([]);
const treeData = ref<MenuItem[]>([]);
const expandedKeys = ref<string[]>([]);

/** 递归收集所有节点 key */
function collectKeys(nodes: MenuItem[]): string[] {
  const keys: string[] = [];
  for (const node of nodes) {
    keys.push(node.id);
    if (node.children?.length) {
      keys.push(...collectKeys(node.children));
    }
  }
  return keys;
}

/** 递归收集所有叶子节点 key */
function collectLeafKeys(nodes: MenuItem[]): string[] {
  const keys: string[] = [];
  for (const node of nodes) {
    if (!node.children?.length) {
      keys.push(node.id);
    } else {
      keys.push(...collectLeafKeys(node.children));
    }
  }
  return keys;
}

/** 打开弹窗 */
async function open(id: string) {
  visible.value = true;
  roleId.value = id;
  checkedKeys.value = [];
  halfCheckedKeys.value = [];

  try {
    // 加载菜单树
    const res = await getMenuTree();
    treeData.value = (res ?? []) as MenuItem[];
    expandedKeys.value = collectKeys(treeData.value);

    // 加载已授权菜单 — backend returns ALL checked IDs (both parent and leaf)
    // But ant-design-vue Tree (without check-strictly) only accepts leaf checked keys
    // Parent nodes will auto-check if all children are checked
    const allCheckedIds = await getRoleMenuIds(id);
    const leafKeys = collectLeafKeys(treeData.value);
    // Only set leaf keys as checkedKeys, parents will auto-derive
    checkedKeys.value = allCheckedIds.filter(k => leafKeys.includes(k));
    // Non-leaf keys that were checked are half-checked (or will auto-resolve)
    halfCheckedKeys.value = allCheckedIds.filter(k => !leafKeys.includes(k));
  } catch {
    message.error('加载数据失败');
  }
}

/** 勾选事件 */
function handleCheck(_checkedKeys: string[], e: any) {
  checkedKeys.value = _checkedKeys as string[];
  halfCheckedKeys.value = e.halfCheckedKeys as string[];
}

/** 提交 */
async function handleOk() {
  confirmLoading.value = true;
  try {
    // Submit both fully checked and half-checked (parent) keys
    await grantRoleMenu({
      id: roleId.value,
      menuIds: [...checkedKeys.value, ...halfCheckedKeys.value],
    });
    message.success('授权成功');
    visible.value = false;
    emit('success');
  } finally {
    confirmLoading.value = false;
  }
}

defineExpose({ open });
</script>

<template>
  <Modal
    v-model:open="visible"
    title="授权菜单"
    :confirm-loading="confirmLoading"
    width="500px"
    @ok="handleOk"
  >
    <div class="max-h-[400px] overflow-auto py-2">
      <Tree
        :checked-keys="checkedKeys"
        v-model:expandedKeys="expandedKeys"
        :tree-data="treeData"
        :field-names="{ title: 'title', key: 'id', children: 'children' }"
        checkable
        @check="handleCheck"
      />
    </div>
  </Modal>
</template>

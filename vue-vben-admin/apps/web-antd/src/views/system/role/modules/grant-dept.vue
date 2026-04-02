<script setup lang="ts">
import { ref } from 'vue';
import { message, Modal, Select, Tree } from 'ant-design-vue';
import { getDeptTree } from '#/api/system/dept';
import { getRoleDeptIds, grantRoleDept } from '#/api/system/role';
import type { DeptItem } from '#/api/system/dept/types';

const emit = defineEmits<{ success: [] }>();

const visible = ref(false);
const confirmLoading = ref(false);
const roleId = ref('');
const dataScope = ref(1);
const checkedDeptKeys = ref<string[]>([]);
const treeData = ref<DeptItem[]>([]);
const expandedKeys = ref<string[]>([]);

const dataScopeOptions = [
  { label: '全部数据', value: 1 },
  { label: '本部门及以下', value: 2 },
  { label: '本部门', value: 3 },
  { label: '仅本人', value: 4 },
  { label: '自定义', value: 5 },
];

/** 递归收集所有节点 key */
function collectKeys(nodes: DeptItem[]): string[] {
  const keys: string[] = [];
  for (const node of nodes) {
    keys.push(node.id);
    if (node.children?.length) {
      keys.push(...collectKeys(node.children));
    }
  }
  return keys;
}

/** 打开弹窗 */
async function open(id: string, currentScope: number) {
  visible.value = true;
  roleId.value = id;
  dataScope.value = currentScope || 1;
  checkedDeptKeys.value = [];

  try {
    const res = await getDeptTree();
    treeData.value = (res ?? []) as DeptItem[];
    expandedKeys.value = collectKeys(treeData.value);

    // Load existing dept IDs for this role
    if (currentScope === 5) {
      try {
        checkedDeptKeys.value = await getRoleDeptIds(id);
      } catch { /* ignore */ }
    }
  } catch {
    message.error('加载部门树失败');
  }
}

/** 提交 */
async function handleOk() {
  confirmLoading.value = true;
  try {
    await grantRoleDept({
      id: roleId.value,
      dataScope: dataScope.value,
      deptIds: dataScope.value === 5 ? checkedDeptKeys.value : [],
    });
    message.success('设置成功');
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
    title="数据权限"
    :confirm-loading="confirmLoading"
    width="500px"
    @ok="handleOk"
  >
    <div class="py-2">
      <div class="mb-4 flex items-center gap-2">
        <span class="shrink-0">数据范围</span>
        <Select
          v-model:value="dataScope"
          :options="dataScopeOptions"
          style="width: 100%"
        />
      </div>
      <div v-if="dataScope === 5" class="max-h-[350px] overflow-auto">
        <Tree
          v-model:checkedKeys="checkedDeptKeys"
          v-model:expandedKeys="expandedKeys"
          :tree-data="treeData"
          :field-names="{ title: 'title', key: 'id', children: 'children' }"
          checkable
        />
      </div>
    </div>
  </Modal>
</template>

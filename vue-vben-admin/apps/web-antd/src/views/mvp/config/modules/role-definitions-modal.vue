<script setup lang="ts">
import { ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import { Alert, Button, Card, Empty, Input, InputNumber, message, Select, Space, Switch } from 'ant-design-vue';

import { getRoleDefinitions, resetRoleDefinitions, type RoleDefinitionItem, saveRoleDefinitions } from '#/api/mvp/workflow';

import { roleLevelOptions } from '../../consts';
import { fallbackRoleDefinitions } from '../../role-definitions';

const emit = defineEmits<{ success: [] }>();

const saving = ref(false);
const definitions = ref<RoleDefinitionItem[]>([]);

const colorOptions = [
  'default',
  'blue',
  'purple',
  'green',
  'orange',
  'cyan',
  'magenta',
  'gold',
  'red',
  'processing',
  'success',
  'warning',
  'error',
].map((value) => ({ label: value, value }));

const roleTypePattern = /^[a-z][a-z0-9_]{1,63}$/;

function createDefinition(sort: number): RoleDefinitionItem {
  return {
    roleType: '',
    displayName: '',
    color: 'default',
    description: '',
    preferredLevels: ['pro', 'max', 'lite'],
    defaultSystemPrompt: '',
    acceptanceJudge: false,
    sort,
  };
}

function addDefinition() {
  let maxSort = 0;
  for (const item of definitions.value) {
    maxSort = Math.max(maxSort, item.sort || 0);
  }
  definitions.value.push(createDefinition(maxSort + 10));
}

function removeDefinition(index: number) {
  definitions.value.splice(index, 1);
}

function duplicateDefinition(index: number) {
  const source = definitions.value[index];
  if (!source) {
    return;
  }
  definitions.value.splice(index + 1, 0, {
    ...source,
    roleType: source.roleType ? `${source.roleType}_copy` : '',
    displayName: source.displayName ? `${source.displayName} 副本` : '',
    sort: (source.sort || (index + 1) * 10) + 1,
  });
}

function restoreDefaultsLocally() {
  definitions.value = fallbackRoleDefinitions.map((item) => ({
    ...item,
    preferredLevels: [...(item.preferredLevels || [])],
  }));
}

async function handleResetToSystemDefaults() {
  try {
    saving.value = true;
    await resetRoleDefinitions();
    message.success('已切回系统默认角色定义');
    definitions.value = fallbackRoleDefinitions.map((item) => ({
      ...item,
      preferredLevels: [...(item.preferredLevels || [])],
    }));
    emit('success');
    modalApi.close();
  } catch (error: any) {
    message.error(error?.message || '重置失败');
  } finally {
    saving.value = false;
  }
}

function normalizeDefinitions(list: RoleDefinitionItem[]) {
  return list.map((item, index) => ({
    roleType: item.roleType.trim(),
    displayName: item.displayName.trim(),
    color: item.color || 'default',
    description: item.description?.trim() || '',
    preferredLevels: (item.preferredLevels || []).filter(Boolean),
    defaultSystemPrompt: item.defaultSystemPrompt?.trim() || '',
    acceptanceJudge: !!item.acceptanceJudge,
    sort: item.sort ?? (index + 1) * 10,
  }));
}

function validateDefinitions(list: RoleDefinitionItem[]) {
  const seen = new Set<string>();
  for (const item of list) {
    if (!item.roleType) {
      throw new Error('角色类型不能为空');
    }
    if (!roleTypePattern.test(item.roleType)) {
      throw new Error(`角色类型 ${item.roleType} 不合法，只允许小写字母、数字、下划线，且必须以字母开头`);
    }
    if (!item.displayName) {
      throw new Error(`角色 ${item.roleType} 缺少展示名`);
    }
    if (seen.has(item.roleType)) {
      throw new Error(`角色类型 ${item.roleType} 重复`);
    }
    seen.add(item.roleType);
  }
}

const [Modal, modalApi] = useVbenModal({
  fullscreenButton: true,
  footer: false,
  async onOpenChange(isOpen: boolean) {
    if (!isOpen) {
      definitions.value = [];
      return;
    }
    modalApi.setState({ title: '角色定义配置' });
    try {
      const res = await getRoleDefinitions();
      definitions.value = (res?.list ?? []).map((item) => ({
        ...item,
        preferredLevels: item.preferredLevels?.length ? item.preferredLevels : ['pro', 'max', 'lite'],
      }));
    } catch {
      message.error('加载角色定义失败');
      definitions.value = [];
    }
  },
});

async function handleSave() {
  try {
    const payload = normalizeDefinitions(definitions.value);
    validateDefinitions(payload);
    saving.value = true;
    await saveRoleDefinitions(payload);
    message.success('角色定义已保存');
    emit('success');
    modalApi.close();
  } catch (error: any) {
    message.error(error?.message || '保存失败');
  } finally {
    saving.value = false;
  }
}
</script>

<template>
  <Modal class="w-[1100px]">
    <div class="space-y-4">
      <Alert
        type="info"
        show-icon
        message="角色定义注册表"
        description="底层保存到 workflow.role_definitions，但这里用可视化方式维护。新增角色后，项目角色/角色预设表单会直接读取这里的定义。"
      />

      <div class="flex items-center justify-between">
        <div class="text-sm text-gray-500">
          共 {{ definitions.length }} 个角色定义
        </div>
        <Space>
          <Button @click="restoreDefaultsLocally">恢复默认模板</Button>
          <Button danger :loading="saving" @click="handleResetToSystemDefaults">清空自定义配置</Button>
          <Button @click="addDefinition">新增角色</Button>
          <Button type="primary" :loading="saving" @click="handleSave">保存配置</Button>
        </Space>
      </div>

      <Empty v-if="definitions.length === 0" description="暂无角色定义，先新增一个" />

      <div v-else class="space-y-4">
        <Card v-for="(item, index) in definitions" :key="`${item.roleType || 'new'}-${index}`" size="small">
          <template #title>
            <div class="flex items-center justify-between">
              <span>{{ item.displayName || item.roleType || `未命名角色 ${index + 1}` }}</span>
              <Space :size="4">
                <Button type="link" @click="duplicateDefinition(index)">复制</Button>
                <Button danger type="link" @click="removeDefinition(index)">删除</Button>
              </Space>
            </div>
          </template>

          <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
            <div>
              <div class="mb-1 text-sm text-gray-500">角色类型</div>
              <Input v-model:value="item.roleType" placeholder="例如 experience_reviewer" />
              <div class="mt-1 text-xs text-gray-400">仅支持小写字母、数字、下划线，且必须以字母开头。</div>
            </div>
            <div>
              <div class="mb-1 text-sm text-gray-500">展示名</div>
              <Input v-model:value="item.displayName" placeholder="例如 体验评审师" />
            </div>
            <div>
              <div class="mb-1 text-sm text-gray-500">颜色</div>
              <Select v-model:value="item.color" :options="colorOptions" class="w-full" />
            </div>
            <div>
              <div class="mb-1 text-sm text-gray-500">排序</div>
              <InputNumber v-model:value="item.sort" :min="0" class="w-full" />
            </div>
            <div>
              <div class="mb-1 text-sm text-gray-500">推荐等级顺序</div>
              <Select
                v-model:value="item.preferredLevels"
                mode="multiple"
                :options="roleLevelOptions"
                class="w-full"
                placeholder="选择推荐等级顺序"
              />
            </div>
            <div>
              <div class="mb-1 text-sm text-gray-500">可作验收评审角色</div>
              <Switch v-model:checked="item.acceptanceJudge" />
            </div>
          </div>

          <div class="mt-4">
            <div class="mb-1 text-sm text-gray-500">角色说明</div>
            <Input.TextArea v-model:value="item.description" :rows="2" placeholder="说明该角色负责什么" />
          </div>

          <div class="mt-4">
            <div class="mb-1 text-sm text-gray-500">默认系统提示词</div>
            <Input.TextArea
              v-model:value="item.defaultSystemPrompt"
              :rows="5"
              placeholder="可选。用于新角色在未显式填写 systemPrompt 时的默认提示词，支持 {{role_display_name}} / {{project_display_name}} / {{role_level_name}} 等占位符。"
            />
          </div>
        </Card>
      </div>
    </div>
  </Modal>
</template>

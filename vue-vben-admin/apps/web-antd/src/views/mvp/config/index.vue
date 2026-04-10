<script setup lang="ts">
import type { ConfigItem } from '#/api/mvp/config/types';

import { computed, h, onMounted, ref } from 'vue';

import { Page, useVbenModal } from '@vben/common-ui';

import { DeleteOutlined, EditOutlined, ReloadOutlined } from '@ant-design/icons-vue';
import {
  Alert,
  Badge,
  Button,
  Card,
  message,
  Modal,
  Space,
  Spin,
  Tag,
  Tooltip,
} from 'ant-design-vue';

import { deleteConfig, getConfigList } from '#/api/mvp/config';

import FormModal from './modules/form.vue';
import RoleDefinitionsModal from './modules/role-definitions-modal.vue';

/** 弹窗组件 */
const [FormModalComp, formModalApi] = useVbenModal({
  connectedComponent: FormModal,
  destroyOnClose: true,
});

const [RoleDefinitionsModalComp, roleDefinitionsModalApi] = useVbenModal({
  connectedComponent: RoleDefinitionsModal,
  destroyOnClose: true,
});

/** 分类定义 */
interface CategoryDef {
  key: string;
  label: string;
  color: string;
  description: string;
  badge?: string;
}

const CATEGORY_DEFS: CategoryDef[] = [
  { key: 'scheduler', label: '调度器', color: 'blue', description: '任务调度和并发控制' },
  { key: 'watchdog', label: '看门狗', color: 'orange', description: '心跳检测与自动重试' },
  { key: 'engine', label: '执行引擎', color: 'green', description: 'Aider/审核超时配置' },
  { key: 'accept', label: '验收', color: 'cyan', description: 'LLM质量评审与人工审核' },
  { key: 'general', label: '自治核心', color: 'purple', description: '⚡ 灰度开关，必须最先开启', badge: '灰度' },
  { key: 'autonomy', label: '自治策略', color: 'geekblue', description: '灰度开关，general 开启后按序开启', badge: '灰度' },
  { key: 'collab', label: '协作通知', color: 'magenta', description: '飞书集成配置' },
];

/** 类型 Tag 颜色映射 */
const TYPE_COLOR: Record<string, string> = {
  int: 'blue',
  string: 'green',
  autonomy: 'purple',
};

/** 数据 */
const loading = ref(false);
const allConfigs = ref<ConfigItem[]>([]);

/** 加载所有配置 */
async function loadConfigs() {
  loading.value = true;
  try {
    const res = await getConfigList({ pageNum: 1, pageSize: 200 } as any);
    allConfigs.value = res?.list ?? [];
  } catch {
    message.error('加载配置失败');
  } finally {
    loading.value = false;
  }
}

onMounted(loadConfigs);

/** 按分类分组，顺序固定 */
const groupedConfigs = computed(() => {
  const map = new Map<string, ConfigItem[]>();
  for (const item of allConfigs.value.filter((entry) => entry.configKey !== 'workflow.role_definitions')) {
    const cat = item.category || 'unknown';
    if (!map.has(cat)) map.set(cat, []);
    map.get(cat)!.push(item);
  }

  const result: { def: CategoryDef; items: ConfigItem[] }[] = [];

  for (const def of CATEGORY_DEFS) {
    if (map.has(def.key)) {
      result.push({ def, items: map.get(def.key)! });
      map.delete(def.key);
    }
  }

  // 剩余未知分类
  for (const [key, items] of map.entries()) {
    result.push({
      def: { key, label: key, color: 'default', description: '其他配置' },
      items,
    });
  }

  return result;
});

const roleDefinitionsConfig = computed(() =>
  allConfigs.value.find((item) => item.configKey === 'workflow.role_definitions'),
);

/** 判断是否是 0/1 的 int 开关 */
function isBoolInt(item: ConfigItem) {
  return item.configType === 'int' && (item.configValue === '0' || item.configValue === '1');
}

/** 截断文字 */
function truncate(str: string | undefined, len: number) {
  if (!str) return '';
  return str.length > len ? str.slice(0, len) + '…' : str;
}

/** 新建 */
function handleCreate() {
  formModalApi.setData(null).open();
}

function handleEditRoleDefinitions() {
  roleDefinitionsModalApi.open();
}

/** 编辑 */
function handleEdit(row: ConfigItem) {
  formModalApi.setData({ id: row.id }).open();
}

/** 删除 */
function handleDelete(row: ConfigItem) {
  Modal.confirm({
    title: '确认删除',
    content: `确定要删除配置项 "${row.configKey}" 吗？`,
    okType: 'danger',
    async onOk() {
      await deleteConfig(row.id);
      message.success('删除成功');
      await loadConfigs();
    },
  });
}

/** 成功回调 */
async function onFormSuccess() {
  await loadConfigs();
}
</script>

<template>
  <Page auto-content-height>
    <FormModalComp @success="onFormSuccess" />
    <RoleDefinitionsModalComp @success="onFormSuccess" />

    <!-- 顶部标题栏 -->
    <div class="mb-4 flex items-center justify-between">
      <div>
        <span class="text-xl font-semibold text-gray-800">引擎配置</span>
        <span class="ml-2 text-sm text-gray-400">共 {{ allConfigs.length }} 项配置</span>
      </div>
      <Space>
        <Button :icon="h(ReloadOutlined)" :loading="loading" @click="loadConfigs">刷新</Button>
        <Button @click="handleEditRoleDefinitions">角色定义</Button>
        <Button type="primary" @click="handleCreate">新建配置</Button>
      </Space>
    </div>

    <Spin :spinning="loading">
      <div class="space-y-4">
        <Card :body-style="{ padding: '16px' }" class="config-group-card">
          <template #title>
            <div class="flex items-center gap-2">
              <Tag color="magenta" class="font-medium">角色定义</Tag>
              <span class="text-sm text-gray-500">workflow.role_definitions 专用配置入口</span>
            </div>
          </template>
          <div class="flex items-center justify-between gap-4">
            <div class="text-sm text-gray-600">
              <div>新增角色、改展示名、调默认提示词，都在这里维护。</div>
              <div class="mt-1 text-xs text-gray-400">
                当前状态：{{ roleDefinitionsConfig ? '已配置' : '使用系统默认定义' }}
              </div>
            </div>
            <Button type="primary" @click="handleEditRoleDefinitions">打开编辑器</Button>
          </div>
        </Card>

        <template v-for="group in groupedConfigs" :key="group.def.key">
          <Card :body-style="{ padding: '0' }" class="config-group-card">
            <!-- 卡片标题 -->
            <template #title>
              <div class="flex items-center gap-2">
                <Tag :color="group.def.color" class="font-medium">{{ group.def.label }}</Tag>
                <span class="text-sm text-gray-500">{{ group.def.description }}</span>
                <Badge
                  v-if="group.def.badge"
                  :count="group.def.badge"
                  :number-style="{ backgroundColor: '#fa8c16', fontSize: '11px', height: '18px', lineHeight: '18px', padding: '0 6px' }"
                  class="ml-1"
                />
                <span class="ml-auto text-xs text-gray-400">{{ group.items.length }} 项</span>
              </div>
            </template>

            <!-- general 分类的灰度说明 -->
            <Alert
              v-if="group.def.key === 'general'"
              class="m-3 rounded"
              type="info"
              show-icon
            >
              <template #message>
                <div class="text-xs leading-6">
                  <div class="mb-1 font-medium">📋 灰度开启顺序（建议）</div>
                  <div>第1步：workflow.autonomy.enabled = 1（开启自治总开关）</div>
                  <div>第2步：workflow.autonomy.audit_only = 1（保持审计模式，观察1-2天）</div>
                  <div>第3步：workflow.autonomy.policy_engine_enabled = 1 + risk_gate_enabled = 1</div>
                  <div>第4步：audit_only = 0（正式接管，慎重！）</div>
                  <div>第5步（可选）：strategy_enabled = 1（开启 L5 策略函数）</div>
                  <div>第6步（可选）：meta_cognition_enabled = 1（开启 L7 元认知观测）</div>
                </div>
              </template>
            </Alert>

            <!-- 配置项列表 -->
            <div
              v-for="(item, index) in group.items"
              :key="item.id"
              class="config-row"
              :class="{ 'border-t border-gray-100': index > 0 }"
            >
              <!-- 左侧：类型Tag + 键名 + 说明 -->
              <div class="config-row-left">
                <Tag
                  :color="TYPE_COLOR[item.configType ?? ''] ?? 'default'"
                  class="config-type-tag shrink-0"
                >
                  {{ item.configType || 'str' }}
                </Tag>
                <div class="min-w-0 flex-1">
                  <div class="config-key">{{ item.configKey }}</div>
                  <Tooltip v-if="item.description" :title="item.description" placement="topLeft">
                    <div class="config-desc">{{ truncate(item.description, 60) }}</div>
                  </Tooltip>
                </div>
              </div>

              <!-- 右侧：值 + 操作 -->
              <div class="config-row-right">
                <!-- 0/1 开关类型 -->
                <template v-if="isBoolInt(item)">
                  <Tag
                    :color="item.configValue === '1' ? 'success' : 'default'"
                    class="config-value-tag"
                  >
                    {{ item.configValue === '1' ? '开启' : '关闭' }}
                  </Tag>
                </template>
                <!-- 普通值 -->
                <template v-else>
                  <Tooltip
                    v-if="item.configValue && item.configValue.length > 20"
                    :title="item.configValue"
                  >
                    <span class="config-value-text">{{ truncate(item.configValue, 20) }}</span>
                  </Tooltip>
                  <span v-else class="config-value-text">{{ item.configValue || '—' }}</span>
                </template>

                <!-- 操作按钮 -->
                <Space :size="4" class="config-actions">
                  <Button
                    :icon="h(EditOutlined)"
                    size="small"
                    type="link"
                    @click.stop="handleEdit(item)"
                  >
                    编辑
                  </Button>
                  <Button
                    :icon="h(DeleteOutlined)"
                    danger
                    size="small"
                    type="link"
                    @click.stop="handleDelete(item)"
                  >
                    删除
                  </Button>
                </Space>
              </div>
            </div>

            <!-- 空状态 -->
            <div v-if="group.items.length === 0" class="py-6 text-center text-gray-400 text-sm">
              暂无配置项
            </div>
          </Card>
        </template>

        <!-- 全局空状态 -->
        <div v-if="!loading && groupedConfigs.length === 0" class="py-16 text-center text-gray-400">
          暂无配置数据，点击右上角"新建配置"添加
        </div>
      </div>
    </Spin>
  </Page>
</template>

<style scoped>
.config-group-card :deep(.ant-card-head) {
  padding: 0 16px;
  min-height: 48px;
}

.config-group-card :deep(.ant-card-head-title) {
  padding: 10px 0;
}

.config-row {
  display: flex;
  align-items: center;
  padding: 10px 16px;
  gap: 12px;
  transition: background-color 0.15s;
}

.config-row:hover {
  background-color: #fafafa;
}

.config-row-left {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  flex: 1;
  min-width: 0;
}

.config-type-tag {
  margin: 2px 0 0;
  font-size: 11px;
  line-height: 18px;
  padding: 0 5px;
  height: 20px;
  border-radius: 3px;
}

.config-key {
  font-size: 13px;
  font-weight: 500;
  color: #262626;
  font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace;
  line-height: 1.5;
}

.config-desc {
  font-size: 12px;
  color: #8c8c8c;
  line-height: 1.4;
  margin-top: 1px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 500px;
  cursor: default;
}

.config-row-right {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-shrink: 0;
}

.config-value-tag {
  font-size: 12px;
  min-width: 44px;
  text-align: center;
}

.config-value-text {
  font-size: 12px;
  color: #595959;
  font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace;
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  display: inline-block;
  vertical-align: middle;
  cursor: default;
}

.config-actions {
  opacity: 0;
  transition: opacity 0.15s;
}

.config-row:hover .config-actions {
  opacity: 1;
}
</style>

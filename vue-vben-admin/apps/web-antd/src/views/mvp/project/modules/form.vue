<script setup lang="ts">
import { computed, h, ref } from 'vue';
import { useRouter } from 'vue-router';

import { useVbenModal } from '@vben/common-ui';

import { QuestionCircleOutlined, TeamOutlined } from '@ant-design/icons-vue';
import { Checkbox, message, Tag, Tooltip } from 'ant-design-vue';

import { useVbenForm, type VbenFormSchema } from '#/adapter/form';
import { getModelList } from '#/api/ai/model';
import {
  getProjectDetail,
  updateProject,
} from '#/api/mvp/project';
import { type CategoryItem, getCategories, getRolePresets, type RolePresetItem, createProject as workflowCreateProject } from '#/api/mvp/workflow';
import { CODING_FAMILY, engineVersionOptions, executionModeMap, projectCategoryOptions, roleLevelMap } from '#/views/mvp/consts';
import { loadRoleTypeMap } from '#/views/mvp/role-definitions';

const emit = defineEmits<{ success: [] }>();

const router = useRouter();

/** 架构师模型选项（仅 capability=architect） */
const architectModelOptions = ref<{ label: string; value: string }[]>([]);
/** 所有模型选项（编辑时用） */
const allModelOptions = ref<{ label: string; value: string }[]>([]);
/** 当前选择的项目分类 categoryCode */
const selectedCategoryCode = ref<string>('software_dev');
/** 当前分类的默认角色预设（用于只读展示） */
const defaultPresets = ref<RolePresetItem[]>([]);
/** 用户选中的预设 ID 集合 */
const selectedPresetIDs = ref<Set<string>>(new Set());
const dynamicRoleTypeMap = ref<Record<string, { color: string; label: string; }>>({});

/** 切换预设选中状态 */
function togglePreset(presetId: string) {
  const newSet = new Set(selectedPresetIDs.value);
  if (newSet.has(presetId)) {
    newSet.delete(presetId);
  } else {
    newSet.add(presetId);
  }
  selectedPresetIDs.value = newSet;
}

/** 动态分类选项（从 API 加载，value=categoryCode, label=displayName） */
const dynamicCategoryOptions = ref<{ familyCode: string; label: string; value: string; }[]>([]);
/** 分类配置缓存 */
const categoryMap = ref<Map<string, CategoryItem>>(new Map());

/** workDir 是否必填（编码类 family 必填，其他可选） */
const isWorkDirRequired = computed(() => {
  const cat = categoryMap.value.get(selectedCategoryCode.value);
  return cat ? cat.familyCode === CODING_FAMILY : selectedCategoryCode.value === 'software_dev';
});

/** 加载分类列表（从后端 API） */
async function loadCategories() {
  try {
    const res = await getCategories();
    const list = res?.list ?? [];
    dynamicCategoryOptions.value = list.map((c) => ({
      label: c.displayName,
      value: c.categoryCode,
      familyCode: c.familyCode,
    }));
    categoryMap.value = new Map(list.map((c) => [c.categoryCode, c]));
  } catch {
    // 兜底使用静态选项（value 仍为中文展示名）
    dynamicCategoryOptions.value = projectCategoryOptions.map((o) => ({
      ...o,
      familyCode: 'coding',
    }));
  }
}

/** 根据 categoryCode 加载角色预设，更新架构师模型选项 */
async function loadPresetsForCategory(categoryCode: string) {
  if (!categoryCode) return;
  selectedCategoryCode.value = categoryCode;
  try {
    const presetRes = await getRolePresets(categoryCode);
    const presets = presetRes?.list ?? [];
    defaultPresets.value = presets;
    selectedPresetIDs.value = new Set();
    formApi.setValues({ architectModelID: undefined });
  } catch { /* ignore */ }
}

/** 渲染带 Tooltip 的表单 label */
function tooltipLabel(label: string, tip: string) {
  return () => h('span', {}, [
    label + ' ',
    h(Tooltip, { title: tip }, {
      default: () => h(QuestionCircleOutlined, { style: { color: '#999', marginLeft: '4px' } }),
    }),
  ]);
}

const isEdit = ref(false);
const editId = ref('');

/** 新建时的表单字段 */
const createSchema: VbenFormSchema[] = [
  {
    component: 'Input',
    fieldName: 'name',
    label: '项目名称',
    rules: 'required',
    componentProps: { placeholder: '请输入项目名称', maxlength: 200 },
  },
  {
    component: 'Select',
    fieldName: 'categoryCode',
    label: '项目分类',
    rules: 'selectRequired',
    componentProps: { options: dynamicCategoryOptions, placeholder: '请选择项目分类', allowClear: true, onChange: (val: string) => loadPresetsForCategory(val) },
    defaultValue: 'software_dev',
  },
  {
    component: 'Textarea',
    fieldName: 'description',
    label: '项目简介',
    componentProps: { placeholder: '请输入项目简介', rows: 4, maxlength: 65_535 },
  },
  {
    component: 'Input',
    fieldName: 'workDir',
    label: tooltipLabel('工作目录', '编码类项目必填（Aider代码编辑的根目录）；非编码类可留空（系统自动生成）'),
    dependencies: {
      triggerFields: ['categoryCode'],
      rules: () => (isWorkDirRequired.value ? 'required' : null),
    },
    componentProps: { placeholder: computed(() => isWorkDirRequired.value ? '如：/www/wwwroot/project/my-app' : '可留空，系统自动生成'), maxlength: 500 },
  },
  {
    component: 'Select',
    fieldName: 'architectModelID',
    label: tooltipLabel('架构师AI模型', '可选。选择后会生成项目级架构师配置；不选则项目对话和任务直接走分类默认预设。'),
    componentProps: { options: architectModelOptions, placeholder: '可不选；需要项目级架构师配置时再选择', allowClear: true, showSearch: true, optionFilterProp: 'label', class: 'w-full' },
  },
  {
    component: 'Select',
    fieldName: 'engineVersion',
    label: tooltipLabel('引擎版本', 'Workflow V2 使用 plan_version + 蓝图，支持阶段化流程'),
    componentProps: { options: engineVersionOptions, placeholder: '请选择引擎版本' },
    defaultValue: 'workflow_v2',
  },
];

/** 编辑时的表单字段 */
const editSchema: VbenFormSchema[] = [
  {
    component: 'Input',
    fieldName: 'name',
    label: '项目名称',
    rules: 'required',
    componentProps: { placeholder: '请输入项目名称', maxlength: 200 },
  },
  {
    component: 'Select',
    fieldName: 'projectCategory',
    label: '项目分类',
    componentProps: { options: projectCategoryOptions, placeholder: '请选择项目分类', allowClear: true },
  },
  {
    component: 'Textarea',
    fieldName: 'description',
    label: '项目简介',
    componentProps: { placeholder: '请输入项目简介', rows: 4, maxlength: 65_535 },
  },
  {
    component: 'Input',
    fieldName: 'workDir',
    label: tooltipLabel('代码工作目录', 'Aider 执行代码编辑时的项目根目录，确认方案前必须填写'),
    componentProps: { placeholder: '如：/www/wwwroot/project/my-app', maxlength: 500 },
  },
  {
    component: 'Textarea',
    fieldName: 'pauseReason',
    label: '暂停原因',
    componentProps: { placeholder: '请输入暂停原因', rows: 4, maxlength: 65_535 },
  },
  {
    component: 'Textarea',
    fieldName: 'globalContext',
    label: tooltipLabel('项目全局上下文', '架构师需求分析+方案设计的压缩摘要'),
    componentProps: { placeholder: '请输入项目全局上下文（架构师需求分析+方案设计的压缩摘要）', rows: 4, maxlength: 65_535 },
  },
  {
    component: 'Select',
    fieldName: 'architectModelID',
    label: '架构师AI模型',
    componentProps: { options: allModelOptions, placeholder: '请选择架构师使用的AI模型', allowClear: true, showSearch: true, optionFilterProp: 'label', class: 'w-full' },
  },
];

/** 表单配置 */
const [Form, formApi] = useVbenForm({
  showDefaultActions: false,
  schema: createSchema,
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
        await updateProject({ id: editId.value, ...values });
        message.success('更新成功');
        emit('success');
        modalApi.close();
      } else {
        // 调用工作流 API 创建项目（发送 categoryCode，后端自动映射 displayName）
        const res = await workflowCreateProject({
          name: values.name,
          categoryCode: values.categoryCode || 'software_dev',
          description: values.description || '',
          workDir: values.workDir,
          architectModelID: values.architectModelID,
          engineVersion: values.engineVersion || 'workflow_v2',
          selectedRoles: [...selectedPresetIDs.value].map((id) => ({ presetID: id })),
        });
        message.success('项目创建成功，正在跳转到对话页面...');
        emit('success');
        modalApi.close();
        // 跳转到对话页面
        router.push({
          path: '/mvp/chat',
          query: {
            projectId: res.projectID,
            conversationId: res.conversationID,
          },
        });
      }
    } finally {
      modalApi.lock(false);
    }
  },
  async onOpenChange(isOpen: boolean) {
    if (isOpen) {
      const data = modalApi.getData<null | { id?: string }>();
      if (data?.id) {
        // 编辑模式：加载所有模型
        isEdit.value = true;
        editId.value = data.id;
        modalApi.setState({ title: '编辑项目' });
        formApi.setState({ schema: editSchema });
        try {
          const modelRes = await getModelList({ pageNum: 1, pageSize: 1000 });
          allModelOptions.value = (modelRes?.list ?? []).map((item: any) => ({
            label: `${item.name} (${item.modelCode})`,
            value: item.id,
          }));
        } catch { /* ignore */ }
        try {
          const detail = await getProjectDetail(data.id);
          if (detail) {
            formApi.setValues(detail);
          }
        } catch {
          message.error('获取详情失败');
        }
      } else {
        // 新建模式：加载分类列表 + 架构师模型 + 预设
        isEdit.value = false;
        editId.value = '';
        modalApi.setState({ title: '新建项目' });
        formApi.setState({ schema: createSchema });
        formApi.resetForm();
        try {
          // 并行加载：分类列表、架构师模型列表、默认预设
          const [roleTypeMap, , modelRes, presetRes] = await Promise.all([
            loadRoleTypeMap(),
            loadCategories(),
            getModelList({ pageNum: 1, pageSize: 1000, capability: 'architect' }),
            getRolePresets('software_dev'),
          ]);
          dynamicRoleTypeMap.value = roleTypeMap;
          // 架构师模型下拉选项
          architectModelOptions.value = (modelRes?.list ?? []).map((item: any) => ({
            label: `${item.name} (${item.modelCode})`,
            value: item.id,
          }));
          // 保存默认预设列表用于展示
          defaultPresets.value = presetRes?.list ?? [];
          selectedPresetIDs.value = new Set();
          formApi.setValues({ architectModelID: undefined });
        } catch { /* ignore */ }
      }
    }
  },
});
</script>

<template>
  <Modal class="w-[600px]">
    <Form />
    <!-- AI 团队角色选择（仅新建模式） -->
    <div v-if="!isEdit && defaultPresets.length > 0" class="mt-4 rounded border border-gray-200 bg-gray-50 p-3">
      <div class="mb-2 flex items-center text-sm font-medium text-gray-600">
        <TeamOutlined class="mr-1" />
        项目角色配置
        <span class="ml-2 text-xs font-normal text-gray-400">可选。仅勾选需要落为项目级覆盖配置的角色；不勾选则运行时走分类默认预设</span>
      </div>
      <div class="flex flex-col gap-2">
        <div
          v-for="preset in defaultPresets"
          :key="preset.id"
          class="flex cursor-pointer items-center gap-2 rounded px-2 py-1 transition-colors hover:bg-gray-100"
          @click="togglePreset(preset.id)"
        >
          <Checkbox
            :checked="selectedPresetIDs.has(preset.id)"
            @click.stop
            @change="togglePreset(preset.id)"
          />
          <Tag :color="dynamicRoleTypeMap[preset.roleType]?.color || 'default'">
            {{ dynamicRoleTypeMap[preset.roleType]?.label || preset.roleType }}
          </Tag>
          <span class="text-xs text-gray-500">
            {{ roleLevelMap[preset.roleLevel]?.label || preset.roleLevel }}
            · {{ executionModeMap[preset.executionMode]?.label || preset.executionMode }}
            · {{ preset.modelName || '未绑定' }}
          </span>
          <Tag v-if="preset.isDefault" class="ml-auto" color="blue" size="small">默认</Tag>
          <Tag v-else class="ml-auto" color="default" size="small">扩展</Tag>
        </div>
      </div>
    </div>
  </Modal>
</template>

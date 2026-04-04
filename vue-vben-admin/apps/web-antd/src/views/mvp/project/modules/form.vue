<script setup lang="ts">
import { h, ref, computed } from 'vue';
import { useRouter } from 'vue-router';
import { useVbenModal } from '@vben/common-ui';
import { useVbenForm } from '#/adapter/form';
import { message, Tooltip, Tag } from 'ant-design-vue';
import { QuestionCircleOutlined, TeamOutlined } from '@ant-design/icons-vue';
import {
  getProjectDetail,
  updateProject,
} from '#/api/mvp/project';
import { createProject as workflowCreateProject, getRolePresets, type RolePresetItem } from '#/api/mvp/workflow';
import { getModelList } from '#/api/ai/model';
import { projectCategoryOptions, engineVersionOptions, roleTypeMap, roleLevelMap, executionModeMap } from '#/views/mvp/consts';

const router = useRouter();

/** 架构师模型选项（仅 capability=architect） */
const architectModelOptions = ref<{ label: string; value: string }[]>([]);
/** 所有模型选项（编辑时用） */
const allModelOptions = ref<{ label: string; value: string }[]>([]);
/** 预设中的架构师默认模型ID */
const presetArchitectModelID = ref<string>('');
/** 当前选择的项目分类 */
const selectedCategory = ref<string>('软件开发');
/** 当前分类的默认角色预设（用于只读展示） */
const defaultPresets = ref<RolePresetItem[]>([]);

/** 编码类分类（workDir 必填） */
const codingCategories = ['软件开发', '游戏开发'];

/** workDir 是否必填（编码类必填，其他可选） */
const isWorkDirRequired = computed(() => codingCategories.includes(selectedCategory.value));

/** 根据项目分类加载角色预设，更新架构师模型选项 */
async function loadPresetsForCategory(category: string) {
  if (!category) return;
  selectedCategory.value = category;
  try {
    const presetRes = await getRolePresets(category);
    const presets = presetRes?.list ?? [];
    defaultPresets.value = presets;
    // 从预设中提取架构师模型
    const architectPreset = presets.find((p) => p.roleType === 'architect');
    if (architectPreset?.modelID) {
      // 确保预设中的模型在下拉列表中
      const exists = architectModelOptions.value.some((o) => o.value === architectPreset.modelID);
      if (!exists) {
        architectModelOptions.value.push({
          label: `${architectPreset.modelName || '预设模型'}`,
          value: architectPreset.modelID,
        });
      }
      presetArchitectModelID.value = architectPreset.modelID;
      formApi.setValues({ architectModelID: architectPreset.modelID });
    }
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

const emit = defineEmits<{ success: [] }>();
const isEdit = ref(false);
const editId = ref('');

/** 新建时的表单字段 */
const createSchema = [
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
    rules: 'selectRequired',
    componentProps: { options: projectCategoryOptions, placeholder: '请选择项目分类', allowClear: true, onChange: (val: string) => loadPresetsForCategory(val) },
    defaultValue: '软件开发',
  },
  {
    component: 'Textarea',
    fieldName: 'description',
    label: '项目简介',
    componentProps: { placeholder: '请输入项目简介', rows: 4, maxlength: 65535 },
  },
  {
    component: 'Input',
    fieldName: 'workDir',
    label: tooltipLabel('工作目录', '编码类项目必填（Aider代码编辑的根目录）；非编码类可留空（系统自动生成）'),
    rules: computed(() => isWorkDirRequired.value ? 'required' : undefined),
    componentProps: { placeholder: computed(() => isWorkDirRequired.value ? '如：/www/wwwroot/project/my-app' : '可留空，系统自动生成'), maxlength: 500 },
  },
  {
    component: 'Select',
    fieldName: 'architectModelID',
    label: tooltipLabel('架构师AI模型', '默认值来自角色预设模板，可修改。仅显示角色为「架构师」的模型'),
    rules: 'selectRequired',
    componentProps: { options: architectModelOptions, placeholder: '请选择架构师使用的AI模型', allowClear: true, showSearch: true, optionFilterProp: 'label', class: 'w-full' },
  },
  {
    component: 'Select',
    fieldName: 'engineVersion',
    label: tooltipLabel('引擎版本', '旧版（Legacy）直接写 mvp_task；新版（WorkflowV2）使用 plan_version + 蓝图，支持阶段化流程'),
    componentProps: { options: engineVersionOptions, placeholder: '请选择引擎版本' },
    defaultValue: 'legacy',
  },
];

/** 编辑时的表单字段 */
const editSchema = [
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
    componentProps: { placeholder: '请输入项目简介', rows: 4, maxlength: 65535 },
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
    componentProps: { placeholder: '请输入暂停原因', rows: 4, maxlength: 65535 },
  },
  {
    component: 'Textarea',
    fieldName: 'globalContext',
    label: tooltipLabel('项目全局上下文', '架构师需求分析+方案设计的压缩摘要'),
    componentProps: { placeholder: '请输入项目全局上下文（架构师需求分析+方案设计的压缩摘要）', rows: 4, maxlength: 65535 },
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
        // 调用工作流 API 创建项目（会自动根据预设创建全部角色配置）
        const res = await workflowCreateProject({
          name: values.name,
          projectCategory: values.projectCategory || '软件开发',
          description: values.description || '',
          workDir: values.workDir,
          architectModelID: values.architectModelID,
          engineVersion: values.engineVersion || 'legacy',
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
      const data = modalApi.getData<{ id?: string } | null>();
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
        // 新建模式：只加载架构师模型 + 读取预设默认值
        isEdit.value = false;
        editId.value = '';
        modalApi.setState({ title: '新建项目' });
        formApi.setState({ schema: createSchema });
        formApi.resetForm();
        try {
          // 并行加载架构师模型列表和预设
          const [modelRes, presetRes] = await Promise.all([
            getModelList({ pageNum: 1, pageSize: 1000, capability: 'architect' }),
            getRolePresets('软件开发'),
          ]);
          // 架构师模型下拉选项
          architectModelOptions.value = (modelRes?.list ?? []).map((item: any) => ({
            label: `${item.name} (${item.modelCode})`,
            value: item.id,
          }));
          // 保存默认预设列表用于展示
          defaultPresets.value = presetRes?.list ?? [];
          // 读取预设中架构师的默认模型
          const architectPreset = defaultPresets.value.find(
            (p) => p.roleType === 'architect',
          );
          if (architectPreset?.modelID) {
            presetArchitectModelID.value = architectPreset.modelID;
            formApi.setValues({ architectModelID: architectPreset.modelID });
          }
        } catch { /* ignore */ }
      }
    }
  },
});
</script>

<template>
  <Modal class="w-[600px]">
    <Form />
    <!-- 默认 AI 团队组成预览（仅新建模式） -->
    <div v-if="!isEdit && defaultPresets.length > 0" class="mt-4 rounded border border-gray-200 bg-gray-50 p-3">
      <div class="mb-2 flex items-center text-sm font-medium text-gray-600">
        <TeamOutlined class="mr-1" />
        默认 AI 团队组成
      </div>
      <div class="flex flex-wrap gap-2">
        <div
          v-for="preset in defaultPresets"
          :key="`${preset.roleType}-${preset.roleLevel}`"
          class="flex items-center gap-1"
        >
          <Tag :color="roleTypeMap[preset.roleType]?.color || 'default'">
            {{ roleTypeMap[preset.roleType]?.label || preset.roleType }}
          </Tag>
          <span class="text-xs text-gray-500">
            {{ roleLevelMap[preset.roleLevel]?.label || preset.roleLevel }}
            · {{ executionModeMap[preset.executionMode]?.label || preset.executionMode }}
            · {{ preset.modelName || '未绑定' }}
          </span>
        </div>
      </div>
    </div>
  </Modal>
</template>

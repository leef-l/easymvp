<script setup lang="ts">
{{- if .HasTooltip}}
import { h, ref } from 'vue';
{{- else}}
import { ref } from 'vue';
{{- end}}
import { useVbenModal } from '@vben/common-ui';
import { useVbenForm } from '#/adapter/form';
{{- if .HasTooltip}}
import { message, Tooltip } from 'ant-design-vue';
import { QuestionCircleOutlined } from '@ant-design/icons-vue';
{{- else}}
import { message } from 'ant-design-vue';
{{- end}}
import {
  get{{.ModelName}}Detail,
  create{{.ModelName}},
  update{{.ModelName}},{{if .HasParentID}}
  get{{.ModelName}}Tree,{{end}}
} from '#/api/{{.AppName}}/{{.ModuleName}}';
{{- if .HasDict}}
import { getDictByType } from '#/api/system/dict';
{{- end}}
{{- if or .HasParentID .HasTreeSelect}}
import type { {{.ModelName}}Item } from '#/api/{{.AppName}}/{{.ModuleName}}/types';

const treeData = ref<{{.ModelName}}Item[]>([]);
{{- end}}
{{- range .Fields}}
{{- if and .IsForeignKey (not .IsHidden) .RefTable}}
{{- if .RefIsTree}}
import { get{{.RefTableCamel}}Tree } from '#/api/{{$.AppName}}/{{.RefTable}}';
{{- else}}
import { get{{.RefTableCamel}}List } from '#/api/{{$.AppName}}/{{.RefTable}}';
{{- end}}
{{- end}}
{{- end}}
{{range .Fields}}
{{- if and (not .IsHidden) (not .IsID) (.IsEnum) (ne .Component "Switch")}}
/** {{.Label}}选项 */
const {{.NameLower}}Options = [
{{- range .EnumValues}}
  { label: '{{.Label}}', value: {{.Value}} },
{{- end}}
];
{{end}}
{{- end}}
{{- range .Fields}}
{{- if and .IsForeignKey (not .IsHidden) .RefTable}}
const {{.NameLower}}Options = ref<{ label: string; value: string }[]>([]);
{{- end}}
{{- end}}
{{- range .Fields}}
{{- if and (not .IsHidden) .DictType}}
const {{.NameLower}}DictOptions = ref<{ label: string; value: string | number }[]>([]);
{{- end}}
{{- end}}
{{- if .HasTooltip}}
/** 渲染带 Tooltip 的表单 label */
function tooltipLabel(label: string, tip: string) {
  return () => h('span', {}, [
    label + ' ',
    h(Tooltip, { title: tip }, {
      default: () => h(QuestionCircleOutlined, { style: { color: '#999', marginLeft: '4px' } }),
    }),
  ]);
}
{{- end}}

const emit = defineEmits<{ success: [] }>();
const isEdit = ref(false);
const editId = ref('');

/** 表单配置 */
const [Form, formApi] = useVbenForm({
  showDefaultActions: false,
  schema: [
{{- range .Fields}}
{{- if and (not .IsHidden) (not .IsID)}}
{{- if eq .Component "Password"}}
    {
      component: 'InputPassword',
      fieldName: '{{.NameLower}}',
      label: {{if .TooltipText}}tooltipLabel('{{.ShortLabel}}', '{{.TooltipText}}'){{else}}'{{.Label}}'{{end}},
      dependencies: {
        triggerFields: ['{{.NameLower}}'],
        rules: () => (isEdit.value ? undefined : 'required'),
        componentProps: () => ({
          placeholder: isEdit.value ? '不填则不修改' : '请输入{{.Label}}',
        }),
      },
    },
{{- else if eq .Component "InputNumber"}}
    {
      component: 'InputNumber',
      fieldName: '{{.NameLower}}',
      label: {{if .TooltipText}}tooltipLabel('{{.ShortLabel}}', '{{.TooltipText}}'){{else}}'{{.Label}}'{{end}},
{{- if .IsRequired}}
      rules: 'required',
{{- end}}
      componentProps: { placeholder: '请输入{{.Label}}', class: 'w-full' },
    },
{{- else if eq .Component "Textarea"}}
    {
      component: 'Textarea',
      fieldName: '{{.NameLower}}',
      label: {{if .TooltipText}}tooltipLabel('{{.ShortLabel}}', '{{.TooltipText}}'){{else}}'{{.Label}}'{{end}},
{{- if .IsRequired}}
      rules: 'required',
{{- end}}
      componentProps: { placeholder: '请输入{{.Label}}', rows: 4{{if gt .MaxLength 0}}, maxlength: {{.MaxLength}}{{end}} },
    },
{{- else if eq .Component "Switch"}}
    {
      component: 'Switch',
      fieldName: '{{.NameLower}}',
      label: {{if .TooltipText}}tooltipLabel('{{.ShortLabel}}', '{{.TooltipText}}'){{else}}'{{.Label}}'{{end}},
      componentProps: { checkedValue: 1, unCheckedValue: 0 },
      defaultValue: {{if .DefaultValue}}{{.DefaultValue}}{{else}}0{{end}},
    },
{{- else if eq .Component "Radio"}}
    {
      component: 'RadioGroup',
      fieldName: '{{.NameLower}}',
      label: {{if .TooltipText}}tooltipLabel('{{.ShortLabel}}', '{{.TooltipText}}'){{else}}'{{.Label}}'{{end}},
{{- if .IsRequired}}
      rules: 'required',
{{- end}}
      componentProps: { options: {{.NameLower}}Options },
    },
{{- else if eq .Component "Select"}}
{{- if .IsEnum}}
    {
      component: 'Select',
      fieldName: '{{.NameLower}}',
      label: {{if .TooltipText}}tooltipLabel('{{.ShortLabel}}', '{{.TooltipText}}'){{else}}'{{.Label}}'{{end}},
{{- if .IsRequired}}
      rules: 'selectRequired',
{{- end}}
      componentProps: { options: {{.NameLower}}Options, placeholder: '请选择{{.Label}}', allowClear: true, class: 'w-full' },
    },
{{- else if .DictType}}
    {
      component: 'Select',
      fieldName: '{{.NameLower}}',
      label: {{if .TooltipText}}tooltipLabel('{{.ShortLabel}}', '{{.TooltipText}}'){{else}}'{{.Label}}'{{end}},
{{- if .IsRequired}}
      rules: 'selectRequired',
{{- end}}
      componentProps: { options: {{.NameLower}}DictOptions, placeholder: '请选择{{.Label}}', allowClear: true, class: 'w-full' },
    },
{{- else if .IsForeignKey}}
{{- if .RefIsTree}}
    {
      component: 'TreeSelect',
      fieldName: '{{.NameLower}}',
      label: {{if .TooltipText}}tooltipLabel('{{.ShortLabel}}', '{{.TooltipText}}'){{else}}'{{.Label}}'{{end}},
{{- if .IsRequired}}
      rules: 'selectRequired',
{{- end}}
      componentProps: {
        treeData: {{.NameLower}}Options.value,
        fieldNames: { label: '{{if .RefDisplayField}}{{.RefDisplayLower}}{{else}}title{{end}}', value: 'id', children: 'children' },
        placeholder: '请选择{{.Label}}',
        allowClear: true,
        treeDefaultExpandAll: true,
        class: 'w-full',
      },
    },
{{- else}}
    {
      component: 'Select',
      fieldName: '{{.NameLower}}',
      label: {{if .TooltipText}}tooltipLabel('{{.ShortLabel}}', '{{.TooltipText}}'){{else}}'{{.Label}}'{{end}},
{{- if .IsRequired}}
      rules: 'selectRequired',
{{- end}}
      componentProps: { options: {{.NameLower}}Options, placeholder: '请选择{{.Label}}', allowClear: true, class: 'w-full' },
    },
{{- end}}
{{- else}}
    {
      component: 'Input',
      fieldName: '{{.NameLower}}',
      label: {{if .TooltipText}}tooltipLabel('{{.ShortLabel}}', '{{.TooltipText}}'){{else}}'{{.Label}}'{{end}},
{{- if .IsRequired}}
      rules: 'required',
{{- end}}
      componentProps: { placeholder: '请输入{{.Label}}' },
    },
{{- end}}
{{- else if eq .Component "SelectMulti"}}
{{- if .IsEnum}}
    {
      component: 'Select',
      fieldName: '{{.NameLower}}',
      label: {{if .TooltipText}}tooltipLabel('{{.ShortLabel}}', '{{.TooltipText}}'){{else}}'{{.Label}}'{{end}},
{{- if .IsRequired}}
      rules: 'selectRequired',
{{- end}}
      componentProps: { options: {{.NameLower}}Options, placeholder: '请选择{{.Label}}', mode: 'multiple', allowClear: true, class: 'w-full' },
    },
{{- else}}
    {
      component: 'Select',
      fieldName: '{{.NameLower}}',
      label: {{if .TooltipText}}tooltipLabel('{{.ShortLabel}}', '{{.TooltipText}}'){{else}}'{{.Label}}'{{end}},
{{- if .IsRequired}}
      rules: 'selectRequired',
{{- end}}
      componentProps: { placeholder: '请输入{{.Label}}', mode: 'tags', allowClear: true, class: 'w-full' },
    },
{{- end}}
{{- else if eq .Component "TreeSelectSingle"}}
    {
      component: 'TreeSelect',
      fieldName: '{{.NameLower}}',
      label: {{if .TooltipText}}tooltipLabel('{{.ShortLabel}}', '{{.TooltipText}}'){{else}}'{{.Label}}'{{end}},
{{- if .IsRequired}}
      rules: 'selectRequired',
{{- end}}
      componentProps: {
        treeData: treeData.value,
        fieldNames: { label: '{{if .RefDisplayLower}}{{.RefDisplayLower}}{{else}}title{{end}}', value: 'id', children: 'children' },
        placeholder: '请选择{{.Label}}',
        allowClear: true,
        treeDefaultExpandAll: true,
        class: 'w-full',
      },
    },
{{- else if eq .Component "TreeSelectMulti"}}
    {
      component: 'TreeSelect',
      fieldName: '{{.NameLower}}',
      label: {{if .TooltipText}}tooltipLabel('{{.ShortLabel}}', '{{.TooltipText}}'){{else}}'{{.Label}}'{{end}},
{{- if .IsRequired}}
      rules: 'selectRequired',
{{- end}}
      componentProps: {
        treeData: treeData.value,
        fieldNames: { label: '{{if .RefDisplayLower}}{{.RefDisplayLower}}{{else}}title{{end}}', value: 'id', children: 'children' },
        placeholder: '请选择{{.Label}}',
        allowClear: true,
        treeCheckable: true,
        treeDefaultExpandAll: true,
        class: 'w-full',
      },
    },
{{- else if eq .Component "DateTimePicker"}}
    {
      component: 'DatePicker',
      fieldName: '{{.NameLower}}',
      label: {{if .TooltipText}}tooltipLabel('{{.ShortLabel}}', '{{.TooltipText}}'){{else}}'{{.Label}}'{{end}},
{{- if .IsRequired}}
      rules: 'required',
{{- end}}
      componentProps: { showTime: true, placeholder: '请选择{{.Label}}', class: 'w-full', valueFormat: 'YYYY-MM-DD HH:mm:ss' },
    },
{{- else if eq .Component "ImageUpload"}}
    {
      component: 'ImageUpload',
      fieldName: '{{.NameLower}}',
      label: '{{.Label}}',
{{- if .IsRequired}}
      rules: 'required',
{{- end}}
      componentProps: { maxCount: 1 },
    },
{{- else if eq .Component "FileUpload"}}
    {
      component: 'FileUpload',
      fieldName: '{{.NameLower}}',
      label: '{{.Label}}',
{{- if .IsRequired}}
      rules: 'required',
{{- end}}
      componentProps: { maxCount: 1 },
    },
{{- else if eq .Component "RichText"}}
    {
      component: 'RichText',
      fieldName: '{{.NameLower}}',
      label: '{{.Label}}',
{{- if .IsRequired}}
      rules: 'required',
{{- end}}
      formItemClass: 'col-span-full',
    },
{{- else if eq .Component "JsonEditor"}}
    {
      component: 'JsonEditor',
      fieldName: '{{.NameLower}}',
      label: '{{.Label}}',
{{- if .IsRequired}}
      rules: 'required',
{{- end}}
      formItemClass: 'col-span-full',
    },
{{- else if eq .Component "IconPicker"}}
    {
      component: 'IconPicker',
      fieldName: '{{.NameLower}}',
      label: {{if .TooltipText}}tooltipLabel('{{.ShortLabel}}', '{{.TooltipText}}'){{else}}'{{.Label}}'{{end}},
{{- if .IsRequired}}
      rules: 'required',
{{- end}}
      componentProps: { placeholder: '请选择图标' },
    },
{{- else if eq .Component "InputUrl"}}
    {
      component: 'Input',
      fieldName: '{{.NameLower}}',
      label: {{if .TooltipText}}tooltipLabel('{{.ShortLabel}}', '{{.TooltipText}}'){{else}}'{{.Label}}'{{end}},
{{- if or .IsRequired (eq .FrontendRules "url")}}
      rules: [
{{- if .IsRequired}}
        { required: true, message: '{{.Label}}不能为空' },
{{- end}}
{{- if eq .FrontendRules "url"}}
        { type: 'url', message: '请输入正确的URL地址' },
{{- end}}
      ],
{{- end}}
      componentProps: { placeholder: '请输入URL地址'{{if gt .MaxLength 0}}, maxlength: {{.MaxLength}}{{end}}, addonBefore: 'https://' },
    },
{{- else}}
    {
      component: 'Input',
      fieldName: '{{.NameLower}}',
      label: {{if .TooltipText}}tooltipLabel('{{.ShortLabel}}', '{{.TooltipText}}'){{else}}'{{.Label}}'{{end}},
{{- if or .IsRequired (eq .FrontendRules "email") (eq .FrontendRules "phone") (eq .FrontendRules "url")}}
      rules: [
{{- if .IsRequired}}
        { required: true, message: '{{.Label}}不能为空' },
{{- end}}
{{- if eq .FrontendRules "email"}}
        { type: 'email', message: '请输入正确的邮箱地址' },
{{- end}}
{{- if eq .FrontendRules "phone"}}
        { pattern: /^1\d{10}$/, message: '请输入正确的手机号' },
{{- end}}
{{- if eq .FrontendRules "url"}}
        { type: 'url', message: '请输入正确的URL地址' },
{{- end}}
      ],
{{- end}}
      componentProps: { placeholder: '请输入{{.Label}}'{{if gt .MaxLength 0}}, maxlength: {{.MaxLength}}{{end}} },
    },
{{- end}}
{{- end}}
{{- end}}
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
        await update{{.ModelName}}({ id: editId.value, ...values });
        message.success('更新成功');
      } else {
        await create{{.ModelName}}(values);
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
{{- if .HasParentID}}
      // 加载树形数据
      try {
        const res = await get{{.ModelName}}Tree();
        treeData.value = [
          { id: '0', title: '顶级节点', children: res ?? [] } as {{.ModelName}}Item,
        ];
        formApi.updateSchema([
          {
            fieldName: '{{range .Fields}}{{if .IsParentID}}{{.NameLower}}{{end}}{{end}}',
            componentProps: { treeData: treeData.value },
          },
        ]);
      } catch {
        // ignore
      }
{{- end}}
{{- range .Fields}}
{{- if and .IsForeignKey (not .IsHidden) .RefTable}}
{{- if .RefIsTree}}
      // 加载{{.Label}}树形数据
      try {
        const {{.RefTableLower}}Res = await get{{.RefTableCamel}}Tree();
        {{.NameLower}}Options.value = {{.RefTableLower}}Res ?? [];
        formApi.updateSchema([
          {
            fieldName: '{{.NameLower}}',
            componentProps: { treeData: {{.NameLower}}Options.value },
          },
        ]);
      } catch {
        // ignore
      }
{{- else}}
      // 加载{{.Label}}选项
      try {
        const {{.RefTableLower}}Res = await get{{.RefTableCamel}}List({ pageNum: 1, pageSize: 1000 });
        {{.NameLower}}Options.value = ({{.RefTableLower}}Res?.list ?? []).map((item: any) => ({
          label: item.{{.RefDisplayLower}} || item.id,
          value: item.id,
        }));
      } catch {
        // ignore
      }
{{- end}}
{{- end}}
{{- end}}
{{- range .Fields}}
{{- if and (not .IsHidden) .DictType}}
      // 加载{{.Label}}字典
      try {
        const dictRes = await getDictByType('{{.DictType}}');
        {{.NameLower}}DictOptions.value = (dictRes ?? []).map((item: any) => ({
          label: item.label,
          value: item.value,
        }));
      } catch {
        // ignore
      }
{{- end}}
{{- end}}
      if (data?.id) {
        isEdit.value = true;
        editId.value = data.id;
        modalApi.setState({ title: '编辑{{.Comment}}' });
        try {
          const detail = await get{{.ModelName}}Detail(data.id);
          if (detail) {
            formApi.setValues(detail);
          }
        } catch {
          message.error('获取详情失败');
        }
      } else {
        isEdit.value = false;
        editId.value = '';
        modalApi.setState({ title: '新建{{.Comment}}' });
        formApi.resetForm();
      }
    }
  },
});
</script>

<template>
  <Modal class="w-[{{if .HasRichText}}800px{{else}}600px{{end}}]">
    <Form />
  </Modal>
</template>

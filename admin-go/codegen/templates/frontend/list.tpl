<script setup lang="ts">
{{- if .HasTooltip}}
import { h } from 'vue';
{{- end}}
import type { VbenFormProps } from '#/adapter/form';
import type { VxeGridProps } from '#/adapter/vxe-table';

import { Page, useVbenModal } from '@vben/common-ui';
{{- if .HasTooltip}}
import { Button, message, Modal, Tag, Tooltip } from 'ant-design-vue';
import { QuestionCircleOutlined } from '@ant-design/icons-vue';
{{- else}}
import { Button, message, Modal, Tag } from 'ant-design-vue';
{{- end}}

import { useVbenVxeGrid } from '#/adapter/vxe-table';
{{- if .HasParentID}}
import { get{{.ModelName}}Tree, delete{{.ModelName}}, export{{.ModelName}}{{if .HasImport}}, import{{.ModelName}}, downloadImportTemplate{{.ModelName}}{{end}} } from '#/api/{{.AppName}}/{{.ModuleName}}';
{{- else}}
import { get{{.ModelName}}List, delete{{.ModelName}}, batchDelete{{.ModelName}}, export{{.ModelName}}{{if .HasImport}}, import{{.ModelName}}, downloadImportTemplate{{.ModelName}}{{end}}{{if .HasBatchEdit}}, batchUpdate{{.ModelName}}{{end}} } from '#/api/{{.AppName}}/{{.ModuleName}}';
{{- end}}
import type { {{.ModelName}}Item } from '#/api/{{.AppName}}/{{.ModuleName}}/types';
import FormModal from './modules/form.vue';
import DetailDrawer from './modules/detail-drawer.vue';

/** 标签颜色池 */
const TAG_COLORS = ['green', 'red', 'blue', 'orange', 'cyan', 'purple', 'geekblue', 'magenta'];
{{range .Fields}}
{{- if and (not .IsHidden) (.IsEnum)}}
/** {{.Label}}选项 */
const {{.NameLower}}Options = [
{{- range .EnumValues}}
  { label: '{{.Label}}', value: {{.Value}} },
{{- end}}
];

/** {{.Label}}映射 */
const {{.NameLower}}Map: Record<number, string> = {
{{- range .EnumValues}}
  {{.Value}}: '{{.Label}}',
{{- end}}
};

/** {{.Label}}颜色 */
function get{{.NameCamel}}Color(val: number): string {
  const keys = [{{range $i, $v := .EnumValues}}{{if $i}}, {{end}}{{$v.Value}}{{end}}];
  const idx = keys.indexOf(val);
  return TAG_COLORS[idx >= 0 ? idx % TAG_COLORS.length : 0] ?? 'default';
}
{{end}}
{{- end}}
{{- if .HasTooltip}}
/** 渲染带 Tooltip 的列标题 */
function tooltipHeader(label: string, tip: string) {
  return () => h('span', {}, [
    label + ' ',
    h(Tooltip, { title: tip }, {
      default: () => h(QuestionCircleOutlined, { style: { color: '#999', marginLeft: '4px' } }),
    }),
  ]);
}
{{- end}}

/** 表单弹窗 */
const [FormModalComp, formModalApi] = useVbenModal({
  connectedComponent: FormModal,
  destroyOnClose: true,
});

/** 详情抽屉 */
const [DetailDrawerComp, detailDrawerApi] = useVbenModal({
  connectedComponent: DetailDrawer,
  destroyOnClose: true,
});

/** 搜索表单配置 */
const formOptions: VbenFormProps = {
  collapsed: false,
  showCollapseButton: true,
  submitOnChange: false,
  submitOnEnter: true,
  schema: [
{{- range .Fields}}
{{- if .IsSearchable}}
    {
      component: 'Input',
      componentProps: { placeholder: '请输入{{.ShortLabel}}', allowClear: true },
      fieldName: '{{.NameLower}}',
      label: '{{.ShortLabel}}',
    },
{{- end}}
{{- end}}
{{- range .Fields}}
{{- if and (not .IsHidden) (not .IsID) (.IsEnum)}}
    {
      component: 'Select',
      componentProps: {
        allowClear: true,
        options: {{.NameLower}}Options,
        placeholder: '请选择{{.Label}}',
        class: 'w-full',
      },
      fieldName: '{{.NameLower}}',
      label: '{{.ShortLabel}}',
    },
{{- end}}
{{- end}}
    {
      component: 'RangePicker',
      fieldName: 'timeRange',
      label: '创建时间',
      componentProps: {
        showTime: true,
        format: 'YYYY-MM-DD HH:mm:ss',
        valueFormat: 'YYYY-MM-DD HH:mm:ss',
        class: 'w-full',
      },
    },
  ],
};

/** 表格列配置 */
const gridOptions: VxeGridProps<{{.ModelName}}Item> = {
  columns: [
{{- if not .HasParentID}}
    { type: 'checkbox', width: 50 },
{{- end}}
    { title: '序号', type: 'seq', width: 50 },
{{- $isTree := .HasParentID}}
{{- $firstDataCol := true}}
{{- range .Fields}}
{{- if and (not .IsHidden) (not .IsID) (not .IsParentID) (not .IsTimeField) (not .IsMultiFK) (not .IsPassword)}}
{{- if .RefFieldJSON}}
    { field: '{{.RefFieldJSON}}', title: '{{.ShortLabel}}'{{if .TooltipText}}, slots: { header: tooltipHeader('{{.ShortLabel}}', '{{.TooltipText}}') }{{end}}{{if and $isTree $firstDataCol}}, treeNode: true{{end}} },
{{- else if .IsEnum}}
    { field: '{{.NameLower}}', title: '{{.ShortLabel}}', width: 120, slots: { default: '{{.NameLower}}_cell' }{{if and $isTree $firstDataCol}}, treeNode: true{{end}} },
{{- else if eq .Component "ImageUpload"}}
    { field: '{{.NameLower}}', title: '{{.ShortLabel}}', width: 100, slots: { default: '{{.NameLower}}_cell' }{{if and $isTree $firstDataCol}}, treeNode: true{{end}} },
{{- else if or (eq .Component "RichText") (eq .Component "JsonEditor")}}
{{- /* 富文本和JSON字段不在列表中显示，不消耗 firstDataCol */}}
{{- else if .IsMoney}}
    { field: '{{.NameLower}}', title: '{{.ShortLabel}}'{{if .TooltipText}}, slots: { header: tooltipHeader('{{.ShortLabel}}', '{{.TooltipText}}') }{{end}}, width: 120, formatter: ({ cellValue }: any) => cellValue != null ? (cellValue / 100).toFixed(2) : '-'{{if and $isTree $firstDataCol}}, treeNode: true{{end}} },
{{- else}}
    { field: '{{.NameLower}}', title: '{{.ShortLabel}}'{{if .TooltipText}}, slots: { header: tooltipHeader('{{.ShortLabel}}', '{{.TooltipText}}') }{{end}}{{if and $isTree $firstDataCol}}, treeNode: true{{end}} },
{{- end}}
{{- if not (or (eq .Component "RichText") (eq .Component "JsonEditor"))}}
{{- $firstDataCol = false}}
{{- end}}
{{- end}}
{{- end}}
{{- range .Fields}}
{{- if and (not .IsHidden) (.IsTimeField)}}
    { field: '{{.NameLower}}', title: '{{.ShortLabel}}'{{if .TooltipText}}, slots: { header: tooltipHeader('{{.ShortLabel}}', '{{.TooltipText}}') }{{end}}, width: 180, formatter: 'formatDateTime' },
{{- end}}
{{- end}}
    { field: 'createdAt', title: '创建时间', width: 180, formatter: 'formatDateTime', sortable: true },
    { title: '操作', width: 240, fixed: 'right', slots: { default: 'action' } },
  ],
{{- if .HasParentID}}
  pagerConfig: { enabled: false },
  treeConfig: {
    childrenField: 'children',
    expandAll: false,
  },
  proxyConfig: {
    ajax: {
      query: async (_params, formValues) => {
        const { timeRange, ...rest } = formValues;
        const params: Record<string, any> = { ...rest };
        if (timeRange && timeRange.length === 2) {
          params.startTime = timeRange[0];
          params.endTime = timeRange[1];
        }
        return await get{{.ModelName}}Tree(params as any) ?? [];
      },
    },
  },
{{- else}}
  height: 'auto',
  pagerConfig: {},
  proxyConfig: {
    ajax: {
      query: async ({ page, sorts }, formValues) => {
        const { timeRange, ...rest } = formValues;
        const params: Record<string, any> = {
          pageNum: page.currentPage,
          pageSize: page.pageSize,
          ...rest,
        };
        if (timeRange && timeRange.length === 2) {
          params.startTime = timeRange[0];
          params.endTime = timeRange[1];
        }
        if (sorts && sorts.length > 0) {
          const sort = sorts[0];
          if (sort && sort.field && sort.order) {
            params.orderBy = sort.field;
            params.orderDir = sort.order;
          }
        }
        const res = await get{{.ModelName}}List(params as any);
        return { items: res?.list ?? [], total: res?.total ?? 0 };
      },
    },
  },
{{- end}}
  sortConfig: {
    remote: true,
    trigger: 'cell',
  },
  toolbarConfig: {
    custom: true,
    refresh: true,
    search: true,
  },
};

const [Grid, gridApi] = useVbenVxeGrid({
  formOptions,
  gridOptions,
});

/** 新建 */
function handleCreate() {
  formModalApi.setData(null).open();
}

/** 查看 */
function handleView(row: {{.ModelName}}Item) {
  detailDrawerApi.setData({ id: row.id }).open();
}

/** 编辑 */
function handleEdit(row: {{.ModelName}}Item) {
  formModalApi.setData({ id: row.id }).open();
}

/** 删除 */
function handleDelete(row: {{.ModelName}}Item) {
  Modal.confirm({
    title: '确认删除',
    content: '确定要删除该{{.Comment}}吗？',
    okType: 'danger',
    async onOk() {
      await delete{{.ModelName}}(row.id);
      message.success('删除成功');
      gridApi.reload();
    },
  });
}

{{- if not .HasParentID}}
/** 批量删除 */
function handleBatchDelete() {
  const rows = gridApi.grid.getCheckboxRecords();
  if (rows.length === 0) {
    message.warning('请先选择要删除的数据');
    return;
  }
  Modal.confirm({
    title: '确认批量删除',
    content: `确定要删除选中的 ${rows.length} 条{{.Comment}}吗？`,
    okType: 'danger',
    async onOk() {
      await batchDelete{{.ModelName}}(rows.map((r: {{.ModelName}}Item) => r.id));
      message.success('批量删除成功');
      gridApi.reload();
    },
  });
}
{{- end}}

/** 导出 */
async function handleExport() {
  try {
    const formValues = await gridApi.formApi.getValues();
    const params: Record<string, any> = { ...formValues };
    if (params.timeRange && params.timeRange.length === 2) {
      params.startTime = params.timeRange[0];
      params.endTime = params.timeRange[1];
      delete params.timeRange;
    }
    const blob = await export{{.ModelName}}(params);
    const url = URL.createObjectURL(blob as any);
    const a = document.createElement('a');
    a.href = url;
    a.download = '{{.Comment}}.csv';
    a.click();
    URL.revokeObjectURL(url);
    message.success('导出成功');
  } catch {
    message.error('导出失败');
  }
}
{{- if .HasImport}}

/** 导入 */
async function handleImport() {
  const input = document.createElement('input');
  input.type = 'file';
  input.accept = '.csv,.xlsx,.xls';
  input.onchange = async () => {
    const file = input.files?.[0];
    if (!file) return;
    const formData = new FormData();
    formData.append('file', file);
    try {
      const res = await import{{.ModelName}}(formData);
      message.success(`导入完成：成功 ${res?.success ?? 0} 条，失败 ${res?.fail ?? 0} 条`);
      gridApi.reload();
    } catch {
      message.error('导入失败');
    }
  };
  input.click();
}

/** 下载导入模板 */
async function handleDownloadTemplate() {
  try {
    const blob = await downloadImportTemplate{{.ModelName}}();
    const url = URL.createObjectURL(blob as any);
    const a = document.createElement('a');
    a.href = url;
    a.download = '{{.Comment}}导入模板.csv';
    a.click();
    URL.revokeObjectURL(url);
  } catch {
    message.error('下载模板失败');
  }
}
{{- end}}
{{- if .HasBatchEdit}}

/** 批量修改状态 */
function handleBatchUpdateStatus() {
  const rows = gridApi.grid.getCheckboxRecords();
  if (rows.length === 0) {
    message.warning('请先选择要修改的数据');
    return;
  }
  Modal.confirm({
    title: '批量修改状态',
    content: `确定要将选中的 ${rows.length} 条数据的状态切换吗？`,
    async onOk() {
      const newStatus = rows[0]?.status === 1 ? 0 : 1;
      await batchUpdate{{.ModelName}}({ ids: rows.map((r: {{.ModelName}}Item) => r.id), status: newStatus });
      message.success('批量修改成功');
      gridApi.reload();
    },
  });
}
{{- end}}
</script>

<template>
  <Page auto-content-height>
    <FormModalComp @success="() => gridApi.reload()" />
    <DetailDrawerComp />
    <Grid>
      <template #toolbar-actions>
        <Button v-auth="['{{.AppName}}:{{.ModuleName}}:create']" type="primary" @click="handleCreate">新建</Button>
{{- if not .HasParentID}}
        <Button v-auth="['{{.AppName}}:{{.ModuleName}}:batch-delete']" danger class="ml-2" @click="handleBatchDelete">批量删除</Button>
{{- end}}
        <Button v-auth="['{{.AppName}}:{{.ModuleName}}:export']" class="ml-2" @click="handleExport">导出</Button>
{{- if .HasImport}}
        <Button v-auth="['{{.AppName}}:{{.ModuleName}}:import']" class="ml-2" @click="handleImport">导入</Button>
        <Button class="ml-2" @click="handleDownloadTemplate">模板下载</Button>
{{- end}}
{{- if .HasBatchEdit}}
        <Button v-auth="['{{.AppName}}:{{.ModuleName}}:batch-update']" class="ml-2" @click="handleBatchUpdateStatus">批量修改状态</Button>
{{- end}}
      </template>
{{- range .Fields}}
{{- if and (not .IsHidden) (not .IsID) (not .IsParentID) (not .IsTimeField) (not .IsMultiFK) (.IsEnum)}}
      <template #{{.NameLower}}_cell="{ row }">
        <Tag :color="get{{.NameCamel}}Color(row.{{.NameLower}})">
          {{"{{"}} {{.NameLower}}Map[row.{{.NameLower}}] || row.{{.NameLower}} {{"}}"}}
        </Tag>
      </template>
{{- else if and (not .IsHidden) (not .IsID) (not .IsParentID) (not .IsTimeField) (not .IsMultiFK) (eq .Component "ImageUpload")}}
      <template #{{.NameLower}}_cell="{ row }">
        <img v-if="row.{{.NameLower}}" :src="row.{{.NameLower}}" style="width: 48px; height: 48px; object-fit: cover; border-radius: 4px;" />
        <span v-else>-</span>
      </template>
{{- end}}
{{- end}}
      <template #action="{ row }">
        <Button v-auth="['{{.AppName}}:{{.ModuleName}}:detail']" type="link" size="small" @click="handleView(row)">查看</Button>
        <Button v-auth="['{{.AppName}}:{{.ModuleName}}:update']" type="link" size="small" @click="handleEdit(row)">编辑</Button>
        <Button v-auth="['{{.AppName}}:{{.ModuleName}}:delete']" type="link" danger size="small" @click="handleDelete(row)">删除</Button>
      </template>
    </Grid>
  </Page>
</template>

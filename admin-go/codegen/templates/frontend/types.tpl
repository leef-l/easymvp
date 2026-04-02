/** {{.Comment}}类型定义 */

/** {{.Comment}}项 */
export interface {{.ModelName}}Item {
  id: string;
{{- range .Fields}}
{{- if and (not .IsHidden) (not .IsID) (not .IsPassword)}}
  {{.NameLower}}{{if not .IsRequired}}?{{end}}: {{if or .IsForeignKey .IsParentID}}string{{else if .IsMultiFK}}string[]{{else}}{{.TSType}}{{end}};
{{- if .RefFieldJSON}}
  {{.RefFieldJSON}}?: string;
{{- end}}
{{- end}}
{{- end}}
  createdAt?: string;
  updatedAt?: string;
{{- if .HasParentID}}
  children?: {{.ModelName}}Item[];
{{- end}}
}

/** {{.Comment}}列表查询参数 */
export interface {{.ModelName}}ListParams {
  pageNum: number;
  pageSize: number;
  orderBy?: string;
  orderDir?: string;
  startTime?: string;
  endTime?: string;
{{- range .Fields}}
{{- if and (not .IsHidden) (not .IsID) (.IsEnum)}}
  {{.NameLower}}?: {{.TSType}};
{{- end}}
{{- end}}
{{- range .Fields}}
{{- if .IsSearchable}}
  {{.NameLower}}?: string;
{{- end}}
{{- end}}
}

{{- if .HasParentID}}
/** {{.Comment}}树形查询参数 */
export interface {{.ModelName}}TreeParams {
  startTime?: string;
  endTime?: string;
{{- range .Fields}}
{{- if and (not .IsHidden) (not .IsID) (not .IsParentID) (.IsEnum)}}
  {{.NameLower}}?: {{.TSType}};
{{- end}}
{{- end}}
{{- range .Fields}}
{{- if .IsSearchable}}
  {{.NameLower}}?: string;
{{- end}}
{{- end}}
}
{{- end}}

/** {{.Comment}}创建参数 */
export interface {{.ModelName}}CreateParams {
{{- range .Fields}}
{{- if and (not .IsHidden) (not .IsID)}}
  {{.NameLower}}{{if not .IsRequired}}?{{end}}: {{if or .IsForeignKey .IsParentID}}string{{else if .IsMultiFK}}string[]{{else}}{{.TSType}}{{end}};
{{- end}}
{{- end}}
}

/** {{.Comment}}更新参数 */
export interface {{.ModelName}}UpdateParams {
  id: string;
{{- range .Fields}}
{{- if and (not .IsHidden) (not .IsID)}}
  {{.NameLower}}{{if not .IsRequired}}?{{end}}: {{if or .IsForeignKey .IsParentID}}string{{else if .IsMultiFK}}string[]{{else}}{{.TSType}}{{end}};
{{- end}}
{{- end}}
}

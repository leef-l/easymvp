package consts
{{range .Fields}}{{if .IsEnum}}
// {{$.ModelName}}{{.NameCamel}} {{.Label}}
const (
{{- $fieldCamel := .NameCamel -}}
{{- range .EnumValues}}
	{{$.ModelName}}{{$fieldCamel}}{{if .NameIdent}}{{.NameIdent}}{{else}}V{{.Value}}{{end}} = {{.Value}} // {{.Label}}
{{- end}}
)
{{end}}{{end}}

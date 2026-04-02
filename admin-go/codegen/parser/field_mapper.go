package parser

import "strings"

// snakeToCamel 将 snake_case 转为 CamelCase
// 例：dept_name → DeptName, parent_id → ParentID, link_url → LinkURL
func snakeToCamel(s string) string {
	parts := strings.Split(s, "_")
	var b strings.Builder
	for _, p := range parts {
		if p == "" {
			continue
		}
		// 特殊缩写处理
		upper := strings.ToUpper(p)
		if upper == "ID" || upper == "URL" || upper == "IP" || upper == "HTTP" || upper == "JSON" || upper == "HTML" {
			b.WriteString(upper)
		} else {
			b.WriteString(strings.ToUpper(p[:1]))
			b.WriteString(p[1:])
		}
	}
	return b.String()
}

// SnakeToCamelSimple 将 snake_case 转为 CamelCase（不做特殊缩写处理，每段仅首字母大写）
// 例：parent_id → ParentId, link_url → LinkUrl, user_role → UserRole
// 供 cmd.tpl 的 ModuleCamel 模板函数和 DAO 风格字段名使用
func SnakeToCamelSimple(s string) string {
	return snakeToCamelDao(s)
}

// snakeToCamelDao 将 snake_case 转为 GoFrame DAO 风格 CamelCase
// 不做特殊缩写处理，每段仅首字母大写
// 例：parent_id → ParentId, link_url → LinkUrl, data_scope → DataScope
func snakeToCamelDao(s string) string {
	parts := strings.Split(s, "_")
	var b strings.Builder
	for _, p := range parts {
		if p == "" {
			continue
		}
		b.WriteString(strings.ToUpper(p[:1]))
		b.WriteString(p[1:])
	}
	return b.String()
}

// snakeToCamelLower 将 snake_case 转为 camelCase（首字母小写）
// 例：dept_name → deptName
func snakeToCamelLower(s string) string {
	camel := snakeToCamel(s)
	if camel == "" {
		return ""
	}
	// 处理连续大写（如 ID → id, URL → url）
	runes := []rune(camel)
	// 找到第一个小写字母之前的大写序列
	i := 0
	for i < len(runes) && runes[i] >= 'A' && runes[i] <= 'Z' {
		i++
	}
	if i == 0 {
		return camel
	}
	if i == 1 {
		// 只有第一个字母是大写
		runes[0] = runes[0] + 32
	} else if i == len(runes) {
		// 全是大写
		for j := range runes {
			runes[j] = runes[j] + 32
		}
	} else {
		// 连续大写后跟小写，如 IDName → idName, URLPath → urlPath
		for j := 0; j < i-1; j++ {
			runes[j] = runes[j] + 32
		}
	}
	return string(runes)
}

// MapGoType 根据数据库类型映射 Go 类型
func MapGoType(dbType string, isID bool) string {
	dbType = strings.ToLower(dbType)

	// ID 字段使用 snowflake.JsonInt64
	if isID {
		return "snowflake.JsonInt64"
	}

	switch {
	case dbType == "bigint unsigned", dbType == "bigint":
		return "int64"
	case dbType == "int", dbType == "int unsigned", dbType == "mediumint", dbType == "mediumint unsigned":
		return "int"
	case dbType == "smallint", dbType == "smallint unsigned":
		return "int"
	case dbType == "tinyint", strings.HasPrefix(dbType, "tinyint("):
		return "int"
	case dbType == "float":
		return "float32"
	case dbType == "double", dbType == "decimal":
		return "float64"
	case strings.HasPrefix(dbType, "varchar"), strings.HasPrefix(dbType, "char"):
		return "string"
	case dbType == "text", dbType == "longtext", dbType == "mediumtext", dbType == "tinytext":
		return "string"
	case dbType == "datetime", dbType == "timestamp", dbType == "date", dbType == "time":
		return "*gtime.Time"
	case dbType == "json":
		return "string"
	case dbType == "blob", dbType == "longblob", dbType == "mediumblob":
		return "[]byte"
	case strings.HasPrefix(dbType, "enum"):
		return "string"
	default:
		return "string"
	}
}

// MapTSType 根据数据库类型映射 TypeScript 类型
func MapTSType(dbType string, isID bool) string {
	dbType = strings.ToLower(dbType)

	// ID 字段前端使用 string（Snowflake 防精度丢失）
	if isID {
		return "string"
	}

	switch {
	case dbType == "bigint unsigned", dbType == "bigint":
		return "string" // bigint 前端也用 string
	case dbType == "int", dbType == "int unsigned",
		dbType == "mediumint", dbType == "mediumint unsigned",
		dbType == "smallint", dbType == "smallint unsigned",
		dbType == "tinyint", strings.HasPrefix(dbType, "tinyint("):
		return "number"
	case dbType == "float", dbType == "double", dbType == "decimal":
		return "number"
	case dbType == "datetime", dbType == "timestamp", dbType == "date", dbType == "time":
		return "string"
	default:
		return "string"
	}
}

// IsHiddenField 判断字段是否在表单中隐藏
// 隐藏字段：id/created_at/updated_at/deleted_at/created_by/dept_id
func IsHiddenField(name string) bool {
	hidden := map[string]bool{
		"id":         true,
		"created_at": true,
		"updated_at": true,
		"deleted_at": true,
		"created_by": true,
		"dept_id":    true,
	}
	return hidden[name]
}

// MapComponent 根据字段信息映射前端组件类型
func MapComponent(field FieldMeta) string {
	name := field.Name

	// parent_id → 树形单选
	if name == "parent_id" {
		return ComponentTreeSelectSingle
	}

	// parent_ids → 树形多选
	if name == "parent_ids" {
		return ComponentTreeSelectMulti
	}

	// *_ids → 多选下拉
	if strings.HasSuffix(name, "_ids") {
		return ComponentSelectMulti
	}

	// *_id → 单选下拉（排除 id 本身和 dept_id）
	if strings.HasSuffix(name, "_id") && name != "id" && name != "dept_id" {
		return ComponentSelect
	}

	// 图片上传
	imageExact := map[string]bool{
		"avatar": true, "cover": true, "logo": true,
		"banner": true, "thumbnail": true, "poster": true,
	}
	if imageExact[name] {
		return ComponentImageUpload
	}
	imageSuffixes := []string{"_image", "_img", "_photo", "_pic", "_cover", "_banner", "_logo", "_thumbnail", "_poster"}
	for _, suffix := range imageSuffixes {
		if strings.HasSuffix(name, suffix) {
			return ComponentImageUpload
		}
	}

	// 文件上传
	if strings.HasSuffix(name, "_file") || strings.HasSuffix(name, "_attachment") {
		return ComponentFileUpload
	}

	// 富文本
	if strings.HasSuffix(name, "_content") || strings.HasSuffix(name, "_body") ||
		strings.HasSuffix(name, "_html") {
		return ComponentRichText
	}

	// JSON 编辑器
	if strings.HasSuffix(name, "_json") || strings.HasSuffix(name, "_config") ||
		strings.HasSuffix(name, "_settings") {
		return ComponentJsonEditor
	}

	// 密码
	if strings.HasSuffix(name, "_password") || strings.HasSuffix(name, "_pwd") ||
		name == "password" {
		return ComponentPassword
	}

	// URL
	if strings.HasSuffix(name, "_url") || strings.HasSuffix(name, "_link") {
		return ComponentInputUrl
	}

	// 时间字段
	if strings.HasSuffix(name, "_at") {
		return ComponentDateTimePicker
	}

	// 数字输入（含金额字段）
	if name == "sort" || name == "order" || strings.HasSuffix(name, "_num") ||
		strings.HasSuffix(name, "_price") || strings.HasSuffix(name, "_amount") ||
		strings.HasSuffix(name, "_income") || strings.HasSuffix(name, "_balance") ||
		name == "price" || name == "amount" || name == "balance" || name == "quantity" ||
		name == "income_total" || name == "income_balance" {
		return ComponentInputNumber
	}

	// 图标
	if name == "icon" {
		return ComponentIconPicker
	}

	// status / is_* 字段，含枚举
	if name == "status" || strings.HasPrefix(name, "is_") {
		if len(field.EnumValues) == 2 {
			return ComponentSwitch
		}
		if len(field.EnumValues) > 0 {
			return ComponentRadio
		}
		return ComponentSwitch
	}

	// type / level / grade 字段，通常有枚举
	if name == "type" || name == "level" || name == "grade" {
		return ComponentSelect
	}

	// 含枚举备注的字段
	if len(field.EnumValues) > 0 {
		return ComponentSelect
	}

	// TEXT/LONGTEXT 类型 → Textarea
	dbLower := strings.ToLower(field.DBType)
	if dbLower == "text" || dbLower == "longtext" || dbLower == "mediumtext" || dbLower == "tinytext" {
		return ComponentTextarea
	}

	// 默认 Input
	return ComponentInput
}

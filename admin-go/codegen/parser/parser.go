package parser

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

// columnInfo 从 information_schema.COLUMNS 查询到的字段信息
type columnInfo struct {
	ColumnName    string
	DataType      string
	ColumnType    string
	IsNullable    string
	ColumnKey     string
	ColumnDefault sql.NullString
	ColumnComment string
	CharMaxLength sql.NullInt64
	Extra         string
}

// Parser 数据库表结构解析器
type Parser struct {
	DSN        string   // "user:pass@tcp(host:port)/dbname"
	SkipFields []string // 额外隐藏的字段列表（从 codegen.yaml skip_fields 加载）
	db         *sql.DB  // 复用数据库连接
	dbName     string   // 缓存的数据库名
}

// New 创建解析器实例
func New(dsn string, skipFields ...[]string) (*Parser, error) {
	p := &Parser{DSN: dsn}
	if len(skipFields) > 0 {
		p.SkipFields = skipFields[0]
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}
	if _, err := db.Exec("SET NAMES utf8mb4"); err != nil {
		db.Close()
		return nil, fmt.Errorf("设置字符集失败: %w", err)
	}
	dbName, err := extractDBName(dsn)
	if err != nil {
		db.Close()
		return nil, err
	}
	p.db = db
	p.dbName = dbName
	return p, nil
}

// Close 释放数据库连接
func (p *Parser) Close() {
	if p.db != nil {
		p.db.Close()
	}
}

// ParseTable 解析单张表
func (p *Parser) ParseTable(tableName string) (*TableMeta, error) {
	db := p.db
	dbName := p.dbName

	// 查询表注释
	tableComment, err := queryTableComment(db, dbName, tableName)
	if err != nil {
		return nil, err
	}

	// 查询字段信息
	columns, err := queryColumns(db, dbName, tableName)
	if err != nil {
		return nil, err
	}

	if len(columns) == 0 {
		return nil, fmt.Errorf("表 %s 不存在或没有字段", tableName)
	}

	// 从表名提取应用名和模块名：{app}_{module}
	appName := ""
	moduleName := tableName
	if idx := strings.Index(tableName, "_"); idx > 0 {
		appName = tableName[:idx]
		moduleName = tableName[idx+1:]
	}

	// 构建 TableMeta
	meta := &TableMeta{
		TableName:    tableName,
		AppName:      appName,
		AppNameCamel: snakeToCamel(appName),
		ModelName:    snakeToCamel(moduleName),
		DaoName:      snakeToCamel(tableName),
		ModuleName:   strings.ToLower(moduleName),
		PackageName:  strings.ReplaceAll(strings.ToLower(moduleName), "_", ""),
		Comment:      tableComment,
	}

	// 构建额外隐藏字段集合
	extraHidden := make(map[string]bool, len(p.SkipFields))
	for _, f := range p.SkipFields {
		extraHidden[f] = true
	}

	for _, col := range columns {
		field := buildFieldMeta(col)
		// 应用 skip_fields 配置中的额外隐藏字段
		if extraHidden[field.Name] {
			field.IsHidden = true
		}
		meta.Fields = append(meta.Fields, field)

		if field.Name == "parent_id" {
			meta.HasParentID = true
		}
		if field.Name == "status" {
			meta.HasStatus = true
		}
		if field.Name == "sort" {
			meta.HasSort = true
		}
		if field.IsPassword {
			meta.HasPassword = true
		}
		if field.TooltipText != "" {
			meta.HasTooltip = true
		}
		if field.Component == "RichText" || field.Component == "JsonEditor" {
			meta.HasRichText = true
		}
		if field.IsMoney {
			meta.HasMoney = true
		}
		if field.Name == "created_by" {
			meta.HasCreatedBy = true
		}
		if field.Name == "dept_id" {
			meta.HasDeptID = true
		}
		if field.DictType != "" {
			meta.HasDict = true
		}
		if field.IsSearchable {
			meta.HasSearchable = true
		}
		if field.Component == ComponentTreeSelectSingle || field.Component == ComponentTreeSelectMulti {
			meta.HasTreeSelect = true
		}
	}

	// HasBatchEdit：有 status 枚举字段就支持批量编辑
	meta.HasBatchEdit = meta.HasStatus
	// HasImport：非树形表默认支持导入
	meta.HasImport = !meta.HasParentID

	// 解析关联字段：对 *_id 外键和 parent_id 查找关联表的显示字段
	for i := range meta.Fields {
		f := &meta.Fields[i]
		if !f.IsForeignKey && !f.IsParentID {
			continue
		}
		// skip_fields 中的外键字段跳过关联表解析（跨应用引用等场景）
		if extraHidden[f.Name] {
			f.IsForeignKey = false
			continue
		}
		// 推断关联表名：parent_id → 自身模块名，xxx_id → xxx
		var refTable string
		if f.IsParentID {
			refTable = moduleName
		} else {
			refTable = strings.TrimSuffix(f.Name, "_id")
		}
		// 查找关联表的显示字段（先尝试带前缀的表名，再尝试不带前缀的）
		displayField := ""
		refTableDB := refTable
		if appName != "" {
			prefixed := appName + "_" + refTable
			displayField = findDisplayField(db, dbName, prefixed)
			if displayField != "" {
				refTableDB = prefixed
			}
		}
		if displayField == "" {
			displayField = findDisplayField(db, dbName, refTable)
		}
		// 关联表不存在或没有可用的显示字段 → 降级为普通字段（跨应用引用等场景）
		if displayField == "" {
			f.IsForeignKey = false
			continue
		}
		refFieldName := snakeToCamel(refTable) + snakeToCamel(displayField)
		// 检查 RefFieldName 是否与已有字段的 CamelCase 名冲突
		collision := false
		for _, other := range meta.Fields {
			if other.NameCamel == refFieldName || other.RefFieldName == refFieldName {
				collision = true
				break
			}
		}
		if collision {
			// 表本身已有同名字段，跳过关联字段生成
			continue
		}
		f.RefTable = refTable
		f.RefTableDB = refTableDB
		f.RefTableCamel = snakeToCamel(refTable)
		f.RefTableLower = snakeToCamelLower(refTable)
		f.RefDisplayField = displayField
		f.RefDisplayCamel = snakeToCamel(displayField)
		f.RefDisplayLower = snakeToCamelLower(displayField)
		f.RefFieldName = refFieldName
		f.RefFieldJSON = snakeToCamelLower(refTable) + snakeToCamel(displayField)
		// 检查关联表是否有 parent_id（树形结构）
		f.RefIsTree = tableHasColumn(db, dbName, refTableDB, "parent_id")
	}

	return meta, nil
}

// findDisplayField 在关联表中按优先级查找显示字段
func findDisplayField(db *sql.DB, dbName, tableName string) string {
	priorities := []string{
		"title", "name", "username", "nickname",
		"real_name", "label", "phone", "mobile",
	}
	for _, col := range priorities {
		if tableHasColumn(db, dbName, tableName, col) {
			return col
		}
	}
	return ""
}

// tableHasColumn 检查表是否存在指定列
func tableHasColumn(db *sql.DB, dbName, tableName, columnName string) bool {
	var count int
	err := db.QueryRow(
		"SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ? AND COLUMN_NAME = ?",
		dbName, tableName, columnName,
	).Scan(&count)
	return err == nil && count > 0
}

// ParseTables 解析多张表
func (p *Parser) ParseTables(tableNames []string) ([]*TableMeta, error) {
	var result []*TableMeta
	for _, name := range tableNames {
		meta, err := p.ParseTable(name)
		if err != nil {
			return nil, fmt.Errorf("解析表 %s 失败: %w", name, err)
		}
		result = append(result, meta)
	}
	return result, nil
}

// extractDBName 从 DSN 中提取数据库名
func extractDBName(dsn string) (string, error) {
	// DSN 格式: user:pass@tcp(host:port)/dbname?params
	slashIdx := strings.LastIndex(dsn, "/")
	if slashIdx < 0 {
		return "", fmt.Errorf("DSN 格式错误，无法提取数据库名: %s", dsn)
	}
	rest := dsn[slashIdx+1:]
	qIdx := strings.Index(rest, "?")
	if qIdx >= 0 {
		rest = rest[:qIdx]
	}
	if rest == "" {
		return "", fmt.Errorf("DSN 中未指定数据库名: %s", dsn)
	}
	return rest, nil
}

// queryTableComment 查询表注释
func queryTableComment(db *sql.DB, dbName, tableName string) (string, error) {
	var comment sql.NullString
	err := db.QueryRow(
		"SELECT TABLE_COMMENT FROM information_schema.TABLES WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?",
		dbName, tableName,
	).Scan(&comment)
	if err != nil {
		return "", fmt.Errorf("查询表 %s 注释失败: %w", tableName, err)
	}
	return comment.String, nil
}

// queryColumns 查询表的所有字段信息
func queryColumns(db *sql.DB, dbName, tableName string) ([]columnInfo, error) {
	rows, err := db.Query(
		`SELECT COLUMN_NAME, DATA_TYPE, COLUMN_TYPE, IS_NULLABLE, COLUMN_KEY,
		        COLUMN_DEFAULT, COLUMN_COMMENT, CHARACTER_MAXIMUM_LENGTH, EXTRA
		 FROM information_schema.COLUMNS
		 WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
		 ORDER BY ORDINAL_POSITION`,
		dbName, tableName,
	)
	if err != nil {
		return nil, fmt.Errorf("查询表 %s 字段失败: %w", tableName, err)
	}
	defer rows.Close()

	var columns []columnInfo
	for rows.Next() {
		var col columnInfo
		if err := rows.Scan(
			&col.ColumnName, &col.DataType, &col.ColumnType, &col.IsNullable,
			&col.ColumnKey, &col.ColumnDefault, &col.ColumnComment,
			&col.CharMaxLength, &col.Extra,
		); err != nil {
			return nil, fmt.Errorf("扫描字段信息失败: %w", err)
		}
		columns = append(columns, col)
	}
	return columns, rows.Err()
}

// buildFieldMeta 根据列信息构建 FieldMeta
func buildFieldMeta(col columnInfo) FieldMeta {
	name := col.ColumnName
	isID := name == "id"
	// 外键判断：_id 后缀 + 排除特殊字段 + 必须是整数类型（varchar/char 类型的 _id 字段视为业务关联ID，非真正外键）
	isIntType := col.DataType == "bigint" || col.DataType == "int" || col.DataType == "smallint" || col.DataType == "tinyint" || col.DataType == "mediumint"
	isForeignKey := strings.HasSuffix(name, "_id") && name != "id" && name != "dept_id" && isIntType
	isMultiFK := strings.HasSuffix(name, "_ids")
	isParentID := name == "parent_id"
	isPassword := name == "password" || strings.HasSuffix(name, "_password") || strings.HasSuffix(name, "_pwd")

	// 解析备注
	label, shortLabel, tooltipText, enums := ParseComment(col.ColumnComment)

	// 构建基础数据库类型（简化，去掉长度信息用于映射）
	dbType := col.ColumnType

	field := FieldMeta{
		Name:         name,
		NameCamel:    snakeToCamel(name),
		NameDao:      snakeToCamelDao(name),
		NameLower:    snakeToCamelLower(name),
		DBType:       dbType,
		GoType:       MapGoType(col.DataType, isID || isForeignKey || isParentID || name == "dept_id" || name == "created_by"),
		TSType:       MapTSType(col.DataType, isID || isForeignKey || isParentID || name == "dept_id" || name == "created_by"),
		Comment:      col.ColumnComment,
		Label:        label,
		ShortLabel:   shortLabel,
		TooltipText:  tooltipText,
		EnumValues:   enums,
		IsRequired:   col.IsNullable == "NO" && col.ColumnDefault.Valid == false && name != "id",
		IsID:         isID,
		IsParentID:   isParentID,
		IsForeignKey: isForeignKey,
		IsMultiFK:    isMultiFK,
		IsTimeField:  strings.HasSuffix(name, "_at"),
		IsHidden:     IsHiddenField(name),
		IsEnum:       len(enums) > 0,
		IsPassword:   isPassword,
		DefaultValue: col.ColumnDefault.String,
	}

	// 判断是否可搜索的文本字段（用于关键词模糊查询）
	searchableNames := map[string]bool{
		"title": true, "name": true, "username": true, "nickname": true,
		"phone": true, "mobile": true, "email": true, "real_name": true,
		"order_no": true, "remark": true,
	}
	goType := field.GoType
	if searchableNames[name] || strings.HasSuffix(name, "_name") || strings.HasSuffix(name, "_title") || strings.HasSuffix(name, "_no") {
		if goType == "string" && !isID && !isForeignKey && !isPassword {
			field.IsSearchable = true
		}
	}

	// 判断是否精确搜索字段（编号类，用 = 而非 LIKE）
	exactSuffixes := []string{"_no", "_code", "_sn"}
	for _, suffix := range exactSuffixes {
		if strings.HasSuffix(name, suffix) {
			field.IsExactSearch = true
			break
		}
	}

	// 判断是否金额字段（单位：分，列表需要分→元格式化）
	moneyNames := map[string]bool{
		"price": true, "amount": true, "balance": true, "quantity": false,
		"income_total": true, "income_balance": true,
	}
	if moneyNames[name] ||
		strings.HasSuffix(name, "_price") || strings.HasSuffix(name, "_amount") ||
		strings.HasSuffix(name, "_balance") || strings.HasSuffix(name, "_income") ||
		strings.HasSuffix(name, "_fee") || strings.HasSuffix(name, "_cost") {
		field.IsMoney = true
	}

	if col.CharMaxLength.Valid {
		field.MaxLength = int(col.CharMaxLength.Int64)
	}

	// 字典类型检测：如果 EnumValues 有 __dict__ 标记，提取字典类型
	if len(field.EnumValues) == 1 && field.EnumValues[0].Value == "__dict__" {
		field.DictType = field.EnumValues[0].Label
		field.EnumValues = nil // 清除标记
		field.IsEnum = false   // 字典字段不算硬编码枚举
	}

	// 自动推导验证规则
	field.ValidationRules, field.FrontendRules = buildValidationRules(field)

	// 映射前端组件
	field.Component = MapComponent(field)

	return field
}

// buildValidationRules 根据字段名和类型自动推导验证规则
func buildValidationRules(f FieldMeta) (goRules []string, frontendRule string) {
	if f.IsID || f.IsHidden {
		return nil, ""
	}
	// 必填
	if f.IsRequired {
		goRules = append(goRules, "required")
	}
	// 邮箱
	if f.Name == "email" || strings.HasSuffix(f.Name, "_email") {
		goRules = append(goRules, "email")
		frontendRule = "email"
	}
	// 手机号
	if f.Name == "phone" || f.Name == "mobile" || strings.HasSuffix(f.Name, "_phone") || strings.HasSuffix(f.Name, "_mobile") {
		goRules = append(goRules, "phone-loose")
		frontendRule = "phone"
	}
	// URL
	if f.Name == "url" || strings.HasSuffix(f.Name, "_url") || strings.HasSuffix(f.Name, "_link") {
		goRules = append(goRules, "url")
		frontendRule = "url"
	}
	// 长度限制（仅 string 类型且有 MaxLength）
	if f.GoType == "string" && f.MaxLength > 0 && !f.IsPassword {
		goRules = append(goRules, fmt.Sprintf("max-length:%d", f.MaxLength))
	}
	// 密码
	if f.IsPassword {
		goRules = append(goRules, "length:6,32")
	}
	return
}

package parser

import "strings"

// extractParentheses 从 label 中提取中文括号（）或英文括号()内的内容
// "排序（升序）" → shortLabel="排序", tooltip="升序"
// "部门名称"     → shortLabel="部门名称", tooltip=""
func extractParentheses(label string) (shortLabel string, tooltip string) {
	// 优先匹配中文括号
	if idx := strings.Index(label, "（"); idx >= 0 {
		if end := strings.Index(label, "）"); end > idx {
			shortLabel = strings.TrimSpace(label[:idx])
			tooltip = strings.TrimSpace(label[idx+len("（") : end])
			return
		}
	}
	// 再匹配英文括号
	if idx := strings.Index(label, "("); idx >= 0 {
		if end := strings.Index(label, ")"); end > idx {
			shortLabel = strings.TrimSpace(label[:idx])
			tooltip = strings.TrimSpace(label[idx+1 : end])
			return
		}
	}
	return label, ""
}

// ParseComment 解析字段备注
// 输入：状态:0=关闭,1=开启
// 输出：label="状态", shortLabel="状态", tooltipText="", enums=[{0,关闭},{1,开启}]
// 输入：排序（升序）
// 输出：label="排序（升序）", shortLabel="排序", tooltipText="升序", enums=[]
// 输入：部门名称
// 输出：label="部门名称", shortLabel="部门名称", tooltipText="", enums=[]
func ParseComment(comment string) (label string, shortLabel string, tooltipText string, enums []EnumValue) {
	comment = strings.TrimSpace(comment)
	if comment == "" {
		return "", "", "", nil
	}

	// 查找冒号分隔符（支持中文冒号和英文冒号）
	sepIdx := -1
	for i, ch := range comment {
		if ch == ':' || ch == '：' {
			sepIdx = i
			break
		}
	}

	// 没有冒号，整个备注就是 label
	if sepIdx < 0 {
		label = comment
		shortLabel, tooltipText = extractParentheses(label)
		return label, shortLabel, tooltipText, nil
	}

	label = strings.TrimSpace(comment[:sepIdx])
	enumPart := strings.TrimSpace(comment[sepIdx+1:])
	if len(label) == 0 {
		label = comment
		shortLabel, tooltipText = extractParentheses(label)
		return label, shortLabel, tooltipText, nil
	}

	shortLabel, tooltipText = extractParentheses(label)

	// 解析枚举部分：0=关闭,1=开启
	if enumPart == "" {
		return label, shortLabel, tooltipText, nil
	}

	// 如果是字典引用（如 dict:gender），不解析枚举
	if strings.HasPrefix(enumPart, "dict:") {
		dictType := strings.TrimPrefix(enumPart, "dict:")
		enums = []EnumValue{{Value: "__dict__", Label: dictType, NameIdent: ""}}
		return label, shortLabel, tooltipText, enums
	}

	pairs := strings.Split(enumPart, ",")
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		eqIdx := strings.Index(pair, "=")
		if eqIdx < 0 {
			continue
		}
		val := strings.TrimSpace(pair[:eqIdx])
		lab := strings.TrimSpace(pair[eqIdx+1:])
		if val != "" && lab != "" {
			enums = append(enums, EnumValue{Value: val, Label: lab, NameIdent: labelToIdent(lab)})
		}
	}

	return label, shortLabel, tooltipText, enums
}

// labelToIdent 将中文枚举标签转为语义化 Go 标识符
func labelToIdent(label string) string {
	m := map[string]string{
		"启用": "Enabled", "禁用": "Disabled",
		"正常": "Normal", "异常": "Abnormal",
		"有效": "Valid", "无效": "Invalid",
		"是": "Yes", "否": "No",
		"男": "Male", "女": "Female",
		"开启": "On", "关闭": "Off",
		"显示": "Show", "隐藏": "Hide",
		"已完成": "Done", "进行中": "InProgress",
		"待处理": "Pending", "已取消": "Cancelled",
		"待审核": "PendingReview", "已通过": "Approved", "已拒绝": "Rejected",
		"待支付": "Unpaid", "已支付": "Paid", "已退款": "Refunded",
		"草稿": "Draft", "已发布": "Published", "已下架": "Offline",
		"目录": "Dir", "菜单": "Menu", "按钮": "Button",
		"普通": "Regular", "VIP": "VIP", "管理员": "Admin",
		"成功": "Success", "失败": "Failed",
		"充值": "Recharge", "消费": "Consume", "提现": "Withdraw",
	}
	if ident, ok := m[label]; ok {
		return ident
	}
	return ""
}

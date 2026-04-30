package service

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
)

// ---------------------------------------------------------------------------
// Severity & EventType enums
// ---------------------------------------------------------------------------

type SecurityAuditSeverity string

const (
	SeverityInfo     SecurityAuditSeverity = "info"
	SeverityWarning  SecurityAuditSeverity = "warning"
	SeverityCritical SecurityAuditSeverity = "critical"
	SeverityAlert    SecurityAuditSeverity = "alert"
)

type SecurityAuditEventType string

const (
	AuditEventAuthentication     SecurityAuditEventType = "authentication"
	AuditEventAuthorization      SecurityAuditEventType = "authorization"
	AuditEventDataAccess         SecurityAuditEventType = "data_access"
	AuditEventConfigChange       SecurityAuditEventType = "config_change"
	AuditEventBrainRunStart      SecurityAuditEventType = "brain_run_start"
	AuditEventBrainRunStop       SecurityAuditEventType = "brain_run_stop"
	AuditEventProjectStateChange SecurityAuditEventType = "project_state_change"
	AuditEventSystemAdmin        SecurityAuditEventType = "system_admin"
)

// ---------------------------------------------------------------------------
// Domain structs
// ---------------------------------------------------------------------------

// SecurityAuditEvent 单条安全审计事件。
type SecurityAuditEvent struct {
	ID          string                 `json:"id"`
	ProjectID   string                 `json:"project_id"`
	EventType   SecurityAuditEventType `json:"event_type"`
	Severity    SecurityAuditSeverity  `json:"severity"`
	Operator    string                 `json:"operator"`
	Resource    string                 `json:"resource"`
	Description string                 `json:"description"`
	Timestamp   time.Time              `json:"timestamp"`
	Metadata    map[string]string      `json:"metadata,omitempty"`
}

// SecurityAuditFilter 查询过滤条件。
type SecurityAuditFilter struct {
	ProjectID string                 `json:"project_id,omitempty"`
	EventType SecurityAuditEventType `json:"event_type,omitempty"`
	Severity  SecurityAuditSeverity  `json:"severity,omitempty"`
	Operator  string                 `json:"operator,omitempty"`
	From      time.Time              `json:"from"`
	To        time.Time              `json:"to"`
	Limit     int                    `json:"limit,omitempty"`
}

// SecurityAuditReport 时段内的安全审计汇总报告。
type SecurityAuditReport struct {
	ProjectID       string                         `json:"project_id"`
	From            time.Time                      `json:"from"`
	To              time.Time                      `json:"to"`
	TotalEvents     int                            `json:"total_events"`
	BySeverity      map[SecurityAuditSeverity]int  `json:"by_severity"`
	ByType          map[SecurityAuditEventType]int `json:"by_type"`
	ComplianceItems []ComplianceItem               `json:"compliance_items"`
	GeneratedAt     time.Time                      `json:"generated_at"`
}

// ComplianceItem 单条合规检查项。
type ComplianceItem struct {
	Rule   string `json:"rule"`
	Passed bool   `json:"passed"`
	Detail string `json:"detail,omitempty"`
}

// ComplianceCheckResult 合规性检查结果。
type ComplianceCheckResult struct {
	ProjectID string           `json:"project_id"`
	CheckedAt time.Time        `json:"checked_at"`
	AllPassed bool             `json:"all_passed"`
	Items     []ComplianceItem `json:"items"`
}

// SecurityAnomaly 检测到的安全异常。
type SecurityAnomaly struct {
	Description string                `json:"description"`
	Severity    SecurityAuditSeverity `json:"severity"`
	EventCount  int                   `json:"event_count"`
	Window      time.Duration         `json:"window"`
	DetectedAt  time.Time             `json:"detected_at"`
}

// ---------------------------------------------------------------------------
// Interface
// ---------------------------------------------------------------------------

type ISecurityAudit interface {
	// RecordEvent 记录一条安全审计事件。
	RecordEvent(ctx context.Context, event SecurityAuditEvent) error
	// QueryEvents 按条件查询审计事件。
	QueryEvents(ctx context.Context, filter SecurityAuditFilter) ([]SecurityAuditEvent, error)
	// GenerateAuditReport 生成指定项目在 [from, to] 时段内的审计报告。
	GenerateAuditReport(ctx context.Context, projectID string, from, to time.Time) (*SecurityAuditReport, error)
	// CheckCompliance 对指定项目执行合规性检查。
	CheckCompliance(ctx context.Context, projectID string) (*ComplianceCheckResult, error)
	// DetectAnomalies 在指定时间窗口内检测异常行为。
	DetectAnomalies(ctx context.Context, projectID string, window time.Duration) ([]SecurityAnomaly, error)
}

var (
	localSecurityAudit     ISecurityAudit
	localSecurityAuditOnce sync.Once
)

type sSecurityAudit struct {
	mu        sync.RWMutex
	events    []SecurityAuditEvent
	maxEvents int
}

func SecurityAudit() ISecurityAudit {
	localSecurityAuditOnce.Do(func() {
		// 默认 10000；若配置存在则覆盖。maxEvents <= 0 表示禁用上限。
		maxEvents := 10000
		ctx := gctx.New()
		if v, err := g.Cfg().Get(ctx, "easymvp.securityAudit.maxEvents", 10000); err == nil && v != nil {
			maxEvents = v.Int()
		}
		localSecurityAudit = &sSecurityAudit{
			maxEvents: maxEvents,
		}
	})
	return localSecurityAudit
}

// ---------------------------------------------------------------------------
// RecordEvent
// ---------------------------------------------------------------------------

func (s *sSecurityAudit) RecordEvent(_ context.Context, event SecurityAuditEvent) error {
	if event.EventType == "" {
		return gerror.New("security audit: event_type is required")
	}
	if event.ProjectID == "" {
		return gerror.New("security audit: project_id is required")
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	if event.ID == "" {
		event.ID = newResourceID("sa")
	}

	s.mu.Lock()
	s.events = append(s.events, event)
	// ring buffer 行为：超过上限时 drop 最早事件。maxEvents <= 0 时禁用上限。
	if s.maxEvents > 0 && len(s.events) > s.maxEvents {
		s.events = s.events[len(s.events)-s.maxEvents:]
	}
	s.mu.Unlock()
	return nil
}

// ---------------------------------------------------------------------------
// QueryEvents
// ---------------------------------------------------------------------------

func (s *sSecurityAudit) QueryEvents(_ context.Context, filter SecurityAuditFilter) ([]SecurityAuditEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []SecurityAuditEvent
	limit := filter.Limit
	if limit <= 0 {
		limit = 200
	}

	// 用 binary search 找第一个 timestamp >= filter.From 的下标，作为倒序遍历的下界。
	// 前提假设：events 大体按 RecordEvent 调用时间追加，timestamp 单调递增。
	// 若调用方传入自定义非单调 timestamp，binary search 行为可能略有偏差（已知折衷）。
	startIdx := 0
	if !filter.From.IsZero() {
		startIdx = sort.Search(len(s.events), func(i int) bool {
			return !s.events[i].Timestamp.Before(filter.From)
		})
	}

	for i := len(s.events) - 1; i >= startIdx && len(result) < limit; i-- {
		ev := s.events[i]
		if matchesFilter(ev, filter) {
			result = append(result, ev)
		}
	}
	return result, nil
}

// ---------------------------------------------------------------------------
// GenerateAuditReport
// ---------------------------------------------------------------------------

func (s *sSecurityAudit) GenerateAuditReport(_ context.Context, projectID string, from, to time.Time) (*SecurityAuditReport, error) {
	if projectID == "" {
		return nil, gerror.New("security audit: project_id is required")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	report := &SecurityAuditReport{
		ProjectID:   projectID,
		From:        from,
		To:          to,
		BySeverity:  make(map[SecurityAuditSeverity]int),
		ByType:      make(map[SecurityAuditEventType]int),
		GeneratedAt: time.Now(),
	}

	for _, ev := range s.events {
		if ev.Timestamp.Before(from) || ev.Timestamp.After(to) {
			continue
		}
		if ev.ProjectID != projectID {
			continue
		}
		report.TotalEvents++
		report.BySeverity[ev.Severity]++
		report.ByType[ev.EventType]++
	}

	report.ComplianceItems = evaluateDefaultComplianceRules(s.events, projectID)
	return report, nil
}

// ---------------------------------------------------------------------------
// CheckCompliance
// ---------------------------------------------------------------------------

func (s *sSecurityAudit) CheckCompliance(_ context.Context, projectID string) (*ComplianceCheckResult, error) {
	if projectID == "" {
		return nil, gerror.New("security audit: project_id is required")
	}

	s.mu.RLock()
	items := evaluateDefaultComplianceRules(s.events, projectID)
	s.mu.RUnlock()

	allPassed := true
	for _, it := range items {
		if !it.Passed {
			allPassed = false
			break
		}
	}

	return &ComplianceCheckResult{
		ProjectID: projectID,
		CheckedAt: time.Now(),
		AllPassed: allPassed,
		Items:     items,
	}, nil
}

// ---------------------------------------------------------------------------
// DetectAnomalies
// ---------------------------------------------------------------------------

func (s *sSecurityAudit) DetectAnomalies(_ context.Context, projectID string, window time.Duration) ([]SecurityAnomaly, error) {
	if projectID == "" {
		return nil, gerror.New("security audit: project_id is required")
	}
	if window <= 0 {
		window = time.Hour
	}

	cutoff := time.Now().Add(-window)

	s.mu.RLock()
	defer s.mu.RUnlock()

	typeCounts := make(map[SecurityAuditEventType]int)
	criticalCount := 0
	for _, ev := range s.events {
		if ev.Timestamp.Before(cutoff) {
			continue
		}
		if ev.ProjectID != projectID {
			continue
		}
		typeCounts[ev.EventType]++
		if ev.Severity == SeverityCritical || ev.Severity == SeverityAlert {
			criticalCount++
		}
	}

	var anomalies []SecurityAnomaly

	// 规则1: 窗口内出现 >= 3 个 critical/alert 级别事件
	if criticalCount >= 3 {
		anomalies = append(anomalies, SecurityAnomaly{
			Description: "高严重级别事件频繁触发",
			Severity:    SeverityAlert,
			EventCount:  criticalCount,
			Window:      window,
			DetectedAt:  time.Now(),
		})
	}

	// 规则2: 单一事件类型在窗口内超过 20 次
	for et, cnt := range typeCounts {
		if cnt >= 20 {
			anomalies = append(anomalies, SecurityAnomaly{
				Description: "事件类型 " + string(et) + " 触发频率异常",
				Severity:    SeverityWarning,
				EventCount:  cnt,
				Window:      window,
				DetectedAt:  time.Now(),
			})
		}
	}

	sort.Slice(anomalies, func(i, j int) bool {
		return anomalies[i].Description < anomalies[j].Description
	})

	return anomalies, nil
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

func matchesFilter(ev SecurityAuditEvent, f SecurityAuditFilter) bool {
	if f.EventType != "" && ev.EventType != f.EventType {
		return false
	}
	if f.Severity != "" && ev.Severity != f.Severity {
		return false
	}
	if f.Operator != "" && ev.Operator != f.Operator {
		return false
	}
	if f.ProjectID != "" && ev.ProjectID != f.ProjectID {
		return false
	}
	if !f.From.IsZero() && ev.Timestamp.Before(f.From) {
		return false
	}
	if !f.To.IsZero() && ev.Timestamp.After(f.To) {
		return false
	}
	return true
}

// recordSecurityAuditNoFail 记录安全审计事件，失败时只打日志不阻塞业务流。
// 用于在 service 主链路上轻量埋点。
func recordSecurityAuditNoFail(ctx context.Context, event SecurityAuditEvent) {
	if err := SecurityAudit().RecordEvent(ctx, event); err != nil {
		g.Log().Errorf(ctx, "security audit record failed: %v", err)
	}
}

// evaluateDefaultComplianceRules 评估默认合规检查规则（调用者需持有读锁）。
func evaluateDefaultComplianceRules(events []SecurityAuditEvent, projectID string) []ComplianceItem {
	var projectEvents []SecurityAuditEvent
	for _, ev := range events {
		if ev.ProjectID == projectID {
			projectEvents = append(projectEvents, ev)
		}
	}

	return []ComplianceItem{
		checkOperationTraceability(projectEvents),
		checkPermissionChangeRecord(projectEvents),
		checkSensitiveDataAccessRecord(projectEvents),
	}
}

// 规则: 操作留痕 — 项目至少存在审计事件
func checkOperationTraceability(events []SecurityAuditEvent) ComplianceItem {
	if len(events) > 0 {
		return ComplianceItem{Rule: "操作留痕", Passed: true, Detail: "存在审计记录"}
	}
	return ComplianceItem{Rule: "操作留痕", Passed: false, Detail: "无任何审计记录"}
}

// 规则: 权限变更记录 — 若存在授权事件则应有对应审计留痕
func checkPermissionChangeRecord(events []SecurityAuditEvent) ComplianceItem {
	for _, ev := range events {
		if ev.EventType == AuditEventAuthorization {
			return ComplianceItem{Rule: "权限变更记录", Passed: true, Detail: "已记录权限变更"}
		}
	}
	return ComplianceItem{Rule: "权限变更记录", Passed: true, Detail: "无权限变更操作（合规）"}
}

// 规则: 敏感数据访问记录
func checkSensitiveDataAccessRecord(events []SecurityAuditEvent) ComplianceItem {
	for _, ev := range events {
		if ev.EventType == AuditEventDataAccess {
			return ComplianceItem{Rule: "敏感数据访问记录", Passed: true, Detail: "已记录数据访问事件"}
		}
	}
	return ComplianceItem{Rule: "敏感数据访问记录", Passed: true, Detail: "无敏感数据访问（合规）"}
}

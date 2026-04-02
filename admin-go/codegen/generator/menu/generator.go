package menu

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"easymvp/codegen/parser"
)

// MenuAppConfig 应用目录的标题、图标和排序配置
type MenuAppConfig struct {
	Title string
	Icon  string
	Sort  int
}

// MenuModuleConfig 模块级菜单配置
type MenuModuleConfig struct {
	Sort   int
	IsShow *int
}

// Config 菜单生成器配置
type Config struct {
	DSN         string
	Force       bool
	DryRun      bool
	MenuApps    map[string]MenuAppConfig    // 从 codegen.yaml 加载的应用目录配置
	MenuModules map[string]MenuModuleConfig // 从 codegen.yaml 加载的模块级配置，key: "appName/moduleName"
}

// Generator 菜单生成器
type Generator struct {
	config Config
}

// New 创建菜单生成器
func New(cfg Config) *Generator {
	return &Generator{config: cfg}
}

// Generate 为一张表生成菜单数据
func (g *Generator) Generate(meta *parser.TableMeta) (int, error) {
	db, err := sql.Open("mysql", g.config.DSN)
	if err != nil {
		return 0, fmt.Errorf("连接数据库失败: %w", err)
	}
	defer db.Close()

	if _, err := db.Exec("SET NAMES utf8mb4"); err != nil {
		return 0, fmt.Errorf("设置字符集失败: %w", err)
	}

	count := 0

	// 1. 查找或创建应用目录
	dirPath := "/" + meta.AppName
	dirID, err := g.ensureDirectory(db, meta.AppName, dirPath)
	if err != nil {
		return count, err
	}

	// 2. 读取模块级配置
	menuSort := 0
	menuIsShow := 1
	moduleKey := meta.AppName + "/" + meta.ModuleName
	if modCfg, ok := g.config.MenuModules[moduleKey]; ok {
		if modCfg.Sort > 0 {
			menuSort = modCfg.Sort
		}
		if modCfg.IsShow != nil {
			menuIsShow = *modCfg.IsShow
		}
	}

	// 3. 插入菜单页
	menuTitle := cleanTitle(meta.Comment)
	menuPath := "/" + meta.AppName + "/" + dashCase(meta.ModuleName)
	menuComponent := meta.AppName + "/" + meta.ModuleName + "/index"
	menuPermission := meta.AppName + ":" + meta.ModuleName + ":list"

	menuID, created, err := g.ensureMenu(db, dirID, menuTitle, menuPath, menuComponent, menuPermission, menuSort, menuIsShow)
	if err != nil {
		return count, err
	}
	if created {
		count++
	}

	// 4. 插入按钮权限
	buttons := []struct {
		suffix     string
		permission string
		sort       int
	}{
		{"新增", meta.AppName + ":" + meta.ModuleName + ":create", 1},
		{"修改", meta.AppName + ":" + meta.ModuleName + ":update", 2},
		{"删除", meta.AppName + ":" + meta.ModuleName + ":delete", 3},
		{"批量删除", meta.AppName + ":" + meta.ModuleName + ":batch-delete", 4},
		{"查看", meta.AppName + ":" + meta.ModuleName + ":detail", 5},
		{"导出", meta.AppName + ":" + meta.ModuleName + ":export", 6},
		{"导入", meta.AppName + ":" + meta.ModuleName + ":import", 7},
		{"批量编辑", meta.AppName + ":" + meta.ModuleName + ":batch-update", 8},
	}

	for _, btn := range buttons {
		btnTitle := menuTitle + btn.suffix
		created, err := g.ensureButton(db, menuID, btnTitle, btn.permission, btn.sort)
		if err != nil {
			return count, err
		}
		if created {
			count++
		}
	}

	return count, nil
}

// ensureDirectory 查找或创建应用目录（type=1）
func (g *Generator) ensureDirectory(db *sql.DB, appName, path string) (int64, error) {
	// 查找已存在的目录
	var id int64
	err := db.QueryRow(
		"SELECT id FROM system_menu WHERE path = ? AND type = 1 AND deleted_at IS NULL",
		path,
	).Scan(&id)

	if err == nil {
		fmt.Printf("  [菜单] 目录已存在: %s (ID: %d)\n", path, id)
		return id, nil
	}

	// 创建目录
	id = generateID()
	title := appName + "管理"
	icon := "AppstoreOutlined"
	sortVal := 50
	if cfg, ok := g.config.MenuApps[appName]; ok {
		if cfg.Title != "" {
			title = cfg.Title
		}
		if cfg.Icon != "" {
			icon = cfg.Icon
		}
		if cfg.Sort > 0 {
			sortVal = cfg.Sort
		}
	}

	if g.config.DryRun {
		fmt.Printf("  [dry-run] INSERT 目录: title=%s, path=%s, icon=%s, sort=%d\n", title, path, icon, sortVal)
		return id, nil
	}

	_, err = db.Exec(
		`INSERT INTO system_menu (id, parent_id, title, type, path, component, permission, icon, sort, is_show, is_cache, status, created_by, dept_id, created_at, updated_at)
		 VALUES (?, 0, ?, 1, ?, NULL, '', ?, ?, 1, 0, 1, 0, 0, NOW(), NOW())`,
		id, title, path, icon, sortVal,
	)
	if err != nil {
		return 0, fmt.Errorf("创建目录失败: %w", err)
	}
	fmt.Printf("  [菜单] 创建目录: %s (ID: %d, sort: %d)\n", title, id, sortVal)
	return id, nil
}

// ensureMenu 查找或创建菜单页（type=2）
func (g *Generator) ensureMenu(db *sql.DB, parentID int64, title, path, component, permission string, sort int, isShow int) (int64, bool, error) {
	var id int64
	err := db.QueryRow(
		"SELECT id FROM system_menu WHERE path = ? AND type = 2 AND deleted_at IS NULL",
		path,
	).Scan(&id)

	if err == nil {
		if g.config.Force {
			_, err = db.Exec(
				`UPDATE system_menu SET title=?, component=?, permission=?, sort=?, is_show=?, updated_at=NOW() WHERE id=?`,
				title, component, permission, sort, isShow, id,
			)
			if err != nil {
				return 0, false, fmt.Errorf("更新菜单失败: %w", err)
			}
			fmt.Printf("  [菜单] 更新菜单: %s (%s)\n", title, path)
			return id, false, nil
		}
		fmt.Printf("  [菜单] 跳过（已存在）: %s (%s)\n", title, path)
		return id, false, nil
	}

	id = generateID()

	if g.config.DryRun {
		fmt.Printf("  [dry-run] INSERT 菜单: title=%s, path=%s, permission=%s, sort=%d, is_show=%d\n", title, path, permission, sort, isShow)
		return id, true, nil
	}

	_, err = db.Exec(
		`INSERT INTO system_menu (id, parent_id, title, type, path, component, permission, icon, sort, is_show, is_cache, status, created_by, dept_id, created_at, updated_at)
		 VALUES (?, ?, ?, 2, ?, ?, ?, '', ?, ?, 0, 1, 0, 0, NOW(), NOW())`,
		id, parentID, title, path, component, permission, sort, isShow,
	)
	if err != nil {
		return 0, false, fmt.Errorf("创建菜单失败: %w", err)
	}
	fmt.Printf("  [菜单] 创建菜单: %s (%s)\n", title, path)
	return id, true, nil
}

// ensureButton 查找或创建按钮权限（type=3）
func (g *Generator) ensureButton(db *sql.DB, parentID int64, title, permission string, sort int) (bool, error) {
	var id int64
	err := db.QueryRow(
		"SELECT id FROM system_menu WHERE permission = ? AND type = 3 AND deleted_at IS NULL",
		permission,
	).Scan(&id)

	if err == nil {
		if g.config.Force {
			_, err = db.Exec(
				`UPDATE system_menu SET title=?, sort=?, updated_at=NOW() WHERE id=?`,
				title, sort, id,
			)
			if err != nil {
				return false, fmt.Errorf("更新按钮失败: %w", err)
			}
			fmt.Printf("  [菜单] 更新按钮: %s (%s)\n", title, permission)
			return false, nil
		}
		fmt.Printf("  [菜单] 跳过（已存在）: %s (%s)\n", title, permission)
		return false, nil
	}

	id = generateID()

	if g.config.DryRun {
		fmt.Printf("  [dry-run] INSERT 按钮: title=%s, permission=%s\n", title, permission)
		return true, nil
	}

	_, err = db.Exec(
		`INSERT INTO system_menu (id, parent_id, title, type, path, component, permission, icon, sort, is_show, is_cache, status, created_by, dept_id, created_at, updated_at)
		 VALUES (?, ?, ?, 3, NULL, NULL, ?, '', ?, 0, 0, 1, 0, 0, NOW(), NOW())`,
		id, parentID, title, permission, sort,
	)
	if err != nil {
		return false, fmt.Errorf("创建按钮失败: %w", err)
	}
	fmt.Printf("  [菜单] 创建按钮: %s (%s)\n", title, permission)
	return true, nil
}

// cleanTitle 从表注释中提取简短标题
func cleanTitle(comment string) string {
	if comment == "" {
		return ""
	}
	// 去掉常见后缀
	for _, suffix := range []string{"表", "管理"} {
		comment = strings.TrimSuffix(comment, suffix)
	}
	return comment
}

// dashCase 将 snake_case 的模块名转为 dash-case（用于 URL path）
func dashCase(s string) string {
	return strings.ReplaceAll(s, "_", "-")
}

// --- 内联 Snowflake ID 生成（与项目 utility/snowflake 算法一致）---

const (
	sfEpoch         = int64(1700000000000)
	sfWorkerBits    = uint(10)
	sfSequenceBits  = uint(12)
	sfSequenceMax   = int64(-1) ^ (int64(-1) << sfSequenceBits)
	sfWorkerShift   = sfSequenceBits
	sfTimestampShift = sfSequenceBits + sfWorkerBits
)

var sfGen = &sfGenerator{workerID: 1}

type sfGenerator struct {
	mu        sync.Mutex
	timestamp int64
	workerID  int64
	sequence  int64
}

func generateID() int64 {
	sfGen.mu.Lock()
	defer sfGen.mu.Unlock()

	now := time.Now().UnixMilli() - sfEpoch
	if now == sfGen.timestamp {
		sfGen.sequence = (sfGen.sequence + 1) & sfSequenceMax
		if sfGen.sequence == 0 {
			for now <= sfGen.timestamp {
				now = time.Now().UnixMilli() - sfEpoch
			}
		}
	} else {
		sfGen.sequence = 0
	}
	sfGen.timestamp = now

	return (now << sfTimestampShift) | (sfGen.workerID << sfWorkerShift) | sfGen.sequence
}

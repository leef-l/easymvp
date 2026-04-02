package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"time"

	"easymvp/codegen/generator/backend"
	"easymvp/codegen/generator/frontend"
	"easymvp/codegen/generator/menu"
	"easymvp/codegen/generator/util"
	"easymvp/codegen/parser"
)

func main() {
	// 命令行参数
	var (
		table    string // 表名，逗号分隔
		only     string // backend | frontend | menu | 空=都生成
		force    bool   // 强制覆盖
		config   string // 配置文件路径
		dryRun   bool   // 只打印不写入
		withMenu bool   // 同时生成菜单
	)

	flag.StringVar(&table, "table", "", "要生成的表名，多个用逗号分隔 (required)")
	flag.StringVar(&only, "only", "", "只生成指定端: backend | frontend | menu")
	flag.BoolVar(&force, "force", false, "强制覆盖已存在文件")
	flag.StringVar(&config, "config", "./codegen.yaml", "配置文件路径")
	flag.BoolVar(&dryRun, "dry-run", false, "只打印将生成的文件列表")
	flag.BoolVar(&withMenu, "menu", false, "同时生成菜单数据到数据库")
	flag.Parse()

	if table == "" {
		fmt.Println("错误: --table 参数不能为空")
		flag.Usage()
		os.Exit(1)
	}

	// 加载配置
	cfg, err := LoadConfig(config)
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 创建解析器
	p, err := parser.New(cfg.Database.DSN(), cfg.SkipFields)
	if err != nil {
		fmt.Printf("初始化解析器失败: %v\n", err)
		os.Exit(1)
	}
	defer p.Close()

	// 解析表名列表
	tableNames := strings.Split(table, ",")
	for i := range tableNames {
		tableNames[i] = strings.TrimSpace(tableNames[i])
	}

	start := time.Now()
	totalFiles := 0

	// 获取当前工作目录（用于计算模板路径）
	cwd, _ := os.Getwd()
	templateDir := filepath.Join(cwd, "templates")

	// 创建全局模板缓存
	tplCache := util.NewTemplateCache()

	// 按应用分组：记录每个应用的模块名和表名
	appModules := make(map[string][]string) // appName -> []moduleName
	appTables := make(map[string][]string)  // appName -> []tableName

	for _, tableName := range tableNames {
		fmt.Printf("\n[codegen] 开始生成表: %s\n", tableName)

		// 解析表结构
		meta, err := p.ParseTable(tableName)
		if err != nil {
			fmt.Printf("[codegen] ✗ 解析表 %s 失败: %v\n", tableName, err)
			continue
		}

		if meta.AppName == "" {
			fmt.Printf("[codegen] ✗ 表名 %s 缺少应用前缀（格式: {app}_{module}，如 system_dept）\n", tableName)
			continue
		}

		// 设置操作日志开关
		meta.EnableOpLog = cfg.OperationLog

		fmt.Printf("[codegen] 应用: %s, 模块: %s, DAO: %s\n", meta.AppName, meta.ModuleName, meta.DaoName)

		// 记录应用的模块和表名
		appModules[meta.AppName] = append(appModules[meta.AppName], meta.ModuleName)
		appTables[meta.AppName] = append(appTables[meta.AppName], meta.TableName)

		// 检查后端应用目录是否存在，不存在则自动创建
		appDir := filepath.Join(cfg.Backend.Output, meta.AppName)
		if _, err := os.Stat(appDir); os.IsNotExist(err) {
			fmt.Printf("[codegen] 应用目录 %s 不存在，正在创建...\n", appDir)
			projectRoot := filepath.Dir(cfg.Backend.Output)
			if projectRoot == "" {
				projectRoot = "."
			}
			initCmd := exec.Command("gf", "init", "app/"+meta.AppName, "-a")
			initCmd.Dir = projectRoot
			initCmd.Stdout = os.Stdout
			initCmd.Stderr = os.Stderr
			if err := initCmd.Run(); err != nil {
				fmt.Printf("[codegen] gf init 执行失败: %v，尝试手动创建目录\n", err)
				if mkErr := os.MkdirAll(appDir, 0755); mkErr != nil {
					fmt.Printf("[codegen] ✗ 创建目录失败: %v\n", mkErr)
					continue
				}
			}
			fmt.Printf("[codegen] 应用 %s 创建完成\n", meta.AppName)
		}

		var files []string

		if dryRun {
			// dry-run 模式：生成到内存并显示 diff
			if only != "frontend" && only != "menu" {
				backendGen := backend.New(backend.Config{
					TemplateDir: filepath.Join(templateDir, "backend"),
					OutputDir:   filepath.Join(cfg.Backend.Output, meta.AppName),
					Cache:       tplCache,
				})
				memFiles, err := backendGen.GenerateToMemory(meta)
				if err != nil {
					fmt.Printf("[codegen] ✗ 后端预览失败: %v\n", err)
				} else {
					printDiff(memFiles)
				}
			}
			if only != "backend" && only != "menu" {
				frontendGen := frontend.New(frontend.Config{
					TemplateDir: filepath.Join(templateDir, "frontend"),
					OutputDir:   cfg.Frontend.Output,
					Cache:       tplCache,
				})
				memFiles, err := frontendGen.GenerateToMemory(meta)
				if err != nil {
					fmt.Printf("[codegen] ✗ 前端预览失败: %v\n", err)
				} else {
					printDiff(memFiles)
				}
			}
		} else {
			// 正常生成模式
			// 生成后端代码
			if only != "frontend" && only != "menu" {
				backendOutput := filepath.Join(cfg.Backend.Output, meta.AppName)
				backendGen := backend.New(backend.Config{
					TemplateDir: filepath.Join(templateDir, "backend"),
					OutputDir:   backendOutput,
					Force:       force,
					Cache:       tplCache,
				})
				generated, err := backendGen.Generate(meta)
				if err != nil {
					fmt.Printf("[codegen] ✗ 后端生成失败: %v\n", err)
				} else {
					for _, f := range generated {
						fmt.Printf("[codegen] 后端: %s\n", f)
					}
					files = append(files, generated...)
				}
			}

			// 生成前端代码
			if only != "backend" && only != "menu" {
				frontendGen := frontend.New(frontend.Config{
					TemplateDir: filepath.Join(templateDir, "frontend"),
					OutputDir:   cfg.Frontend.Output,
					Force:       force,
					Cache:       tplCache,
				})
				generated, err := frontendGen.Generate(meta)
				if err != nil {
					fmt.Printf("[codegen] ✗ 前端生成失败: %v\n", err)
				} else {
					for _, f := range generated {
						fmt.Printf("[codegen] 前端: %s\n", f)
					}
					files = append(files, generated...)
				}
			}
		}

		fmt.Printf("[codegen] 表 %s 生成完成，共 %d 个文件\n", tableName, len(files))
		totalFiles += len(files)

		// 生成菜单数据
		if only == "menu" || withMenu {
			menuApps := make(map[string]menu.MenuAppConfig, len(cfg.MenuApps))
			for k, v := range cfg.MenuApps {
				menuApps[k] = menu.MenuAppConfig{Title: v.Title, Icon: v.Icon, Sort: v.Sort}
			}
			menuModules := make(map[string]menu.MenuModuleConfig, len(cfg.MenuModules))
			for k, v := range cfg.MenuModules {
				modCfg := menu.MenuModuleConfig{Sort: v.Sort}
				if v.IsShow != nil {
					modCfg.IsShow = v.IsShow
				}
				menuModules[k] = modCfg
			}
			menuGen := menu.New(menu.Config{
				DSN:         cfg.Database.DSN(),
				Force:       force,
				DryRun:      dryRun,
				MenuApps:    menuApps,
				MenuModules: menuModules,
			})
			menuCount, err := menuGen.Generate(meta)
			if err != nil {
				fmt.Printf("[codegen] ✗ 菜单生成失败: %v\n", err)
			} else {
				fmt.Printf("[codegen] 表 %s 菜单生成完成，新增 %d 条\n", tableName, menuCount)
				totalFiles += menuCount
			}
		}
	}

	// ========== 后置生成：按应用生成 DAO / main.go / cmd.go / middleware ==========
	if only != "frontend" && only != "menu" && !dryRun {
		for appName, newModules := range appModules {
			appDir := filepath.Join(cfg.Backend.Output, appName)
			fmt.Printf("\n[codegen] ===== 应用 %s 后置生成 =====\n", appName)

			// 1. 扫描已有的 logic 和 controller 目录，合并模块列表
			allModules := scanExistingModules(appDir, newModules)

			// 2. 收集所有表名（已有 + 新增）用于 hack/config.yaml
			allTables := scanExistingTables(appDir, appName, appTables[appName])

			// 3. 生成 hack/config.yaml
			hackDir := filepath.Join(appDir, "hack")
			hackFile := filepath.Join(hackDir, "config.yaml")
			if err := os.MkdirAll(hackDir, 0755); err != nil {
				fmt.Printf("[codegen] ✗ 创建 hack 目录失败: %v\n", err)
			} else {
				hackData := map[string]string{
					"DBLink": cfg.Database.DSNForHack(),
					"Tables": strings.Join(allTables, ","),
				}
				if err := renderTemplate(
					filepath.Join(templateDir, "backend", "hack_config.tpl"),
					hackFile,
					hackData,
					true, // hack/config.yaml 总是覆盖
				); err != nil {
					fmt.Printf("[codegen] ✗ 生成 hack/config.yaml 失败: %v\n", err)
				} else {
					fmt.Printf("[codegen] hack/config.yaml\n")
					totalFiles++
				}
			}

			// 4. 执行 gf gen dao
			fmt.Printf("[codegen] 执行 gf gen dao (应用: %s)...\n", appName)
			daoCmd := exec.Command("gf", "gen", "dao")
			daoCmd.Dir = appDir
			daoCmd.Stdout = os.Stdout
			daoCmd.Stderr = os.Stderr
			if err := daoCmd.Run(); err != nil {
				fmt.Printf("[codegen] gf gen dao 执行失败: %v\n", err)
			} else {
				fmt.Printf("[codegen] gf gen dao 完成\n")
			}

			// 5. 生成 main.go
			mainFile := filepath.Join(appDir, "main.go")
			mainData := map[string]interface{}{
				"AppName": appName,
				"Modules": allModules,
			}
			if err := renderTemplate(
				filepath.Join(templateDir, "backend", "main.tpl"),
				mainFile,
				mainData,
				force,
			); err != nil {
				fmt.Printf("[codegen] ✗ 生成 main.go 失败: %v\n", err)
			} else {
				fmt.Printf("[codegen] main.go\n")
				totalFiles++
			}

			// 6. 生成 internal/cmd/cmd.go
			cmdDir := filepath.Join(appDir, "internal", "cmd")
			if err := os.MkdirAll(cmdDir, 0755); err != nil {
				fmt.Printf("[codegen] ✗ 创建 cmd 目录失败: %v\n", err)
			} else {
				cmdFile := filepath.Join(cmdDir, "cmd.go")
				cmdData := map[string]interface{}{
					"AppName": appName,
					"Modules": allModules,
				}
				if err := renderTemplate(
					filepath.Join(templateDir, "backend", "cmd.tpl"),
					cmdFile,
					cmdData,
					force,
				); err != nil {
					fmt.Printf("[codegen] ✗ 生成 cmd.go 失败: %v\n", err)
				} else {
					fmt.Printf("[codegen] internal/cmd/cmd.go\n")
					totalFiles++
				}
			}

			// 7. 复制 middleware/auth.go（如果不存在）
			mwDir := filepath.Join(appDir, "internal", "middleware")
			mwFile := filepath.Join(mwDir, "auth.go")
			if _, err := os.Stat(mwFile); os.IsNotExist(err) {
				if err := os.MkdirAll(mwDir, 0755); err != nil {
					fmt.Printf("[codegen] ✗ 创建 middleware 目录失败: %v\n", err)
				} else {
					tplPath := filepath.Join(templateDir, "backend", "middleware_auth.tpl")
					content, err := os.ReadFile(tplPath)
					if err != nil {
						fmt.Printf("[codegen] ✗ 读取 middleware 模板失败: %v\n", err)
					} else {
						if err := os.WriteFile(mwFile, content, 0644); err != nil {
							fmt.Printf("[codegen] ✗ 写入 middleware/auth.go 失败: %v\n", err)
						} else {
							fmt.Printf("[codegen] internal/middleware/auth.go\n")
							totalFiles++
						}
					}
				}
			} else {
				fmt.Printf("[codegen] 跳过（已存在）: internal/middleware/auth.go\n")
			}

			// 8. 确保 internal/packed/packed.go 存在
			packedDir := filepath.Join(appDir, "internal", "packed")
			packedFile := filepath.Join(packedDir, "packed.go")
			if _, err := os.Stat(packedFile); os.IsNotExist(err) {
				if err := os.MkdirAll(packedDir, 0755); err != nil {
					fmt.Printf("[codegen] ✗ 创建 packed 目录失败: %v\n", err)
				} else {
					if err := os.WriteFile(packedFile, []byte("package packed\n"), 0644); err != nil {
						fmt.Printf("[codegen] ✗ 写入 packed.go 失败: %v\n", err)
					} else {
						fmt.Printf("[codegen] internal/packed/packed.go\n")
						totalFiles++
					}
				}
			}
		}
	}

	elapsed := time.Since(start)
	fmt.Printf("\n[codegen] 全部完成！共生成 %d 个文件，耗时 %.1fs\n", totalFiles, elapsed.Seconds())
}

// printDiff 打印文件 diff 预览
func printDiff(files map[string][]byte) {
	for path, newContent := range files {
		existing, err := os.ReadFile(path)
		if err != nil {
			// 新文件
			fmt.Printf("\n\033[32m+ 新文件: %s (%d bytes)\033[0m\n", path, len(newContent))
			continue
		}
		if bytes.Equal(existing, newContent) {
			fmt.Printf("  无变化: %s\n", path)
			continue
		}
		// 有差异
		fmt.Printf("\n\033[33m~ 有变化: %s\033[0m\n", path)
		oldLines := bytes.Split(existing, []byte("\n"))
		newLines := bytes.Split(newContent, []byte("\n"))
		fmt.Printf("  原文件: %d 行 -> 新文件: %d 行\n", len(oldLines), len(newLines))
	}
}

// scanExistingModules 扫描应用目录下已有的 logic 子目录，合并新模块，返回去重排序后的列表
func scanExistingModules(appDir string, newModules []string) []string {
	moduleSet := make(map[string]bool)
	for _, m := range newModules {
		moduleSet[m] = true
	}

	// 扫描 internal/logic/ 下的子目录（只包含有 .go 文件的目录）
	logicDir := filepath.Join(appDir, "internal", "logic")
	entries, err := os.ReadDir(logicDir)
	if err == nil {
		for _, e := range entries {
			if e.IsDir() {
				// 检查该目录下是否有 .go 文件，避免引入空目录或仅存在于 controller 的模块
				subEntries, _ := os.ReadDir(filepath.Join(logicDir, e.Name()))
				hasGo := false
				for _, se := range subEntries {
					if !se.IsDir() && filepath.Ext(se.Name()) == ".go" {
						hasGo = true
						break
					}
				}
				if hasGo {
					moduleSet[e.Name()] = true
				}
			}
		}
	}

	// 去重排序
	var modules []string
	for m := range moduleSet {
		modules = append(modules, m)
	}
	sort.Strings(modules)
	return modules
}

// scanExistingTables 扫描已有的 hack/config.yaml 中的表名，合并新表名，返回去重排序后的列表
func scanExistingTables(appDir string, appName string, newTables []string) []string {
	tableSet := make(map[string]bool)
	for _, t := range newTables {
		tableSet[t] = true
	}

	// 尝试从已有的 hack/config.yaml 中提取 tables 字段
	hackFile := filepath.Join(appDir, "hack", "config.yaml")
	data, err := os.ReadFile(hackFile)
	if err == nil {
		content := string(data)
		// 简单解析: 找到 tables: "xxx,yyy" 行
		for _, line := range strings.Split(content, "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "tables:") {
				val := strings.TrimPrefix(line, "tables:")
				val = strings.TrimSpace(val)
				val = strings.Trim(val, "\"")
				for _, t := range strings.Split(val, ",") {
					t = strings.TrimSpace(t)
					if t != "" {
						tableSet[t] = true
					}
				}
			}
		}
	}

	// 也扫描 internal/dao/internal/ 下的 .go 文件名作为表名推断
	// DAO 文件名就是完整表名（如 play_activity.go -> play_activity）
	daoInternalDir := filepath.Join(appDir, "internal", "dao", "internal")
	entries, err := os.ReadDir(daoInternalDir)
	if err == nil {
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".go") {
				name := strings.TrimSuffix(e.Name(), ".go")
				if name != "" {
					tableSet[name] = true
				}
			}
		}
	}

	var tables []string
	for t := range tableSet {
		tables = append(tables, t)
	}
	sort.Strings(tables)
	return tables
}

// renderTemplate 渲染模板到文件，overwrite 控制是否覆盖已有文件
// 内置 ModuleCamel 模板函数供所有模板使用
func renderTemplate(tplPath, outPath string, data interface{}, overwrite bool) error {
	if !overwrite {
		if _, err := os.Stat(outPath); err == nil {
			fmt.Printf("  跳过（已存在）: %s\n", outPath)
			return nil
		}
	}
	funcMap := template.FuncMap{
		"ModuleCamel": parser.SnakeToCamelSimple,
		"PackageName": func(s string) string {
			return strings.ReplaceAll(strings.ToLower(s), "_", "")
		},
	}
	tpl, err := template.New(filepath.Base(tplPath)).Funcs(funcMap).ParseFiles(tplPath)
	if err != nil {
		return fmt.Errorf("解析模板 %s 失败: %v", tplPath, err)
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("渲染模板失败: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		return fmt.Errorf("创建目录失败: %v", err)
	}
	return os.WriteFile(outPath, buf.Bytes(), 0644)
}

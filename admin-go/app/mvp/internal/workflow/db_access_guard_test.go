package workflow

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

func TestForbiddenDatabaseEntrypointsStayOutOfOrchestrationLayers(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}

	workflowDir := filepath.Dir(thisFile)
	internalDir := filepath.Dir(workflowDir)
	moduleDir := filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(workflowDir))))

	roots := []string{
		filepath.Join(internalDir, "controller", "chat"),
		workflowDir,
	}

	var findings []string
	for _, root := range roots {
		rootFindings, err := collectForbiddenDatabaseEntrypoints(root, moduleDir)
		if err != nil {
			t.Fatalf("scan %s: %v", root, err)
		}
		findings = append(findings, rootFindings...)
	}

	if len(findings) > 0 {
		t.Fatalf("forbidden database entrypoints found:\n%s", strings.Join(findings, "\n"))
	}
}

func collectForbiddenDatabaseEntrypoints(root, moduleDir string) ([]string, error) {
	var findings []string

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			switch d.Name() {
			case "repo", "configrepo":
				if path != root {
					return filepath.SkipDir
				}
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		fileFindings, err := collectForbiddenDatabaseEntrypointsInFile(path, moduleDir)
		if err != nil {
			return err
		}
		findings = append(findings, fileFindings...)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return findings, nil
}

func collectForbiddenDatabaseEntrypointsInFile(path, moduleDir string) ([]string, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		return nil, err
	}

	gAliases, dotImport, daoImports := collectGuardedImports(file)
	if len(gAliases) == 0 && !dotImport && len(daoImports) == 0 {
		return nil, nil
	}

	relPath, relErr := filepath.Rel(moduleDir, path)
	if relErr != nil {
		relPath = path
	}
	relPath = filepath.ToSlash(relPath)

	findings := make([]string, 0, len(daoImports)+1)
	for _, importPath := range daoImports {
		findings = append(findings, fmt.Sprintf("%s imports forbidden dao package %q", relPath, importPath))
	}

	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		if isForbiddenDBCall(call, gAliases, dotImport) {
			position := fset.Position(call.Pos())
			findings = append(findings, fmt.Sprintf("%s:%d calls forbidden DB entrypoint", relPath, position.Line))
		}
		return true
	})

	return findings, nil
}

func collectGuardedImports(file *ast.File) (map[string]struct{}, bool, []string) {
	gAliases := make(map[string]struct{})
	daoImports := make([]string, 0)
	dotImport := false

	for _, spec := range file.Imports {
		importPath, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			continue
		}

		switch {
		case importPath == "github.com/gogf/gf/v2/frame/g":
			if spec.Name != nil {
				switch spec.Name.Name {
				case ".":
					dotImport = true
				case "_":
				default:
					gAliases[spec.Name.Name] = struct{}{}
				}
				continue
			}
			gAliases["g"] = struct{}{}
		case strings.HasSuffix(importPath, "/internal/dao"), strings.Contains(importPath, "/internal/dao/"):
			daoImports = append(daoImports, importPath)
		}
	}

	return gAliases, dotImport, daoImports
}

func isForbiddenDBCall(call *ast.CallExpr, gAliases map[string]struct{}, dotImport bool) bool {
	switch fn := call.Fun.(type) {
	case *ast.SelectorExpr:
		ident, ok := fn.X.(*ast.Ident)
		if !ok || fn.Sel.Name != "DB" {
			return false
		}
		_, ok = gAliases[ident.Name]
		return ok
	case *ast.Ident:
		return dotImport && fn.Name == "DB"
	default:
		return false
	}
}

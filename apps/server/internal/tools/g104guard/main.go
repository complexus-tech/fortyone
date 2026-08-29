// Command g104guard verifies a complete gosec report while allowing one narrow
// legacy exception: a G104 finding on a standalone pkg/web response write. It
// deliberately consumes JSON without printing source snippets, which can
// contain request or credential material.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const webPackagePath = "github.com/complexus-tech/projects-api/pkg/web"

type gosecReport struct {
	GolangErrors map[string]json.RawMessage `json:"Golang errors"`
	Issues       []gosecIssue               `json:"Issues"`
	Stats        gosecStats                 `json:"Stats"`
}

type gosecIssue struct {
	RuleID string `json:"rule_id"`
	File   string `json:"file"`
	Line   string `json:"line"`
	Column string `json:"column"`
}

type gosecStats struct {
	Files int `json:"files"`
}

type location struct {
	RuleID string
	Path   string
	Line   int
	Column int
}

type auditResult struct {
	Allowed  int
	Rejected []location
}

type parsedSource struct {
	fileSet *token.FileSet
	file    *ast.File
}

func main() {
	rootPath := flag.String("root", ".", "server source root")
	flag.Parse()
	if flag.NArg() != 0 {
		fatal(errors.New("g104guard does not accept positional arguments"))
	}

	rootAbsolute, err := filepath.Abs(*rootPath)
	if err != nil {
		fatal(fmt.Errorf("resolve source root: %w", err))
	}
	root, err := os.OpenRoot(rootAbsolute)
	if err != nil {
		fatal(fmt.Errorf("open source root: %w", err))
	}
	defer func() {
		if closeErr := root.Close(); closeErr != nil {
			fatal(fmt.Errorf("close source root: %w", closeErr))
		}
	}()

	result, err := auditGosec(os.Stdin, rootAbsolute, root.FS())
	if err != nil {
		fatal(err)
	}
	if err := writeAudit(os.Stdout, result); err != nil {
		fatal(err)
	}
	if len(result.Rejected) != 0 {
		os.Exit(1)
	}
}

func auditGosec(input io.Reader, rootPath string, sourceFS fs.FS) (auditResult, error) {
	decoder := json.NewDecoder(input)
	var report gosecReport
	if err := decoder.Decode(&report); err != nil {
		return auditResult{}, errors.New("decode gosec report")
	}
	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return auditResult{}, errors.New("gosec report contains trailing data")
	}
	if len(report.GolangErrors) != 0 {
		return auditResult{}, fmt.Errorf("gosec could not analyze %d package(s)", len(report.GolangErrors))
	}
	if report.Stats.Files <= 0 {
		return auditResult{}, errors.New("gosec report did not analyze any files")
	}

	rootAbsolute, err := filepath.Abs(rootPath)
	if err != nil {
		return auditResult{}, errors.New("resolve gosec source root")
	}
	parsedFiles := make(map[string]parsedSource)
	seen := make(map[location]struct{}, len(report.Issues))
	result := auditResult{Rejected: make([]location, 0)}
	for _, issue := range report.Issues {
		item, err := normalizeLocation(rootAbsolute, issue)
		if err != nil {
			return auditResult{}, err
		}
		if _, duplicate := seen[item]; duplicate {
			return auditResult{}, fmt.Errorf("gosec returned duplicate finding at %s:%d:%d", item.Path, item.Line, item.Column)
		}
		seen[item] = struct{}{}

		if item.RuleID != "G104" {
			result.Rejected = append(result.Rejected, item)
			continue
		}

		parsed, ok := parsedFiles[item.Path]
		if !ok {
			source, readErr := fs.ReadFile(sourceFS, item.Path)
			if readErr != nil {
				return auditResult{}, fmt.Errorf("read G104 source %s", item.Path)
			}
			fileSet := token.NewFileSet()
			file, parseErr := parser.ParseFile(fileSet, item.Path, source, 0)
			if parseErr != nil {
				return auditResult{}, fmt.Errorf("parse G104 source %s", item.Path)
			}
			parsed = parsedSource{fileSet: fileSet, file: file}
			parsedFiles[item.Path] = parsed
		}

		if approvedResponseWrite(parsed, item) {
			result.Allowed++
			continue
		}
		result.Rejected = append(result.Rejected, item)
	}

	sort.Slice(result.Rejected, func(left, right int) bool {
		if result.Rejected[left].Path != result.Rejected[right].Path {
			return result.Rejected[left].Path < result.Rejected[right].Path
		}
		if result.Rejected[left].Line != result.Rejected[right].Line {
			return result.Rejected[left].Line < result.Rejected[right].Line
		}
		if result.Rejected[left].Column != result.Rejected[right].Column {
			return result.Rejected[left].Column < result.Rejected[right].Column
		}
		return result.Rejected[left].RuleID < result.Rejected[right].RuleID
	})
	return result, nil
}

func normalizeLocation(rootPath string, issue gosecIssue) (location, error) {
	if !validRuleID(issue.RuleID) {
		return location{}, errors.New("gosec returned an invalid rule identifier")
	}
	line, err := strconv.Atoi(issue.Line)
	if err != nil || line <= 0 {
		return location{}, errors.New("gosec returned an invalid finding line")
	}
	column, err := strconv.Atoi(issue.Column)
	if err != nil || column <= 0 {
		return location{}, errors.New("gosec returned an invalid finding column")
	}

	filePath := filepath.Clean(issue.File)
	if !filepath.IsAbs(filePath) {
		filePath = filepath.Join(rootPath, filePath)
	}
	relative, err := filepath.Rel(rootPath, filePath)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return location{}, errors.New("gosec finding is outside the source root")
	}
	relative = filepath.ToSlash(relative)
	if !fs.ValidPath(relative) || path.Ext(relative) != ".go" {
		return location{}, errors.New("gosec returned an invalid source path")
	}
	return location{RuleID: issue.RuleID, Path: relative, Line: line, Column: column}, nil
}

func validRuleID(ruleID string) bool {
	if len(ruleID) < 2 || len(ruleID) > 8 || ruleID[0] != 'G' {
		return false
	}
	for _, character := range ruleID[1:] {
		if character < '0' || character > '9' {
			return false
		}
	}
	return true
}

func approvedResponseWrite(source parsedSource, finding location) bool {
	webBindings := make(map[string]struct{})
	for _, importSpec := range source.file.Imports {
		importPath, err := strconv.Unquote(importSpec.Path.Value)
		if err != nil || importPath != webPackagePath {
			continue
		}
		binding := path.Base(importPath)
		if importSpec.Name != nil {
			binding = importSpec.Name.Name
		}
		if binding != "_" && binding != "." {
			webBindings[binding] = struct{}{}
		}
	}
	if len(webBindings) == 0 {
		return false
	}

	approved := false
	ast.Inspect(source.file, func(node ast.Node) bool {
		if approved {
			return false
		}
		statement, ok := node.(*ast.ExprStmt)
		if !ok {
			return true
		}
		position := source.fileSet.Position(statement.Pos())
		if position.Line != finding.Line || position.Column != finding.Column {
			return true
		}
		call, ok := statement.X.(*ast.CallExpr)
		if !ok || len(call.Args) != 4 {
			return false
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || (selector.Sel.Name != "Respond" && selector.Sel.Name != "RespondError") {
			return false
		}
		identifier, ok := selector.X.(*ast.Ident)
		if !ok || identifier.Obj != nil {
			return false
		}
		_, approved = webBindings[identifier.Name]
		return false
	})
	return approved
}

func writeAudit(output io.Writer, result auditResult) error {
	if _, err := fmt.Fprintf(
		output,
		"gosec guard: allowed %d standalone pkg/web response write(s); rejected %d finding(s)\n",
		result.Allowed,
		len(result.Rejected),
	); err != nil {
		return fmt.Errorf("write gosec summary: %w", err)
	}
	for _, finding := range result.Rejected {
		if _, err := fmt.Fprintf(
			output,
			"gosec guard: rejected %s at %s:%d:%d\n",
			finding.RuleID,
			finding.Path,
			finding.Line,
			finding.Column,
		); err != nil {
			return fmt.Errorf("write gosec finding: %w", err)
		}
	}
	return nil
}

func fatal(err error) {
	_, _ = fmt.Fprintln(os.Stderr, "g104guard:", err)
	os.Exit(1)
}

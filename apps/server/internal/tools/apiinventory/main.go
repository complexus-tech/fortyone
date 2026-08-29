// Command apiinventory generates a deterministic map of API route ownership,
// middleware, tests, persistence adoption, and module hotspots.
package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io/fs"
	"os"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const inventoryPath = "docs/inventory/api.md"

var namedQueryPattern = regexp.MustCompile(`(?m)^-- name: [A-Za-z][A-Za-z0-9_]* :(one|many|exec|execrows|execresult|copyfrom|batchexec|batchmany|batchone)\s*$`)

type route struct {
	Method         string
	Path           string
	Module         string
	Handler        string
	Middleware     []string
	CredentialMode string
	WorkspaceGuard bool
	PolicyGuard    string
	RateLimitGuard bool
	SourcePath     string
	Line           int
}

type moduleStats struct {
	Name                   string
	Routes                 int
	TestFiles              int
	TestFunctions          int
	NamedQueries           int
	SQLCGenerated          bool
	SQLXProductionFiles    []string
	LargestHandwrittenFile string
	LargestLineCount       int
}

func main() {
	write := flag.Bool("write", false, "write the generated API inventory")
	flag.Parse()
	if flag.NArg() != 0 {
		fatal(errors.New("apiinventory does not accept positional arguments"))
	}

	root, err := os.OpenRoot(".")
	if err != nil {
		fatal(fmt.Errorf("open server root: %w", err))
	}
	defer root.Close()

	generated, err := generate(root.FS())
	if err != nil {
		fatal(err)
	}
	if *write {
		if err := root.MkdirAll(path.Dir(inventoryPath), 0o755); err != nil {
			fatal(fmt.Errorf("create inventory directory: %w", err))
		}
		if err := writeRootFile(root, inventoryPath, generated); err != nil {
			fatal(err)
		}
		return
	}

	current, err := root.ReadFile(inventoryPath)
	if err != nil {
		fatal(fmt.Errorf("read API inventory; run make inventory-generate: %w", err))
	}
	if !bytes.Equal(current, generated) {
		fatal(errors.New("docs/inventory/api.md is stale; run make inventory-generate"))
	}
}

func generate(root fs.FS) ([]byte, error) {
	routes, err := collectRoutes(root)
	if err != nil {
		return nil, err
	}
	modules, err := collectModuleStats(root, routes)
	if err != nil {
		return nil, err
	}
	return render(routes, modules), nil
}

func collectRoutes(root fs.FS) ([]route, error) {
	routeFiles := make([]string, 0)
	err := fs.WalkDir(root, "internal", func(filePath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		if strings.HasSuffix(filePath, "/http/routes.go") {
			routeFiles = append(routeFiles, filePath)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk route files: %w", err)
	}
	sort.Strings(routeFiles)

	routes := make([]route, 0)
	for _, routeFile := range routeFiles {
		fileRoutes, err := parseRouteFile(root, routeFile)
		if err != nil {
			return nil, err
		}
		routes = append(routes, fileRoutes...)
	}
	sort.Slice(routes, func(left, right int) bool {
		if routes[left].Module != routes[right].Module {
			return routes[left].Module < routes[right].Module
		}
		if routes[left].Path != routes[right].Path {
			return routes[left].Path < routes[right].Path
		}
		return routes[left].Method < routes[right].Method
	})

	registered := make(map[string]route, len(routes))
	for _, item := range routes {
		key := item.Method + " " + item.Path
		if previous, duplicate := registered[key]; duplicate {
			return nil, fmt.Errorf(
				"duplicate route %s at %s:%d and %s:%d",
				key,
				previous.SourcePath,
				previous.Line,
				item.SourcePath,
				item.Line,
			)
		}
		registered[key] = item
	}
	return routes, nil
}

func parseRouteFile(root fs.FS, filePath string) ([]route, error) {
	source, err := fs.ReadFile(root, filePath)
	if err != nil {
		return nil, fmt.Errorf("read route file %s: %w", filePath, err)
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, source, 0)
	if err != nil {
		return nil, fmt.Errorf("parse route file %s: %w", filePath, err)
	}

	module := moduleForPath(filePath)
	bindings := collectMiddlewareBindings(fset, file)
	routes := make([]route, 0)
	var inspectErr error
	ast.Inspect(file, func(node ast.Node) bool {
		if inspectErr != nil {
			return false
		}
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || !isAppReceiver(selector.X) || !isHTTPMethod(selector.Sel.Name) {
			return true
		}
		if len(call.Args) < 2 {
			inspectErr = fmt.Errorf("route call at %s:%d needs path and handler", filePath, fset.Position(call.Pos()).Line)
			return false
		}
		pathLiteral, ok := call.Args[0].(*ast.BasicLit)
		if !ok || pathLiteral.Kind != token.STRING {
			inspectErr = fmt.Errorf("route path at %s:%d must be a string literal", filePath, fset.Position(call.Pos()).Line)
			return false
		}
		routePath, err := strconv.Unquote(pathLiteral.Value)
		if err != nil || !strings.HasPrefix(routePath, "/") {
			inspectErr = fmt.Errorf("invalid route path at %s:%d", filePath, fset.Position(call.Pos()).Line)
			return false
		}
		middleware := make([]string, 0, len(call.Args)-2)
		resolvedMiddleware := make([]string, 0, len(call.Args)-2)
		for _, expression := range call.Args[2:] {
			rendered := expressionString(fset, expression)
			middleware = append(middleware, rendered)
			resolvedMiddleware = append(resolvedMiddleware, resolveMiddlewareExpression(rendered, bindings))
		}
		credentialMode, workspaceGuard, policyGuard, rateLimitGuard := classifyRouteGuards(resolvedMiddleware)
		routes = append(routes, route{
			Method:         strings.ToUpper(selector.Sel.Name),
			Path:           routePath,
			Module:         module,
			Handler:        expressionString(fset, call.Args[1]),
			Middleware:     middleware,
			CredentialMode: credentialMode,
			WorkspaceGuard: workspaceGuard,
			PolicyGuard:    policyGuard,
			RateLimitGuard: rateLimitGuard,
			SourcePath:     filePath,
			Line:           fset.Position(call.Pos()).Line,
		})
		return true
	})
	if inspectErr != nil {
		return nil, inspectErr
	}
	return routes, nil
}

func collectModuleStats(root fs.FS, routes []route) ([]moduleStats, error) {
	byName := make(map[string]*moduleStats)
	for _, item := range routes {
		stats := ensureModule(byName, item.Module)
		stats.Routes++
	}

	err := fs.WalkDir(root, "internal/modules", func(filePath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		module := moduleForPath(filePath)
		stats := ensureModule(byName, module)

		switch path.Ext(filePath) {
		case ".sql":
			if !strings.Contains(filePath, "/repository/queries/") {
				return nil
			}
			raw, err := fs.ReadFile(root, filePath)
			if err != nil {
				return err
			}
			stats.NamedQueries += len(namedQueryPattern.FindAll(raw, -1))
			return nil
		case ".go":
		default:
			return nil
		}

		raw, err := fs.ReadFile(root, filePath)
		if err != nil {
			return err
		}
		if strings.Contains(filePath, "/repository/sqlc/") && isGeneratedGo(raw) {
			stats.SQLCGenerated = true
			return nil
		}
		if strings.HasSuffix(filePath, "_test.go") {
			stats.TestFiles++
			count, err := countTestFunctions(filePath, raw)
			if err != nil {
				return err
			}
			stats.TestFunctions += count
			return nil
		}
		if isGeneratedGo(raw) {
			return nil
		}

		lineCount := bytes.Count(raw, []byte("\n"))
		if len(raw) > 0 && raw[len(raw)-1] != '\n' {
			lineCount++
		}
		if lineCount > stats.LargestLineCount {
			stats.LargestLineCount = lineCount
			stats.LargestHandwrittenFile = filePath
		}
		usesSQLX, err := importsPackage(filePath, raw, "github.com/jmoiron/sqlx")
		if err != nil {
			return err
		}
		if usesSQLX {
			stats.SQLXProductionFiles = append(stats.SQLXProductionFiles, filePath)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("collect module inventory: %w", err)
	}

	names := make([]string, 0, len(byName))
	for name := range byName {
		names = append(names, name)
	}
	sort.Strings(names)
	modules := make([]moduleStats, 0, len(names))
	for _, name := range names {
		stats := byName[name]
		sort.Strings(stats.SQLXProductionFiles)
		modules = append(modules, *stats)
	}
	return modules, nil
}

func render(routes []route, modules []moduleStats) []byte {
	var output strings.Builder
	output.WriteString("<!-- Code generated by internal/tools/apiinventory; DO NOT EDIT. -->\n\n")
	output.WriteString("# API code inventory\n\n")
	output.WriteString("This deterministic inventory shows where routes and persistence live in the API codebase. Run `make inventory-generate` after moving a route, test, or query, or changing a persistence dependency; `make inventory-check` rejects drift. Counts describe code shape, not test quality or security approval.\n\n")
	requiredAuth, optionalAuth, noAuth, workspaceGuarded, policyGuarded, rateLimited, webhookRoutes := summarizeRouteGuards(routes)
	output.WriteString(fmt.Sprintf("Current snapshot: **%d routes across %d modules**. Registered middleware classifies %d routes with required user authentication, %d with optional authentication, %d without user-auth middleware, %d with current workspace-membership resolution, %d with an explicit role/scope guard, and %d with a route-level rate limit. The no-user-auth set includes %d webhook routes whose provider signature/replay policy must be verified in their handler contract. These are registration facts, not proof of complete service/resource authorization.\n\n", len(routes), len(modules), requiredAuth, optionalAuth, noAuth, workspaceGuarded, policyGuarded, rateLimited, webhookRoutes))
	output.WriteString("Credential configuration is indexed in [`docs/configuration.md`](../configuration.md), migration compatibility in [`docs/database/migration-operations.md`](../database/migration-operations.md), and architectural exceptions in the enforced debt baseline.\n\n")
	output.WriteString("## Module map\n\n")
	output.WriteString("| Module | Routes | Test files / functions | Named SQLC queries | Generated SQLC | Production SQLx files | Largest handwritten production file |\n")
	output.WriteString("| --- | ---: | ---: | ---: | --- | ---: | --- |\n")
	for _, stats := range modules {
		largest := "—"
		if stats.LargestHandwrittenFile != "" {
			largest = fmt.Sprintf("[%s](../../%s#L1) (%d)", path.Base(stats.LargestHandwrittenFile), stats.LargestHandwrittenFile, stats.LargestLineCount)
		}
		output.WriteString(fmt.Sprintf(
			"| `%s` | %d | %d / %d | %d | %s | %d | %s |\n",
			stats.Name,
			stats.Routes,
			stats.TestFiles,
			stats.TestFunctions,
			stats.NamedQueries,
			yesNo(stats.SQLCGenerated),
			len(stats.SQLXProductionFiles),
			largest,
		))
	}

	output.WriteString("\n## Route ownership\n\n")
	currentModule := ""
	for _, item := range routes {
		if item.Module != currentModule {
			currentModule = item.Module
			output.WriteString("### " + currentModule + "\n\n")
			output.WriteString("| Method | Path | Registered guards | Handler | Registered middleware | Source |\n")
			output.WriteString("| --- | --- | --- | --- | --- | --- |\n")
		}
		middleware := "—"
		if len(item.Middleware) > 0 {
			middleware = "`" + escapeCode(strings.Join(item.Middleware, " → ")) + "`"
		}
		output.WriteString(fmt.Sprintf(
			"| `%s` | `%s` | `%s` | `%s` | %s | [routes.go](../../%s#L%d) |\n",
			item.Method,
			escapeCode(item.Path),
			escapeCode(renderRouteGuards(item)),
			escapeCode(item.Handler),
			middleware,
			item.SourcePath,
			item.Line,
		))
	}
	return []byte(output.String())
}

func collectMiddlewareBindings(fset *token.FileSet, file *ast.File) map[string]string {
	bindings := make(map[string]string)
	ast.Inspect(file, func(node ast.Node) bool {
		assignment, ok := node.(*ast.AssignStmt)
		if !ok || len(assignment.Lhs) != len(assignment.Rhs) {
			return true
		}
		for index, left := range assignment.Lhs {
			identifier, ok := left.(*ast.Ident)
			if !ok || identifier.Name == "_" {
				continue
			}
			bindings[identifier.Name] = expressionString(fset, assignment.Rhs[index])
		}
		return true
	})
	return bindings
}

func resolveMiddlewareExpression(expression string, bindings map[string]string) string {
	resolved, ok := bindings[expression]
	if !ok {
		return expression
	}
	return expression + "=" + resolved
}

func classifyRouteGuards(middleware []string) (credentialMode string, workspaceGuard bool, policyGuard string, rateLimitGuard bool) {
	credentialMode = "none"
	policies := make([]string, 0, 2)
	for _, expression := range middleware {
		lower := strings.ToLower(expression)
		switch {
		case strings.Contains(lower, "optionalauth"):
			credentialMode = "optional"
		case credentialMode == "none" && (strings.Contains(expression, ".Auth(") || strings.Contains(expression, "MachineAuth") || strings.EqualFold(expression, "auth")):
			credentialMode = "required"
		}
		if strings.Contains(expression, ".Workspace(") || strings.EqualFold(expression, "workspace") {
			workspaceGuard = true
		}
		if strings.Contains(expression, "RequireMinimumRole") {
			switch {
			case strings.Contains(expression, "RoleAdmin"):
				policies = append(policies, "role>=admin")
			case strings.Contains(expression, "RoleMember"):
				policies = append(policies, "role>=member")
			case strings.Contains(expression, "RoleGuest"):
				policies = append(policies, "role>=guest")
			default:
				policies = append(policies, "role")
			}
		}
		if strings.Contains(expression, "RequireScopes") {
			policies = append(policies, "scope")
		}
		if strings.Contains(lower, "ratelimit") {
			rateLimitGuard = true
		}
	}
	if len(policies) == 0 {
		policyGuard = "none"
	} else {
		policyGuard = strings.Join(policies, "+")
	}
	return credentialMode, workspaceGuard, policyGuard, rateLimitGuard
}

func summarizeRouteGuards(routes []route) (requiredAuth, optionalAuth, noAuth, workspace, policy, rateLimited, webhooks int) {
	for _, item := range routes {
		switch item.CredentialMode {
		case "required":
			requiredAuth++
		case "optional":
			optionalAuth++
		default:
			noAuth++
		}
		if item.WorkspaceGuard {
			workspace++
		}
		if item.PolicyGuard != "none" {
			policy++
		}
		if item.RateLimitGuard {
			rateLimited++
		}
		if strings.HasPrefix(item.Path, "/webhooks/") {
			webhooks++
		}
	}
	return requiredAuth, optionalAuth, noAuth, workspace, policy, rateLimited, webhooks
}

func renderRouteGuards(item route) string {
	guards := []string{"auth=" + item.CredentialMode}
	if item.WorkspaceGuard {
		guards = append(guards, "workspace=current")
	}
	if item.PolicyGuard != "none" {
		guards = append(guards, item.PolicyGuard)
	}
	if item.RateLimitGuard {
		guards = append(guards, "rate-limited")
	}
	if strings.HasPrefix(item.Path, "/webhooks/") {
		guards = append(guards, "webhook-contract")
	}
	return strings.Join(guards, "; ")
}

func moduleForPath(filePath string) string {
	parts := strings.Split(filePath, "/")
	if len(parts) >= 3 && parts[0] == "internal" && parts[1] == "modules" {
		return parts[2]
	}
	if len(parts) >= 2 && parts[0] == "internal" {
		return parts[1]
	}
	return "unknown"
}

func ensureModule(modules map[string]*moduleStats, name string) *moduleStats {
	stats := modules[name]
	if stats == nil {
		stats = &moduleStats{Name: name}
		modules[name] = stats
	}
	return stats
}

func isAppReceiver(expression ast.Expr) bool {
	identifier, ok := expression.(*ast.Ident)
	return ok && identifier.Name == "app"
}

func isHTTPMethod(name string) bool {
	switch name {
	case "Get", "Post", "Put", "Patch", "Delete":
		return true
	default:
		return false
	}
}

func expressionString(fset *token.FileSet, expression ast.Expr) string {
	var output bytes.Buffer
	if err := printer.Fprint(&output, fset, expression); err != nil {
		return "unknown"
	}
	return strings.Join(strings.Fields(output.String()), " ")
}

func countTestFunctions(filePath string, source []byte) (int, error) {
	file, err := parser.ParseFile(token.NewFileSet(), filePath, source, 0)
	if err != nil {
		return 0, fmt.Errorf("parse test file %s: %w", filePath, err)
	}
	count := 0
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Recv != nil {
			continue
		}
		if strings.HasPrefix(function.Name.Name, "Test") || strings.HasPrefix(function.Name.Name, "Fuzz") || strings.HasPrefix(function.Name.Name, "Benchmark") {
			count++
		}
	}
	return count, nil
}

func importsPackage(filePath string, source []byte, importPath string) (bool, error) {
	file, err := parser.ParseFile(token.NewFileSet(), filePath, source, parser.ImportsOnly)
	if err != nil {
		return false, fmt.Errorf("parse imports for %s: %w", filePath, err)
	}
	for _, spec := range file.Imports {
		value, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			return false, fmt.Errorf("parse import in %s: %w", filePath, err)
		}
		if value == importPath {
			return true, nil
		}
	}
	return false, nil
}

func isGeneratedGo(source []byte) bool {
	prefix := source
	if len(prefix) > 512 {
		prefix = prefix[:512]
	}
	return bytes.Contains(prefix, []byte("Code generated")) && bytes.Contains(prefix, []byte("DO NOT EDIT"))
}

func escapeCode(value string) string {
	value = strings.ReplaceAll(value, "|", "\\|")
	return strings.ReplaceAll(value, "`", "\\`")
}

func yesNo(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}

func writeRootFile(root *os.Root, filePath string, data []byte) error {
	file, err := root.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("open %s: %w", filePath, err)
	}
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		return fmt.Errorf("write %s: %w", filePath, err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close %s: %w", filePath, err)
	}
	return nil
}

func fatal(err error) {
	_, _ = fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

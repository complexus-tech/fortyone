package architecture_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const (
	architectureBaselineVersion = 1
	architectureBaselinePath    = "testdata/debt-baseline.json"
	handwrittenFileLineLimit    = 700

	ruleCrossModuleHTTPImport     = "cross_module_http_import"
	ruleCrossModuleServiceImport  = "cross_module_service_import"
	ruleDirectSQL                 = "direct_sql_outside_persistence"
	ruleGeneratedOpenAPILeak      = "generated_openapi_leak"
	ruleGeneratedSQLCLeak         = "generated_sqlc_leak"
	ruleModuleDependencyCycle     = "module_dependency_cycle"
	ruleOversizedHandwrittenFile  = "oversized_handwritten_file"
	ruleRawDBRouteConfig          = "raw_database_route_config"
	ruleRepositoryAuthContextRead = "repository_auth_context_read"
	ruleRepositoryImport          = "repository_repository_import"
	ruleRepositoryServiceImport   = "repository_service_import"
	ruleRepositoryServiceHTTP     = "repository_service_http_import"
	ruleServiceRepositoryImport   = "service_repository_import"
	ruleSQLXImport                = "sqlx_import"
	ruleUnsafeRawRequestBodyRead  = "unsafe_raw_request_body_read"
	sqlxImportPath                = "github.com/jmoiron/sqlx"
	standardDatabaseSQLImportPath = "database/sql"
	projectImportPrefix           = "github.com/complexus-tech/projects-api/"
)

var countedArchitectureRules = []string{
	ruleCrossModuleHTTPImport,
	ruleCrossModuleServiceImport,
	ruleDirectSQL,
	ruleGeneratedOpenAPILeak,
	ruleGeneratedSQLCLeak,
	ruleModuleDependencyCycle,
	ruleRawDBRouteConfig,
	ruleRepositoryAuthContextRead,
	ruleRepositoryImport,
	ruleRepositoryServiceHTTP,
	ruleRepositoryServiceImport,
	ruleServiceRepositoryImport,
	ruleSQLXImport,
	ruleUnsafeRawRequestBodyRead,
}

var ruleRemediation = map[string]string{
	ruleCrossModuleHTTPImport:     "Depend on a caller-owned capability interface or a transport-neutral domain type; never reuse another module's handler package.",
	ruleCrossModuleServiceImport:  "Define the capability needed by the caller as a narrow interface in the caller's service package, implement it in the owning module, and wire the adapter in bootstrap.",
	ruleDirectSQL:                 "Move application SQL to internal/modules/<module>/repository/queries (or the database platform/migrations when that is the real owner) and expose a typed repository operation.",
	ruleGeneratedOpenAPILeak:      "Map generated OpenAPI request and response values in the HTTP adapter; service and domain packages must expose handwritten transport-neutral types.",
	ruleGeneratedSQLCLeak:         "Map generated sqlc values inside the owning repository; services, domain packages, HTTP handlers, and other modules must consume handwritten types.",
	ruleModuleDependencyCycle:     "Break the cycle with a caller-owned capability port, domain event, or platform primitive, then compose the concrete implementation only in bootstrap.",
	ruleOversizedHandwrittenFile:  "Split the file by cohesive use case or capability. Generated files are detected by their standard DO NOT EDIT header and are exempt.",
	ruleRawDBRouteConfig:          "Construct database-backed middleware or repositories in bootstrap and inject narrow ports into the route config instead of a raw DB handle.",
	ruleRepositoryAuthContextRead: "Resolve and authorize the actor in the service/policy layer, then pass explicit actor or tenant identifiers into the repository operation.",
	ruleRepositoryImport:          "Replace the concrete repository dependency with one aggregate-owned operation or a narrow caller-owned port; repositories must not orchestrate sibling repositories.",
	ruleRepositoryServiceImport:   "Move persistence models into the repository or a transport-neutral domain package and pass explicit values; persistence must not depend on the use-case layer.",
	ruleRepositoryServiceHTTP:     "Move tracing/context helpers to a transport-neutral platform package and keep repository/service packages independent of HTTP and pkg/web.",
	ruleServiceRepositoryImport:   "Declare the required persistence behavior as a narrow interface in the service package and inject the concrete repository adapter from bootstrap.",
	ruleSQLXImport:                "Migrate the call site behind a pgx/sqlc repository boundary. SQLx is transitional debt and new imports are not accepted.",
	ruleUnsafeRawRequestBodyRead:  "Read signed or raw request bytes once with web.ReadBoundedBody (or an equivalent MaxBytesReader that rejects overflow); io.LimitReader silently truncates and is not an ingress limit.",
}

var applicationSQLPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?is)^SELECT\s+(?:EXISTS\s*\(|.+\s+FROM\s+|[a-z_][a-z0-9_]*\s*\(|\$[0-9]+|[0-9]+\b)`),
	regexp.MustCompile(`(?is)^INSERT\s+INTO\s+[a-z_\"][a-z0-9_\".]*\s*(?:\(|VALUES\b|SELECT\b)`),
	regexp.MustCompile(`(?is)^UPDATE\s+[a-z_\"][a-z0-9_\".]*\s+(?:(?:AS\s+)?[a-z_][a-z0-9_]*\s+)?SET\s+`),
	regexp.MustCompile(`(?is)^DELETE\s+FROM\s+[a-z_\"][a-z0-9_\".]*\b`),
	regexp.MustCompile(`(?is)^WITH\s+(?:RECURSIVE\s+)?[a-z_\"][a-z0-9_\"]*(?:\s*\([^)]*\))?\s+AS\s*\(`),
	regexp.MustCompile(`(?is)^(?:CREATE\s+(?:TABLE|INDEX|UNIQUE\s+INDEX)|ALTER\s+TABLE|DROP\s+(?:TABLE|INDEX)|TRUNCATE|MERGE\s+INTO|GRANT|REVOKE)\s+`),
}

type architectureFinding struct {
	Rule   string
	Path   string
	Line   int
	Detail string
	Value  int
}

func (f architectureFinding) String() string {
	location := f.Path
	if f.Line > 0 {
		location += ":" + strconv.Itoa(f.Line)
	}
	return location + " - " + f.Detail
}

type architectureBaseline struct {
	Version        int                       `json:"version"`
	RulePathCounts map[string]map[string]int `json:"rulePathCounts"`
	OversizedFiles map[string]int            `json:"oversizedFiles"`
}

func scanArchitecture(root string) ([]architectureFinding, error) {
	findings := make([]architectureFinding, 0)
	moduleEdges := make([]moduleImportEdge, 0)
	fset := token.NewFileSet()

	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			switch entry.Name() {
			case ".git", ".tools", "tmp", "vendor":
				return filepath.SkipDir
			default:
				return nil
			}
		}

		relativePath, err := filepath.Rel(root, path)
		if err != nil {
			return fmt.Errorf("make %s relative to %s: %w", path, root, err)
		}
		relativePath = filepath.ToSlash(relativePath)

		switch filepath.Ext(path) {
		case ".go":
			source, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("read %s: %w", relativePath, err)
			}
			fileFindings, err := inspectGoSource(fset, relativePath, source)
			if err != nil {
				return err
			}
			findings = append(findings, fileFindings...)
			if !strings.HasSuffix(relativePath, "_test.go") && !isGeneratedSource(source) {
				edges, err := inspectModuleImportEdges(fset, relativePath, source)
				if err != nil {
					return err
				}
				moduleEdges = append(moduleEdges, edges...)
			}
		case ".sql":
			if isApprovedSQLPath(relativePath) {
				return nil
			}
			source, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("read %s: %w", relativePath, err)
			}
			if looksLikeSQL(string(source)) {
				findings = append(findings, architectureFinding{
					Rule:   ruleDirectSQL,
					Path:   relativePath,
					Line:   firstSQLLine(source),
					Detail: "SQL file is outside an approved repository queries, database platform, or migrations directory",
				})
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scan server architecture: %w", err)
	}

	findings = append(findings, findModuleCycleFindings(moduleEdges)...)
	sortFindings(findings)
	return findings, nil
}

func inspectGoSource(fset *token.FileSet, path string, source []byte) ([]architectureFinding, error) {
	if isGeneratedSource(source) {
		return nil, nil
	}

	file, err := parser.ParseFile(fset, path, source, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}

	findings := make([]architectureFinding, 0)
	lineCount := physicalLineCount(source)
	if lineCount > handwrittenFileLineLimit {
		findings = append(findings, architectureFinding{
			Rule:   ruleOversizedHandwrittenFile,
			Path:   path,
			Line:   1,
			Detail: fmt.Sprintf("handwritten file has %d lines; the refactor gate is %d", lineCount, handwrittenFileLineLimit),
			Value:  lineCount,
		})
	}

	importsByAlias := make(map[string]string, len(file.Imports))
	for _, spec := range file.Imports {
		importPath, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			return nil, fmt.Errorf("unquote import in %s: %w", path, err)
		}
		alias := filepath.Base(importPath)
		if spec.Name != nil {
			alias = spec.Name.Name
		}
		importsByAlias[alias] = importPath

		line := fset.Position(spec.Pos()).Line
		if importPath == sqlxImportPath {
			findings = append(findings, architectureFinding{
				Rule: ruleSQLXImport, Path: path, Line: line,
				Detail: "imports transitional SQLx package " + importPath,
			})
		}
		if isGeneratedSQLCImport(importPath) && !isOwningRepositoryImport(path, importPath) {
			findings = append(findings, architectureFinding{
				Rule: ruleGeneratedSQLCLeak, Path: path, Line: line,
				Detail: "imports repository-generated package " + importPath,
			})
		}
		if isRepositoryOrServicePath(path) && isHTTPDependency(importPath) {
			findings = append(findings, architectureFinding{
				Rule: ruleRepositoryServiceHTTP, Path: path, Line: line,
				Detail: "persistence/business layer imports transport dependency " + importPath,
			})
		}
		if isCrossModuleHTTPImport(path, importPath) {
			findings = append(findings, architectureFinding{
				Rule: ruleCrossModuleHTTPImport, Path: path, Line: line,
				Detail: "imports another module's HTTP package " + importPath,
			})
		}
		findings = append(findings, inspectLayerImport(path, importPath, line)...)
	}

	if isRouteConfigFile(path) {
		findings = append(findings, inspectRouteConfigFields(fset, path, file, importsByAlias)...)
	}
	findings = append(findings, inspectRepositoryAuthContextReads(fset, path, file, importsByAlias)...)
	findings = append(findings, inspectUnsafeRawRequestBodyReads(fset, path, file, importsByAlias)...)
	if !strings.HasSuffix(path, "_test.go") && !isApprovedSQLPath(path) {
		ast.Inspect(file, func(node ast.Node) bool {
			literal, ok := node.(*ast.BasicLit)
			if !ok || literal.Kind != token.STRING {
				return true
			}
			value, err := strconv.Unquote(literal.Value)
			if err != nil || !looksLikeSQL(value) {
				return true
			}
			findings = append(findings, architectureFinding{
				Rule: ruleDirectSQL, Path: path, Line: fset.Position(literal.Pos()).Line,
				Detail: "contains direct SQL beginning " + strconv.Quote(sqlPreview(value)),
			})
			return true
		})
	}

	sortFindings(findings)
	return findings, nil
}

func inspectRouteConfigFields(fset *token.FileSet, path string, file *ast.File, imports map[string]string) []architectureFinding {
	findings := make([]architectureFinding, 0)
	for _, declaration := range file.Decls {
		generic, ok := declaration.(*ast.GenDecl)
		if !ok || generic.Tok != token.TYPE {
			continue
		}
		for _, rawSpec := range generic.Specs {
			spec, ok := rawSpec.(*ast.TypeSpec)
			if !ok || !isRouteConfigTypeName(spec.Name.Name) {
				continue
			}
			structure, ok := spec.Type.(*ast.StructType)
			if !ok {
				continue
			}
			for _, field := range structure.Fields.List {
				databaseType, ok := rawDatabaseType(field.Type, imports)
				if !ok {
					continue
				}
				fieldName := "<embedded>"
				if len(field.Names) > 0 {
					fieldName = field.Names[0].Name
				}
				findings = append(findings, architectureFinding{
					Rule: ruleRawDBRouteConfig, Path: path, Line: fset.Position(field.Pos()).Line,
					Detail: fmt.Sprintf("route config field %s.%s exposes raw %s", spec.Name.Name, fieldName, databaseType),
				})
			}
		}
	}
	return findings
}

func rawDatabaseType(expression ast.Expr, imports map[string]string) (string, bool) {
	for {
		switch typed := expression.(type) {
		case *ast.StarExpr:
			expression = typed.X
		case *ast.SelectorExpr:
			alias, ok := typed.X.(*ast.Ident)
			if !ok {
				return "", false
			}
			importPath := imports[alias.Name]
			switch importPath {
			case sqlxImportPath:
				return importPath + "." + typed.Sel.Name, typed.Sel.Name == "DB" || typed.Sel.Name == "Tx"
			case standardDatabaseSQLImportPath:
				return importPath + "." + typed.Sel.Name, typed.Sel.Name == "DB" || typed.Sel.Name == "Tx"
			case "github.com/jackc/pgx/v5":
				return importPath + "." + typed.Sel.Name, typed.Sel.Name == "Conn" || typed.Sel.Name == "Tx"
			case "github.com/jackc/pgx/v5/pgxpool":
				return importPath + "." + typed.Sel.Name, typed.Sel.Name == "Pool"
			default:
				return "", false
			}
		default:
			return "", false
		}
	}
}

func loadArchitectureBaseline(path string) (architectureBaseline, []byte, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return architectureBaseline{}, nil, fmt.Errorf("read architecture baseline: %w", err)
	}
	var baseline architectureBaseline
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&baseline); err != nil {
		return architectureBaseline{}, nil, fmt.Errorf("decode architecture baseline: %w", err)
	}
	if baseline.Version != architectureBaselineVersion {
		return architectureBaseline{}, nil, fmt.Errorf("architecture baseline version is %d; expected %d", baseline.Version, architectureBaselineVersion)
	}
	if err := validateArchitectureBaseline(baseline); err != nil {
		return architectureBaseline{}, nil, err
	}
	return baseline, raw, nil
}

func validateArchitectureBaseline(baseline architectureBaseline) error {
	if baseline.RulePathCounts == nil || baseline.OversizedFiles == nil {
		return fmt.Errorf("architecture baseline maps must not be null")
	}
	knownRules := make(map[string]struct{}, len(countedArchitectureRules))
	for _, rule := range countedArchitectureRules {
		knownRules[rule] = struct{}{}
		if baseline.RulePathCounts[rule] == nil {
			return fmt.Errorf("architecture baseline is missing rule %q", rule)
		}
	}
	for rule, paths := range baseline.RulePathCounts {
		if _, ok := knownRules[rule]; !ok {
			return fmt.Errorf("architecture baseline contains unknown rule %q", rule)
		}
		for path, count := range paths {
			if err := validateBaselinePath(path); err != nil {
				return fmt.Errorf("architecture baseline rule %s: %w", rule, err)
			}
			if count <= 0 {
				return fmt.Errorf("architecture baseline rule %s path %s has non-positive count %d", rule, path, count)
			}
		}
	}
	for path, lines := range baseline.OversizedFiles {
		if err := validateBaselinePath(path); err != nil {
			return fmt.Errorf("architecture oversized-file baseline: %w", err)
		}
		if lines <= handwrittenFileLineLimit {
			return fmt.Errorf("architecture oversized-file baseline path %s has %d lines; entries must exceed the %d-line gate", path, lines, handwrittenFileLineLimit)
		}
	}
	return nil
}

func validateBaselinePath(path string) error {
	if path == "" || filepath.IsAbs(path) || filepath.ToSlash(path) != path || filepath.Clean(path) != path || path == "." || strings.HasPrefix(path, "../") {
		return fmt.Errorf("path %q must be a clean, relative, forward-slash server path", path)
	}
	return nil
}

func snapshotArchitectureDebt(findings []architectureFinding) architectureBaseline {
	baseline := architectureBaseline{
		Version:        architectureBaselineVersion,
		RulePathCounts: make(map[string]map[string]int, len(countedArchitectureRules)),
		OversizedFiles: make(map[string]int),
	}
	for _, rule := range countedArchitectureRules {
		baseline.RulePathCounts[rule] = make(map[string]int)
	}
	for _, finding := range findings {
		if finding.Rule == ruleOversizedHandwrittenFile {
			baseline.OversizedFiles[finding.Path] = finding.Value
			continue
		}
		baseline.RulePathCounts[finding.Rule][finding.Path]++
	}
	return baseline
}

func compareArchitectureDebt(expected architectureBaseline, findings []architectureFinding) []string {
	actual := snapshotArchitectureDebt(findings)
	findingIndex := indexFindings(findings)
	differences := make([]string, 0)

	for _, rule := range countedArchitectureRules {
		paths := unionSortedKeys(expected.RulePathCounts[rule], actual.RulePathCounts[rule])
		for _, path := range paths {
			want := expected.RulePathCounts[rule][path]
			got := actual.RulePathCounts[rule][path]
			if got <= want {
				continue
			}
			differences = append(differences, renderDebtGrowth(rule, path, architectureScope(path), want, got, findingIndex[rule+"\x00"+path]))
		}
	}

	paths := unionSortedKeys(expected.OversizedFiles, actual.OversizedFiles)
	for _, path := range paths {
		want := expected.OversizedFiles[path]
		got := actual.OversizedFiles[path]
		if got <= want {
			continue
		}
		differences = append(differences, renderDebtGrowth(ruleOversizedHandwrittenFile, path, architectureScope(path), want, got, findingIndex[ruleOversizedHandwrittenFile+"\x00"+path]))
	}

	sort.Strings(differences)
	return differences
}

func renderDebtGrowth(rule, path, scope string, baseline, current int, findings []architectureFinding) string {
	var message strings.Builder
	fmt.Fprintf(&message, "[%s] architecture debt grew for %s (%s): baseline=%d current=%d.\n", rule, path, scope, baseline, current)
	for _, finding := range findings {
		fmt.Fprintf(&message, "  - %s\n", finding.String())
	}
	fmt.Fprintf(&message, "  Remediation: %s\n", ruleRemediation[rule])
	message.WriteString("  Do not increase debt-baseline.json to accept new debt; update it only when current debt is removed or an architecture decision explicitly changes the rule.")
	return message.String()
}

func marshalArchitectureBaseline(baseline architectureBaseline) ([]byte, error) {
	raw, err := json.MarshalIndent(baseline, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal architecture baseline: %w", err)
	}
	return append(raw, '\n'), nil
}

func indexFindings(findings []architectureFinding) map[string][]architectureFinding {
	index := make(map[string][]architectureFinding)
	for _, finding := range findings {
		key := finding.Rule + "\x00" + finding.Path
		index[key] = append(index[key], finding)
	}
	return index
}

func sortFindings(findings []architectureFinding) {
	sort.Slice(findings, func(i, j int) bool {
		left, right := findings[i], findings[j]
		if left.Rule != right.Rule {
			return left.Rule < right.Rule
		}
		if left.Path != right.Path {
			return left.Path < right.Path
		}
		if left.Line != right.Line {
			return left.Line < right.Line
		}
		return left.Detail < right.Detail
	})
}

func unionSortedKeys[V comparable](left, right map[string]V) []string {
	keys := make(map[string]struct{}, len(left)+len(right))
	for key := range left {
		keys[key] = struct{}{}
	}
	for key := range right {
		keys[key] = struct{}{}
	}
	result := make([]string, 0, len(keys))
	for key := range keys {
		result = append(result, key)
	}
	sort.Strings(result)
	return result
}

func isGeneratedSource(source []byte) bool {
	prefix := source
	if len(prefix) > 2048 {
		prefix = prefix[:2048]
	}
	return bytes.Contains(prefix, []byte("Code generated")) && bytes.Contains(prefix, []byte("DO NOT EDIT"))
}

func physicalLineCount(source []byte) int {
	if len(source) == 0 {
		return 0
	}
	lines := bytes.Count(source, []byte{'\n'})
	if source[len(source)-1] != '\n' {
		lines++
	}
	return lines
}

func isApprovedSQLPath(path string) bool {
	parts := strings.Split(filepath.ToSlash(path), "/")
	if len(parts) >= 2 && parts[0] == "internal" && parts[1] == "migrations" {
		return true
	}
	if len(parts) >= 4 && parts[0] == "internal" && parts[1] == "modules" && parts[3] == "repository" {
		return true
	}
	if len(parts) >= 4 && parts[0] == "internal" && parts[1] == "platform" && parts[3] == "repository" {
		// Shared platform capabilities own persistence behind the same explicit
		// repository boundary as product modules.
		return true
	}
	if filepath.ToSlash(path) == "internal/testkit/postgres.go" {
		// The hermetic PostgreSQL harness owns database lifecycle DDL; it is test
		// infrastructure rather than application persistence.
		return true
	}
	if strings.HasPrefix(filepath.ToSlash(path), "internal/tools/migrationstate/") ||
		strings.HasPrefix(filepath.ToSlash(path), "internal/tools/sqlccontract/queries/") {
		// These pinned build tools validate migration/sqlc infrastructure. They do
		// not execute application persistence paths at runtime.
		return true
	}
	return strings.HasPrefix(filepath.ToSlash(path), "internal/platform/database/")
}

func looksLikeSQL(value string) bool {
	value = strings.TrimSpace(stripLeadingSQLComments(value))
	if value == "" {
		return false
	}
	for _, pattern := range applicationSQLPatterns {
		if pattern.MatchString(value) {
			return true
		}
	}
	return false
}

func stripLeadingSQLComments(value string) string {
	for {
		value = strings.TrimSpace(value)
		switch {
		case strings.HasPrefix(value, "--"):
			lineEnd := strings.IndexByte(value, '\n')
			if lineEnd < 0 {
				return ""
			}
			value = value[lineEnd+1:]
		case strings.HasPrefix(value, "/*"):
			commentEnd := strings.Index(value[2:], "*/")
			if commentEnd < 0 {
				return ""
			}
			value = value[commentEnd+4:]
		default:
			return value
		}
	}
}

func sqlPreview(value string) string {
	fields := strings.Fields(stripLeadingSQLComments(value))
	preview := strings.Join(fields, " ")
	if len(preview) > 72 {
		preview = preview[:72] + "..."
	}
	return preview
}

func firstSQLLine(source []byte) int {
	for index, line := range bytes.Split(source, []byte{'\n'}) {
		if looksLikeSQL(string(line)) {
			return index + 1
		}
	}
	return 1
}

func isGeneratedSQLCImport(importPath string) bool {
	parts := strings.Split(strings.TrimPrefix(importPath, modulePrefix), "/")
	return len(parts) == 4 &&
		(parts[0] == "modules" || parts[0] == "platform") &&
		parts[2] == "repository" &&
		parts[3] == "sqlc"
}

func isOwningRepositoryImport(path, importPath string) bool {
	parts := strings.Split(strings.TrimPrefix(importPath, modulePrefix), "/")
	if len(parts) < 4 || (parts[0] != "modules" && parts[0] != "platform") {
		return false
	}
	return strings.HasPrefix(
		filepath.ToSlash(path),
		"internal/"+parts[0]+"/"+parts[1]+"/repository/",
	)
}

func isRepositoryOrServicePath(path string) bool {
	parts := strings.Split(filepath.ToSlash(path), "/")
	return len(parts) >= 4 &&
		parts[0] == "internal" &&
		(parts[1] == "modules" || parts[1] == "platform") &&
		(parts[3] == "repository" || parts[3] == "service")
}

func isHTTPDependency(importPath string) bool {
	if importPath == projectImportPrefix+"pkg/web" {
		return true
	}
	if !strings.HasPrefix(importPath, modulePrefix) {
		return false
	}
	return pathHasSegment(strings.TrimPrefix(importPath, modulePrefix), "http")
}

func isCrossModuleHTTPImport(path, importPath string) bool {
	sourceModule := moduleFromRelativePath(path)
	targetParts := strings.Split(strings.TrimPrefix(importPath, modulePrefix), "/")
	return sourceModule != "" && len(targetParts) >= 3 && targetParts[0] == "modules" && targetParts[2] == "http" && targetParts[1] != sourceModule
}

func moduleFromRelativePath(path string) string {
	parts := strings.Split(filepath.ToSlash(path), "/")
	if len(parts) >= 3 && parts[0] == "internal" && parts[1] == "modules" {
		return parts[2]
	}
	return ""
}

func pathHasSegment(path, segment string) bool {
	for _, part := range strings.Split(filepath.ToSlash(path), "/") {
		if part == segment {
			return true
		}
	}
	return false
}

func isRouteConfigFile(path string) bool {
	return strings.HasSuffix(path, ".go") && pathHasSegment(path, "http")
}

func isRouteConfigTypeName(name string) bool {
	return name == "Config" || strings.HasSuffix(name, "RouteConfig") || strings.HasSuffix(name, "RoutesConfig")
}

func architectureScope(path string) string {
	parts := strings.Split(filepath.ToSlash(path), "/")
	if len(parts) >= 3 && parts[0] == "internal" && parts[1] == "modules" {
		return "module:" + parts[2]
	}
	if len(parts) >= 3 && parts[0] == "internal" {
		return "internal:" + parts[1]
	}
	if len(parts) >= 2 && parts[0] == "pkg" {
		return "pkg:" + parts[1]
	}
	if len(parts) >= 2 && parts[0] == "cmd" {
		return "cmd:" + parts[1]
	}
	return "server-root"
}

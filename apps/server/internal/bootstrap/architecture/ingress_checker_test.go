package architecture_test

import (
	"fmt"
	"go/ast"
	"go/token"
	"path/filepath"
	"strings"
)

const (
	standardHTTPImportPath = "net/http"
	standardIOImportPath   = "io"
	legacyIOImportPath     = "io/ioutil"
	standardJSONImportPath = "encoding/json"
	webImportPath          = projectImportPrefix + "pkg/web"
	authImportPath         = projectImportPrefix + "internal/platform/auth"
	middlewareImportPath   = projectImportPrefix + "internal/platform/http/middleware"
)

func inspectRepositoryAuthContextReads(
	fset *token.FileSet,
	path string,
	file *ast.File,
	imports map[string]string,
) []architectureFinding {
	source, ok := packageLayerFromRelativePath(path)
	if !ok || source.Layer != "repository" || strings.HasSuffix(path, "_test.go") {
		return nil
	}

	findings := make([]architectureFinding, 0)
	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		importPath, function := authContextCall(call.Fun, imports)
		if !isAuthContextReader(importPath, function) {
			return true
		}
		findings = append(findings, architectureFinding{
			Rule: ruleRepositoryAuthContextRead, Path: path, Line: fset.Position(call.Pos()).Line,
			Detail: fmt.Sprintf("repository reads request identity with %s.%s", importPath, function),
		})
		return true
	})
	return findings
}

func authContextCall(expression ast.Expr, imports map[string]string) (string, string) {
	switch function := expression.(type) {
	case *ast.SelectorExpr:
		alias, ok := function.X.(*ast.Ident)
		if !ok {
			return "", ""
		}
		return imports[alias.Name], function.Sel.Name
	case *ast.Ident:
		return imports["."], function.Name
	default:
		return "", ""
	}
}

func isAuthContextReader(importPath, function string) bool {
	switch importPath {
	case authImportPath:
		return function == "GetActor" || function == "GetUserID"
	case middlewareImportPath:
		return function == "GetActor" || function == "GetUserID" || function == "GetWorkspace"
	default:
		return false
	}
}

func inspectUnsafeRawRequestBodyReads(
	fset *token.FileSet,
	path string,
	file *ast.File,
	imports map[string]string,
) []architectureFinding {
	if strings.HasSuffix(path, "_test.go") || !isModuleHTTPPath(path) {
		return nil
	}

	findings := make([]architectureFinding, 0)
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Body == nil {
			continue
		}
		requestNames := httpRequestParameterNames(function, imports)
		if len(requestNames) == 0 {
			continue
		}
		bodyKinds := requestBodyBindings(function.Body, requestNames, imports)

		ast.Inspect(function.Body, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			importPath, functionName, ok := importedFunction(call.Fun, imports)
			if !ok {
				return true
			}

			switch {
			case importPath == standardIOImportPath && functionName == "LimitReader" &&
				len(call.Args) > 0 && requestBodyKindOf(call.Args[0], requestNames, bodyKinds, imports) == requestBodyRaw:
				findings = append(findings, architectureFinding{
					Rule: ruleUnsafeRawRequestBodyRead, Path: path, Line: fset.Position(call.Pos()).Line,
					Detail: "io.LimitReader truncates an HTTP request body instead of rejecting overflow",
				})
			case (importPath == standardIOImportPath || importPath == legacyIOImportPath) &&
				functionName == "ReadAll" && len(call.Args) > 0 &&
				requestBodyKindOf(call.Args[0], requestNames, bodyKinds, imports) == requestBodyRaw:
				findings = append(findings, architectureFinding{
					Rule: ruleUnsafeRawRequestBodyRead, Path: path, Line: fset.Position(call.Pos()).Line,
					Detail: importPath + ".ReadAll reads an HTTP request body without a rejecting byte limit",
				})
			case importPath == standardIOImportPath && (functionName == "Copy" || functionName == "CopyBuffer") &&
				len(call.Args) > 1 && requestBodyKindOf(call.Args[1], requestNames, bodyKinds, imports) == requestBodyRaw:
				findings = append(findings, architectureFinding{
					Rule: ruleUnsafeRawRequestBodyRead, Path: path, Line: fset.Position(call.Pos()).Line,
					Detail: "io." + functionName + " streams an HTTP request body without a rejecting byte limit",
				})
			case importPath == standardJSONImportPath && functionName == "NewDecoder" && len(call.Args) > 0 &&
				requestBodyKindOf(call.Args[0], requestNames, bodyKinds, imports) == requestBodyRaw:
				findings = append(findings, architectureFinding{
					Rule: ruleUnsafeRawRequestBodyRead, Path: path, Line: fset.Position(call.Pos()).Line,
					Detail: "json.NewDecoder consumes an HTTP request body without a rejecting byte limit",
				})
			}
			return true
		})
	}
	return findings
}

func isModuleHTTPPath(path string) bool {
	parts := strings.Split(filepath.ToSlash(path), "/")
	return len(parts) >= 5 && parts[0] == "internal" && parts[1] == "modules" && parts[3] == "http"
}

func httpRequestParameterNames(function *ast.FuncDecl, imports map[string]string) map[string]struct{} {
	names := make(map[string]struct{})
	if function.Type.Params == nil {
		return names
	}
	for _, field := range function.Type.Params.List {
		if !isHTTPRequestType(field.Type, imports) {
			continue
		}
		for _, name := range field.Names {
			names[name.Name] = struct{}{}
		}
	}
	return names
}

func isHTTPRequestType(expression ast.Expr, imports map[string]string) bool {
	pointer, ok := expression.(*ast.StarExpr)
	if !ok {
		return false
	}
	selector, ok := pointer.X.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != "Request" {
		return false
	}
	alias, ok := selector.X.(*ast.Ident)
	return ok && imports[alias.Name] == standardHTTPImportPath
}

type requestBodyKind uint8

const (
	requestBodyUnknown requestBodyKind = iota
	requestBodyRaw
	requestBodyBounded
	requestBodyTruncated
)

func requestBodyBindings(
	body *ast.BlockStmt,
	requestNames map[string]struct{},
	imports map[string]string,
) map[string]requestBodyKind {
	bindings := make(map[string]requestBodyKind)
	ast.Inspect(body, func(node ast.Node) bool {
		switch statement := node.(type) {
		case *ast.AssignStmt:
			for index, left := range statement.Lhs {
				name, ok := left.(*ast.Ident)
				if !ok || len(statement.Rhs) == 0 {
					continue
				}
				rightIndex := index
				if len(statement.Rhs) == 1 {
					rightIndex = 0
				}
				if rightIndex >= len(statement.Rhs) {
					continue
				}
				if kind := requestBodyKindOf(statement.Rhs[rightIndex], requestNames, bindings, imports); kind != requestBodyUnknown {
					bindings[name.Name] = kind
				}
			}
		case *ast.ValueSpec:
			for index, name := range statement.Names {
				if len(statement.Values) == 0 {
					continue
				}
				valueIndex := index
				if len(statement.Values) == 1 {
					valueIndex = 0
				}
				if valueIndex >= len(statement.Values) {
					continue
				}
				if kind := requestBodyKindOf(statement.Values[valueIndex], requestNames, bindings, imports); kind != requestBodyUnknown {
					bindings[name.Name] = kind
				}
			}
		}
		return true
	})
	return bindings
}

func requestBodyKindOf(
	expression ast.Expr,
	requestNames map[string]struct{},
	bindings map[string]requestBodyKind,
	imports map[string]string,
) requestBodyKind {
	if isRequestBodyExpression(expression, requestNames) {
		return requestBodyRaw
	}
	if identifier, ok := expression.(*ast.Ident); ok {
		return bindings[identifier.Name]
	}
	call, ok := expression.(*ast.CallExpr)
	if !ok {
		return requestBodyUnknown
	}
	importPath, function, ok := importedFunction(call.Fun, imports)
	if !ok {
		return requestBodyUnknown
	}
	switch {
	case importPath == standardHTTPImportPath && function == "MaxBytesReader" && len(call.Args) >= 2 &&
		requestBodyKindOf(call.Args[1], requestNames, bindings, imports) == requestBodyRaw:
		return requestBodyBounded
	case importPath == webImportPath && function == "ReadBoundedBody" && len(call.Args) >= 2:
		request, ok := call.Args[1].(*ast.Ident)
		if ok {
			if _, exists := requestNames[request.Name]; exists {
				return requestBodyBounded
			}
		}
	case importPath == standardIOImportPath && function == "LimitReader" && len(call.Args) > 0 &&
		requestBodyKindOf(call.Args[0], requestNames, bindings, imports) == requestBodyRaw:
		return requestBodyTruncated
	}
	return requestBodyUnknown
}

func isRequestBodyExpression(expression ast.Expr, requestNames map[string]struct{}) bool {
	selector, ok := expression.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != "Body" {
		return false
	}
	request, ok := selector.X.(*ast.Ident)
	if !ok {
		return false
	}
	_, exists := requestNames[request.Name]
	return exists
}

func importedFunction(expression ast.Expr, imports map[string]string) (string, string, bool) {
	selector, ok := expression.(*ast.SelectorExpr)
	if !ok {
		return "", "", false
	}
	alias, ok := selector.X.(*ast.Ident)
	if !ok {
		return "", "", false
	}
	importPath, ok := imports[alias.Name]
	return importPath, selector.Sel.Name, ok
}

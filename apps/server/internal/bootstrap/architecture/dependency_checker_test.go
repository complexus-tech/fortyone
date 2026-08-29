package architecture_test

import (
	"fmt"
	"go/parser"
	"go/token"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type packageLayer struct {
	Kind  string
	Owner string
	Layer string
}

type moduleImportEdge struct {
	Source     string
	Target     string
	Path       string
	Line       int
	ImportPath string
}

func inspectLayerImport(path, importPath string, line int) []architectureFinding {
	// Test packages may deliberately import concrete collaborators to build
	// fixtures. They do not create a production dependency edge and therefore
	// are outside the module-boundary ratchet.
	if strings.HasSuffix(path, "_test.go") {
		return nil
	}
	source, ok := packageLayerFromRelativePath(path)
	if !ok {
		return nil
	}
	findings := make([]architectureFinding, 0, 2)
	if (source.Layer == "service" || source.Layer == "domain") && isGeneratedOpenAPIImport(importPath) {
		findings = append(findings, architectureFinding{
			Rule: ruleGeneratedOpenAPILeak, Path: path, Line: line,
			Detail: "business layer imports generated OpenAPI package " + importPath,
		})
	}

	target, ok := packageLayerFromImportPath(importPath)
	if !ok {
		return findings
	}

	switch {
	case source.Layer == "repository" && target.Layer == "service":
		findings = append(findings, architectureFinding{
			Rule: ruleRepositoryServiceImport, Path: path, Line: line,
			Detail: "repository imports concrete service package " + importPath,
		})
	case source.Layer == "repository" && target.Layer == "repository":
		if !isOwningGeneratedSQLCImport(path, importPath) {
			findings = append(findings, architectureFinding{
				Rule: ruleRepositoryImport, Path: path, Line: line,
				Detail: "repository imports concrete repository package " + importPath,
			})
		}
	case source.Layer == "service" && target.Layer == "repository":
		findings = append(findings, architectureFinding{
			Rule: ruleServiceRepositoryImport, Path: path, Line: line,
			Detail: "service imports concrete repository adapter " + importPath,
		})
	}

	if source.Kind == "modules" && source.Layer == "service" &&
		target.Kind == "modules" && target.Layer == "service" && source.Owner != target.Owner {
		findings = append(findings, architectureFinding{
			Rule: ruleCrossModuleServiceImport, Path: path, Line: line,
			Detail: fmt.Sprintf("module %s service imports module %s concrete service package %s", source.Owner, target.Owner, importPath),
		})
	}

	return findings
}

func packageLayerFromRelativePath(path string) (packageLayer, bool) {
	parts := strings.Split(filepath.ToSlash(path), "/")
	if len(parts) < 4 || parts[0] != "internal" || (parts[1] != "modules" && parts[1] != "platform") {
		return packageLayer{}, false
	}
	if len(parts) == 4 {
		return packageLayer{Kind: parts[1], Owner: parts[2], Layer: "domain"}, true
	}
	return packageLayer{Kind: parts[1], Owner: parts[2], Layer: parts[3]}, true
}

func packageLayerFromImportPath(importPath string) (packageLayer, bool) {
	if !strings.HasPrefix(importPath, projectImportPrefix+"internal/") {
		return packageLayer{}, false
	}
	parts := strings.Split(strings.TrimPrefix(importPath, projectImportPrefix+"internal/"), "/")
	if len(parts) < 2 || (parts[0] != "modules" && parts[0] != "platform") {
		return packageLayer{}, false
	}
	if len(parts) == 2 {
		return packageLayer{Kind: parts[0], Owner: parts[1], Layer: "domain"}, true
	}
	return packageLayer{Kind: parts[0], Owner: parts[1], Layer: parts[2]}, true
}

func isOwningGeneratedSQLCImport(path, importPath string) bool {
	if !isGeneratedSQLCImport(importPath) || !isOwningRepositoryImport(path, importPath) {
		return false
	}
	source, sourceOK := packageLayerFromRelativePath(path)
	target, targetOK := packageLayerFromImportPath(importPath)
	return sourceOK && targetOK && source.Kind == target.Kind && source.Owner == target.Owner
}

func isGeneratedOpenAPIImport(importPath string) bool {
	if !strings.HasPrefix(importPath, projectImportPrefix) {
		return false
	}
	parts := strings.Split(strings.TrimPrefix(importPath, projectImportPrefix), "/")
	seenOpenAPI := false
	for _, part := range parts {
		switch {
		case strings.EqualFold(part, "openapi"):
			seenOpenAPI = true
		case seenOpenAPI && (part == "generated" || part == "gen"):
			return true
		}
	}
	return false
}

func inspectModuleImportEdges(fset *token.FileSet, path string, source []byte) ([]moduleImportEdge, error) {
	sourceModule := moduleFromRelativePath(path)
	if sourceModule == "" {
		return nil, nil
	}
	file, err := parser.ParseFile(fset, path, source, parser.ImportsOnly)
	if err != nil {
		return nil, fmt.Errorf("parse module imports in %s: %w", path, err)
	}

	edges := make([]moduleImportEdge, 0)
	for _, spec := range file.Imports {
		importPath, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			return nil, fmt.Errorf("unquote module import in %s: %w", path, err)
		}
		targetParts := strings.Split(strings.TrimPrefix(importPath, projectImportPrefix), "/")
		if len(targetParts) < 3 || targetParts[0] != "internal" || targetParts[1] != "modules" {
			continue
		}
		targetModule := targetParts[2]
		if targetModule == "" || targetModule == sourceModule {
			continue
		}
		edges = append(edges, moduleImportEdge{
			Source: sourceModule, Target: targetModule, Path: path,
			Line: fset.Position(spec.Pos()).Line, ImportPath: importPath,
		})
	}
	sortModuleEdges(edges)
	return edges, nil
}

func findModuleCycleFindings(edges []moduleImportEdge) []architectureFinding {
	if len(edges) == 0 {
		return nil
	}
	sortModuleEdges(edges)
	graph := moduleAdjacency(edges)
	components := stronglyConnectedModules(graph)
	findings := make([]architectureFinding, 0)
	for _, component := range components {
		if len(component) < 2 {
			continue
		}
		members := make(map[string]struct{}, len(component))
		for _, module := range component {
			members[module] = struct{}{}
		}
		for _, edge := range edges {
			if _, ok := members[edge.Source]; !ok {
				continue
			}
			if _, ok := members[edge.Target]; !ok {
				continue
			}
			cycle := []string{edge.Source, edge.Target}
			if path, ok := modulePath(graph, members, edge.Target, edge.Source); ok && len(path) > 1 {
				cycle = append([]string{edge.Source}, path...)
			}
			findings = append(findings, architectureFinding{
				Rule: ruleModuleDependencyCycle, Path: edge.Path, Line: edge.Line,
				Detail: fmt.Sprintf("module import %s -> %s participates in cycle %s", edge.Source, edge.Target, strings.Join(cycle, " -> ")),
			})
		}
	}
	sortFindings(findings)
	return findings
}

func moduleAdjacency(edges []moduleImportEdge) map[string][]string {
	sets := make(map[string]map[string]struct{})
	for _, edge := range edges {
		if sets[edge.Source] == nil {
			sets[edge.Source] = make(map[string]struct{})
		}
		sets[edge.Source][edge.Target] = struct{}{}
		if sets[edge.Target] == nil {
			sets[edge.Target] = make(map[string]struct{})
		}
	}
	graph := make(map[string][]string, len(sets))
	for source, targets := range sets {
		graph[source] = make([]string, 0, len(targets))
		for target := range targets {
			graph[source] = append(graph[source], target)
		}
		sort.Strings(graph[source])
	}
	return graph
}

func stronglyConnectedModules(graph map[string][]string) [][]string {
	modules := make([]string, 0, len(graph))
	for module := range graph {
		modules = append(modules, module)
	}
	sort.Strings(modules)

	index := 0
	indices := make(map[string]int, len(graph))
	lowLinks := make(map[string]int, len(graph))
	onStack := make(map[string]bool, len(graph))
	stack := make([]string, 0, len(graph))
	components := make([][]string, 0)
	var visit func(string)
	visit = func(module string) {
		indices[module] = index
		lowLinks[module] = index
		index++
		stack = append(stack, module)
		onStack[module] = true

		for _, target := range graph[module] {
			targetIndex, visited := indices[target]
			if !visited {
				visit(target)
				if lowLinks[target] < lowLinks[module] {
					lowLinks[module] = lowLinks[target]
				}
			} else if onStack[target] && targetIndex < lowLinks[module] {
				lowLinks[module] = targetIndex
			}
		}

		if lowLinks[module] != indices[module] {
			return
		}
		component := make([]string, 0)
		for {
			last := len(stack) - 1
			member := stack[last]
			stack = stack[:last]
			onStack[member] = false
			component = append(component, member)
			if member == module {
				break
			}
		}
		sort.Strings(component)
		components = append(components, component)
	}

	for _, module := range modules {
		if _, visited := indices[module]; !visited {
			visit(module)
		}
	}
	sort.Slice(components, func(i, j int) bool {
		return strings.Join(components[i], "\x00") < strings.Join(components[j], "\x00")
	})
	return components
}

func modulePath(
	graph map[string][]string,
	allowed map[string]struct{},
	start string,
	target string,
) ([]string, bool) {
	type pathNode struct {
		module string
		path   []string
	}
	queue := []pathNode{{module: start, path: []string{start}}}
	seen := map[string]struct{}{start: {}}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if current.module == target {
			return current.path, true
		}
		for _, next := range graph[current.module] {
			if _, ok := allowed[next]; !ok {
				continue
			}
			if _, ok := seen[next]; ok {
				continue
			}
			seen[next] = struct{}{}
			nextPath := append(append([]string(nil), current.path...), next)
			queue = append(queue, pathNode{module: next, path: nextPath})
		}
	}
	return nil, false
}

func sortModuleEdges(edges []moduleImportEdge) {
	sort.Slice(edges, func(i, j int) bool {
		left, right := edges[i], edges[j]
		if left.Source != right.Source {
			return left.Source < right.Source
		}
		if left.Target != right.Target {
			return left.Target < right.Target
		}
		if left.Path != right.Path {
			return left.Path < right.Path
		}
		if left.Line != right.Line {
			return left.Line < right.Line
		}
		return left.ImportPath < right.ImportPath
	})
}

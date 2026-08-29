// Command configreference generates the API/worker environment matrix from the
// Go configuration schemas and verifies that .env.example stays complete.
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
	"reflect"
	"regexp"
	"sort"
	"strings"
)

const (
	apiConfigPath       = "cmd/api/config.go"
	workerConfigPath    = "internal/bootstrap/worker/config.go"
	examplePath         = ".env.example"
	referenceOutputPath = "docs/configuration.md"
)

var exampleLinePattern = regexp.MustCompile(`^([A-Z][A-Z0-9_]*)=(.*)$`)

type binding struct {
	Name       string
	Process    string
	Field      string
	Type       string
	Default    string
	SourcePath string
	Line       int
}

type exampleValue struct {
	Value string
	Notes string
	Line  int
}

type variable struct {
	Name     string
	Bindings []binding
	Example  exampleValue
}

func main() {
	write := flag.Bool("write", false, "write the generated configuration reference")
	flag.Parse()
	if flag.NArg() != 0 {
		fatal(errors.New("configreference does not accept positional arguments"))
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
		if err := writeRootFile(root, referenceOutputPath, generated); err != nil {
			fatal(err)
		}
		return
	}

	current, err := root.ReadFile(referenceOutputPath)
	if err != nil {
		fatal(fmt.Errorf("read generated configuration reference; run make config-generate: %w", err))
	}
	if !bytes.Equal(current, generated) {
		fatal(errors.New("docs/configuration.md is stale; run make config-generate"))
	}
}

func generate(root fs.FS) ([]byte, error) {
	apiBindings, err := parseConfig(root, apiConfigPath, "API")
	if err != nil {
		return nil, err
	}
	workerBindings, err := parseConfig(root, workerConfigPath, "worker")
	if err != nil {
		return nil, err
	}
	examples, err := parseExamples(root)
	if err != nil {
		return nil, err
	}

	variables, err := mergeAndValidate(append(apiBindings, workerBindings...), examples)
	if err != nil {
		return nil, err
	}
	return renderReference(variables), nil
}

func parseConfig(root fs.FS, path string, process string) ([]binding, error) {
	source, err := fs.ReadFile(root, path)
	if err != nil {
		return nil, fmt.Errorf("read %s config: %w", process, err)
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, source, 0)
	if err != nil {
		return nil, fmt.Errorf("parse %s config: %w", process, err)
	}

	types := make(map[string]ast.Expr)
	for _, declaration := range file.Decls {
		generic, ok := declaration.(*ast.GenDecl)
		if !ok || generic.Tok != token.TYPE {
			continue
		}
		for _, rawSpec := range generic.Specs {
			spec, ok := rawSpec.(*ast.TypeSpec)
			if ok {
				types[spec.Name.Name] = spec.Type
			}
		}
	}

	configExpression, found := types["Config"]
	if !found {
		return nil, fmt.Errorf("%s does not declare Config", path)
	}
	structure, ok := configExpression.(*ast.StructType)
	if !ok {
		return nil, fmt.Errorf("%s Config must be a struct", path)
	}

	bindings := make([]binding, 0)
	if err := walkConfigStruct(fset, path, process, "Config", structure, types, map[string]bool{"Config": true}, &bindings); err != nil {
		return nil, err
	}
	sort.Slice(bindings, func(left, right int) bool { return bindings[left].Name < bindings[right].Name })
	return bindings, nil
}

func walkConfigStruct(
	fset *token.FileSet,
	path string,
	process string,
	prefix string,
	structure *ast.StructType,
	types map[string]ast.Expr,
	visiting map[string]bool,
	bindings *[]binding,
) error {
	for _, field := range structure.Fields.List {
		if len(field.Names) != 1 {
			continue
		}
		fieldName := field.Names[0].Name
		fieldPath := prefix + "." + fieldName
		tag := reflect.StructTag("")
		if field.Tag != nil {
			rawTag, err := strconvUnquote(field.Tag.Value)
			if err != nil {
				return fmt.Errorf("parse struct tag at %s:%d: %w", path, fset.Position(field.Pos()).Line, err)
			}
			tag = reflect.StructTag(rawTag)
		}

		if environmentName := strings.TrimSpace(tag.Get("env")); environmentName != "" {
			*bindings = append(*bindings, binding{
				Name:       environmentName,
				Process:    process,
				Field:      fieldPath,
				Type:       expressionString(fset, field.Type),
				Default:    tag.Get("default"),
				SourcePath: path,
				Line:       fset.Position(field.Pos()).Line,
			})
			continue
		}

		switch expression := field.Type.(type) {
		case *ast.StructType:
			if err := walkConfigStruct(fset, path, process, fieldPath, expression, types, visiting, bindings); err != nil {
				return err
			}
		case *ast.Ident:
			if visiting[expression.Name] {
				return fmt.Errorf("configuration type cycle through %s", expression.Name)
			}
			nested, ok := types[expression.Name].(*ast.StructType)
			if !ok {
				continue
			}
			visiting[expression.Name] = true
			if err := walkConfigStruct(fset, path, process, fieldPath, nested, types, visiting, bindings); err != nil {
				return err
			}
			delete(visiting, expression.Name)
		}
	}
	return nil
}

func parseExamples(root fs.FS) (map[string]exampleValue, error) {
	raw, err := fs.ReadFile(root, examplePath)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", examplePath, err)
	}
	examples := make(map[string]exampleValue)
	comments := make([]string, 0)
	for index, rawLine := range strings.Split(string(raw), "\n") {
		line := strings.TrimSpace(rawLine)
		switch {
		case strings.HasPrefix(line, "#"):
			comment := strings.TrimSpace(strings.TrimPrefix(line, "#"))
			if comment != "" {
				comments = append(comments, comment)
			}
			continue
		case line == "":
			comments = nil
			continue
		}

		match := exampleLinePattern.FindStringSubmatch(line)
		if match == nil {
			return nil, fmt.Errorf("%s:%d is not a comment, blank line, or KEY=value entry", examplePath, index+1)
		}
		if previous, exists := examples[match[1]]; exists {
			return nil, fmt.Errorf("%s defines %s at lines %d and %d", examplePath, match[1], previous.Line, index+1)
		}
		examples[match[1]] = exampleValue{
			Value: strings.TrimSpace(match[2]),
			Notes: strings.Join(comments, " "),
			Line:  index + 1,
		}
		comments = nil
	}
	return examples, nil
}

func mergeAndValidate(bindings []binding, examples map[string]exampleValue) ([]variable, error) {
	byName := make(map[string][]binding)
	perProcess := make(map[string]map[string]struct{})
	for _, item := range bindings {
		if perProcess[item.Process] == nil {
			perProcess[item.Process] = make(map[string]struct{})
		}
		if _, duplicate := perProcess[item.Process][item.Name]; duplicate {
			return nil, fmt.Errorf("%s configuration binds %s more than once", item.Process, item.Name)
		}
		perProcess[item.Process][item.Name] = struct{}{}
		byName[item.Name] = append(byName[item.Name], item)
	}

	missing := make([]string, 0)
	for name := range byName {
		if _, exists := examples[name]; !exists {
			missing = append(missing, name)
		}
	}
	unknown := make([]string, 0)
	for name := range examples {
		if _, exists := byName[name]; !exists {
			unknown = append(unknown, name)
		}
	}
	sort.Strings(missing)
	sort.Strings(unknown)
	if len(missing) > 0 || len(unknown) > 0 {
		return nil, fmt.Errorf(".env.example drift: missing=%v unknown=%v", missing, unknown)
	}
	names := make([]string, 0, len(byName))
	for name := range byName {
		names = append(names, name)
	}
	sort.Strings(names)
	variables := make([]variable, 0, len(names))
	for _, name := range names {
		items := byName[name]
		sort.Slice(items, func(left, right int) bool { return items[left].Process < items[right].Process })
		variables = append(variables, variable{Name: name, Bindings: items, Example: examples[name]})
	}
	return variables, nil
}

func renderReference(variables []variable) []byte {
	var output strings.Builder
	output.WriteString("<!-- Code generated by internal/tools/configreference; DO NOT EDIT. -->\n\n")
	output.WriteString("# API and worker configuration\n\n")
	output.WriteString("This matrix is generated from the API and worker Go configuration schemas and `.env.example`. Run `make config-generate` after changing a binding and `make config-check` to verify drift. Values are deployment configuration, not a self-hosting compatibility promise.\n\n")
	output.WriteString("A blank example means the deployment must choose a value when the enabled capability requires it. Production startup applies stronger secret and transport checks than the parser defaults. Sensitive defaults are shown only as development placeholders; their values are never copied into this document. API and worker processes can have different defaults because they are deployed with separate task environments.\n\n")

	categories := categoryOrder()
	for _, category := range categories {
		matching := make([]variable, 0)
		for _, item := range variables {
			if variableCategory(item.Name) == category {
				matching = append(matching, item)
			}
		}
		if len(matching) == 0 {
			continue
		}
		output.WriteString("## " + category + "\n\n")
		output.WriteString("| Variable | Process | Type | Default | Sensitive | Schema field | Notes |\n")
		output.WriteString("| --- | --- | --- | --- | --- | --- | --- |\n")
		for _, item := range matching {
			output.WriteString(renderVariableRow(item))
		}
		output.WriteString("\n")
	}
	return []byte(strings.TrimRight(output.String(), "\n") + "\n")
}

func renderVariableRow(item variable) string {
	processes := make([]string, 0, len(item.Bindings))
	types := make([]string, 0, len(item.Bindings))
	defaults := make([]string, 0, len(item.Bindings))
	fields := make([]string, 0, len(item.Bindings))
	sensitive := isSensitive(item.Name)
	for _, itemBinding := range item.Bindings {
		processes = appendUnique(processes, itemBinding.Process)
		types = appendUnique(types, "`"+itemBinding.Type+"`")
		defaultValue := "—"
		if itemBinding.Default != "" {
			if sensitive {
				defaultValue = "development placeholder"
			} else {
				defaultValue = "`" + escapeTable(itemBinding.Default) + "`"
			}
		}
		if len(item.Bindings) > 1 {
			defaultValue = itemBinding.Process + ": " + defaultValue
		}
		defaults = appendUnique(defaults, defaultValue)
		fields = append(fields, fmt.Sprintf("[%s %s](../%s#L%d)", itemBinding.Process, itemBinding.Field, itemBinding.SourcePath, itemBinding.Line))
	}

	return fmt.Sprintf(
		"| `%s` | %s | %s | %s | %s | %s | %s |\n",
		item.Name,
		strings.Join(processes, ", "),
		strings.Join(types, " / "),
		strings.Join(defaults, "<br>"),
		yesNo(sensitive),
		strings.Join(fields, "<br>"),
		escapeTable(item.Example.Notes),
	)
}

func categoryOrder() []string {
	return []string{
		"Runtime and lifecycle",
		"Authentication and security",
		"Database",
		"Redis and queues",
		"Email",
		"AI",
		"Provider integrations",
		"Storage",
		"Observability",
		"Other",
	}
}

func variableCategory(name string) string {
	switch {
	case name == "APP_ENVIRONMENT", strings.HasPrefix(name, "APP_API_"), strings.HasPrefix(name, "APP_WORKER_"):
		return "Runtime and lifecycle"
	case strings.HasPrefix(name, "APP_AUTH_"), strings.Contains(name, "TOKEN_HMAC"), name == "FEEDBACK_INGRESS_SECRET":
		return "Authentication and security"
	case strings.HasPrefix(name, "APP_DB_"):
		return "Database"
	case strings.HasPrefix(name, "APP_REDIS_"), strings.Contains(name, "QUEUES"):
		return "Redis and queues"
	case strings.HasPrefix(name, "APP_EMAIL_"), strings.HasPrefix(name, "APP_BREVO_"):
		return "Email"
	case strings.HasPrefix(name, "OPENAI_"):
		return "AI"
	case strings.HasPrefix(name, "GITHUB_"), strings.HasPrefix(name, "SLACK_"), strings.HasPrefix(name, "FIGMA_"), name == "APP_GITHUB_APP_ID", name == "GOOGLE_CLIENT_ID", strings.HasPrefix(name, "STRIPE_"):
		return "Provider integrations"
	case strings.HasPrefix(name, "APP_STORAGE_"), strings.HasPrefix(name, "STORAGE_"), strings.HasPrefix(name, "APP_AWS_"), strings.HasPrefix(name, "APP_AZURE_"):
		return "Storage"
	case strings.HasPrefix(name, "APP_TRACING_"):
		return "Observability"
	default:
		return "Other"
	}
}

func isSensitive(name string) bool {
	upper := strings.ToUpper(name)
	if strings.HasSuffix(upper, "_HMAC_KEY_ID") || strings.HasSuffix(upper, "_TOKENS_PER_DAY") {
		return false
	}
	return strings.Contains(upper, "SECRET") ||
		strings.Contains(upper, "PASSWORD") ||
		strings.Contains(upper, "PRIVATE_KEY") ||
		strings.Contains(upper, "API_KEY") ||
		strings.Contains(upper, "ACCESS_KEY") ||
		strings.Contains(upper, "ACCOUNT_KEY") ||
		strings.Contains(upper, "HMAC_KEY") ||
		strings.Contains(upper, "TOKEN") ||
		strings.Contains(upper, "CONNECTION_STRING") ||
		name == "APP_TRACING_HEADERS"
}

func expressionString(fset *token.FileSet, expression ast.Expr) string {
	var output bytes.Buffer
	if err := printer.Fprint(&output, fset, expression); err != nil {
		return "unknown"
	}
	return output.String()
}

func strconvUnquote(value string) (string, error) {
	if len(value) < 2 || value[0] != '`' || value[len(value)-1] != '`' {
		return "", errors.New("configuration tags must use raw string literals")
	}
	return value[1 : len(value)-1], nil
}

func appendUnique(values []string, candidate string) []string {
	for _, value := range values {
		if value == candidate {
			return values
		}
	}
	return append(values, candidate)
}

func escapeTable(value string) string {
	value = strings.ReplaceAll(strings.TrimSpace(value), "|", "\\|")
	value = strings.ReplaceAll(value, "\n", " ")
	return value
}

func yesNo(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}

func writeRootFile(root *os.Root, path string, data []byte) error {
	file, err := root.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("open %s: %w", path, err)
	}
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		return fmt.Errorf("write %s: %w", path, err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close %s: %w", path, err)
	}
	return nil
}

func fatal(err error) {
	_, _ = fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

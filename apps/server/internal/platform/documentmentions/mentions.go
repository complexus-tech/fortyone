// Package documentmentions recognizes identity-bearing Markdown links without
// turning quoted examples, code, or raw HTML into notification instructions.
package documentmentions

import (
	"regexp"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

const MaximumRecipients = 10

var tokenPattern = regexp.MustCompile(`\[@(?:\\.|[^\]\\\r\n])+\]\((mention:[^\s)]+|agent:maya)\)`)
var labelEscaper = strings.NewReplacer(`\`, `\\`, `[`, `\[`, `]`, `\]`, `(`, `\(`, `)`, `\)`, "\n", " ", "\r", " ")

type token struct {
	start, end    int
	target, label string
}
type Result struct {
	Body    string
	UserIDs []uuid.UUID
	Maya    bool
}

func tokens(body string) []token {
	source := []byte(body)
	root := goldmark.DefaultParser().Parse(text.NewReader(source))
	allowed := map[int]bool{}
	_ = ast.Walk(root, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch node.Kind() {
		case ast.KindBlockquote, ast.KindCodeBlock, ast.KindFencedCodeBlock, ast.KindCodeSpan, ast.KindHTMLBlock:
			return ast.WalkSkipChildren, nil
		case ast.KindParagraph:
			for child := node.FirstChild(); child != nil; child = child.NextSibling() {
				if child.Kind() == ast.KindRawHTML {
					return ast.WalkSkipChildren, nil
				}
			}
		}
		if link, ok := node.(*ast.Link); ok {
			if first, ok := link.FirstChild().(*ast.Text); ok {
				allowed[first.Segment.Start-1] = true
			}
		}
		return ast.WalkContinue, nil
	})
	var result []token
	for _, match := range tokenPattern.FindAllStringSubmatchIndex(body, -1) {
		if !allowed[match[0]] {
			continue
		}
		value := body[match[0]:match[1]]
		split := strings.LastIndex(value, "](")
		result = append(result, token{match[0], match[1], body[match[2]:match[3]], value[2:split]})
	}
	return result
}

func Candidates(body string) Result {
	result := Result{Body: body}
	for _, token := range tokens(body) {
		if token.target == "agent:maya" {
			result.Maya = true
			continue
		}
		raw := strings.TrimPrefix(token.target, "mention:")
		id, err := uuid.Parse(raw)
		if err == nil && id != uuid.Nil && id.String() == raw && !slices.Contains(result.UserIDs, id) && len(result.UserIDs) < MaximumRecipients {
			result.UserIDs = append(result.UserIDs, id)
		}
	}
	return result
}

func Canonicalize(body string, names map[uuid.UUID]string) Result {
	result := Result{}
	var output strings.Builder
	position := 0
	for _, token := range tokens(body) {
		output.WriteString(body[position:token.start])
		position = token.end
		if token.target == "agent:maya" {
			output.WriteString("[@Maya](agent:maya)")
			result.Maya = true
			continue
		}
		raw := strings.TrimPrefix(token.target, "mention:")
		id, err := uuid.Parse(raw)
		name, allowed := names[id]
		seen := slices.Contains(result.UserIDs, id)
		if err != nil || id == uuid.Nil || id.String() != raw || !allowed || strings.TrimSpace(name) == "" || (!seen && len(result.UserIDs) >= MaximumRecipients) {
			output.WriteString("@" + labelEscaper.Replace(token.label))
			continue
		}
		output.WriteString("[@" + labelEscaper.Replace(name) + "](mention:" + id.String() + ")")
		if !seen {
			result.UserIDs = append(result.UserIDs, id)
		}
	}
	output.WriteString(body[position:])
	result.Body = output.String()
	return result
}

package agentadapter

import (
	"bytes"
	stdhtml "html"
	"net/url"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	markdownhtml "github.com/yuin/goldmark/renderer/html"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var commentMarkdown = goldmark.New(
	goldmark.WithExtensions(extension.Strikethrough, extension.Linkify, extension.Table),
	goldmark.WithRendererOptions(markdownhtml.WithHardWraps()),
)

// replyHTML accepts Markdown, never trusted HTML. Raw HTML stays disabled in
// the parser, then a second allowlist excludes media, mentions and attributes
// that could bypass the normal attachment/mention write paths.
func replyHTML(text string) string {
	text = strings.TrimSpace(text)
	if runes := []rune(text); len(runes) > 6000 {
		text = string(runes[:6000]) + "…"
	}
	for {
		output := renderReplyMarkdown(text)
		if utf8.RuneCountInString(output) <= 9500 {
			return output
		}
		// Reparse a shorter source so the persisted markup remains balanced.
		runes := []rune(text)
		text = string(runes[:len(runes)/2]) + "…"
	}
}

func renderReplyMarkdown(text string) string {
	var rendered bytes.Buffer
	if err := commentMarkdown.Convert([]byte(text), &rendered); err != nil {
		return plainReplyHTML(text)
	}
	nodes, err := html.ParseFragment(&rendered, &html.Node{Type: html.ElementNode, Data: "div", DataAtom: atom.Div})
	if err != nil {
		return plainReplyHTML(text)
	}
	var output strings.Builder
	for _, node := range nodes {
		writeReplyNode(&output, node)
	}
	if value := strings.TrimSpace(output.String()); value != "" && value != "<p></p>" {
		return value
	}
	// An answer containing only omitted raw HTML is displayed literally rather
	// than publishing an empty comment or trusting the model's markup.
	return plainReplyHTML(text)
}

func plainReplyHTML(text string) string {
	return "<p>" + strings.ReplaceAll(stdhtml.EscapeString(text), "\n", "<br>") + "</p>"
}

func writeReplyNode(output *strings.Builder, node *html.Node) {
	if node.Type == html.TextNode {
		output.WriteString(stdhtml.EscapeString(node.Data))
		return
	}
	if node.Type != html.ElementNode {
		return
	}
	name := node.Data
	switch name {
	case "script", "style", "iframe", "object", "embed", "svg", "math", "template", "form", "input", "button", "video", "audio", "source", "picture":
		return
	case "img":
		alt := replyAttribute(node, "alt")
		if alt == "" {
			alt = "[Image omitted]"
		}
		output.WriteString(stdhtml.EscapeString(alt))
		return
	case "p", "br", "strong", "em", "del", "h1", "h2", "h3", "h4", "h5", "h6", "ul", "ol", "li", "blockquote", "code", "pre", "hr", "table", "thead", "tbody", "tr", "th", "td", "a":
	default:
		writeReplyChildren(output, node)
		return
	}
	if name == "a" && !safeReplyLink(replyAttribute(node, "href")) {
		writeReplyChildren(output, node)
		return
	}
	output.WriteByte('<')
	output.WriteString(name)
	if name == "a" {
		output.WriteString(` href="` + stdhtml.EscapeString(replyAttribute(node, "href")) + `" rel="noopener noreferrer"`)
	}
	if name == "ol" {
		if start, err := strconv.ParseUint(replyAttribute(node, "start"), 10, 32); err == nil && start > 0 {
			output.WriteString(` start="` + strconv.FormatUint(start, 10) + `"`)
		}
	}
	output.WriteByte('>')
	if name == "br" || name == "hr" {
		return
	}
	writeReplyChildren(output, node)
	output.WriteString("</" + name + ">")
}

func writeReplyChildren(output *strings.Builder, node *html.Node) {
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		writeReplyNode(output, child)
	}
}

func replyAttribute(node *html.Node, key string) string {
	for _, attribute := range node.Attr {
		if attribute.Namespace == "" && attribute.Key == key {
			return attribute.Val
		}
	}
	return ""
}

func safeReplyLink(value string) bool {
	if value == "" || strings.HasPrefix(value, "//") || strings.Contains(value, `\`) {
		return false
	}
	for _, character := range value {
		if unicode.IsControl(character) || unicode.IsSpace(character) {
			return false
		}
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.User != nil {
		return false
	}
	switch strings.ToLower(parsed.Scheme) {
	case "":
		return parsed.Host == ""
	case "http", "https":
		return parsed.Hostname() != ""
	case "mailto":
		return parsed.Opaque != "" && parsed.Host == ""
	default:
		return false
	}
}

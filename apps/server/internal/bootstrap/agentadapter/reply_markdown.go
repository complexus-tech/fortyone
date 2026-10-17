package agentadapter

import (
	"bytes"
	stdhtml "html"
	"net/url"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"
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
// the parser, then a second allowlist excludes media and untrusted attributes.
// Only server-resolved member links can enter the separate mention write path.
func replyHTML(text string) string {
	return renderMentionReply(text, nil).HTML
}

func boundedReplyText(text string) string {
	// Leading indentation is Markdown code, where an example must never ping.
	text = strings.Trim(text, "\r\n")
	if runes := []rune(text); len(runes) > 6000 {
		text = string(runes[:6000]) + "…"
	}
	return text
}

func renderReplyMarkdown(text string) string {
	return renderReplyMarkdownWithMentions(text, &replyMentions{})
}

func renderMentionReply(text string, targets map[uuid.UUID]string) renderedReply {
	text = boundedReplyText(text)
	for {
		mentions := &replyMentions{targets: targets}
		output := renderReplyMarkdownWithMentions(text, mentions)
		if utf8.RuneCountInString(output) <= 9500 {
			return renderedReply{HTML: output, MentionedUserIDs: mentions.ids}
		}
		// Reparse both markup and recipients together: omitted text cannot notify.
		runes := []rune(text)
		text = string(runes[:len(runes)/2]) + "…"
	}
}

func renderReplyMarkdownWithMentions(text string, mentions *replyMentions) string {
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
		writeReplyNode(&output, node, mentions)
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

func writeReplyNode(output *strings.Builder, node *html.Node, mentions *replyMentions) {
	if node.Type == html.TextNode {
		output.WriteString(stdhtml.EscapeString(node.Data))
		return
	}
	if node.Type != html.ElementNode {
		return
	}
	name := node.Data
	if name == "blockquote" || name == "pre" || name == "code" {
		mentions.blockedDepth++
		defer func() { mentions.blockedDepth-- }()
	}
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
		writeReplyChildren(output, node, mentions)
		return
	}
	if name == "a" && mentions.write(output, replyAttribute(node, "href")) {
		return
	}
	if name == "a" && !safeReplyLink(replyAttribute(node, "href")) {
		writeReplyChildren(output, node, mentions)
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
	writeReplyChildren(output, node, mentions)
	output.WriteString("</" + name + ">")
}

func writeReplyChildren(output *strings.Builder, node *html.Node, mentions *replyMentions) {
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		writeReplyNode(output, child, mentions)
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

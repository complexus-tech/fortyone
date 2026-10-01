package storieshttp

import (
	htmlparser "golang.org/x/net/html"
	"html"
	"net/url"
	"strings"
)

var importedCommentElements = map[string]bool{
	"p": true, "br": true, "strong": true, "em": true, "b": true, "i": true, "u": true, "s": true, "code": true, "pre": true,
	"blockquote": true, "ul": true, "ol": true, "li": true, "a": true, "h1": true, "h2": true, "h3": true, "h4": true,
	"table": true, "thead": true, "tbody": true, "tr": true, "td": true, "th": true, "hr": true, "span": true, "div": true,
}
var droppedCommentElements = map[string]bool{"script": true, "style": true, "iframe": true, "object": true, "svg": true, "math": true, "template": true}

// Imported HTML renders in the same rich comment surface as native editor
// output. Preserve its supported content vocabulary, without executable HTML,
// synthesized mentions, CSS, or provider-controlled event attributes.
func safeImportedCommentHTML(value string) string {
	nodes, err := htmlparser.ParseFragment(strings.NewReader(value), nil)
	if err != nil {
		return html.EscapeString(value)
	}
	var output strings.Builder
	var appendNode func(*htmlparser.Node)
	appendNode = func(node *htmlparser.Node) {
		if node.Type == htmlparser.TextNode {
			output.WriteString(html.EscapeString(node.Data))
			return
		}
		if node.Type != htmlparser.ElementNode {
			for child := node.FirstChild; child != nil; child = child.NextSibling {
				appendNode(child)
			}
			return
		}
		if droppedCommentElements[node.Data] {
			return
		}
		allowed := importedCommentElements[node.Data]
		if allowed {
			output.WriteByte('<')
			output.WriteString(node.Data)
			if node.Data == "a" {
				for _, attribute := range node.Attr {
					if attribute.Key != "href" {
						continue
					}
					link, err := url.Parse(strings.TrimSpace(attribute.Val))
					if err == nil && (link.Scheme == "https" || link.Scheme == "http" || link.Scheme == "mailto") && link.User == nil {
						output.WriteString(` href="`)
						output.WriteString(html.EscapeString(link.String()))
						output.WriteString(`" rel="noopener noreferrer"`)
					}
				}
			}
			output.WriteByte('>')
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			appendNode(child)
		}
		if allowed && node.Data != "br" && node.Data != "hr" {
			output.WriteString("</")
			output.WriteString(node.Data)
			output.WriteByte('>')
		}
	}
	for _, node := range nodes {
		appendNode(node)
	}
	return output.String()
}

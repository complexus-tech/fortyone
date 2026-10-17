package chatrepository

import (
	"encoding/json"
	"net/mail"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
	"golang.org/x/net/publicsuffix"
)

const maximumMessageLinks = 50

// Match a whole explicit scheme before considering a bare domain, so an unsafe
// ftp/data/javascript address cannot accidentally yield a nested safe resource.
var linkCandidates = regexp.MustCompile(`(?i)(?:(?:[a-z][a-z0-9+.-]*://|mailto:|javascript:|data:|vbscript:)[^\s<>"\x60]+|(?:[a-z0-9.!#$%&'*+/=?^_{}|~-]+@)?(?:[\pL\pN](?:[\pL\pN-]*[\pL\pN])?\.)+(?:[\pL]{2,63}|xn--[a-z0-9-]+)(?::[0-9]+)?(?:[/?#][^\s<>"\x60]+)?)`)

type linkNode struct {
	Type    string     `json:"type"`
	Text    string     `json:"text"`
	Content []linkNode `json:"content"`
	Marks   []linkMark `json:"marks"`
}

type linkMark struct {
	Type  string `json:"type"`
	Attrs struct {
		Href string `json:"href"`
	} `json:"attrs"`
}

// Resources follow the rendered rich document when present. Falling back to
// its plain-text mirror would index code and explicit-link labels twice.
func extractLinks(plain string, rich json.RawMessage) []string {
	result := make([]string, 0)
	seen := make(map[string]bool)
	add := func(value string, inferred bool) {
		href, ok := resourceLinkHref(value, inferred)
		if !ok || seen[href] || len(result) >= maximumMessageLinks {
			return
		}
		seen[href] = true
		result = append(result, href)
	}
	addPlain := func(value string) {
		for _, candidate := range plainResourceLinks(value) {
			add(candidate, true)
		}
	}
	var document linkNode
	if json.Unmarshal(rich, &document) != nil || document.Type != "doc" {
		addPlain(plain)
		return result
	}
	var walk func(linkNode)
	walk = func(node linkNode) {
		if node.Type == "codeBlock" || len(result) >= maximumMessageLinks {
			return
		}
		if slices.ContainsFunc(node.Marks, func(mark linkMark) bool { return mark.Type == "code" }) {
			return
		}
		linked := false
		for _, mark := range node.Marks {
			if mark.Type == "link" {
				linked = true
				add(mark.Attrs.Href, false)
			}
		}
		if !linked {
			addPlain(node.Text)
		}
		for _, child := range node.Content {
			walk(child)
		}
	}
	walk(document)
	return result
}

type linkRange struct{ start, end int }

func plainResourceLinks(value string) []string {
	matches := linkCandidates.FindAllStringIndex(value, -1)
	if len(matches) == 0 {
		return nil
	}
	excluded := resourceCodeRanges([]byte(value))
	result := make([]string, 0, len(matches))
	for _, match := range matches {
		if match[0] > 0 {
			previous, _ := utf8.DecodeLastRuneInString(value[:match[0]])
			if unicode.IsLetter(previous) || unicode.IsDigit(previous) || strings.ContainsRune("_@./-", previous) {
				continue
			}
		}
		if slices.ContainsFunc(excluded, func(r linkRange) bool { return match[0] < r.end && match[1] > r.start }) {
			continue
		}
		candidate := strings.TrimRight(value[match[0]:match[1]], ".,;:!?'")
		for len(candidate) > 0 {
			closing := candidate[len(candidate)-1]
			opening := map[byte]byte{')': '(', ']': '[', '}': '{'}[closing]
			if opening == 0 || strings.Count(candidate, string(closing)) <= strings.Count(candidate, string(opening)) {
				break
			}
			candidate = strings.TrimRight(candidate[:len(candidate)-1], ".,;:!?'")
		}
		result = append(result, candidate)
	}
	return result
}

// Goldmark supplies code ranges without interpreting the surrounding prose as
// HTML or fetching URLs. Unfinished fenced blocks remain code through EOF.
func resourceCodeRanges(source []byte) []linkRange {
	var ranges []linkRange
	root := goldmark.New().Parser().Parse(text.NewReader(source))
	_ = ast.Walk(root, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch node := node.(type) {
		case *ast.CodeSpan:
			for child := node.FirstChild(); child != nil; child = child.NextSibling() {
				if part, ok := child.(*ast.Text); ok {
					ranges = append(ranges, linkRange{part.Segment.Start, part.Segment.Stop})
				}
			}
			return ast.WalkSkipChildren, nil
		case *ast.FencedCodeBlock:
			if node.Info != nil {
				ranges = append(ranges, linkRange{node.Info.Segment.Start, node.Info.Segment.Stop})
			}
		}
		if node.Kind() == ast.KindCodeBlock || node.Kind() == ast.KindFencedCodeBlock {
			for i := 0; i < node.Lines().Len(); i++ {
				line := node.Lines().At(i)
				ranges = append(ranges, linkRange{line.Start, line.Stop})
			}
			return ast.WalkSkipChildren, nil
		}
		return ast.WalkContinue, nil
	})
	return ranges
}

func resourceLinkHref(value string, inferred bool) (string, bool) {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > 2048 || strings.ContainsFunc(value, unicode.IsControl) {
		return "", false
	}
	if inferred && !strings.Contains(value, "://") && !strings.HasPrefix(strings.ToLower(value), "mailto:") {
		if strings.Contains(value, "@") {
			value = "mailto:" + value
		} else {
			parsed, err := url.Parse("https://" + value)
			if err != nil {
				return "", false
			}
			host := strings.ToLower(parsed.Hostname())
			tld := host[strings.LastIndex(host, ".")+1:]
			if _, icann := publicsuffix.PublicSuffix(tld); !icann {
				return "", false
			}
			value = "https://" + value
		}
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.User != nil {
		return "", false
	}
	parsed.Scheme = strings.ToLower(parsed.Scheme)
	switch parsed.Scheme {
	case "http", "https":
		if parsed.Hostname() == "" {
			return "", false
		}
		parsed.Host = strings.ToLower(parsed.Host)
		if (parsed.Scheme == "https" && parsed.Port() == "443") || (parsed.Scheme == "http" && parsed.Port() == "80") {
			parsed.Host = strings.TrimSuffix(parsed.Host, ":"+parsed.Port())
		}
		if parsed.Path == "" {
			parsed.Path = "/"
		}
	case "mailto":
		address, err := mail.ParseAddress(parsed.Opaque)
		if err != nil || address.Address != parsed.Opaque {
			return "", false
		}
	default:
		return "", false
	}
	return parsed.String(), true
}

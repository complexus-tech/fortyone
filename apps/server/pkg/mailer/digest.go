package mailer

import (
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"html"
	"html/template"
	"net/url"
	"strings"
	"unicode"
	"unicode/utf16"

	"golang.org/x/text/unicode/norm"
)

// DigestDetailLimit applies before copy generation as well as during rendering.
const DigestDetailLimit = 5

type EmailActor struct {
	ID        uuid.UUID
	Name      string
	AvatarURL string
}
type DigestRow struct {
	Icon  string
	Label string
	Text  string
	URL   string
	Actor EmailActor
	More  bool
}
type Digest struct {
	Intro string
	Rows  []DigestRow
}

// Keep this palette and FNV-1a seed normalization aligned with packages/lib/src/avatar-color.ts.
var avatarColors = [...][2]string{{"#E67E22", "#111827"}, {"#C0392B", "#FFFFFF"}, {"#8E44AD", "#FFFFFF"}, {"#27AE60", "#111827"}, {"#4A90E2", "#111827"}, {"#30336B", "#FFFFFF"}}

func avatarColor(name string) [2]string {
	seed := cases.Lower(language.Und).String(strings.Join(strings.FieldsFunc(norm.NFKC.String(name), func(r rune) bool { return r == '\ufeff' || (unicode.IsSpace(r) && r != '\u0085') }), " "))
	if seed == "" {
		seed = "user"
	}
	hash := uint32(0x811c9dc5)
	for _, unit := range utf16.Encode([]rune(seed)) {
		hash ^= uint32(unit)
		hash *= 0x01000193
	}
	return avatarColors[hash%uint32(len(avatarColors))]
}

func avatarInitials(name string) string {
	words := strings.Fields(name)
	if len(words) == 0 {
		return "U"
	}
	first := []rune(words[0])
	if len(words) == 1 {
		return strings.ToUpper(string(first[:min(2, len(first))]))
	}
	return strings.ToUpper(string(first[0]) + string([]rune(words[len(words)-1])[0]))
}

func safeAvatarURL(value string) string {
	parsed, err := url.Parse(strings.TrimSpace(value))
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil {
		return ""
	}
	return parsed.String()
}

// actorText is renderer-owned markup; all text and URL attributes are escaped.
// Unlike safeHTML, this function accepts typed actor data rather than model HTML.
func actorText(text string, actor EmailActor) template.HTML {
	name := strings.TrimSpace(actor.Name)
	names := []string{name}
	if parts := strings.Fields(name); len(parts) > 1 {
		names = append(names, parts[0])
	}
	for _, displayName := range names {
		if displayName == "" || !(text == displayName || strings.HasPrefix(text, displayName+" ") || strings.HasPrefix(text, displayName+":")) {
			continue
		}
		colors := avatarColor(name)
		style := fmt.Sprintf("display:inline-block;width:20px;height:20px;line-height:20px;vertical-align:text-bottom;border-radius:50%%;background-color:%s;color:%s;font-family:Inter,Arial,Helvetica,sans-serif;font-size:8px;font-weight:500;text-align:center;", colors[0], colors[1])
		initials := html.EscapeString(avatarInitials(name))
		badge := `<span aria-hidden="true" style="` + style + `">` + initials + `</span>`
		if src := safeAvatarURL(actor.AvatarURL); src != "" {
			badge = `<img src="` + html.EscapeString(src) + `" width="20" height="20" alt="` + initials + `" style="` + style + `border:0;object-fit:cover;">`
		}
		first, rest, _ := strings.Cut(displayName, " ")
		suffix := ""
		if rest != "" {
			suffix = " " + html.EscapeString(rest)
		}
		// #nosec G203 -- markup and CSS are fixed; all variable text and attributes are escaped.
		return template.HTML(`<span style="white-space:nowrap;">` + badge + `<span style="display:inline-block;width:4px;font-size:0;">&nbsp;</span>` + html.EscapeString(first) + `</span>` + suffix + html.EscapeString(text[len(displayName):]))
	}
	// #nosec G203 -- text is escaped without adding markup.
	return template.HTML(html.EscapeString(text))
}

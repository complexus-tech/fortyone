package emailreply

import (
	"html"
	"strings"

	"github.com/complexus-tech/projects-api/pkg/mailer"
)

const mayaReplyLogoURL = "https://fortyone.app/images/logo.png"

// renderMayaReplyHTML places trusted, already-escaped email-agent blocks in
// the same restrained product shell as FortyOne's other transactional email.
// Model-authored values never enter attributes, styles, or URLs.
func renderMayaReplyHTML(subject, content string) string {
	content = styleMayaReplyBlocks(content)
	var output strings.Builder
	output.WriteString(`<!doctype html><html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1.0"><title>`)
	output.WriteString(html.EscapeString(strings.TrimSpace(subject)))
	output.WriteString(`</title></head><body style="`)
	output.WriteString(mailer.EmailStyleString("body"))
	output.WriteString(`"><div style="`)
	output.WriteString(mailer.EmailStyleString("shell"))
	output.WriteString(`"><div style="`)
	output.WriteString(mailer.EmailStyleString("inner"))
	output.WriteString(`"><div style="`)
	output.WriteString(mailer.EmailStyleString("brand"))
	output.WriteString(`"><img src="`)
	output.WriteString(mayaReplyLogoURL)
	output.WriteString(`" alt="FortyOne" style="`)
	output.WriteString(mailer.EmailStyleString("logo"))
	output.WriteString(`"></div><div class="email-body" style="`)
	output.WriteString(mailer.EmailStyleString("bodyBlock"))
	output.WriteString(`">`)
	output.WriteString(content)
	output.WriteString(`</div><div style="`)
	output.WriteString(mailer.EmailStyleString("footer"))
	output.WriteString(`"><p style="`)
	output.WriteString(mailer.EmailStyleString("footerTextLast"))
	output.WriteString(`">FortyOne by Complexus LLC</p></div></div></div></body></html>`)
	return output.String()
}

// styleMayaReplyBlocks adds only constant, product-owned inline styles to the
// fixed element vocabulary emitted by emailagent.RenderHTML. Model text is
// already escaped and remains untouched.
func styleMayaReplyBlocks(content string) string {
	paragraph := `<p style="` + mailer.EmailStyleString("text") + `">`
	callout := `<div role="note" style="` + mailer.EmailStyleString("quietPanel") + `"><p style="` + mailer.EmailStyleString("textNoMargin") + `">`
	list := `<ul style="margin: 0 0 18px; padding-left: 22px; color: #3b3a38; font-family: Geist, Helvetica, Arial, sans-serif; font-size: 15px; line-height: 1.62;">`
	item := `<li style="margin: 0 0 8px;">`

	content = strings.ReplaceAll(content, `<div role="note"><p>`, callout)
	content = strings.ReplaceAll(content, `<p>`, paragraph)
	content = strings.ReplaceAll(content, `<ul>`, list)
	content = strings.ReplaceAll(content, `<li>`, item)
	return content
}

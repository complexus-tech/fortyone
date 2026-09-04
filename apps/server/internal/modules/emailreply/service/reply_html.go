package emailreply

import (
	"strings"

	"github.com/complexus-tech/projects-api/pkg/mailer"
)

// renderMayaReplyHTML uses the same layout as file-backed transactional mail.
func renderMayaReplyHTML(subject, content string) (string, error) {
	return mailer.RenderMayaReply(subject, styleMayaReplyBlocks(content))
}

// styleMayaReplyBlocks adds only constant, product-owned inline styles to the
// fixed element vocabulary emitted by emailagent.RenderHTML. Model text is
// already escaped and remains untouched.
func styleMayaReplyBlocks(content string) string {
	paragraph := `<p style="` + mailer.EmailStyleString("text") + `">`
	callout := `<div role="note" style="` + mailer.EmailStyleString("quietPanel") + `"><p style="` + mailer.EmailStyleString("textNoMargin") + `">`
	list := `<ul style="` + mailer.EmailStyleString("replyList") + `">`
	item := `<li style="` + mailer.EmailStyleString("replyItem") + `">`

	content = strings.ReplaceAll(content, `<div role="note"><p>`, callout)
	content = strings.ReplaceAll(content, `<p>`, paragraph)
	content = strings.ReplaceAll(content, `<ul>`, list)
	content = strings.ReplaceAll(content, `<li>`, item)
	return content
}

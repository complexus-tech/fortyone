package jobs

import (
	"fmt"
	"html"

	"github.com/complexus-tech/projects-api/pkg/mailer"
)

func formatCompactNotificationRows(intro string, rows []string) string {
	textStyle := mailer.EmailStyleString("notificationText")
	listStyle := mailer.EmailStyleString("notificationList")
	messageStyle := mailer.EmailStyleString("notificationMessage")

	content := fmt.Sprintf(`
		<div style="%s">
			<p style="%s">%s</p>
			<div style="%s">
	`, textStyle, textStyle, html.EscapeString(intro), listStyle)

	for index, row := range rows {
		itemStyle := mailer.EmailStyleString("notificationItem")
		if index == 0 {
			itemStyle = mailer.EmailStyleString("notificationItemFirst")
		}
		content += fmt.Sprintf(`
			<div style="%s">
				<p style="%s">%s</p>
			</div>
		`, itemStyle, messageStyle, row)
	}

	return content + "</div></div>"
}

func formatEmailStrong(value string) string {
	return fmt.Sprintf(`<strong style="%s">%s</strong>`, mailer.EmailStyleString("detailValue"), html.EscapeString(value))
}

func formatEmailLink(url string, label string) string {
	return fmt.Sprintf(`<a href="%s" style="%s">%s</a>`, html.EscapeString(url), mailer.EmailStyleString("notificationLink"), html.EscapeString(label))
}

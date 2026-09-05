package mailer

import (
	"html/template"
)

const emailFontStack = `Inter, Arial, Helvetica, sans-serif`

// Email-safe colors from the approved warm email design.
const (
	emailColorBackground        = `#fff6ef`
	emailColorForeground        = `#25150e`
	emailColorForegroundInverse = `#ffffff`
	emailColorTextSecondary     = `#25150e`
	emailColorTextMuted         = `#72645d`
	emailColorSurfaceMuted      = `#fff7f1`
	emailColorBorder            = `#f1e4dc`
	emailColorButton            = `#25150e`
)

var emailStyles = map[string]string{
	"replyList":             `margin: 0 0 12px; padding-left: 20px; color: ` + emailColorForeground + `; font-family: ` + emailFontStack + `; font-size: 15px; line-height: 21px;`,
	"replyItem":             `margin: 0 0 6px;`,
	"signature":             `margin: 16px 0 0; color: ` + emailColorForeground + `; font-family: ` + emailFontStack + `; font-size: 15px; line-height: 21px;`,
	"body":                  `margin: 0; padding: 0; background-color: ` + emailColorBackground + `; color: ` + emailColorTextSecondary + `; font-family: ` + emailFontStack + `; font-size: 15px; line-height: 21px;`,
	"shell":                 `width: 100%; max-width: 640px; margin: 0 auto; background-color: ` + emailColorBackground + `;`,
	"inner":                 `padding: 48px 48px 38px;`,
	"brand":                 `display: block; margin: 0 0 26px;`,
	"logo":                  `display: block; border: 0; width: 126px; height: 31px;`,
	"eyebrow":               `margin: 0 0 14px; color: ` + emailColorTextMuted + `; font-family: ` + emailFontStack + `; font-size: 12px; font-weight: 600; line-height: 1.4; text-transform: uppercase;`,
	"heading":               `margin: 0 0 12px; max-width: 420px; color: ` + emailColorForeground + `; font-family: ` + emailFontStack + `; font-size: 23px; font-weight: 600; line-height: 29px; letter-spacing: -0.5px;`,
	"bodyBlock":             `margin: 0;`,
	"text":                  `margin: 0 0 12px; color: ` + emailColorTextSecondary + `; font-family: ` + emailFontStack + `; font-size: 15px; line-height: 21px;`,
	"textNoMargin":          `margin: 0; color: ` + emailColorTextSecondary + `; font-family: ` + emailFontStack + `; font-size: 15px; line-height: 21px;`,
	"textSmallTop":          `margin: 12px 0 0; color: ` + emailColorTextSecondary + `; font-family: ` + emailFontStack + `; font-size: 15px; line-height: 21px;`,
	"buttonContainer":       `margin: 0;`,
	"button":                `display: block; padding: 10px 12px; border: 1px solid ` + emailColorButton + `; border-radius: 6px; background-color: ` + emailColorButton + `; color: #ffffff; font-family: ` + emailFontStack + `; font-size: 14px; font-weight: 600; line-height: 22px; text-align: center; text-decoration: none; mso-hide: all;`,
	"panel":                 `margin: 18px 0 0; padding: 0; background-color: transparent;`,
	"quietPanel":            `margin: 14px 0; padding: 16px 0; border-top: 1px solid ` + emailColorBorder + `; border-bottom: 1px solid ` + emailColorBorder + `;`,
	"notificationPanel":     `margin: 14px 0 0; padding: 0; color: ` + emailColorForeground + `; font-family: ` + emailFontStack + `; font-size: 15px; line-height: 21px;`,
	"panelLabel":            `margin: 0 0 8px; color: ` + emailColorTextMuted + `; font-family: ` + emailFontStack + `; font-size: 11px; font-weight: 600; line-height: 17px; letter-spacing: 1.6px; text-transform: uppercase;`,
	"panelLabelTight":       `margin: 0 0 4px; color: ` + emailColorTextMuted + `; font-family: ` + emailFontStack + `; font-size: 15px; font-weight: 600; line-height: 1.4;`,
	"panelValue":            `margin: 0; color: ` + emailColorForeground + `; font-family: ` + emailFontStack + `; font-size: 16px; font-weight: 400; line-height: 21px;`,
	"panelTitle":            `margin: 0 0 6px; color: ` + emailColorForeground + `; font-family: ` + emailFontStack + `; font-size: 15px; font-weight: 600; line-height: 21px;`,
	"detailList":            `display: block; margin: 0; padding: 0; list-style: none;`,
	"detailRow":             `display: block; padding: 0 0 10px;`,
	"detailLabel":           `display: inline-block; min-width: 116px; color: ` + emailColorTextMuted + `; font-family: ` + emailFontStack + `; font-size: 15px; line-height: 21px;`,
	"detailValue":           `color: ` + emailColorForeground + `; font-family: ` + emailFontStack + `; font-size: 15px; font-weight: 600; line-height: 21px;`,
	"codePanel":             `margin: 16px 0; padding: 20px; border-radius: 8px; background-color: ` + emailColorSurfaceMuted + `;`,
	"codePanelCentered":     `margin: 16px 0; padding: 20px; border-radius: 8px; background-color: ` + emailColorSurfaceMuted + `; text-align: center;`,
	"verificationCode":      `margin: 0; color: ` + emailColorForeground + `; font-family: SFMono-Regular, Consolas, monospace; font-size: 30px; font-weight: 600; line-height: 1.1;`,
	"securityNote":          `margin: 16px 0 0; color: ` + emailColorTextMuted + `; font-family: ` + emailFontStack + `; font-size: 13px; line-height: 21px;`,
	"secondaryActions":      `margin-top: 32px; margin-bottom: 42px;`,
	"secondaryActionLink":   `color: ` + emailColorTextMuted + `; font-family: ` + emailFontStack + `; font-size: 15px; line-height: 21px; text-decoration: underline; text-decoration-thickness: 1px; text-underline-offset: 3px;`,
	"workspaceLink":         `display: inline-block; overflow-wrap: anywhere; color: ` + emailColorForeground + `; font-family: ` + emailFontStack + `; font-size: 15px; line-height: 21px; text-decoration: underline; text-decoration-thickness: 1px; text-underline-offset: 3px;`,
	"notificationText":      `margin: 0 0 12px; color: ` + emailColorTextSecondary + `; font-family: ` + emailFontStack + `; font-size: 15px; line-height: 21px;`,
	"notificationList":      `margin: 0; padding: 0; list-style: none;`,
	"notificationItemFirst": `padding: 12px 0; border-top: 1px solid ` + emailColorBorder + `; color: ` + emailColorTextSecondary + `; font-family: ` + emailFontStack + `; font-size: 15px; line-height: 21px;`,
	"notificationItem":      `padding: 12px 0; border-top: 1px solid ` + emailColorBorder + `; color: ` + emailColorTextSecondary + `; font-family: ` + emailFontStack + `; font-size: 15px; line-height: 21px;`,
	"notificationSublist":   `margin: 6px 0 12px; padding: 0; list-style: none;`,
	"notificationSubitem":   `padding: 8px 0; border-top: 0; color: ` + emailColorTextMuted + `; font-family: ` + emailFontStack + `; font-size: 15px; line-height: 21px;`,
	"notificationMessage":   `margin: 0; color: ` + emailColorTextSecondary + `; font-family: ` + emailFontStack + `; font-size: 15px; line-height: 21px;`,
	"notificationLink":      `display: inline; color: ` + emailColorForeground + `; font-family: ` + emailFontStack + `; font-size: 15px; font-weight: 600; line-height: 21px; text-decoration: none;`,
	"footer":                `margin-top: 46px; padding-top: 26px; border-top: 1px solid ` + emailColorBorder + `; color: ` + emailColorTextMuted + `; font-family: ` + emailFontStack + `; font-size: 15px; line-height: 21px;`,
	"footerText":            `margin: 0 0 8px; color: ` + emailColorTextMuted + `; font-family: ` + emailFontStack + `; font-size: 13px; line-height: 21px;`,
	"footerTextLast":        `margin: 0; color: ` + emailColorTextMuted + `; font-family: ` + emailFontStack + `; font-size: 13px; line-height: 21px;`,
}

func emailStyle(name string) template.CSS {
	// #nosec G203 -- EmailStyleString only returns values from the immutable
	// emailStyles allowlist; caller input can select a key but cannot add CSS.
	return template.CSS(EmailStyleString(name))
}

func EmailStyleString(name string) string {
	if style, ok := emailStyles[name]; ok {
		return style
	}
	return ""
}

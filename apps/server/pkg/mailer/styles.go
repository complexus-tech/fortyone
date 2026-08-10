package mailer

import (
	"html/template"
)

const emailFontStack = `Geist, Helvetica, Arial, sans-serif`

// Email clients do not consistently support OKLCH, so these colors are the
// sRGB hex equivalents of the shared light and dark theme tokens.
const (
	emailColorBackground        = `#fefefd` // background: oklch(0.997 0.002 95)
	emailColorForeground        = `#0b0a08` // foreground: oklch(0.145 0.006 95)
	emailColorForegroundInverse = `#fefefd` // foreground-inverse: oklch(0.997 0.002 95)
	emailColorTextSecondary     = `#3b3a38` // text-secondary: oklch(0.35 0.005 95)
	emailColorTextMuted         = `#565552` // text-muted: oklch(0.45 0.005 95)
	emailColorSurfaceMuted      = `#f7f7f4` // surface-muted: oklch(0.975 0.003 95)
	emailColorBorder            = `#e5e4e2` // border: oklch(0.92 0.004 95)
	emailColorButton            = `#14120b` // dark background: oklch(0.1821 0.0139 94)
)

var emailStyles = map[string]string{
	"body":                  `margin: 0; padding: 0; background-color: ` + emailColorBackground + `; color: ` + emailColorTextSecondary + `; font-family: ` + emailFontStack + `; font-size: 15px; line-height: 1.62;`,
	"shell":                 `width: 100%; max-width: 640px; margin: 0 auto; background-color: ` + emailColorBackground + `;`,
	"inner":                 `padding: 48px 48px 38px;`,
	"brand":                 `display: block; margin: 0 0 40px;`,
	"logo":                  `display: block; width: auto; height: 42px; border-radius: 8px;`,
	"eyebrow":               `margin: 0 0 14px; color: ` + emailColorTextMuted + `; font-family: ` + emailFontStack + `; font-size: 12px; font-weight: 600; line-height: 1.4; text-transform: uppercase;`,
	"heading":               `margin: 0; max-width: 520px; color: ` + emailColorForeground + `; font-family: ` + emailFontStack + `; font-size: 26px; font-weight: 600; line-height: 1.16;`,
	"bodyBlock":             `margin-top: 28px;`,
	"text":                  `margin: 0 0 18px; color: ` + emailColorTextSecondary + `; font-family: ` + emailFontStack + `; font-size: 15px; line-height: 1.62;`,
	"textNoMargin":          `margin: 0; color: ` + emailColorTextSecondary + `; font-family: ` + emailFontStack + `; font-size: 15px; line-height: 1.62;`,
	"textSmallTop":          `margin: 12px 0 0; color: ` + emailColorTextSecondary + `; font-family: ` + emailFontStack + `; font-size: 15px; line-height: 1.62;`,
	"buttonContainer":       `margin: 30px 0;`,
	"button":                `display: inline-block; padding: 10px 18px; border-radius: 8px; background-color: ` + emailColorButton + `; color: ` + emailColorForegroundInverse + `; font-family: ` + emailFontStack + `; font-size: 15px; font-weight: 400; line-height: 1.35; text-align: center; text-decoration: none;`,
	"panel":                 `margin: 28px 0; padding: 22px 24px; border-radius: 8px; background-color: ` + emailColorSurfaceMuted + `;`,
	"quietPanel":            `margin: 28px 0; padding: 22px 0; border-top: 1px solid ` + emailColorBorder + `; border-bottom: 1px solid ` + emailColorBorder + `; background-color: transparent; border-radius: 0;`,
	"notificationPanel":     `margin: 28px 0; padding: 0; background-color: transparent; color: ` + emailColorTextSecondary + `; font-family: ` + emailFontStack + `; font-size: 15px; line-height: 1.62;`,
	"panelLabel":            `margin: 0 0 10px; color: ` + emailColorTextMuted + `; font-family: ` + emailFontStack + `; font-size: 15px; font-weight: 600; line-height: 1.4;`,
	"panelLabelTight":       `margin: 0 0 4px; color: ` + emailColorTextMuted + `; font-family: ` + emailFontStack + `; font-size: 15px; font-weight: 600; line-height: 1.4;`,
	"panelValue":            `margin: 0; color: ` + emailColorForeground + `; font-family: ` + emailFontStack + `; font-size: 15px; font-weight: 500; line-height: 1.62;`,
	"panelTitle":            `margin: 0 0 12px; color: ` + emailColorForeground + `; font-family: ` + emailFontStack + `; font-size: 15px; font-weight: 600; line-height: 1.3;`,
	"detailList":            `display: block; margin: 0; padding: 0; list-style: none;`,
	"detailRow":             `display: block; padding: 0 0 10px;`,
	"detailLabel":           `display: inline-block; min-width: 116px; color: ` + emailColorTextMuted + `; font-family: ` + emailFontStack + `; font-size: 15px; line-height: 1.5;`,
	"detailValue":           `color: ` + emailColorForeground + `; font-family: ` + emailFontStack + `; font-size: 15px; font-weight: 600; line-height: 1.5;`,
	"codePanel":             `margin: 28px 0; padding: 24px; border-radius: 8px; background-color: ` + emailColorSurfaceMuted + `;`,
	"codePanelCentered":     `margin: 28px 0; padding: 24px; border-radius: 8px; background-color: ` + emailColorSurfaceMuted + `; text-align: center;`,
	"verificationCode":      `margin: 0; color: ` + emailColorForeground + `; font-family: SFMono-Regular, Consolas, monospace; font-size: 30px; font-weight: 600; line-height: 1.1;`,
	"securityNote":          `margin: 28px 0 0; padding-top: 22px; border-top: 1px solid ` + emailColorBorder + `; color: ` + emailColorTextMuted + `; font-family: ` + emailFontStack + `; font-size: 15px; line-height: 1.62;`,
	"secondaryActions":      `margin-top: 32px; margin-bottom: 42px;`,
	"secondaryActionLink":   `color: ` + emailColorTextMuted + `; font-family: ` + emailFontStack + `; font-size: 15px; line-height: 1.5; text-decoration: underline; text-decoration-thickness: 1px; text-underline-offset: 3px;`,
	"workspaceLink":         `display: inline-block; overflow-wrap: anywhere; color: ` + emailColorForeground + `; font-family: ` + emailFontStack + `; font-size: 15px; line-height: 1.5; text-decoration: underline; text-decoration-thickness: 1px; text-underline-offset: 3px;`,
	"notificationText":      `margin: 0 0 14px; color: ` + emailColorTextSecondary + `; font-family: ` + emailFontStack + `; font-size: 15px; line-height: 1.62;`,
	"notificationList":      `margin: 0; padding: 0;`,
	"notificationItemFirst": `padding: 0 0 10px; border-top: 0; color: ` + emailColorTextSecondary + `; font-family: ` + emailFontStack + `; font-size: 15px; line-height: 1.5;`,
	"notificationItem":      `padding: 10px 0; border-top: 1px solid ` + emailColorBorder + `; color: ` + emailColorTextSecondary + `; font-family: ` + emailFontStack + `; font-size: 15px; line-height: 1.5;`,
	"notificationSublist":   `margin: 6px 0 12px; padding: 0; list-style: none;`,
	"notificationSubitem":   `padding: 8px 0; border-top: 0; color: ` + emailColorTextMuted + `; font-family: ` + emailFontStack + `; font-size: 15px; line-height: 1.5;`,
	"notificationMessage":   `margin: 0; color: ` + emailColorTextSecondary + `; font-family: ` + emailFontStack + `; font-size: 15px; line-height: 1.55;`,
	"notificationLink":      `display: inline; color: ` + emailColorForeground + `; font-family: ` + emailFontStack + `; font-size: 15px; font-weight: 600; line-height: 1.55; text-decoration: none;`,
	"footer":                `margin-top: 46px; padding-top: 26px; border-top: 1px solid ` + emailColorBorder + `; color: ` + emailColorTextMuted + `; font-family: ` + emailFontStack + `; font-size: 15px; line-height: 1.5;`,
	"footerText":            `margin: 0 0 6px; color: ` + emailColorTextMuted + `; font-family: ` + emailFontStack + `; font-size: 14px; line-height: 1.5;`,
	"footerTextLast":        `margin: 0; color: ` + emailColorTextMuted + `; font-family: ` + emailFontStack + `; font-size: 14px; line-height: 1.5;`,
}

func emailStyle(name string) template.CSS {
	return template.CSS(EmailStyleString(name))
}

func EmailStyleString(name string) string {
	if style, ok := emailStyles[name]; ok {
		return style
	}
	return ""
}

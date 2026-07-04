package mailer

import (
	"html/template"
)

const emailFontStack = `Geist, Helvetica, Arial, sans-serif`

var emailStyles = map[string]string{
	"body":                  `margin: 0; padding: 0; background-color: #ffffff; color: #303030; font-family: ` + emailFontStack + `; font-size: 15px; line-height: 1.62;`,
	"shell":                 `width: 100%; max-width: 640px; margin: 0 auto; background-color: #ffffff;`,
	"inner":                 `padding: 48px 48px 38px;`,
	"brand":                 `display: block; margin: 0 0 40px;`,
	"logo":                  `display: block; width: auto; height: 42px; border-radius: 8px;`,
	"eyebrow":               `margin: 0 0 14px; color: #6f6c67; font-family: ` + emailFontStack + `; font-size: 12px; font-weight: 600; line-height: 1.4; text-transform: uppercase;`,
	"heading":               `margin: 0; max-width: 520px; color: #111111; font-family: ` + emailFontStack + `; font-size: 26px; font-weight: 600; line-height: 1.16;`,
	"bodyBlock":             `margin-top: 28px;`,
	"text":                  `margin: 0 0 18px; color: #303030; font-family: ` + emailFontStack + `; font-size: 15px; line-height: 1.62;`,
	"textNoMargin":          `margin: 0; color: #303030; font-family: ` + emailFontStack + `; font-size: 15px; line-height: 1.62;`,
	"textSmallTop":          `margin: 12px 0 0; color: #303030; font-family: ` + emailFontStack + `; font-size: 15px; line-height: 1.62;`,
	"buttonContainer":       `margin: 30px 0;`,
	"button":                `display: inline-block; padding: 13px 22px; border-radius: 8px; background-color: #111111; color: #ffffff; font-family: ` + emailFontStack + `; font-size: 15px; font-weight: 400; line-height: 1.35; text-align: center; text-decoration: none;`,
	"panel":                 `margin: 28px 0; padding: 22px 24px; border-radius: 8px; background-color: #f7f7f7;`,
	"quietPanel":            `margin: 28px 0; padding: 22px 0; border-top: 1px solid #e5e5e5; border-bottom: 1px solid #e5e5e5; background-color: transparent; border-radius: 0;`,
	"notificationPanel":     `margin: 28px 0; padding: 0; background-color: transparent; color: #303030; font-family: ` + emailFontStack + `; font-size: 15px; line-height: 1.62;`,
	"panelLabel":            `margin: 0 0 10px; color: #6f6c67; font-family: ` + emailFontStack + `; font-size: 15px; font-weight: 600; line-height: 1.4;`,
	"panelLabelTight":       `margin: 0 0 4px; color: #6f6c67; font-family: ` + emailFontStack + `; font-size: 15px; font-weight: 600; line-height: 1.4;`,
	"panelValue":            `margin: 0; color: #111111; font-family: ` + emailFontStack + `; font-size: 15px; font-weight: 500; line-height: 1.62;`,
	"panelTitle":            `margin: 0 0 12px; color: #111111; font-family: ` + emailFontStack + `; font-size: 15px; font-weight: 600; line-height: 1.3;`,
	"detailList":            `display: block; margin: 0; padding: 0; list-style: none;`,
	"detailRow":             `display: block; padding: 0 0 10px;`,
	"detailLabel":           `display: inline-block; min-width: 116px; color: #6f6c67; font-family: ` + emailFontStack + `; font-size: 15px; line-height: 1.5;`,
	"detailValue":           `color: #111111; font-family: ` + emailFontStack + `; font-size: 15px; font-weight: 600; line-height: 1.5;`,
	"codePanel":             `margin: 28px 0; padding: 24px; border-radius: 8px; background-color: #f7f7f7;`,
	"codePanelCentered":     `margin: 28px 0; padding: 24px; border-radius: 8px; background-color: #f7f7f7; text-align: center;`,
	"verificationCode":      `margin: 0; color: #111111; font-family: SFMono-Regular, Consolas, monospace; font-size: 30px; font-weight: 600; line-height: 1.1;`,
	"securityNote":          `margin: 28px 0 0; padding-top: 22px; border-top: 1px solid #e5e5e5; color: #6f6c67; font-family: ` + emailFontStack + `; font-size: 15px; line-height: 1.62;`,
	"secondaryActions":      `margin-top: 32px; margin-bottom: 42px;`,
	"secondaryActionLink":   `color: #6f6c67; font-family: ` + emailFontStack + `; font-size: 15px; line-height: 1.5; text-decoration: underline; text-decoration-thickness: 1px; text-underline-offset: 3px;`,
	"workspaceLink":         `display: inline-block; overflow-wrap: anywhere; color: #111111; font-family: ` + emailFontStack + `; font-size: 15px; line-height: 1.5; text-decoration: underline; text-decoration-thickness: 1px; text-underline-offset: 3px;`,
	"notificationText":      `margin: 0 0 14px; color: #303030; font-family: ` + emailFontStack + `; font-size: 15px; line-height: 1.62;`,
	"notificationList":      `margin: 0; padding: 0;`,
	"notificationItemFirst": `padding: 0 0 10px; border-top: 0; color: #303030; font-family: ` + emailFontStack + `; font-size: 15px; line-height: 1.5;`,
	"notificationItem":      `padding: 10px 0; border-top: 1px solid #e5e5e5; color: #303030; font-family: ` + emailFontStack + `; font-size: 15px; line-height: 1.5;`,
	"notificationSublist":   `margin: 6px 0 12px; padding: 0; list-style: none;`,
	"notificationSubitem":   `padding: 8px 0; border-top: 0; color: #6f6c67; font-family: ` + emailFontStack + `; font-size: 15px; line-height: 1.5;`,
	"notificationMessage":   `margin: 0; color: #4f4c48; font-family: ` + emailFontStack + `; font-size: 15px; line-height: 1.55;`,
	"notificationLink":      `display: inline; color: #111111; font-family: ` + emailFontStack + `; font-size: 15px; font-weight: 600; line-height: 1.55; text-decoration: none;`,
	"footer":                `margin-top: 46px; padding-top: 26px; border-top: 1px solid #e5e5e5; color: #6f6c67; font-family: ` + emailFontStack + `; font-size: 15px; line-height: 1.5;`,
	"footerText":            `margin: 0 0 6px; color: #6f6c67; font-family: ` + emailFontStack + `; font-size: 14px; line-height: 1.5;`,
	"footerTextLast":        `margin: 0; color: #6f6c67; font-family: ` + emailFontStack + `; font-size: 14px; line-height: 1.5;`,
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

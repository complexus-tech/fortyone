import { mkdir, readFile, writeFile } from "node:fs/promises";
import { fileURLToPath } from "node:url";
import { join } from "node:path";
import { emails } from "./content.mjs";

const root = fileURLToPath(new URL(".", import.meta.url));
const ink = "#25150e";
const muted = "#72645d";
const paper = "#fff1e7";
const rule = "#f1e4dc";
const fontStack = "Inter,Arial,Helvetica,sans-serif";
const fontFaces = [
  [400, "Regular"],
  [500, "Medium"],
  [600, "SemiBold"],
  [700, "Bold"],
]
  .map(
    ([weight, name]) =>
      `@font-face{font-family:Inter;font-style:normal;font-weight:${weight};font-display:swap;src:url('../assets/fonts/Inter-${name}.woff2') format('woff2');mso-font-alt:Arial;}`,
  )
  .join("");
// Use the same inner measure for images and Outlook's fixed-width buttons.
const cardWidth = (email) => (email.image ? 480 : 520);
const contentGutter = 32;
const escape = (value) =>
  String(value)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#39;");
const destination = (path) => `https://example.com/fortyone/${path}`;
const table = (content, attrs = "") =>
  `<table role="presentation" border="0" cellpadding="0" cellspacing="0" width="100%" ${attrs}>${content}</table>`;
const paragraph = (text, style = "") =>
  `<p style="margin:0 0 12px;color:${ink};font-size:15px;line-height:21px;${style}">${escape(text)}</p>`;
const label = (text) =>
  `<p style="margin:0 0 8px;color:${muted};font-size:11px;line-height:17px;font-weight:600;letter-spacing:1.6px;text-transform:uppercase;">${escape(text)}</p>`;
const quote = (text, color = "#ffb18e") =>
  table(
    `<tr><td style="border-left:2px solid ${color};padding:0 0 0 16px;">${paragraph(text, "margin:0;")}</td></tr>`,
  );
const signature = () =>
  `<p style="margin:16px 0 0;font-size:15px;line-height:21px;color:${ink};"><strong>Maya</strong><br><span style="font-size:14px;color:${muted};">Your AI agent at FortyOne</span></p>`;
const columns = (items) =>
  table(
    `<tr>${items.map(([key, value]) => `<td width="50%" valign="top" style="padding:0 12px 0 0;">${label(key)}<p style="margin:0;font-size:16px;line-height:21px;overflow-wrap:anywhere;">${escape(value)}</p></td>`).join("")}</tr>`,
  );
const seam = () =>
  `<tr><td height="24" style="padding:0;">${table(`<tr><td width="12" style="width:12px;">${table(`<tr><td height="24" bgcolor="${paper}" style="height:24px;border-radius:0 12px 12px 0;font-size:0;line-height:0;">&nbsp;</td></tr>`)}</td><td valign="middle" style="padding:0;">${table(`<tr><td style="height:1px;border-top:1px dashed #eea27e;font-size:0;line-height:0;">&nbsp;</td></tr>`)}</td><td width="12" style="width:12px;">${table(`<tr><td height="24" bgcolor="${paper}" style="height:24px;border-radius:12px 0 0 12px;font-size:0;line-height:0;">&nbsp;</td></tr>`)}</td></tr>`)}</td></tr>`;

function button(action, width) {
  const url = escape(destination(action.path));
  return `<!--[if mso]><v:roundrect xmlns:v="urn:schemas-microsoft-com:vml" xmlns:w="urn:schemas-microsoft-com:office:word" href="${url}" style="height:44px;v-text-anchor:middle;width:${width}px;" arcsize="13%" stroke="f" fillcolor="${ink}"><w:anchorlock/><center style="color:#ffffff;font-family:Arial,sans-serif;font-size:14px;font-weight:bold;">${escape(action.label)}</center></v:roundrect><![endif]--><!--[if !mso]><!--><a href="${url}" target="_blank" style="display:block;background-color:${ink};border:1px solid ${ink};border-radius:6px;color:#ffffff;font-family:${fontStack};font-size:14px;font-weight:600;line-height:22px;padding:10px 12px;text-align:center;text-decoration:none;mso-hide:all;">${escape(action.label)}</a><!--<![endif]-->`;
}

function bodyContent(email, width) {
  let result = email.image
    ? `<img class="hero" src="../assets/${escape(email.image)}" width="${width}" alt="${escape(email.imageAlt)}" style="display:block;border:0;border-radius:8px;width:100%;max-width:${width}px;height:auto;margin:0 0 24px;color:${muted};font-size:14px;">`
    : "";
  result += `<h1 style="margin:0 0 12px;color:${ink};font-size:23px;line-height:29px;font-weight:600;letter-spacing:-0.5px;max-width:420px;">${escape(email.title)}</h1>`;
  result += email.paragraphs.map((text) => paragraph(text)).join("");
  if (email.details)
    result += table(
      `<tr><td style="padding:18px 0 0;">${columns(email.details)}</td></tr>`,
    );
  if (email.warning)
    result += `<h2 style="margin:20px 0 8px;color:#aa421b;font-size:17px;line-height:23px;font-weight:600;">${escape(email.warning.title)}</h2>${paragraph(email.warning.text)}`;
  if (email.rows)
    result += table(
      email.rows
        .map(
          ([key, value]) =>
            `<tr><td class="detail-label" width="42%" valign="top" style="padding:12px 8px 12px 0;border-bottom:1px solid ${rule};font-size:15px;line-height:21px;">${escape(key)}</td><td class="detail-value" style="padding:12px 0;border-bottom:1px solid ${rule};font-size:15px;line-height:21px;overflow-wrap:anywhere;word-break:break-word;">${escape(value)}</td></tr>`,
        )
        .join(""),
      `style="margin:18px 0;"`,
    );
  if (email.updates)
    result += table(
      email.updates
        .map(
          (update) =>
            `<tr><td style="padding:12px 0;border-top:1px solid ${rule};">${update.id ? label(update.id) : ""}<h2 style="margin:0 0 6px;font-size:15px;line-height:21px;font-weight:600;">${escape(update.title)}</h2>${paragraph(update.text, "margin:0;font-size:15px;line-height:21px;")}${update.quote ? `<div style="padding-top:10px;">${quote(update.quote)}</div>` : ""}</td></tr>`,
        )
        .join(""),
      `style="margin-top:14px;"`,
    );
  if (email.conversation)
    result += table(
      `<tr><td style="padding:16px 0;border-top:1px solid ${rule};border-bottom:1px solid ${rule};">${label("Your comment")}${quote(email.conversation.original, "#dedad7")}</td></tr><tr><td style="padding:16px 0 0;"><h2 style="margin:0 0 8px;font-size:15px;line-height:21px;">${escape(email.conversation.author)}</h2>${quote(email.conversation.reply)}</td></tr>`,
      `style="margin-top:14px;"`,
    );
  if (email.proposal)
    result += table(
      `<tr><td style="padding:16px 0;border-top:1px solid ${rule};">${label("Objective")}<h2 style="margin:0 0 16px;font-size:17px;line-height:23px;">${escape(email.proposal.objective)}</h2>${table(`<tr><td width="50%" valign="top" style="padding-right:12px;"><p style="margin:0 0 8px;line-height:21px;">Current health</p><p style="margin:0;color:#a44932;font-size:16px;line-height:21px;font-weight:600;">${escape(email.proposal.before)}</p></td><td valign="top"><p style="margin:0 0 8px;line-height:21px;">New health</p><p style="margin:0;color:#32674e;font-size:16px;line-height:21px;font-weight:600;">${escape(email.proposal.after)}</p></td></tr>`)}</td></tr><tr><td style="padding:16px 0;border-top:1px solid ${rule};border-bottom:1px solid ${rule};">${paragraph("Check-in to add", "margin:0 0 12px;")}${quote(email.proposal.checkIn)}</td></tr>`,
      `style="margin-top:14px;"`,
    );
  if (email.closing) {
    const closing = paragraph(email.closing, "margin:16px 0 0;");
    result += email.emphasis
      ? closing.replace(
          escape(email.emphasis),
          `<strong>${escape(email.emphasis)}</strong>`,
        )
      : closing;
  }
  if (email.maya && !email.confirmation) result += signature();
  return result;
}

function render(email) {
  const width = cardWidth(email);
  const innerWidth = width - contentGutter * 2;
  const footer = email.footerLink
    ? `<a href="${destination("settings/notifications")}" target="_blank" style="color:${muted};text-decoration:underline;">${escape(email.footerLink)}</a>`
    : escape(email.footer || "");
  const actionContent = email.confirmation
    ? `<p style="margin:0 0 12px;font-size:15px;line-height:21px;font-weight:600;">Reply <span style="font-family:Consolas,monospace;">CONFIRM</span> to apply this change.</p><p style="margin:0;font-size:15px;line-height:21px;">Reply <span style="font-family:Consolas,monospace;">CANCEL</span> to leave it unchanged.</p>${signature()}`
    : `${button(email.action, innerWidth)}${email.helper ? `<p style="margin:18px 0 0;text-align:center;color:${muted};font-size:13px;line-height:21px;">${escape(email.helper)}</p>` : ""}${email.fallback ? `<p style="margin:18px 0 0;text-align:center;font-size:14px;line-height:22px;"><a href="${destination("workspace")}" target="_blank" style="color:${muted};text-decoration:underline;">${escape(email.fallback)}</a></p>` : ""}`;
  return `<!doctype html>
<html lang="en" xmlns="http://www.w3.org/1999/xhtml" xmlns:v="urn:schemas-microsoft-com:vml" xmlns:o="urn:schemas-microsoft-com:office:office">
<head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><meta http-equiv="X-UA-Compatible" content="IE=edge"><meta name="x-apple-disable-message-reformatting"><meta name="color-scheme" content="light"><meta name="supported-color-schemes" content="light"><title>${escape(email.subject)}</title>
<!--[if mso]><noscript><xml><o:OfficeDocumentSettings><o:AllowPNG/><o:PixelsPerInch>96</o:PixelsPerInch></o:OfficeDocumentSettings></xml></noscript><![endif]-->
<!--[if !mso]><!--><style>${fontFaces}</style><!--<![endif]-->
<!--[if mso]><style>body,table,td,p,h1,h2,a,strong,span{font-family:Arial,sans-serif!important;}</style><![endif]-->
<style>html{min-height:100%;background-color:${paper};}body{margin:0!important;padding:0!important;width:100%!important;-webkit-text-size-adjust:100%;-ms-text-size-adjust:100%;}table,td{mso-table-lspace:0pt;mso-table-rspace:0pt;}img{-ms-interpolation-mode:bicubic;}a[x-apple-data-detectors]{color:inherit!important;text-decoration:none!important;}@media screen and (max-width:560px){.outer{padding:26px 16px!important;}.content{padding:24px 24px 14px!important;}.actions{padding:16px 24px 28px!important;}.brand{padding-bottom:24px!important;}.email-card{width:100%!important;}.email-footer{padding:24px 8px 0!important;}h1{font-size:21px!important;line-height:27px!important;}.hero{margin-bottom:22px!important;}}@media screen and (max-width:420px){.detail-label,.detail-value{display:block!important;width:100%!important;}.detail-label{padding:14px 0 4px!important;border-bottom:0!important;color:#72645d!important;}.detail-value{padding:0 0 14px!important;}}</style>
</head><body bgcolor="${paper}" style="margin:0;padding:0;background-color:${paper};color:${ink};font-family:${fontStack};font-size:15px;line-height:21px;">
<div style="display:none;font-size:1px;color:${paper};line-height:1px;max-height:0;max-width:0;opacity:0;overflow:hidden;mso-hide:all;">${escape(email.preheader)}${"&#847; &zwnj; &nbsp;".repeat(24)}</div>
${table(`<tr><td class="outer" align="center" style="padding:32px 24px;background-color:${paper};font-family:${fontStack};"><!--[if mso]><table role="presentation" border="0" cellpadding="0" cellspacing="0" width="${width}"><tr><td><![endif]-->${table(`<tr><td class="brand" align="center" style="padding:0 0 26px;"><img src="../assets/wordmark.png" width="126" height="31" alt="fortyone" style="display:block;border:0;width:126px;height:31px;color:${ink};font-size:24px;font-weight:bold;"></td></tr><tr><td>${table(`<tr><td class="content" style="padding:28px ${contentGutter}px 14px;">${bodyContent(email, innerWidth)}</td></tr>${seam()}<tr><td class="actions" style="padding:16px ${contentGutter}px 26px;">${actionContent}</td></tr>`, `class="email-card" bgcolor="#ffffff" style="width:100%;background-color:#ffffff;border-radius:16px;border-collapse:separate;"`)}</td></tr><tr><td class="email-footer" align="center" style="padding:26px 20px 0;color:${muted};font-size:13px;line-height:21px;">${footer ? `<p style="margin:0 0 8px;">${footer}</p>` : ""}<p style="margin:0;">FortyOne by Complexus LLC</p></td></tr>`, `style="width:100%;max-width:${width}px;"`)}<!--[if mso]></td></tr></table><![endif]--></td></tr>`, `bgcolor="${paper}" style="background-color:${paper};"`)}
</body></html>`;
}

function plainText(email) {
  const parts = [email.title, ...email.paragraphs];
  if (email.details) parts.push(...email.details.map(([k, v]) => `${k}: ${v}`));
  if (email.warning) parts.push(email.warning.title, email.warning.text);
  if (email.rows) parts.push(...email.rows.map(([k, v]) => `${k}: ${v}`));
  for (const update of email.updates || [])
    parts.push(
      [update.id, update.title, update.text, update.quote]
        .filter(Boolean)
        .join("\n"),
    );
  if (email.conversation)
    parts.push(
      `Your comment: ${email.conversation.original}`,
      `${email.conversation.author}: ${email.conversation.reply}`,
    );
  if (email.proposal)
    parts.push(
      email.proposal.objective,
      `Current health: ${email.proposal.before}\nNew health: ${email.proposal.after}`,
      `Check-in to add: ${email.proposal.checkIn}`,
    );
  if (email.closing) parts.push(email.closing);
  if (email.confirmation)
    parts.push(
      "Reply CONFIRM to apply this change.\nReply CANCEL to leave it unchanged.",
    );
  if (email.maya) parts.push("Maya\nYour AI agent at FortyOne");
  if (email.action)
    parts.push(`${email.action.label}: ${destination(email.action.path)}`);
  if (email.helper) parts.push(email.helper);
  if (email.fallback)
    parts.push(`${email.fallback}: ${destination("workspace")}`);
  if (email.footer) parts.push(email.footer);
  if (email.footerLink)
    parts.push(`${email.footerLink}: ${destination("settings/notifications")}`);
  parts.push("FortyOne by Complexus LLC");
  return parts.join("\n\n") + "\n";
}

await mkdir(join(root, "emails"), { recursive: true });
const manifest = [];
for (const email of emails) {
  const html = render(email);
  await writeFile(join(root, "emails", `${email.id}.html`), html);
  await writeFile(join(root, "emails", `${email.id}.txt`), plainText(email));
  manifest.push({
    ...email,
    html: `emails/${email.id}.html`,
    text: `emails/${email.id}.txt`,
    bytes: Buffer.byteLength(html),
    sender: email.maya
      ? "Maya, AI Agent <maya@example.com>"
      : "FortyOne <notifications@example.com>",
    recipient: email.recipient || "Alex <alex@example.com>",
    replyTo: email.maya
      ? "Maya <maya-thread@example.com> (sample)"
      : "Not enabled in this preview",
  });
}
await writeFile(
  join(root, "metadata.json"),
  JSON.stringify(manifest, null, 2) + "\n",
);
const gallery = await readFile(join(root, "gallery.html"), "utf8");
await writeFile(
  join(root, "index.html"),
  gallery.replace(
    "<!-- MANIFEST -->",
    `<script id="manifest" type="application/json">${JSON.stringify(manifest).replaceAll("<", "\\u003c")}</script>`,
  ),
);
console.log(
  `Built ${manifest.length} HTML emails, plain-text alternatives, metadata, and gallery.`,
);

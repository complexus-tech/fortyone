type StandaloneDocumentHTMLInput = {
  contentHtml: string;
  title: string;
  updatedAt: string;
};

const escapeHTML = (value: string) =>
  value
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#039;");

export const createStandaloneDocumentHTML = ({
  contentHtml,
  title,
  updatedAt,
}: StandaloneDocumentHTMLInput) =>
  `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>${escapeHTML(title)}</title>
  <style>
    :root { color-scheme: light; font-family: ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; }
    body { color: #171717; margin: 0; }
    article { margin: 0 auto; max-width: 760px; padding: 64px 28px 96px; }
    h1 { font-size: 2.75rem; letter-spacing: -0.035em; line-height: 1.08; margin: 0 0 12px; }
    .updated { color: #737373; margin: 0 0 48px; }
    .content { font-size: 1.075rem; line-height: 1.7; }
    .content h2 { font-size: 1.65rem; line-height: 1.2; margin: 2.25rem 0 0.75rem; }
    .content h3 { font-size: 1.3rem; margin: 2rem 0 0.65rem; }
    .content .tableWrapper { border-radius: 8px; corner-shape: squircle; margin: 1.5rem 0; max-width: 100%; overflow-x: auto; }
    .content table { border-collapse: separate; border-spacing: 0; margin: 0; min-width: 100%; width: max-content; }
    .content td, .content th { border: 1px solid #d4d4d4; min-width: 140px; padding: 10px 12px; text-align: left; vertical-align: top; }
    .content th { background: #fafafa; font-weight: 600; }
    .content blockquote { border-left: 3px solid #a3a3a3; color: #525252; margin-left: 0; padding-left: 18px; }
    .content img, .content video { border: 1px solid #e5e5e5; border-radius: 13px; corner-shape: squircle; height: auto; max-width: 100%; }
    .content mark { border-radius: 3px; padding: 0 0.08em; }
    @media print { article { padding-top: 24px; } }
  </style>
</head>
<body>
  <article>
    <h1>${escapeHTML(title)}</h1>
    <p class="updated">Updated ${escapeHTML(updatedAt)}</p>
    <div class="content">${contentHtml}</div>
  </article>
</body>
</html>`;

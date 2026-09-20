import sanitizeHtml, { defaults, simpleTransform } from "sanitize-html";

// Public document content never renders workspace chrome, relationships, member
// details, history, or unsanitized HTML. Media stays behind the revocable token.
export function publicDocumentHTML(
  html: string,
  token: string,
  apiUrl: string,
) {
  const sanitizedHTML = sanitizeHtml(html, {
    allowedTags: [...defaults.allowedTags, "img", "video", "input", "mark"],
    allowedAttributes: {
      a: ["href", "title", "rel", "target"],
      img: [
        "src",
        "alt",
        "title",
        "width",
        "data-width",
        "data-align",
        "style",
      ],
      video: ["src", "controls", "preload"],
      input: ["type", "checked", "disabled"],
      td: ["colspan", "rowspan"],
      th: ["colspan", "rowspan"],
      ul: ["data-type"],
      li: ["data-type", "data-checked"],
      mark: ["style"],
      span: ["style"],
    },
    allowedStyles: {
      img: {
        width: [/^\d+(?:\.\d+)?(?:px|%)$/],
        "max-width": [/^100%$/],
        display: [/^block$/],
        "margin-left": [/^(?:auto|0)$/],
        "margin-right": [/^(?:auto|0)$/],
      },
      mark: {
        "background-color": [
          /^#[0-9a-f]{6}$/i,
          /^rgb\(\s*\d{1,3}\s*,\s*\d{1,3}\s*,\s*\d{1,3}\s*\)$/i,
        ],
      },
      span: {
        color: [
          /^#[0-9a-f]{6}$/i,
          /^rgb\(\s*\d{1,3}\s*,\s*\d{1,3}\s*,\s*\d{1,3}\s*\)$/i,
        ],
      },
    },
    allowedSchemes: ["https", "http", "mailto"],
    allowProtocolRelative: false,
    transformTags: {
      a: simpleTransform("a", {
        rel: "noopener noreferrer",
        target: "_blank",
      }),
      input: (_tag, attrs) => ({
        tagName: "input",
        attribs: {
          type: "checkbox",
          disabled: "",
          ...(Object.hasOwn(attrs, "checked") ? { checked: "" } : {}),
        },
      }),
      img: (tagName, attribs) => ({
        tagName,
        attribs: rewriteMedia(attribs, token, apiUrl),
      }),
      video: (tagName, attribs) => ({
        tagName,
        attribs: {
          ...rewriteMedia(attribs, token, apiUrl),
          controls: "",
          preload: "metadata",
        },
      }),
    },
  });

  return wrapPublicDocumentTables(
    normalizeLegacyMarkdownArtifacts(sanitizedHTML),
  );
}

const MARKDOWN_TABLE_ROW = /^\s*\|.+\|\s*$/u;
const MARKDOWN_TABLE_SEPARATOR_CELL = /^:?-{3,}:?$/u;
const MARKDOWN_TABLE_BLOCK = /(?:<p>\s*\|(?:(?!<\/p>)[\s\S])+\|\s*<\/p>){2,}/gu;
/* eslint-disable prefer-named-capture-group -- the app target is below ES2018. */
const PARAGRAPH_CONTENT = /<p>([\s\S]*?)<\/p>/gu;
const TABLE_ELEMENT = /<table(?:\s[^>]*)?>[\s\S]*?<\/table>/gu;

function wrapPublicDocumentTables(html: string) {
  return html.replace(
    TABLE_ELEMENT,
    (table) => `<div class="tableWrapper">${table}</div>`,
  );
}

const parseMarkdownTableRow = (value: string) =>
  value
    .trim()
    .replace(/^\||\|$/gu, "")
    .split("|")
    .map((cell) => cell.trim());

function normalizeLegacyMarkdownArtifacts(html: string) {
  const normalizedLists = html
    .replace(/<ul([^>]*)>([\s\S]*?)<\/ul>/gu, (_match, attributes, items) => {
      const normalizedItems = String(items).replace(
        /(<li[^>]*>\s*<p[^>]*>)\s*[-+*]\s+/gu,
        "$1",
      );
      return `<ul${String(attributes)}>${normalizedItems}</ul>`;
    })
    .replace(/<ol([^>]*)>([\s\S]*?)<\/ol>/gu, (_match, attributes, items) => {
      const normalizedItems = String(items).replace(
        /(<li[^>]*>\s*<p[^>]*>)\s*\d+[.)]\s+/gu,
        "$1",
      );
      return `<ol${String(attributes)}>${normalizedItems}</ol>`;
    });

  return normalizedLists.replace(MARKDOWN_TABLE_BLOCK, (block) => {
    const rows = [...block.matchAll(PARAGRAPH_CONTENT)].map((match) =>
      parseMarkdownTableRow(match[1]),
    );
    if (
      rows.length < 2 ||
      !MARKDOWN_TABLE_ROW.test(`|${rows[0].join("|")}|`) ||
      !rows[1].every((cell) => MARKDOWN_TABLE_SEPARATOR_CELL.test(cell))
    )
      return block;

    const header = rows[0].map((cell) => `<th>${cell}</th>`).join("");
    const body = rows
      .slice(2)
      .map(
        (row) => `<tr>${row.map((cell) => `<td>${cell}</td>`).join("")}</tr>`,
      )
      .join("");
    return `<table><thead><tr>${header}</tr></thead><tbody>${body}</tbody></table>`;
  });
}
/* eslint-enable prefer-named-capture-group -- End the ES2017 compatibility parser exception. */

function rewriteMedia(
  attributes: Record<string, string>,
  token: string,
  apiUrl: string,
) {
  const source = attributes.src;
  const match =
    /\/workspaces\/[^/]+\/documents\/[0-9a-f-]+\/media\/[0-9a-f-]{36}(?=[?#]|$)/.exec(
      source,
    );
  return match
    ? {
        ...attributes,
        src: `${apiUrl}/public/documents/${token}/media/${match[0].slice(-36)}`,
      }
    : attributes;
}

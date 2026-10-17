import { marked, Renderer } from "marked";
import type { Token } from "marked";
import sanitizeHtml from "sanitize-html";

const escapeText = (text: string) =>
  text.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
const escapeAttribute = (text: string) =>
  escapeText(text).replace(/"/g, "&quot;");
const USER_MENTION =
  /^mention:[\da-f]{8}-[\da-f]{4}-[\da-f]{4}-[\da-f]{4}-[\da-f]{12}$/i;
export function hasDocumentMayaMention(body: string): boolean {
  const contains = (tokens: Token[]): boolean =>
    tokens.some((token) => {
      if (["blockquote", "code", "codespan", "html"].includes(token.type))
        return false;
      if (token.type === "link" && token.href === "agent:maya") return true;
      if ("items" in token && Array.isArray(token.items))
        return contains(token.items as Token[]);
      return (
        "tokens" in token &&
        Array.isArray(token.tokens) &&
        contains(token.tokens)
      );
    });
  return contains(marked.lexer(body));
}
const renderer = new Renderer();
// The API body is Markdown/plain text, never an HTML document or file upload.
renderer.html = ({ text }) => escapeText(text);
renderer.image = ({ text }) => escapeText(text);
renderer.link = function renderLink(token) {
  const userId = USER_MENTION.test(token.href)
    ? token.href.slice(8).toLowerCase()
    : undefined;
  const agent = token.href === "agent:maya";
  if (userId || agent) {
    const label = agent ? "Maya" : token.text.replace(/^@/, "");
    return `<span data-type="mention" data-mention-kind="${agent ? "agent" : "user"}" data-id="${userId ?? "maya"}" data-label="${escapeAttribute(label)}"${agent ? ' data-agent-key="maya"' : ""}>@${escapeText(label)}</span>`;
  }
  return `<a href="${escapeAttribute(token.href)}"${token.title ? ` title="${escapeAttribute(token.title)}"` : ""} target="_blank" rel="noopener noreferrer">${this.parser.parseInline(token.tokens)}</a>`;
};

export const renderDocumentComment = (body: string) =>
  sanitizeHtml(
    marked.parse(body, { async: false, breaks: true, gfm: true, renderer }),
    {
      allowedTags: [
        "p",
        "br",
        "strong",
        "em",
        "s",
        "del",
        "a",
        "span",
        "ul",
        "ol",
        "li",
        "blockquote",
        "pre",
        "code",
        "h1",
        "h2",
        "h3",
        "h4",
        "h5",
        "h6",
        "hr",
        "table",
        "thead",
        "tbody",
        "tr",
        "th",
        "td",
      ],
      allowedAttributes: {
        a: ["href", "title", "rel", "target"],
        span: [
          "data-type",
          "data-id",
          "data-label",
          "data-mention-kind",
          "data-agent-key",
        ],
        ol: ["start"],
      },
      allowedSchemes: ["http", "https", "mailto"],
      allowProtocolRelative: false,
    },
  );

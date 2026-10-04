import { Parser } from "htmlparser2";

const HIDDEN_ELEMENTS = new Set([
  "head",
  "iframe",
  "noscript",
  "object",
  "script",
  "style",
  "svg",
  "template",
]);
const BLOCK_ELEMENTS = new Set([
  "blockquote",
  "br",
  "div",
  "h1",
  "h2",
  "h3",
  "h4",
  "h5",
  "h6",
  "hr",
  "li",
  "ol",
  "p",
  "pre",
  "table",
  "td",
  "th",
  "tr",
  "ul",
]);

/** Extract visible text without a DOM, preserving encoded literal text once. */
export const htmlToPlainText = (html: string): string => {
  const parts: string[] = [];
  let hiddenDepth = 0;
  const parser = new Parser(
    {
      onopentag(name) {
        if (hiddenDepth > 0 || HIDDEN_ELEMENTS.has(name)) {
          hiddenDepth++;
        } else if (BLOCK_ELEMENTS.has(name)) {
          parts.push(" ");
        }
      },
      ontext(text) {
        if (hiddenDepth === 0) parts.push(text);
      },
      onclosetag(name) {
        if (hiddenDepth > 0) {
          hiddenDepth--;
        } else if (BLOCK_ELEMENTS.has(name)) {
          parts.push(" ");
        }
      },
    },
    { decodeEntities: true },
  );
  parser.end(html);
  return parts.join("").replace(/\s+/g, " ").trim();
};

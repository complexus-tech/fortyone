import sanitizeHtml, { defaults, simpleTransform } from "sanitize-html";
import {
  looksLikeMarkdown,
  markdownToRichTextHTML,
} from "@/lib/tiptap/markdown";

const mentionAttributes = [
  "class",
  "data-type",
  "data-id",
  "data-label",
  "data-mention-kind",
  "data-agent-key",
];

export const sanitizeCommentHTML = (html: string) =>
  sanitizeHtml(html, {
    allowedTags: [...defaults.allowedTags, "input", "label", "s", "u", "img"],
    allowedAttributes: {
      a: ["href", "title", "target", "rel", ...mentionAttributes],
      span: mentionAttributes,
      img: ["src", "alt", "title", "width", "height"],
      code: ["class"],
      ol: ["start"],
      ul: ["data-type"],
      li: ["data-type", "data-checked"],
      th: ["colspan", "rowspan"],
      td: ["colspan", "rowspan"],
      input: ["type", "checked", "disabled"],
    },
    allowedClasses: { a: ["mention"], span: ["mention"], code: ["language-*"] },
    allowedSchemes: ["http", "https", "mailto"],
    allowProtocolRelative: false,
    transformTags: {
      a: simpleTransform("a", { rel: "noopener noreferrer" }),
      input: (_tag, attributes) => ({
        tagName: "input",
        attribs: {
          type: "checkbox",
          disabled: "",
          ...(Object.hasOwn(attributes, "checked") ? { checked: "" } : {}),
        },
      }),
    },
  });

/** Only legacy plain paragraphs are Markdown; rich documents keep their existing structure. */
export const renderCommentHTML = (html: string) => {
  const safeHTML = sanitizeCommentHTML(html);
  const document = new DOMParser().parseFromString(safeHTML, "text/html");
  const hasRichContent = Array.from(document.body.querySelectorAll("*")).some(
    (element) => !["P", "BR"].includes(element.tagName),
  );
  if (hasRichContent) return safeHTML;
  document.body.querySelectorAll("br").forEach((element) => {
    element.replaceWith("\n");
  });
  const text = Array.from(document.body.childNodes)
    .map((node) => node.textContent ?? "")
    .join("\n\n");
  return looksLikeMarkdown(text)
    ? sanitizeCommentHTML(markdownToRichTextHTML(text))
    : safeHTML;
};

import DOMPurify from "dompurify";
import { isSafeLink, isSafeMediaUrl } from "./content";

const TAGS = [
  "p",
  "br",
  "strong",
  "b",
  "em",
  "i",
  "u",
  "s",
  "strike",
  "del",
  "h1",
  "h2",
  "h3",
  "h4",
  "h5",
  "h6",
  "blockquote",
  "code",
  "pre",
  "hr",
  "ul",
  "ol",
  "li",
  "label",
  "input",
  "div",
  "span",
  "a",
  "img",
  "video",
  "table",
  "thead",
  "tbody",
  "tr",
  "th",
  "td",
  "colgroup",
  "col",
];
const UNSAFE_TAGS = new Set([
  "script",
  "style",
  "iframe",
  "object",
  "embed",
  "svg",
  "math",
  "form",
  "button",
  "meta",
  "link",
]);
const SUPPORTED_DATA_TYPES = new Set(["taskList", "taskItem", "mention"]);

/** Detect unsupported document constructs before a schema could silently discard them. */
export const getUnsupportedRichText = (html: string) => {
  const document = new DOMParser().parseFromString(html, "text/html");
  const unsupported = new Set<string>();
  for (const element of Array.from(document.body.querySelectorAll("*"))) {
    const tag = element.tagName.toLowerCase();
    if (
      UNSAFE_TAGS.has(tag) ||
      element.closest("script,style,iframe,object,embed,svg,math,form")
    )
      continue;
    if (!TAGS.includes(tag)) unsupported.add(tag);
    const type = element.getAttribute("data-type");
    if (type && !SUPPORTED_DATA_TYPES.has(type)) unsupported.add(type);
    if (tag === "video" && !element.hasAttribute("data-document-media-video"))
      unsupported.add("video format");
    if (
      tag === "div" &&
      !element.parentElement?.matches('li[data-type="taskItem"]')
    )
      unsupported.add("custom block");
    if (
      element.hasAttribute("style") &&
      !["img", "video", "table", "th", "td", "col"].includes(tag)
    )
      unsupported.add("custom text styling");
  }
  return [...unsupported];
};

/** Runs only in the editor's DOM runtime, never in the native JavaScript runtime. */
export const sanitizeRichText = (html: string) => {
  const clean = DOMPurify.sanitize(html, {
    ALLOWED_TAGS: TAGS,
    ALLOWED_ATTR: [
      "href",
      "src",
      "alt",
      "title",
      "class",
      "colspan",
      "rowspan",
      "colwidth",
      "start",
      "type",
      "checked",
      "data-type",
      "data-checked",
      "data-id",
      "data-label",
      "data-mention-suggestion-char",
      "data-width",
      "data-align",
      "data-attachment-id",
      "data-document-media-video",
    ],
    ALLOW_DATA_ATTR: false,
    FORBID_TAGS: [...UNSAFE_TAGS],
  });
  const document = new DOMParser().parseFromString(clean, "text/html");
  for (const anchor of Array.from(document.body.querySelectorAll("a"))) {
    const href = anchor.getAttribute("href") ?? "";
    const profileLink = /^\/profile\/[a-zA-Z0-9-]+$/.test(href);
    if (!isSafeLink(href) && !profileLink) anchor.removeAttribute("href");
    anchor.setAttribute("rel", "noopener noreferrer");
  }
  for (const media of Array.from(document.body.querySelectorAll("img,video"))) {
    if (!isSafeMediaUrl(media.getAttribute("src") ?? ""))
      media.removeAttribute("src");
    if (media.tagName === "VIDEO") {
      media.setAttribute("controls", "");
      media.setAttribute("preload", "none");
    }
  }
  return document.body.innerHTML;
};

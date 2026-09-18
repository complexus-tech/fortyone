"use client";

import { useMemo } from "react";
import sanitizeHtml, { simpleTransform } from "sanitize-html";

const DESCRIPTION_TAGS = [
  "p",
  "div",
  "span",
  "br",
  "hr",
  "a",
  "h1",
  "h2",
  "h3",
  "h4",
  "h5",
  "h6",
  "strong",
  "b",
  "em",
  "i",
  "u",
  "s",
  "strike",
  "del",
  "sub",
  "sup",
  "ul",
  "ol",
  "li",
  "blockquote",
  "pre",
  "code",
  "table",
  "thead",
  "tbody",
  "tfoot",
  "tr",
  "th",
  "td",
];
const HTML_DESCRIPTION = new RegExp(
  `</?(?:${DESCRIPTION_TAGS.join("|")})(?:\\s[^>]*|\\s*/?)>`,
  "i",
);

export function CalendarEventDescription({
  description,
}: {
  description: string;
}) {
  const html = useMemo(() => {
    if (!HTML_DESCRIPTION.test(description)) return null;
    return sanitizeHtml(description, {
      allowedTags: DESCRIPTION_TAGS,
      allowedAttributes: {
        a: ["href", "title", "target", "rel"],
        ol: ["start"],
        li: ["value"],
        th: ["colspan", "rowspan"],
        td: ["colspan", "rowspan"],
      },
      allowedSchemes: ["https", "http", "mailto", "tel"],
      allowProtocolRelative: false,
      transformTags: {
        a: simpleTransform("a", {
          target: "_blank",
          rel: "noopener noreferrer",
        }),
      },
    });
  }, [description]);

  if (html === null) {
    return (
      <div className="text-text-muted text-base break-words whitespace-pre-wrap">
        {description}
      </div>
    );
  }

  return (
    <div
      className="prose dark:prose-invert text-text-muted prose-headings:text-foreground prose-a:text-foreground prose-p:my-2 prose-ul:my-2 prose-ol:my-2 prose-headings:mt-4 prose-headings:mb-2 prose-a:underline max-w-none overflow-x-auto text-base leading-relaxed break-words"
      dangerouslySetInnerHTML={{ __html: html }}
    />
  );
}

import type { JSONContent } from "@tiptap/core";
import { find } from "linkifyjs";
import type { ChatLink } from "./types";

const MAX_MESSAGE_LINKS = 50;

export const getChatLinkHref = (value: unknown): string | null => {
  if (typeof value !== "string") return null;
  try {
    const url = new URL(value.trim());
    return ["https:", "http:", "mailto:"].includes(url.protocol)
      ? url.href
      : null;
  } catch {
    return null;
  }
};

type TextRange = { start: number; end: number };

/** Keep legacy plain-text code examples literal, including unfinished fences. */
const codeRanges = (text: string): TextRange[] => {
  const ranges: TextRange[] = [];
  // The app targets ES2017, while its supported browsers handle named groups.
  // eslint-disable-next-line prefer-regex-literals -- Named literals fail the ES2017 TypeScript target.
  const fences = new RegExp(
    "(?:^|\\n)[ \\t]*(?<marker>`{3,}|~{3,})[^\\n]*(?:\\n|$)",
    "g",
  );
  let opening: RegExpExecArray | null;
  while ((opening = fences.exec(text))) {
    const start = opening.index;
    const marker = opening.groups?.marker;
    if (!marker) continue;
    const closing = new RegExp(
      `(?:^|\\n)[ \\t]*${marker[0]}{${marker.length},}[ \\t]*(?=\\n|$)`,
      "g",
    );
    closing.lastIndex = Math.max(start, fences.lastIndex - 1);
    const match = closing.exec(text);
    const end = match ? closing.lastIndex : text.length;
    ranges.push({ start, end });
    fences.lastIndex = end;
  }
  const ticks = /`+/g;
  let inline: RegExpExecArray | null;
  while ((inline = ticks.exec(text))) {
    const start = inline.index;
    if (ranges.some((range) => start >= range.start && start < range.end))
      continue;
    const marker = inline[0];
    // eslint-disable-next-line prefer-template -- SWC emits unescaped backticks when minifying this pattern in a template literal.
    const closing = new RegExp("(?<!`)" + marker + "(?!`)", "g");
    closing.lastIndex = ticks.lastIndex;
    const match = closing.exec(text);
    if (!match) continue;
    ranges.push({ start: inline.index, end: closing.lastIndex });
    ticks.lastIndex = closing.lastIndex;
  }
  return ranges;
};

export const findChatLinks = (text: string) => {
  const excluded = codeRanges(text);
  return find(text, { defaultProtocol: "https" }).flatMap((match) => {
    const href = getChatLinkHref(match.href);
    if (
      !href ||
      excluded.some(
        (range) => match.start < range.end && match.end > range.start,
      )
    )
      return [];
    return [{ start: match.start, end: match.end, text: match.value, href }];
  });
};

/** Preview the same destinations the message renderer exposes, in document order. */
export const extractMessageLinks = (
  text: string,
  richContent: JSONContent | null,
): ChatLink[] => {
  const links = new Map<string, ChatLink>();
  const add = (value: unknown) => {
    const href = getChatLinkHref(value);
    if (!href || links.has(href) || links.size >= MAX_MESSAGE_LINKS) return;
    links.set(href, { id: href, url: href, title: href });
  };
  const addText = (value: string) => {
    for (const link of findChatLinks(value)) add(link.href);
  };
  const visit = (node: JSONContent) => {
    if (
      links.size >= MAX_MESSAGE_LINKS ||
      node.type === "codeBlock" ||
      node.marks?.some((mark) => mark.type === "code")
    )
      return;
    if (node.type === "text") {
      const link = node.marks?.find((mark) => mark.type === "link");
      if (link) add(link.attrs?.href);
      else addText(node.text ?? "");
    }
    node.content?.forEach(visit);
  };
  if (richContent?.type === "doc") visit(richContent);
  else addText(text);
  return [...links.values()];
};

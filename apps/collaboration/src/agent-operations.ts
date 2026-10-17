import * as Y from "yjs";

export type AgentDocumentOperations = {
  version: 1;
  title?: string;
  replacements?: { before: string; after: string }[];
  appendParagraphs?: string[];
};

export class AgentEditError extends Error {
  constructor(
    readonly code: string,
    readonly status: "conflict" | "forbidden" | "failed" = "failed",
  ) {
    super(code);
  }
}

function record(value: unknown): value is Record<string, unknown> {
  return Boolean(value) && typeof value === "object" && !Array.isArray(value);
}

function plainText(
  value: unknown,
  maximum: number,
  allowEmpty = false,
): value is string {
  return (
    typeof value === "string" &&
    (allowEmpty || value.trim().length > 0) &&
    [...value].length <= maximum &&
    !/[\uD800-\uDFFF]/u.test(value) &&
    !/[\u0000-\u0008\u000b\u000c\u000e-\u001f]/u.test(value)
  );
}

export function parseAgentOperations(value: unknown): AgentDocumentOperations {
  if (
    !record(value) ||
    value.version !== 1 ||
    Object.keys(value).some(
      (key) =>
        !["version", "title", "replacements", "appendParagraphs"].includes(key),
    ) ||
    Buffer.byteLength(JSON.stringify(value)) > 65_536
  )
    throw new AgentEditError("invalid_operations");
  if (value.title !== undefined && !plainText(value.title, 255))
    throw new AgentEditError("invalid_title");
  if (
    value.replacements !== undefined &&
    (!Array.isArray(value.replacements) ||
      value.replacements.length > 32 ||
      value.replacements.some(
        (replacement) =>
          !record(replacement) ||
          Object.keys(replacement).some(
            (key) => !["before", "after"].includes(key),
          ) ||
          !plainText(replacement.before, 16_000) ||
          !plainText(replacement.after, 16_000, true),
      ))
  )
    throw new AgentEditError("invalid_replacements");
  if (
    value.appendParagraphs !== undefined &&
    (!Array.isArray(value.appendParagraphs) ||
      value.appendParagraphs.length > 32 ||
      value.appendParagraphs.some(
        (paragraph) =>
          !plainText(paragraph, 16_000) || /[\r\n]/u.test(paragraph),
      ))
  )
    throw new AgentEditError("invalid_paragraphs");
  if (
    value.title === undefined &&
    !(value.replacements as unknown[] | undefined)?.length &&
    !(value.appendParagraphs as unknown[] | undefined)?.length
  )
    throw new AgentEditError("empty_operations");
  return value as AgentDocumentOperations;
}

function textNodes(fragment: Y.XmlFragment | Y.XmlElement): Y.XmlText[] {
  return fragment.toArray().flatMap((node) => {
    if (node instanceof Y.XmlText) return [node];
    if (
      node instanceof Y.XmlElement &&
      !["image", "documentVideo", "mention", "documentReference"].includes(
        node.nodeName,
      )
    )
      return textNodes(node);
    return [];
  });
}

function textContent(node: Y.XmlText): string {
  return node
    .toDelta()
    .map((part: { insert: unknown }) =>
      typeof part.insert === "string" ? part.insert : "",
    )
    .join("");
}

function replaceText(document: Y.Doc, before: string, after: string) {
  const matches: { node: Y.XmlText; index: number }[] = [];
  for (const node of textNodes(document.getXmlFragment("default"))) {
    const text = textContent(node);
    let offset = 0;
    for (;;) {
      const index = text.indexOf(before, offset);
      if (index < 0) break;
      matches.push({ node, index });
      offset = index + 1;
    }
  }
  if (matches.length !== 1)
    throw new AgentEditError("text_not_unique", "conflict");
  if (before === after) return false;
  const { node, index } = matches[0];
  // Only change the differing span: untouched formatting and links retain their
  // original CRDT identity, and concurrent clients can merge unaffected edits.
  let prefix = 0;
  while (
    prefix < before.length &&
    prefix < after.length &&
    before[prefix] === after[prefix]
  )
    prefix++;
  // Yjs offsets are UTF-16 units, but a CRDT item must never be split inside
  // a surrogate pair: doing so replaces the retained half with U+FFFD.
  if (prefix > 0 && (before.codePointAt(prefix - 1) ?? 0) > 0xffff) prefix--;
  let suffix = 0;
  while (
    suffix < before.length - prefix &&
    suffix < after.length - prefix &&
    before[before.length - 1 - suffix] === after[after.length - 1 - suffix]
  )
    suffix++;
  if (
    suffix > 0 &&
    ((before.codePointAt(before.length - suffix - 1) ?? 0) > 0xffff ||
      (after.codePointAt(after.length - suffix - 1) ?? 0) > 0xffff)
  )
    suffix--;
  const start = index + prefix;
  const removed = before.length - prefix - suffix;
  const inserted = after.slice(prefix, after.length - suffix);
  let position = 0;
  let attributes: Record<string, unknown> | undefined;
  for (const part of node.toDelta()) {
    const length = typeof part.insert === "string" ? part.insert.length : 1;
    const overlaps =
      removed > 0
        ? position < start + removed && position + length > start
        : position < start && position + length > start;
    if (overlaps && part.attributes?.link)
      throw new AgentEditError("protected_reference", "conflict");
    if (position <= start && position + length > start)
      attributes = part.attributes;
    position += length;
  }
  node.delete(start, removed);
  if (inserted)
    node.insert(start, inserted, attributes?.link ? {} : attributes);
  return true;
}

/** Mutates only validated prose/title, keeping media, mentions and other blocks. */
export function applyAgentOperations(
  document: Y.Doc,
  input: unknown,
): string[] {
  const operations = parseAgentOperations(input);
  const changed = new Set<string>();
  document.transact(() => {
    if (operations.title !== undefined) {
      const title = document.getText("title");
      const next = operations.title.trim();
      if (title.toString() !== next) {
        title.delete(0, title.length);
        title.insert(0, next);
        changed.add("title");
      }
    }
    for (const replacement of operations.replacements ?? []) {
      if (replaceText(document, replacement.before, replacement.after))
        changed.add("content");
    }
    for (const body of operations.appendParagraphs ?? []) {
      const paragraph = new Y.XmlElement("paragraph");
      const text = new Y.XmlText();
      text.insert(0, body);
      paragraph.insert(0, [text]);
      document.getXmlFragment("default").push([paragraph]);
      changed.add("content");
    }
  });
  return [...changed];
}

import type { Node as ProseMirrorNode } from "@tiptap/pm/model";

export type CommentAnchorRange = { from: number; to: number };

export function matchesCommentQuote(
  document: ProseMirrorNode,
  range: CommentAnchorRange,
  quote: string,
) {
  return (
    Boolean(quote.trim()) &&
    range.from >= 1 &&
    range.to > range.from &&
    range.to <= document.content.size &&
    document.textBetween(range.from, range.to, "\n\n").trim() === quote.trim()
  );
}

/** Reattach only an exact, unique quote; ambiguous or removed passages stay detached. */
export function findCommentQuote(
  document: ProseMirrorNode,
  quote: string,
): CommentAnchorRange | null {
  const target = quote.trim();
  if (!target) return null;
  let text = "";
  let firstBlock = true;
  const spans: {
    start: number;
    end: number;
    from: number;
    to: number;
    isText: boolean;
  }[] = [];
  document.descendants((node, position) => {
    let value = "";
    if (node.isText) value = node.text ?? "";
    else if (node.isLeaf) value = node.type.spec.leafText?.(node) ?? "";
    // Mirror ProseMirror's textBetween block separators used when quoting.
    if (node.isBlock && ((node.isLeaf && value) || node.isTextblock)) {
      if (!firstBlock) {
        text += "\n\n";
      }
      firstBlock = false;
    }
    if (value)
      spans.push({
        start: text.length,
        end: text.length + value.length,
        from: position,
        to: position + node.nodeSize,
        isText: node.isText,
      });
    text += value;
  });
  const index = text.indexOf(target);
  if (index < 0 || text.includes(target, index + 1)) return null;
  const start = spans.find((span) => span.start <= index && span.end > index);
  const endIndex = index + target.length - 1;
  const end = spans.find(
    (span) => span.start <= endIndex && span.end > endIndex,
  );
  if (!start || !end) return null;
  const range = {
    from: start.from + (start.isText ? index - start.start : 0),
    to: end.isText ? end.from + endIndex - end.start + 1 : end.to,
  };
  return matchesCommentQuote(document, range, target) ? range : null;
}

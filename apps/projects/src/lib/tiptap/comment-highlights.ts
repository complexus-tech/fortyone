import { Extension, type Editor } from "@tiptap/core";
import { Plugin, PluginKey } from "@tiptap/pm/state";
import { Decoration, DecorationSet } from "@tiptap/pm/view";
import { findCommentQuote, matchesCommentQuote } from "./comment-anchors";

export type CommentHighlight = {
  from: number;
  id: string;
  to: number;
  quote?: string;
};

const commentHighlightsKey = new PluginKey<DecorationSet>(
  "documentCommentHighlights",
);
const commentDraftHighlightKey = new PluginKey<DecorationSet>(
  "documentCommentDraftHighlight",
);

const createHighlightPlugin = (key: PluginKey<DecorationSet>) =>
  new Plugin({
    key,
    props: {
      decorations: (state) => key.getState(state) ?? DecorationSet.empty,
    },
    state: {
      init: () => DecorationSet.empty,
      apply: (transaction, current) => {
        const highlights = transaction.getMeta(key) as
          | CommentHighlight[]
          | undefined;
        if (!highlights) {
          const mapped = current.map(transaction.mapping, transaction.doc);
          return mapped.remove(
            mapped.find().filter((decoration) => {
              const quote = decoration.spec.quote as string | undefined;
              return (
                quote !== undefined &&
                !matchesCommentQuote(transaction.doc, decoration, quote)
              );
            }),
          );
        }
        const maximumPosition = transaction.doc.content.size;
        return DecorationSet.create(
          transaction.doc,
          highlights
            .filter(
              ({ from, to }) => from > 0 && to > from && to <= maximumPosition,
            )
            .map(({ from, id, to, quote }) =>
              Decoration.inline(
                from,
                to,
                {
                  class: "document-comment-highlight",
                  "data-comment-thread-id": id,
                },
                { id, quote },
              ),
            ),
        );
      },
    },
  });

export const CommentHighlights = Extension.create({
  name: "documentCommentHighlights",
  addProseMirrorPlugins() {
    return [
      createHighlightPlugin(commentHighlightsKey),
      createHighlightPlugin(commentDraftHighlightKey),
    ];
  },
});

/** Prefer transaction-mapped anchors over stale numeric API positions. */
export const resolveCommentHighlight = (
  editor: Editor,
  highlight: CommentHighlight,
) => {
  const quote = highlight.quote;
  if (quote === undefined) return highlight;
  const mapped = commentHighlightsKey
    .getState(editor.state)
    ?.find(undefined, undefined, (spec) => spec.id === highlight.id);
  const current = mapped?.find((range) =>
    matchesCommentQuote(editor.state.doc, range, quote),
  );
  if (current) return { ...highlight, from: current.from, to: current.to };
  if (matchesCommentQuote(editor.state.doc, highlight, quote)) return highlight;
  const found = findCommentQuote(editor.state.doc, quote);
  return found ? { ...highlight, ...found } : null;
};

export const updateCommentHighlights = (
  editor: Editor,
  highlights: CommentHighlight[],
) => {
  editor.view.dispatch(
    editor.state.tr.setMeta(
      commentHighlightsKey,
      highlights
        .map((highlight) => resolveCommentHighlight(editor, highlight))
        .filter(
          (highlight): highlight is CommentHighlight => highlight !== null,
        ),
    ),
  );
};

/** Draft changes must not replace thread decorations already mapped by edits. */
export const updateCommentDraftHighlight = (
  editor: Editor,
  highlight: CommentHighlight | null,
) => {
  editor.view.dispatch(
    editor.state.tr.setMeta(
      commentDraftHighlightKey,
      highlight ? [highlight] : [],
    ),
  );
};

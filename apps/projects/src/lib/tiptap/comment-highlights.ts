import { Extension, type Editor } from "@tiptap/core";
import { Plugin, PluginKey } from "@tiptap/pm/state";
import { Decoration, DecorationSet } from "@tiptap/pm/view";

export type CommentHighlight = {
  from: number;
  id: string;
  to: number;
};

const commentHighlightsKey = new PluginKey<DecorationSet>(
  "documentCommentHighlights",
);

export const CommentHighlights = Extension.create({
  name: "documentCommentHighlights",
  addProseMirrorPlugins() {
    return [
      new Plugin({
        key: commentHighlightsKey,
        props: {
          decorations: (state) =>
            commentHighlightsKey.getState(state) ?? DecorationSet.empty,
        },
        state: {
          init: () => DecorationSet.empty,
          apply: (transaction, current) => {
            const highlights = transaction.getMeta(commentHighlightsKey) as
              | CommentHighlight[]
              | undefined;
            if (!highlights)
              return current.map(transaction.mapping, transaction.doc);
            const maximumPosition = transaction.doc.content.size;
            return DecorationSet.create(
              transaction.doc,
              highlights
                .filter(
                  ({ from, to }) =>
                    from > 0 && to > from && to <= maximumPosition,
                )
                .map(({ from, id, to }) =>
                  Decoration.inline(from, to, {
                    class: "document-comment-highlight",
                    "data-comment-thread-id": id,
                  }),
                ),
            );
          },
        },
      }),
    ];
  },
});

export const updateCommentHighlights = (
  editor: Editor,
  highlights: CommentHighlight[],
) => {
  editor.view.dispatch(
    editor.state.tr.setMeta(commentHighlightsKey, highlights),
  );
};

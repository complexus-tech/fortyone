import type { Editor } from "@tiptap/core";
import { act, render, waitFor } from "@testing-library/react";
import {
  EditorContent,
  NodeViewWrapper,
  ReactNodeViewRenderer,
  useEditor,
} from "@tiptap/react";
import StarterKit from "@tiptap/starter-kit";
import { DocumentMention } from "@fortyone/document-editor";
import { CommentHighlights } from "@/lib/tiptap/comment-highlights";
import type { DocumentCommentThread } from "./types";
import { useDocumentCommentInteractions } from "./use-document-comment-interactions";

const params = new URLSearchParams();
jest.mock("next/navigation", () => ({ useSearchParams: () => params }));
const closeRelatedWork = jest.fn();
const Mention = DocumentMention.extend({
  addNodeView: () =>
    ReactNodeViewRenderer(() => (
      <NodeViewWrapper as="span">@Alice</NodeViewWrapper>
    )),
});
const extensions = [StarterKit, CommentHighlights, Mention];
function Harness({
  threads,
  onReady,
}: {
  threads: DocumentCommentThread[];
  onReady: (editor: Editor) => void;
}) {
  const editor = useEditor({
    extensions,
    immediatelyRender: false,
    content:
      '<p>Quote <span data-type="mention" data-mention-kind="user" data-id="11111111-1111-4111-8111-111111111111" data-label="Alice">@Alice</span> end.</p>',
    onCreate: ({ editor: createdEditor }) => {
      onReady(createdEditor);
    },
  });
  useDocumentCommentInteractions({
    editor,
    commentThreads: threads,
    collaborationReady: true,
    closeRelatedWork,
  });
  return <EditorContent editor={editor} />;
}

test("metadata highlight refresh with React mention nodes has no flushSync lifecycle warning", async () => {
  const onReady = jest.fn((editor: Editor) => editor);
  const error = jest
    .spyOn(console, "error")
    .mockImplementation(() => undefined);
  try {
    const view = render(<Harness onReady={onReady} threads={[]} />);
    await waitFor(() => {
      expect(onReady).toHaveBeenCalled();
    });
    error.mockClear();
    const doc = onReady.mock.calls[0][0].state.doc;
    const thread = {
      id: "thread",
      anchorStart: 1,
      anchorEnd: doc.content.size - 1,
      quote: doc.textBetween(1, doc.content.size - 1, "\n\n").trim(),
      resolvedAt: null,
    } as DocumentCommentThread;
    await act(async () => {
      view.rerender(<Harness onReady={onReady} threads={[thread]} />);
    });
    expect(
      error.mock.calls.filter((call) => String(call[0]).includes("flushSync")),
    ).toHaveLength(0);
    expect(
      view.container.querySelector('[data-comment-thread-id="thread"]'),
    ).not.toBeNull();
  } finally {
    error.mockRestore();
  }
});

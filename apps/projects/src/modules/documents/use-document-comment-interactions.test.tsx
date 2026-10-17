import type { Editor } from "@tiptap/core";
import {
  act,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
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
  content = '<p>Quote <span data-type="mention" data-mention-kind="user" data-id="11111111-1111-4111-8111-111111111111" data-label="Alice">@Alice</span> end.</p>',
}: {
  threads: DocumentCommentThread[];
  onReady: (editor: Editor) => void;
  content?: string;
}) {
  const editor = useEditor({
    extensions,
    immediatelyRender: false,
    content,
    onCreate: ({ editor: createdEditor }) => {
      onReady(createdEditor);
    },
  });
  const comments = useDocumentCommentInteractions({
    editor,
    commentThreads: threads,
    collaborationReady: true,
    closeRelatedWork,
  });
  return (
    <>
      <EditorContent editor={editor} />
      <button onClick={comments.openComments} type="button">
        All comments
      </button>
      <button onClick={comments.closeComments} type="button">
        Close comments
      </button>
      <output aria-label="Comment scope">
        {comments.commentThreadScope?.join(",") ?? "all"}
      </output>
      <output aria-label="Comments open">
        {String(comments.isCommentsOpen)}
      </output>
      <output aria-label="Active thread">
        {comments.activeCommentThreadId ?? "none"}
      </output>
    </>
  );
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

test("highlight clicks scope comments to the same mapped passage without including partially overlapping sections", async () => {
  const onReady = jest.fn((editor: Editor) => editor);
  const content = "<p>Quote the first section.</p><p>Another section.</p>";
  const view = render(
    <Harness content={content} onReady={onReady} threads={[]} />,
  );
  await waitFor(() => {
    expect(onReady).toHaveBeenCalled();
  });
  const editor = onReady.mock.calls[0][0];
  const first = editor.state.doc.firstChild!;
  const firstStart = 1;
  const firstEnd = first.nodeSize - 1;
  const secondStart = first.nodeSize + 1;
  const secondEnd = editor.state.doc.content.size - 1;
  const createThread = (id: string, from: number, to: number) =>
    ({
      id,
      anchorStart: from,
      anchorEnd: to,
      quote: editor.state.doc.textBetween(from, to, "\n\n").trim(),
      resolvedAt: null,
    }) as DocumentCommentThread;
  const threads = [
    createThread("passage", firstStart, firstEnd),
    createThread("same-passage", firstStart, firstEnd),
    createThread("overlap", firstStart, firstStart + 5),
    createThread("unrelated", secondStart, secondEnd),
  ];
  await act(async () => {
    view.rerender(
      <Harness content={content} onReady={onReady} threads={threads} />,
    );
  });
  act(() => {
    editor.commands.insertContentAt(0, "<p>New introduction.</p>");
  });
  fireEvent.click(
    view.container.querySelector('[data-comment-thread-id="passage"]')!,
  );
  expect(screen.getByLabelText("Comment scope")).toHaveTextContent(
    "passage,same-passage",
  );
  expect(screen.getByLabelText("Comments open")).toHaveTextContent("true");
  expect(closeRelatedWork).toHaveBeenCalled();

  fireEvent.click(
    view.container.querySelector('[data-comment-thread-id="unrelated"]')!,
  );
  expect(screen.getByLabelText("Comment scope")).toHaveTextContent("unrelated");
  expect(screen.getByLabelText("Active thread")).toHaveTextContent("unrelated");
  fireEvent.click(screen.getByRole("button", { name: "Close comments" }));
  expect(screen.getByLabelText("Comments open")).toHaveTextContent("false");
  fireEvent.click(screen.getByRole("button", { name: "All comments" }));
  expect(screen.getByLabelText("Comment scope")).toHaveTextContent("all");
  expect(screen.getByLabelText("Active thread")).toHaveTextContent("none");
  expect(screen.getByLabelText("Comments open")).toHaveTextContent("true");
});

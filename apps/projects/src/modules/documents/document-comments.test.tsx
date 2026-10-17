import type * as ReactTypes from "react";
import type { Editor } from "@tiptap/react";
import { fireEvent, render, screen } from "@testing-library/react";
import { useDocumentCommentMutations, useDocumentComments } from "./hooks";
import { DocumentCommentsPanel } from "./document-comments";
import type { DocumentCommentThread } from "./types";

jest.mock("./use-document-mentions", () => ({
  useDocumentMentionRecipients: () => ({ data: { mayaId: "maya-id" } }),
}));
jest.mock("@/components/ui/maya-avatar", () => ({
  MayaAvatar: () => <span data-testid="maya-avatar" />,
}));
jest.mock("./document-comment-maya", () => ({
  useDocumentMayaRuns: () => ({
    data: { enabled: true, runs: [] },
    isError: false,
    refetch: jest.fn(),
  }),
  useDocumentMayaRefresh: jest.fn(),
  isDocumentMayaActive: (run: { status: string }) =>
    ["queued", "running"].includes(run.status),
}));
jest.mock("./hooks", () => ({
  useDocumentCommentMutations: jest.fn(),
  useDocumentComments: jest.fn(),
}));
jest.mock("./document-comment-composer", () => ({
  DocumentCommentComposer: () => <div>Selection composer</div>,
  DocumentCommentReplyComposer: ({
    replyToCommentId,
  }: {
    replyToCommentId?: string;
  }) => (
    <div data-reply-target={replyToCommentId} data-testid="reply-composer">
      Reply composer
    </div>
  ),
}));
jest.mock("icons", () => ({
  CloseIcon: () => null,
  Comment01Icon: () => null,
  ArrowDown2Icon: () => null,
  CheckIcon: () => null,
  QuoteIcon: () => null,
}));
jest.mock("ui", () => {
  const React = jest.requireActual<typeof ReactTypes>("react");
  return {
    Avatar: ({ name }: { name: string }) => (
      <span data-testid="avatar">{name}</span>
    ),
    Button: ({
      children,
      asIcon: _asIcon,
      leftIcon,
      color: _color,
      size: _size,
      variant: _variant,
      ...props
    }: ReactTypes.ButtonHTMLAttributes<HTMLButtonElement> & {
      asIcon?: boolean;
      leftIcon?: ReactTypes.ReactNode;
      color?: string;
      size?: string;
      variant?: string;
    }) =>
      React.createElement(
        "button",
        { type: "button", ...props },
        leftIcon,
        children,
      ),
  };
});
const thread: DocumentCommentThread = {
  id: "thread",
  documentId: "doc",
  quote: "Selected passage",
  anchorStart: 1,
  anchorEnd: 17,
  createdBy: "alice",
  createdAt: "2026-10-16T12:00:00Z",
  resolvedAt: null,
  resolvedBy: null,
  comments: [
    {
      id: "root",
      body: "Make **outcomes** concrete",
      createdBy: "alice",
      authorName: "Alice",
      authorAvatar: null,
      createdAt: "2026-10-16T12:00:00Z",
      updatedAt: "2026-10-16T12:00:00Z",
    },
    {
      id: "reply",
      body: "I agree with this",
      createdBy: "bob",
      authorName: "Bob",
      authorAvatar: null,
      createdAt: "2026-10-16T12:01:00Z",
      updatedAt: "2026-10-16T12:01:00Z",
    },
  ],
};
const resolve = jest.fn().mockResolvedValue(undefined);
const refetch = jest.fn();
const scrollIntoView = jest.fn();
const panel = (
  props: Partial<ReactTypes.ComponentProps<typeof DocumentCommentsPanel>> = {},
) => (
  <DocumentCommentsPanel
    documentId="doc"
    draft={null}
    editor={null}
    onClose={jest.fn()}
    onDraftChange={jest.fn()}
    {...props}
  />
);
beforeEach(() => {
  jest.clearAllMocks();
  Element.prototype.scrollIntoView = scrollIntoView;
  jest.mocked(useDocumentComments).mockReturnValue({
    data: [thread],
    isPending: false,
    isError: false,
    isFetching: false,
    refetch,
  } as unknown as ReturnType<typeof useDocumentComments>);
  jest.mocked(useDocumentCommentMutations).mockReturnValue({
    resolve: { mutateAsync: resolve, isPending: false },
    isThreadPending: () => false,
  } as unknown as ReturnType<typeof useDocumentCommentMutations>);
});

test("renders Markdown and keeps reply participants collapsed until expanded", () => {
  render(panel());
  expect(screen.getByText("outcomes").tagName).toBe("STRONG");
  expect(screen.queryByText("I agree with this")).not.toBeInTheDocument();
  expect(
    screen.getAllByTestId("avatar").map((node) => node.textContent),
  ).toEqual(["Alice", "Bob"]);
  fireEvent.click(screen.getByRole("button", { name: "1 reply" }));
  expect(screen.getByText("I agree with this")).toBeInTheDocument();
  expect(screen.getByRole("button", { name: "Hide replies" })).toHaveAttribute(
    "aria-expanded",
    "true",
  );
  fireEvent.click(screen.getByRole("button", { name: "Hide replies" }));
  expect(screen.queryByText("I agree with this")).not.toBeInTheDocument();
});

test("active threads expand and quote navigation both notifies selection and locates the exact passage", () => {
  const run = jest.fn();
  const chain = {
    focus: jest.fn().mockReturnThis(),
    setTextSelection: jest.fn().mockReturnThis(),
    scrollIntoView: jest.fn().mockReturnThis(),
    run,
  };
  const editor = {
    state: {
      doc: { content: { size: 100 }, textBetween: () => "Selected passage" },
    },
    chain: () => chain,
  } as unknown as Editor;
  const onThreadSelect = jest.fn();
  render(panel({ activeThreadId: "thread", editor, onThreadSelect }));
  expect(screen.getByText("I agree with this")).toBeInTheDocument();
  expect(scrollIntoView).toHaveBeenCalled();
  fireEvent.click(screen.getByRole("button", { name: "Selected passage" }));
  expect(onThreadSelect).toHaveBeenCalledWith(thread);
  expect(chain.setTextSelection).toHaveBeenCalledWith({ from: 1, to: 17 });
});

test("does not select a different passage after document text changes", () => {
  const chain = jest.fn();
  render(
    panel({
      editor: {
        state: {
          doc: {
            content: { size: 100 },
            textBetween: () => "Replacement",
            descendants: jest.fn(),
          },
        },
        chain,
      } as unknown as Editor,
    }),
  );
  fireEvent.click(screen.getByRole("button", { name: "Selected passage" }));
  expect(chain).not.toHaveBeenCalled();
  expect(screen.getByRole("alert")).toHaveTextContent(
    "This passage has changed",
  );
});

test("provides an explicit retry and does not misrepresent a load failure as an empty document", () => {
  jest.mocked(useDocumentComments).mockReturnValue({
    data: [],
    isPending: false,
    isError: true,
    isFetching: false,
    refetch,
  } as unknown as ReturnType<typeof useDocumentComments>);
  render(panel());
  expect(screen.queryByText("No open comments")).not.toBeInTheDocument();
  fireEvent.click(screen.getByRole("button", { name: "Retry" }));
  expect(refetch).toHaveBeenCalledTimes(1);
});

test("prevents optimistic or pending thread writes and hides mutations for readers", () => {
  jest.mocked(useDocumentCommentMutations).mockReturnValue({
    resolve: { mutateAsync: resolve, isPending: false },
    isThreadPending: () => true,
  } as unknown as ReturnType<typeof useDocumentCommentMutations>);
  const view = render(panel());
  expect(screen.getByRole("button", { name: "Reply" })).toBeDisabled();
  expect(screen.getByRole("button", { name: "Resolve" })).toBeDisabled();
  view.rerender(
    panel({
      canComment: false,
      draft: { from: 1, to: 17, text: "Selected passage" },
    }),
  );
  expect(
    screen.queryByRole("button", { name: "Resolve" }),
  ).not.toBeInTheDocument();
  expect(screen.queryByText("Selection composer")).not.toBeInTheDocument();
});

test("Maya replies and collapsed participants use her stable identity avatar", () => {
  jest.mocked(useDocumentComments).mockReturnValue({
    data: [
      {
        ...thread,
        comments: [
          thread.comments[0],
          {
            ...thread.comments[1],
            createdBy: "maya-id",
            authorName: "Agent display name",
          },
        ],
      },
    ],
    isPending: false,
    isError: false,
    isFetching: false,
    refetch,
  } as unknown as ReturnType<typeof useDocumentComments>);
  render(panel());
  expect(screen.getAllByTestId("maya-avatar")).toHaveLength(1);
  fireEvent.click(screen.getByRole("button", { name: "1 reply" }));
  expect(screen.getAllByTestId("maya-avatar")).toHaveLength(1);
  expect(screen.getByText("Agent display name")).toBeInTheDocument();
});

test("a reply being drafted keeps its target when a newer comment arrives", () => {
  const view = render(panel());
  fireEvent.click(screen.getByRole("button", { name: "Reply" }));
  expect(screen.getByTestId("reply-composer")).toHaveAttribute(
    "data-reply-target",
    "reply",
  );
  jest.mocked(useDocumentComments).mockReturnValue({
    data: [
      {
        ...thread,
        comments: [
          ...thread.comments,
          { ...thread.comments[1], id: "new-reply" },
        ],
      },
    ],
    isPending: false,
    isError: false,
    isFetching: false,
    refetch,
  } as unknown as ReturnType<typeof useDocumentComments>);
  view.rerender(panel());
  expect(screen.getByTestId("reply-composer")).toHaveAttribute(
    "data-reply-target",
    "reply",
  );
  fireEvent.click(screen.getByRole("button", { name: "Hide replies" }));
  fireEvent.click(screen.getByRole("button", { name: "Reply" }));
  expect(screen.getByTestId("reply-composer")).toHaveAttribute(
    "data-reply-target",
    "new-reply",
  );
});

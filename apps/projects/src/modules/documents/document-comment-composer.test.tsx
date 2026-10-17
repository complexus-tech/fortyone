import type * as NodeCrypto from "node:crypto";
import type * as ReactTypes from "react";
import type * as TiptapReact from "@tiptap/react";
import {
  act,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { DocumentCommentComposer } from "./document-comment-composer";
import { useDocumentCommentMutations } from "./hooks";
import type { DocumentCommentSelection } from "./types";

let currentEditor: TiptapReact.Editor | null = null;
jest.mock("./use-document-mentions", () => ({
  useDocumentMentions: jest.fn(),
}));
jest.mock("./hooks", () => ({ useDocumentCommentMutations: jest.fn() }));
jest.mock("@/lib/auth/client", () => ({
  useSession: () => ({ data: { user: { id: "alice" } } }),
}));
jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({ workspaceSlug: "first" }),
}));
jest.mock("@/shell/desktop/workspace-scope", () => ({
  useDesktopWorkspaceScope: () => null,
}));
jest.mock("@tiptap/react/menus", () => ({ BubbleMenu: () => null }));
jest.mock("icons", () => ({
  BoldIcon: () => null,
  ItalicIcon: () => null,
  CodeIcon: () => null,
  UnorderedListIcon: () => null,
}));
jest.mock("ui", () => {
  const React = jest.requireActual<typeof ReactTypes>("react");
  const { EditorContent } =
    jest.requireActual<typeof TiptapReact>("@tiptap/react");
  return {
    TextEditor: ({ editor }: { editor: TiptapReact.Editor }) => {
      currentEditor = editor;
      return <EditorContent editor={editor} />;
    },
    Button: ({
      children,
      asIcon: _asIcon,
      active: _active,
      color: _color,
      size: _size,
      variant: _variant,
      ...props
    }: ReactTypes.ButtonHTMLAttributes<HTMLButtonElement> & {
      asIcon?: boolean;
      active?: boolean;
      color?: string;
      size?: string;
      variant?: string;
    }) => React.createElement("button", { type: "button", ...props }, children),
  };
});
const create = jest.fn();
const draft: DocumentCommentSelection = {
  from: 1,
  to: 17,
  text: "Selected passage",
  draftKey: "stable-selection",
};
const onCreated = jest.fn();
const composer = (selection = draft) => (
  <DocumentCommentComposer
    documentId="doc"
    draft={selection}
    onCancel={jest.fn()}
    onCreated={onCreated}
  />
);
const type = async (html: string) => {
  await waitFor(() => {
    expect(screen.getByRole("textbox", { name: "Comment" })).toHaveAttribute(
      "contenteditable",
      "true",
    );
  });
  act(() => {
    currentEditor!.commands.setContent(html);
  });
};
beforeEach(() => {
  Object.defineProperty(crypto, "randomUUID", {
    configurable: true,
    value: jest.requireActual<typeof NodeCrypto>("node:crypto").randomUUID,
  });
  jest.clearAllMocks();
  localStorage.clear();
  currentEditor = null;
  create.mockResolvedValue({ id: "created-thread" });
  jest.mocked(useDocumentCommentMutations).mockReturnValue({
    create: { mutateAsync: create, isPending: false },
    isCreatingDraft: () => false,
  } as unknown as ReturnType<typeof useDocumentCommentMutations>);
});

test("preserves a failed rich-text draft across close and reopen, then clears only after successful retry", async () => {
  create.mockRejectedValueOnce(new Error("Offline"));
  const view = render(composer());
  await type("<p>Keep <strong>this draft</strong></p>");
  fireEvent.click(screen.getByRole("button", { name: "Send" }));
  await screen.findByRole("alert");
  const firstCreationId = create.mock.calls[0][0].creationId;
  expect(create).toHaveBeenLastCalledWith(
    expect.objectContaining({ body: "Keep **this draft**" }),
  );
  expect(onCreated).not.toHaveBeenCalled();
  expect(screen.getByRole("textbox")).toHaveTextContent("Keep this draft");
  view.unmount();
  render(composer());
  await waitFor(() => {
    expect(screen.getByRole("textbox")).toHaveTextContent("Keep this draft");
  });
  fireEvent.click(screen.getByRole("button", { name: "Send" }));
  await waitFor(() => {
    expect(onCreated).toHaveBeenCalledWith({ id: "created-thread" });
  });
  expect(localStorage.length).toBe(0);
  expect(create.mock.calls[1][0].creationId).toBe(firstCreationId);
});

test("submits with Cmd/Ctrl+Enter and synchronously prevents duplicate sends while locking editing", async () => {
  let complete: (value: unknown) => void = () => undefined;
  create.mockImplementation(
    () =>
      new Promise((resolve) => {
        complete = resolve;
      }),
  );
  render(composer());
  await type("<p>Submit once</p>");
  const textbox = screen.getByRole("textbox");
  fireEvent.keyDown(textbox, { key: "Enter", ctrlKey: true });
  fireEvent.submit(textbox.closest("form")!);
  await waitFor(() => {
    expect(create).toHaveBeenCalledTimes(1);
  });
  expect(textbox).toHaveAttribute("contenteditable", "false");
  expect(screen.getByRole("button", { name: "Sending…" })).toBeDisabled();
  expect(screen.getByRole("button", { name: "Cancel" })).toBeDisabled();
  await act(async () => {
    complete({ id: "created-thread" });
  });
  expect(onCreated).toHaveBeenCalledTimes(1);
});

test("a mapped selection retains the same draft and a detached anchor prevents sending", async () => {
  const view = render(composer());
  await type("<p>Still drafting</p>");
  view.rerender(composer({ ...draft, from: 20, to: 36 }));
  expect(screen.getByRole("textbox")).toHaveTextContent("Still drafting");
  view.rerender(composer({ ...draft, from: 20, to: 36, isDetached: true }));
  expect(screen.getByRole("button", { name: "Send" })).toBeDisabled();
  expect(screen.getByRole("alert")).toHaveTextContent(
    "The selected text has changed",
  );
  fireEvent.keyDown(screen.getByRole("textbox"), {
    key: "Enter",
    metaKey: true,
  });
  expect(create).not.toHaveBeenCalled();
});

test("formatting controls change real editor content and serialize Markdown", async () => {
  render(composer());
  await type("<p>Important</p>");
  act(() => {
    currentEditor!.commands.setTextSelection({ from: 1, to: 10 });
  });
  fireEvent.click(screen.getByRole("button", { name: "Bold" }));
  expect(screen.getByRole("textbox").querySelector("strong")).toHaveTextContent(
    "Important",
  );
  fireEvent.click(screen.getByRole("button", { name: "Send" }));
  await waitFor(() => {
    expect(create).toHaveBeenCalledWith(
      expect.objectContaining({ body: "**Important**" }),
    );
  });
});

test("clears a remounted locked composer when the original request succeeds", async () => {
  let complete: (value: unknown) => void = () => undefined;
  create.mockImplementation(
    () =>
      new Promise((resolve) => {
        complete = resolve;
      }),
  );
  const first = render(composer());
  await type("<p>One pending comment</p>");
  fireEvent.click(screen.getByRole("button", { name: "Send" }));
  await waitFor(() => {
    expect(create).toHaveBeenCalledTimes(1);
  });
  first.unmount();
  jest.mocked(useDocumentCommentMutations).mockReturnValue({
    create: { mutateAsync: create, isPending: false },
    isCreatingDraft: () => true,
  } as unknown as ReturnType<typeof useDocumentCommentMutations>);
  render(composer());
  await waitFor(() => {
    expect(screen.getByRole("textbox")).toHaveTextContent(
      "One pending comment",
    );
  });
  expect(screen.getByRole("button", { name: "Send" })).toBeDisabled();
  await act(async () => {
    complete({ id: "created-thread" });
  });
  await waitFor(() => {
    expect(screen.getByRole("textbox")).toHaveTextContent("");
    expect(screen.getByRole("textbox")).not.toHaveTextContent(
      "One pending comment",
    );
  });
  expect(onCreated).not.toHaveBeenCalled();
  expect(localStorage.length).toBe(0);
});

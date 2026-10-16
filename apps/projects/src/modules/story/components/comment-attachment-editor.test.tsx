/* global afterEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ReactNode } from "react";
import { act, fireEvent, render, screen } from "@testing-library/react";
import { Editor, EditorContent } from "@tiptap/react";
import Document from "@tiptap/extension-document";
import Paragraph from "@tiptap/extension-paragraph";
import Text from "@tiptap/extension-text";
import { CommentAttachment } from "./comment-attachment-extension";
import { COMMENT_ATTACHMENT_NODE } from "./comment-attachments";

jest.mock("player.style/sutro/react", () => ({
  __esModule: true,
  default: ({ children }: { children: ReactNode }) => <div>{children}</div>,
}));
jest.mock("@/hooks/owner", () => ({
  useIsAdminOrOwner: () => ({ isAdminOrOwner: false }),
}));
jest.mock("@/components/ui", () => ({ ConfirmDialog: () => null }));
jest.mock("@/lib/api-url", () => ({
  getApiUrl: () => "https://api.example.test",
}));

const FILE = {
  id: "7d3a62e4-a7cc-4489-99b6-a986195525c5",
  filename: "Review.pdf",
  url: "/attachments/review.pdf",
  mimeType: "application/pdf",
  size: 128,
  uploadedBy: "author",
  createdAt: "2026-10-16T12:00:00Z",
};
const editors: Editor[] = [];

const mountEditor = (editable: boolean) => {
  const editor = new Editor({
    editable,
    extensions: [Document, Paragraph, Text, CommentAttachment],
    content: {
      type: "doc",
      content: [
        { type: COMMENT_ATTACHMENT_NODE, attrs: FILE },
        { type: "paragraph" },
      ],
    },
  });
  editors.push(editor);
  render(<EditorContent editor={editor} />);
  return editor;
};

afterEach(() => {
  act(() => {
    editors.splice(0).forEach((editor) => {
      editor.destroy();
    });
  });
});

describe("comment attachment editor controls", () => {
  it("allows removal after draft recovery unlocks without emitting an editor update", async () => {
    const editor = mountEditor(false);
    const remove = await screen.findByLabelText(`Remove ${FILE.filename}`);
    const before = editor.getHTML();
    fireEvent.click(remove);
    expect(editor.getHTML()).toBe(before);
    const transaction = jest.fn();
    const update = jest.fn();
    editor.on("transaction", transaction);
    editor.on("update", update);

    act(() => {
      editor.setEditable(true, false);
    });

    expect(transaction).not.toHaveBeenCalled();
    expect(update).not.toHaveBeenCalled();
    expect(editor.view.dom).toHaveAttribute("contenteditable", "true");
    expect(screen.getByLabelText(`Remove ${FILE.filename}`)).toBe(remove);
    fireEvent.click(remove);
    expect(editor.getHTML()).toBe("<p></p>");
    expect(editor.isEmpty).toBe(true);
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });

  it("guards an existing control during submission and isolates removal from native editor selection", async () => {
    const editor = mountEditor(true);
    const remove = await screen.findByLabelText(`Remove ${FILE.filename}`);
    const before = editor.getHTML();
    act(() => {
      editor.setEditable(false, false);
    });
    expect(editor.view.dom).toHaveAttribute("contenteditable", "false");
    fireEvent.click(remove);
    expect(editor.getHTML()).toBe(before);

    act(() => {
      editor.setEditable(true, false);
    });
    const pointerDown = jest.fn();
    const mouseDown = jest.fn();
    editor.view.dom.addEventListener("pointerdown", pointerDown);
    editor.view.dom.addEventListener("mousedown", mouseDown);
    // The icon is the actual pointer target; ProseMirror does not recognize it
    // as a button and otherwise begins selecting the soon-to-be-deleted atom.
    const icon = remove.querySelector("svg")!;
    fireEvent.pointerDown(icon);
    fireEvent.mouseDown(icon, { button: 0 });
    expect(pointerDown).not.toHaveBeenCalled();
    expect(mouseDown).not.toHaveBeenCalled();
    fireEvent.mouseUp(icon, { button: 0 });
    fireEvent.click(icon);

    expect(editor.getHTML()).toBe("<p></p>");
    expect(editor.isEmpty).toBe(true);
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });
});

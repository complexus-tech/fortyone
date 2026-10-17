/* global afterEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type * as ReactTypes from "react";
import type { BubbleMenu } from "@tiptap/react/menus";
import { Editor } from "@tiptap/core";
import { createDocumentExtensions } from "@fortyone/document-editor";
import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
} from "@testing-library/react";
import { TextEditor } from "ui";

type BubbleMenuProps = ReactTypes.ComponentProps<typeof BubbleMenu>;
let mockBubbleMenuProps: BubbleMenuProps | undefined;

// Keep the real editor, selection policy, toolbar, and shared controls. Floating
// placement is tested in the browser; this adapter exposes its visibility policy.
jest.mock(
  require.resolve("@tiptap/react/menus", {
    paths: [`${process.cwd()}/../../packages/ui`],
  }),
  () => {
    const React = jest.requireActual<typeof ReactTypes>("react");
    return {
      BubbleMenu: (props: BubbleMenuProps) => {
        mockBubbleMenuProps = props;
        return React.createElement(
          "div",
          { "data-testid": "selection-toolbar" },
          props.children,
        );
      },
    };
  },
);
jest.mock("ui", () => ({
  ...jest.requireActual("../../../../../packages/ui/src/editor"),
}));

const editors: Editor[] = [];
const createEditor = (editable = false) => {
  const editor = new Editor({
    extensions: createDocumentExtensions(),
    content: "<p>Review this launch plan.</p>",
    editable,
  });
  editors.push(editor);
  return editor;
};

const selectBrowserPassage = (editor: Editor) => {
  const text = editor.view.dom.querySelector("p")?.firstChild;
  if (!text) throw new Error("The document paragraph was not mounted");
  const range = document.createRange();
  range.setStart(text, 0);
  range.setEnd(text, 6);
  act(() => {
    const selection = window.getSelection();
    selection?.removeAllRanges();
    selection?.addRange(range);
    document.dispatchEvent(new Event("selectionchange"));
  });
};

const selectPassage = (editor: Editor) => {
  act(() => {
    editor.commands.setTextSelection({ from: 1, to: 7 });
  });
  selectBrowserPassage(editor);
};

const toolbarShouldShow = (editor: Editor) => {
  const shouldShow = mockBubbleMenuProps?.shouldShow;
  if (!shouldShow) throw new Error("The selection toolbar was not mounted");
  return shouldShow({
    editor,
    element: screen.getByTestId("selection-toolbar"),
    from: editor.state.selection.from,
    to: editor.state.selection.to,
    state: editor.state,
    oldState: editor.state,
    view: editor.view,
  });
};

afterEach(() => {
  cleanup();
  editors.splice(0).forEach((editor) => {
    editor.destroy();
  });
  window.getSelection()?.removeAllRanges();
  mockBubbleMenuProps = undefined;
});

describe("document reviewer selection toolbar", () => {
  it("allows commenting on a read-only browser selection without editor focus", () => {
    const editor = createEditor();
    const onComment = jest.fn();
    const onCreate = jest.fn();
    render(
      <TextEditor
        bubbleMenuCommentAction={onComment}
        bubbleMenuCreateActions={[
          { id: "ticket", label: "Create ticket", onSelect: onCreate },
        ]}
        editor={editor}
      />,
    );
    expect(editor.state.selection.empty).toBe(true);
    selectBrowserPassage(editor);

    expect(editor.view.hasFocus()).toBe(false);
    expect(document.activeElement).toBe(document.body);
    expect(window.getSelection()?.toString()).toBe("Review");
    expect(editor.state.selection.from).toBe(1);
    expect(editor.state.selection.to).toBe(7);
    expect(toolbarShouldShow(editor)).toBe(true);
    expect(screen.getAllByRole("button")).toHaveLength(1);
    expect(
      screen.queryByRole("button", { name: "Bold" }),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Text style" }),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Create from selection" }),
    ).not.toBeInTheDocument();

    const comment = screen.getByRole("button", { name: "Comment" });
    const mouseDown = new MouseEvent("mousedown", {
      bubbles: true,
      cancelable: true,
    });
    fireEvent(comment, mouseDown);
    expect(mouseDown.defaultPrevented).toBe(true);
    fireEvent.click(comment);
    expect(onComment).toHaveBeenCalledWith({ from: 1, to: 7, text: "Review" });
    expect(onCreate).not.toHaveBeenCalled();
    expect(editor.getHTML()).toBe("<p>Review this launch plan.</p>");
  });

  it("does not offer a toolbar when the reviewer has no comment action", () => {
    const editor = createEditor();
    render(<TextEditor editor={editor} />);
    selectPassage(editor);
    expect(toolbarShouldShow(editor)).toBe(false);
    expect(screen.queryByRole("button")).not.toBeInTheDocument();
  });

  it("removes comment controls when permission changes on the mounted editor", () => {
    const editor = createEditor();
    const { rerender } = render(
      <TextEditor bubbleMenuCommentAction={jest.fn()} editor={editor} />,
    );
    selectPassage(editor);
    expect(toolbarShouldShow(editor)).toBe(true);

    rerender(<TextEditor editor={editor} />);
    expect(toolbarShouldShow(editor)).toBe(false);
    expect(
      screen.queryByRole("button", { name: "Comment" }),
    ).not.toBeInTheDocument();
  });

  it("immediately removes formatting when the same editor becomes comment-only", () => {
    const editor = createEditor(true);
    const onComment = jest.fn();
    const createActions = [
      { id: "ticket", label: "Create ticket", onSelect: jest.fn() },
    ];
    const { rerender } = render(
      <TextEditor
        allowFormatting
        bubbleMenuCommentAction={onComment}
        bubbleMenuCreateActions={createActions}
        editor={editor}
      />,
    );
    selectPassage(editor);
    expect(screen.getByRole("button", { name: "Bold" })).toBeInTheDocument();

    // React receives the new permissions before the page effect synchronizes
    // Tiptap's mutable editable flag. Formatting must disappear in that render.
    rerender(
      <TextEditor
        allowFormatting={false}
        bubbleMenuCommentAction={onComment}
        bubbleMenuCreateActions={createActions}
        editor={editor}
      />,
    );
    expect(editor.isEditable).toBe(true);
    expect(screen.getAllByRole("button")).toHaveLength(1);
    expect(
      screen.queryByRole("button", { name: "Bold" }),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Create from selection" }),
    ).not.toBeInTheDocument();

    act(() => {
      editor.setEditable(false, false);
    });
    selectPassage(editor);
    expect(toolbarShouldShow(editor)).toBe(true);
    fireEvent.click(screen.getByRole("button", { name: "Comment" }));
    expect(onComment).toHaveBeenCalledWith({ from: 1, to: 7, text: "Review" });
    expect(editor.getHTML()).toBe("<p>Review this launch plan.</p>");
  });

  it("hides the reviewer toolbar for a collapsed or unrelated browser selection", () => {
    const editor = createEditor();
    render(
      <>
        <p data-testid="outside-document">Other page content</p>
        <TextEditor bubbleMenuCommentAction={jest.fn()} editor={editor} />
      </>,
    );
    selectPassage(editor);
    expect(toolbarShouldShow(editor)).toBe(true);

    const outside = document.createRange();
    outside.selectNodeContents(screen.getByTestId("outside-document"));
    act(() => {
      window.getSelection()?.removeAllRanges();
      window.getSelection()?.addRange(outside);
      document.dispatchEvent(new Event("selectionchange"));
    });
    expect(editor.state.selection.empty).toBe(true);
    expect(toolbarShouldShow(editor)).toBe(false);

    selectBrowserPassage(editor);
    act(() => {
      window.getSelection()?.collapseToStart();
      document.dispatchEvent(new Event("selectionchange"));
    });
    expect(editor.state.selection.empty).toBe(true);
    expect(toolbarShouldShow(editor)).toBe(false);
  });

  it("retains formatting and creation controls for editable documents", () => {
    const editor = createEditor(true);
    render(
      <TextEditor
        bubbleMenuCommentAction={jest.fn()}
        bubbleMenuCreateActions={[
          { id: "ticket", label: "Create ticket", onSelect: jest.fn() },
        ]}
        editor={editor}
      />,
    );
    selectPassage(editor);
    act(() => {
      editor.view.focus();
    });
    expect(editor.view.hasFocus()).toBe(true);
    expect(toolbarShouldShow(editor)).toBe(true);
    expect(
      screen.getByRole("button", { name: "Text style" }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "Create from selection" }),
    ).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Bold" }));
    expect(editor.getHTML()).toBe(
      "<p><strong>Review</strong> this launch plan.</p>",
    );
  });
});

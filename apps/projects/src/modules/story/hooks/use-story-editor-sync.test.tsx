/* global afterEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */
import type { ReactNode } from "react";
import { Editor } from "@tiptap/core";
import Document from "@tiptap/extension-document";
import Paragraph from "@tiptap/extension-paragraph";
import Text from "@tiptap/extension-text";
import { act, renderHook, waitFor } from "@testing-library/react";
import {
  QueryClient,
  QueryClientProvider,
  useMutation,
} from "@tanstack/react-query";
import {
  useStoryEditorSaving,
  useStoryEditorSync,
} from "./use-story-editor-sync";

jest.mock("@/shell/desktop/desktop-outbox-coordinator", () => ({
  useDesktopOutbox: () => null,
}));

const editors: Editor[] = [];
const createEditor = () => {
  const editor = new Editor({
    extensions: [Document, Paragraph, Text],
    content: "<p>Original description</p>",
  });
  editors.push(editor);
  return editor;
};
const blur = (editor: Editor) => {
  editor.isFocused = false;
  editor.emit("blur", {
    editor,
    event: new FocusEvent("blur"),
    transaction: editor.state.tr,
  });
};

afterEach(() => {
  editors.splice(0).forEach((editor) => {
    editor.destroy();
  });
});

describe("story editor remote updates", () => {
  it("waits for both overlapping title and description saves", async () => {
    const client = new QueryClient();
    let finishTitle: () => void = () => undefined;
    let finishDescription: () => void = () => undefined;
    const title = new Promise<void>((resolve) => {
      finishTitle = resolve;
    });
    const description = new Promise<void>((resolve) => {
      finishDescription = resolve;
    });
    const { result, unmount } = renderHook(
      () => ({
        saving: useStoryEditorSaving("story"),
        mutation: useMutation({
          mutationFn: ({ field }: { storyId: string; field: string }) =>
            field === "title" ? title : description,
        }),
      }),
      {
        wrapper: ({ children }: { children: ReactNode }) => (
          <QueryClientProvider client={client}>{children}</QueryClientProvider>
        ),
      },
    );
    act(() => {
      result.current.mutation.mutate({
        storyId: "story",
        field: "description",
      });
      result.current.mutation.mutate({ storyId: "story", field: "title" });
    });
    await waitFor(() => {
      expect(result.current.saving).toBe(true);
    });
    await act(async () => {
      finishTitle();
      await title;
    });
    await waitFor(() => {
      expect(result.current.mutation.isPending).toBe(false);
    });
    expect(result.current.saving).toBe(true);
    await act(async () => {
      finishDescription();
      await description;
    });
    await waitFor(() => {
      expect(result.current.saving).toBe(false);
    });
    unmount();
    client.clear();
  });

  it("applies an external description after a focused reader blurs", () => {
    const editor = createEditor();
    const { rerender } = renderHook(
      ({ text }) =>
        useStoryEditorSync({
          editor,
          storyId: "story",
          text,
          html: `<p>${text}</p>`,
          saving: false,
        }),
      { initialProps: { text: "Original description" } },
    );
    editor.isFocused = true;
    rerender({ text: "Updated by Maya" });
    expect(editor.getText()).toBe("Original description");
    act(() => {
      blur(editor);
    });
    expect(editor.getText()).toBe("Updated by Maya");
  });

  it("preserves a local draft through optimistic changes, blur, and a failed save", () => {
    const editor = createEditor();
    const { rerender } = renderHook(
      ({ text, saving }) =>
        useStoryEditorSync({
          editor,
          storyId: "story",
          text,
          html: `<p>${text}</p>`,
          saving,
        }),
      { initialProps: { text: "Original description", saving: false } },
    );
    act(() => editor.commands.setContent("<p>My unsaved draft</p>"));
    rerender({ text: "My unsaved draft", saving: true });
    act(() => {
      blur(editor);
    });
    rerender({ text: "Original description", saving: false });
    expect(editor.getText()).toBe("My unsaved draft");
    rerender({ text: "Updated by Maya", saving: false });
    expect(editor.getText()).toBe("My unsaved draft");

    // Once the user's own save is acknowledged, later remote changes can flow.
    rerender({ text: "My unsaved draft", saving: false });
    rerender({ text: "Updated by Maya", saving: false });
    expect(editor.getText()).toBe("Updated by Maya");
  });

  it("updates a plain-text description and clears an empty description", () => {
    const editor = createEditor();
    const { rerender } = renderHook(
      ({ text }) =>
        useStoryEditorSync({
          editor,
          storyId: "story",
          text,
          html: "",
          saving: false,
        }),
      { initialProps: { text: "New plain text" } },
    );
    expect(editor.getText()).toBe("New plain text");
    rerender({ text: "" });
    expect(editor.isEmpty).toBe(true);
  });

  it("replaces a focused reused editor when switching stories", () => {
    const editor = createEditor();
    const { rerender } = renderHook(
      ({ storyId, text }) =>
        useStoryEditorSync({
          editor,
          storyId,
          text,
          html: `<p>${text}</p>`,
          saving: false,
        }),
      { initialProps: { storyId: "first", text: "Original description" } },
    );
    editor.isFocused = true;
    act(() => editor.commands.setContent("<p>First story draft</p>"));
    rerender({ storyId: "second", text: "Second story description" });
    expect(editor.getText()).toBe("Second story description");
  });
});

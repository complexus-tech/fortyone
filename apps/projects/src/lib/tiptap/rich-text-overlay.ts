import type { Editor } from "@tiptap/core";

export const getRichTextOverlayRoot = (editor: Editor) =>
  editor.view.dom.closest<HTMLElement>(".rich-document-editor") ??
  editor.view.dom.closest<HTMLElement>('[role="dialog"]') ??
  document.body;

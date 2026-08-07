/* global afterEach, describe, expect, it -- Jest globals are provided by the projects test runner. */

import type { Editor } from "@tiptap/core";
import { getRichTextOverlayRoot } from "./rich-text-overlay";

const createEditor = (dom: HTMLElement) =>
  ({ view: { dom } }) as unknown as Editor;

describe("getRichTextOverlayRoot", () => {
  afterEach(() => {
    document.body.replaceChildren();
  });

  it("keeps editor overlays inside the active rich-text surface", () => {
    const overlayRoot = document.createElement("div");
    const editorDom = document.createElement("div");
    overlayRoot.className = "rich-document-editor";
    overlayRoot.append(editorDom);
    document.body.append(overlayRoot);

    expect(getRichTextOverlayRoot(createEditor(editorDom))).toBe(overlayRoot);
  });

  it("falls back to the owning dialog when the editor has no rich-text wrapper", () => {
    const dialog = document.createElement("div");
    const editorDom = document.createElement("div");
    dialog.setAttribute("role", "dialog");
    dialog.append(editorDom);
    document.body.append(dialog);

    expect(getRichTextOverlayRoot(createEditor(editorDom))).toBe(dialog);
  });

  it("uses the document body outside an editor surface or dialog", () => {
    const editorDom = document.createElement("div");
    document.body.append(editorDom);

    expect(getRichTextOverlayRoot(createEditor(editorDom))).toBe(document.body);
  });
});

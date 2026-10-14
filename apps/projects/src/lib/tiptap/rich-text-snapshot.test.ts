/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { Editor, Node } from "@tiptap/core";
import { createDocumentExtensions } from "@fortyone/document-editor";
import {
  getPersistableRichTextContent,
  hasPendingRichTextMedia,
} from "./rich-text-media";
import { createRichTextSnapshot } from "./rich-text-snapshot";

describe("rich-text snapshots", () => {
  it("preserves formatted HTML and custom text serializers after the live editor changes and is destroyed", () => {
    const mention = Node.create({
      name: "snapshotMention",
      inline: true,
      group: "inline",
      atom: true,
      addAttributes: () => ({ label: { default: "Maya" } }),
      renderHTML: ({ node }) => ["span", {}, `@${node.attrs.label}`],
      renderText: ({ node }) => `@${node.attrs.label}`,
    });
    const editor = new Editor({
      extensions: [...createDocumentExtensions(), mention],
      content:
        "<h2>Heading</h2><p><strong>Bold</strong> and <em>italic</em></p>" +
        "<ul><li><p>First</p></li><li><p>Second</p></li></ul>" +
        "<table><tbody><tr><td><p>Cell</p></td></tr></tbody></table>",
    });
    editor.commands.insertContent({
      type: "snapshotMention",
      attrs: { label: "Maya" },
    });
    const expected = getPersistableRichTextContent(editor);
    const snapshot = createRichTextSnapshot(editor);

    editor.commands.setContent("A different ticket", { emitUpdate: false });
    editor.destroy();

    expect(getPersistableRichTextContent(snapshot)).toEqual(expected);
    expect(expected.contentText).toContain("@Maya");
    expect(expected.contentHtml).toContain("<strong>Bold</strong>");
    expect(expected.contentHtml).toContain("<table");
  });

  it("retains the queued pending-media state instead of reading a later document", () => {
    const editor = new Editor({
      extensions: createDocumentExtensions(),
      content: {
        type: "doc",
        content: [
          {
            type: "paragraph",
            content: [{ type: "text", text: "Draft" }],
          },
          {
            type: "image",
            attrs: {
              isUploading: true,
              src: "blob:pending-image",
              uploadId: "image-upload",
            },
          },
          {
            type: "documentVideo",
            attrs: {
              isUploading: true,
              src: "blob:pending-video",
              uploadId: "video-upload",
            },
          },
        ],
      },
    });
    const snapshot = createRichTextSnapshot(editor);
    editor.commands.setContent("Next ticket", { emitUpdate: false });
    editor.destroy();

    expect(hasPendingRichTextMedia(snapshot)).toBe(true);
    expect(getPersistableRichTextContent(snapshot)).toEqual({
      contentHtml: "<p>Draft</p>",
      contentText: "Draft",
    });
  });

  it("keeps empty content empty", () => {
    const editor = new Editor({
      extensions: createDocumentExtensions(),
      content: "",
    });
    const snapshot = createRichTextSnapshot(editor);
    editor.destroy();

    expect(getPersistableRichTextContent(snapshot)).toEqual({
      contentHtml: "",
      contentText: "",
    });
  });
});

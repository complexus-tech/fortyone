import { Editor } from "@tiptap/core";
import Document from "@tiptap/extension-document";
import Link from "@tiptap/extension-link";
import Paragraph from "@tiptap/extension-paragraph";
import Text from "@tiptap/extension-text";
import type { GoogleDriveFileReference } from "@/shared/google-drive/types";
import { replacePastedGoogleDriveURLLabel } from "./google-drive-description-link";

const rawURL = "https://docs.google.com/spreadsheets/d/sheet-1/edit";
const file = {
  availability: "available",
  createdAt: "2026-09-04T08:00:00Z",
  id: "reference-1",
  mimeType: "application/vnd.google-apps.spreadsheet",
  name: "Launch plan",
  targetId: "story-1",
  targetType: "story",
  updatedAt: "2026-09-04T08:00:00Z",
  webViewLink: rawURL,
} satisfies GoogleDriveFileReference;

describe("replacePastedGoogleDriveURLLabel", () => {
  it("relabels the closest pasted URL while preserving it as a link", () => {
    const editor = new Editor({
      content: `<p><a href="${rawURL}">${rawURL}</a></p>`,
      extensions: [Document, Paragraph, Text, Link],
    });

    expect(
      replacePastedGoogleDriveURLLabel({
        approximatePosition: 1,
        editor,
        file,
        rawURL,
      }),
    ).toBe(true);
    expect(editor.getText()).toBe("Launch plan");
    expect(editor.getHTML()).toContain('class="google-drive-smart-link"');
    expect(editor.getHTML()).toContain(`href="${rawURL}"`);

    editor.destroy();
  });
});

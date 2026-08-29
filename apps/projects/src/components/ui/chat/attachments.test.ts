/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { readFileSync } from "node:fs";
import { join } from "node:path";

const readSource = (path: string) =>
  readFileSync(join(process.cwd(), path), "utf8");

describe("Maya chat attachments", () => {
  it("renders the hidden file input required by react-dropzone", () => {
    const source = readSource("src/components/ui/chat/chat-input.tsx");

    expect(source).toContain("const { getInputProps, open } = useDropzone");
    expect(source).toContain("<input");
    expect(source).toContain("{...getInputProps({");
    expect(source).toContain("onClick={open}");
  });

  it("sends the AI SDK FileUIPart contract with the original filename", () => {
    const source = readSource("src/modules/maya/hooks/use-maya-chat.ts");

    expect(source).toContain('type: "file"');
    expect(source).toContain("filename: file.name");
    expect(source).toContain("url: await fileToBase64(file)");
    expect(source).not.toMatch(/\n\s+name: file\.name/);
  });

  it("renders clickable previews for sent images and PDFs", () => {
    const displaySource = readSource(
      "src/components/ui/chat/attachments-display.tsx",
    );
    const previewSource = readSource(
      "src/modules/story/components/story-attachment-preview.tsx",
    );

    expect(displaySource).toContain('part.mediaType.startsWith("image/")');
    expect(displaySource).not.toContain('startsWith("/image")');
    expect(displaySource).toContain('part.mediaType === "application/pdf"');
    expect(displaySource).toContain("<StoryAttachmentPreview");
    expect(previewSource).toContain("setIsOpen(true)");
    expect(previewSource).toContain("<ObjectViewer");
  });

  it("uses a theme-independent high-contrast remove button over images", () => {
    const previewSource = readSource(
      "src/modules/story/components/story-attachment-preview.tsx",
    );

    expect(previewSource).toContain("border-white/20 bg-black/75");
    expect(previewSource).toContain("hover:bg-black/90");
    expect(previewSource).toContain(
      '<CloseIcon className="h-4 text-white" strokeWidth={3} />',
    );
  });
});

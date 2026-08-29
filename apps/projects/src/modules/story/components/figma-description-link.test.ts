/* eslint-disable no-undef -- Jest globals are provided by the Projects test environment. */

import { Editor } from "@tiptap/core";
import Document from "@tiptap/extension-document";
import Link from "@tiptap/extension-link";
import Paragraph from "@tiptap/extension-paragraph";
import Text from "@tiptap/extension-text";
import type { FigmaArtifact } from "@/modules/settings/workspace/integrations/figma/types";
import {
  getFigmaLinkLabel,
  getStandaloneFigmaURL,
  replacePastedFigmaURLLabel,
} from "./figma-description-link";

const RAW_URL = "https://www.figma.com/design/file-key/file-name?node-id=12-34";

const artifact: FigmaArtifact = {
  canonicalUrl:
    "https://www.figma.com/design/file-key/file-name?node-id=12%3A34",
  fileKey: "file-key",
  fileName: "Product foundations",
  lastModified: null,
  nodeId: "12:34",
  nodeName: "Checkout flow",
  nodeType: "FRAME",
  originalUrl: RAW_URL,
  thumbnailUrl: null,
  version: null,
};

const createEditor = (content: string) =>
  new Editor({
    content,
    extensions: [Document, Paragraph, Text, Link],
  });

describe("Figma description links", () => {
  it("recognizes only standalone HTTP Figma URLs", () => {
    expect(getStandaloneFigmaURL(`  ${RAW_URL}\n`)).toBe(RAW_URL);
    expect(getStandaloneFigmaURL(`Design: ${RAW_URL}`)).toBeNull();
    expect(
      getStandaloneFigmaURL("https://figma.com.example/design/file-key"),
    ).toBeNull();
    expect(
      getStandaloneFigmaURL("ftp://www.figma.com/design/file-key"),
    ).toBeNull();
  });

  it("prefers a frame name and falls back to the file name", () => {
    expect(getFigmaLinkLabel(artifact)).toBe("Checkout flow");
    expect(getFigmaLinkLabel({ ...artifact, nodeName: null })).toBe(
      "Product foundations",
    );
  });

  it("replaces a raw pasted URL with the resolved name and favicon class", () => {
    const editor = createEditor(`<p>${RAW_URL}</p>`);

    expect(
      replacePastedFigmaURLLabel({
        approximatePosition: RAW_URL.length + 1,
        artifact,
        editor,
        rawURL: RAW_URL,
      }),
    ).toBe(true);

    expect(editor.getText()).toBe("Checkout flow");
    expect(editor.getHTML()).toContain('class="figma-smart-link"');
    expect(editor.getHTML()).toContain(`href="${artifact.canonicalUrl}"`);
    expect(editor.getHTML()).toContain('title="Checkout flow"');
    editor.destroy();
  });

  it("does not overwrite link text that the user has already changed", () => {
    const editor = createEditor(
      `<p><a href="${RAW_URL}">Existing design label</a></p>`,
    );

    expect(
      replacePastedFigmaURLLabel({
        approximatePosition: 1,
        artifact,
        editor,
        rawURL: RAW_URL,
      }),
    ).toBe(false);
    expect(editor.getText()).toBe("Existing design label");
    editor.destroy();
  });
});

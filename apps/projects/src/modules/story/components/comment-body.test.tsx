/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */
import { Editor } from "@tiptap/core";
import { createRef } from "react";
import Document from "@tiptap/extension-document";
import Paragraph from "@tiptap/extension-paragraph";
import Text from "@tiptap/extension-text";
import { render, screen } from "@testing-library/react";
import { CommentAttachment } from "./comment-attachment-extension";
import {
  COMMENT_ATTACHMENT_NODE,
  commentAttachmentUrl,
} from "./comment-attachments";
import { CommentBody } from "./comment-body";

jest.mock("@/lib/api-url", () => ({
  getApiUrl: () => "https://api.example.test",
}));
jest.mock("./comment-attachment-list", () => ({
  CommentAttachmentList: ({ files }: { files: { filename: string }[] }) => (
    <ul aria-label="Attachments">
      {files.map((file) => (
        <li key={file.filename}>{file.filename}</li>
      ))}
    </ul>
  ),
}));

describe("comment body", () => {
  it("anchors visibility to the first line of the reply body without affecting layout", () => {
    const visibilityRef = createRef<HTMLDivElement>();
    const { container } = render(
      <CommentBody
        html="<p>Maya's response</p>"
        visibilityRef={visibilityRef}
      />,
    );
    expect(visibilityRef.current).toHaveAttribute("aria-hidden", "true");
    expect(visibilityRef.current).toHaveClass("absolute", "top-0", "h-6");
    expect(visibilityRef.current?.parentElement).toBe(container.firstChild);
    expect(visibilityRef.current?.parentElement).toHaveTextContent(
      "Maya's response",
    );
  });

  it("extracts persisted image and file previews before sanitizing the text body", () => {
    const files = [
      {
        id: "7d3a62e4-a7cc-4489-99b6-a986195525c5",
        filename: "Screenshot.png",
        mimeType: "image/png",
      },
      {
        id: "8d3a62e4-a7cc-4489-99b6-a986195525c5",
        filename: "Plan.pdf",
        mimeType: "application/pdf",
      },
    ];
    const editor = new Editor({
      extensions: [Document, Paragraph, Text, CommentAttachment],
      content: {
        type: "doc",
        content: [
          {
            type: "paragraph",
            content: [{ type: "text", text: "**Review** these files" }],
          },
          ...files.map((file) => ({
            type: COMMENT_ATTACHMENT_NODE,
            attrs: {
              ...file,
              url: commentAttachmentUrl("acme", "story", file.id),
              size: 128,
              uploadedBy: "author",
              createdAt: "2026-10-16T12:00:00Z",
            },
          })),
        ],
      },
    });
    const { container } = render(<CommentBody html={editor.getHTML()} />);
    expect(screen.getByRole("list", { name: "Attachments" })).toHaveTextContent(
      "Screenshot.pngPlan.pdf",
    );
    expect(container.querySelector("strong")).toHaveTextContent("Review");
    expect(
      container.querySelector("[data-comment-attachment]"),
    ).not.toBeInTheDocument();
    editor.destroy();
  });
});

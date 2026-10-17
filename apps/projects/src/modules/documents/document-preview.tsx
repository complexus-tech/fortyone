"use client";

import { useEditor } from "@tiptap/react";
import { createDocumentExtensions } from "@fortyone/document-editor";
import { TextEditor } from "ui";

export const DocumentPreview = ({ html }: { html: string }) => {
  const editor = useEditor(
    {
      extensions: createDocumentExtensions(),
      content: html,
      editable: false,
      immediatelyRender: false,
    },
    [html],
  );
  return (
    <TextEditor
      className="rich-document-editor text-[1.1rem] leading-7"
      editor={editor}
      hideBubbleMenu
    />
  );
};

"use client";

import { EditorContent, useEditor } from "@tiptap/react";
import { createDocumentExtensions } from "@fortyone/document-editor";

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
    <EditorContent
      className="rich-document-editor prose prose-lg dark:prose-invert max-w-none"
      editor={editor}
    />
  );
};

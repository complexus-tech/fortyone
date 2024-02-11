"use client";

import { EditorContent, useEditor } from "@tiptap/react";
import StarterKit from "@tiptap/starter-kit";

export const TextEditor = ({ content = "" }: { content?: string }) => {
  const editor = useEditor({
    extensions: [StarterKit],
    content,
  });
  return (
    <EditorContent
      className="text-lg text-gray-250 dark:text-gray-200/80"
      editor={editor}
      placeholder="Issue description"
    />
  );
};

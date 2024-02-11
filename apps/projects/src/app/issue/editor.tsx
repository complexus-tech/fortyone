"use client";

import { useEditor, EditorContent } from "@tiptap/react";
import StarterKit from "@tiptap/starter-kit";

export const Editor = () => {
  const editor = useEditor({
    extensions: [StarterKit],
    content: `
    <p>Change the color of the button to red. This will hold the description of the issue. 
    It will be a long one so we can test the overflow of the container.</p>
    <p>Use the color palette from the design system to change the color of the button. The pallette is available in the design system documentation. Use the color palette from the design system to change the color of the button. 
    The pallette is available in the design system documentation🌎️.</p>`,
  });

  return (
    <EditorContent
      className="text-lg text-gray-250 dark:text-gray-200/80"
      editor={editor}
    />
  );
};

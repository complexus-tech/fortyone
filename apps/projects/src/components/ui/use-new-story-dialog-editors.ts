import type { Editor } from "@tiptap/core";
import Document from "@tiptap/extension-document";
import Paragraph from "@tiptap/extension-paragraph";
import Placeholder from "@tiptap/extension-placeholder";
import Text from "@tiptap/extension-text";
import { useEditor } from "@tiptap/react";
import { marked } from "marked";
import { createRichTextExtensions } from "@/lib/tiptap/rich-text-extensions";

export const useNewStoryDialogEditors = ({
  description,
  onMediaFiles,
  onMediaRequest,
  onStoryTitleChange,
  storyTerm,
}: {
  description?: string;
  onMediaFiles: (editor: Editor, files: File[]) => void;
  onMediaRequest: () => void;
  onStoryTitleChange: (title: string) => void;
  storyTerm: string;
}) => {
  const titleEditor = useEditor({
    extensions: [
      Document,
      Paragraph,
      Text,
      Placeholder.configure({ placeholder: "Enter title..." }),
    ],
    content: "",
    editable: true,
    autofocus: true,
    immediatelyRender: false,
    onUpdate: ({ editor }) => {
      onStoryTitleChange(editor.getText());
    },
  });

  const descriptionEditor = useEditor({
    extensions: createRichTextExtensions({
      onMediaFiles,
      onMediaRequest,
      placeholder: `${storyTerm} description — type / for commands`,
    }),
    content: marked.parse(description || "", { gfm: true }),
    editable: true,
    immediatelyRender: false,
  });

  return { descriptionEditor, titleEditor };
};

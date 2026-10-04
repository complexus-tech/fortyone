import type { Editor } from "@tiptap/core";
import { useEffect } from "react";
import Document from "@tiptap/extension-document";
import Paragraph from "@tiptap/extension-paragraph";
import Placeholder from "@tiptap/extension-placeholder";
import Text from "@tiptap/extension-text";
import { useEditor } from "@tiptap/react";
import { marked } from "marked";
import { createRichTextExtensions } from "@/lib/tiptap/rich-text-extensions";

export const useNewStoryDialogEditors = ({
  description,
  editable = true,
  onDescriptionChange,
  onMediaFiles,
  onMediaRequest,
  onStoryTitleChange,
  storyTerm,
}: {
  description?: string;
  editable?: boolean;
  onDescriptionChange?: (value: {
    description: string;
    descriptionHTML: string;
  }) => void;
  onMediaFiles?: (editor: Editor, files: File[]) => void;
  onMediaRequest?: () => void;
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
    editable,
    autofocus: true,
    immediatelyRender: false,
    editorProps: {
      attributes: { "aria-label": `${storyTerm} title`, role: "textbox" },
    },
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
    editable,
    immediatelyRender: false,
    editorProps: {
      attributes: {
        "aria-label": "Description",
        "aria-multiline": "true",
        role: "textbox",
      },
    },
    onUpdate: ({ editor }) => {
      onDescriptionChange?.({
        description: editor.getText(),
        descriptionHTML: editor.getHTML(),
      });
    },
  });

  useEffect(() => {
    titleEditor?.setEditable(editable, false);
    descriptionEditor?.setEditable(editable, false);
  }, [descriptionEditor, editable, titleEditor]);

  return { descriptionEditor, titleEditor };
};

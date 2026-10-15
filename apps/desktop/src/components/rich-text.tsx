import { useEffect } from "react";
import type { Doc } from "yjs";
import Collaboration from "@tiptap/extension-collaboration";
import { useEditor, EditorContent } from "@tiptap/react";
import {
  createDocumentExtensions,
  DocumentImage,
  DocumentVideo,
} from "@fortyone/document-editor";
import { createMediaNodeView, type MediaScope } from "../media-node-view";

export function RichText({
  html,
  onChange,
  readOnly = false,
  label = "Description",
  collaborationDocument,
  mediaScope,
}: {
  html: string;
  onChange: (html: string, text: string) => void;
  readOnly?: boolean;
  label?: string;
  collaborationDocument?: Doc;
  mediaScope: MediaScope;
}) {
  const editor = useEditor({
    extensions: [
      ...createDocumentExtensions({
        collaborative: Boolean(collaborationDocument),
        image: DocumentImage.extend({
          addNodeView() {
            return ({ node }) => createMediaNodeView(node, mediaScope, false);
          },
        }),
        video: DocumentVideo.extend({
          addNodeView() {
            return ({ node }) => createMediaNodeView(node, mediaScope, true);
          },
        }),
      }),
      ...(collaborationDocument
        ? [Collaboration.configure({ document: collaborationDocument })]
        : []),
    ],
    content: collaborationDocument ? undefined : html,
    editable: !readOnly,
    editorProps: {
      attributes: {
        "aria-label": label,
        role: "textbox",
        "aria-multiline": "true",
        class: "rich-text-content",
      },
    },
    onCreate: ({ editor: current }) => {
      if (collaborationDocument) onChange(current.getHTML(), current.getText());
    },
    onUpdate: ({ editor: current }) =>
      onChange(current.getHTML(), current.getText()),
  });
  useEffect(() => {
    editor?.setEditable(!readOnly);
  }, [editor, readOnly]);
  return (
    <div className="rich-text">
      {!readOnly ? (
        <div
          className="editor-toolbar"
          role="toolbar"
          aria-label="Text formatting"
        >
          <button
            type="button"
            aria-label="Bold"
            aria-pressed={editor?.isActive("bold") ?? false}
            onClick={() => editor?.chain().focus().toggleBold().run()}
          >
            <b>B</b>
          </button>
          <button
            type="button"
            aria-label="Italic"
            aria-pressed={editor?.isActive("italic") ?? false}
            onClick={() => editor?.chain().focus().toggleItalic().run()}
          >
            <i>I</i>
          </button>
          <button
            type="button"
            aria-label="Heading"
            onClick={() =>
              editor?.chain().focus().toggleHeading({ level: 2 }).run()
            }
          >
            H2
          </button>
          <button
            type="button"
            onClick={() => editor?.chain().focus().toggleBulletList().run()}
          >
            List
          </button>
          <button
            type="button"
            onClick={() => editor?.chain().focus().toggleTaskList().run()}
          >
            Checklist
          </button>
          <button
            type="button"
            onClick={() => editor?.chain().focus().undo().run()}
          >
            Undo
          </button>
        </div>
      ) : null}
      <EditorContent editor={editor} />
    </div>
  );
}

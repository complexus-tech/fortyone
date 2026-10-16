import type { Editor } from "@tiptap/core";
import type { NodeViewProps } from "@tiptap/react";
import type { DOMOutputSpec } from "@tiptap/pm/model";
import { Extension, Node } from "@tiptap/core";
import { NodeViewWrapper, ReactNodeViewRenderer } from "@tiptap/react";
import { Plugin } from "@tiptap/pm/state";
import type { StoryAttachment } from "../types";
import { CommentAttachmentList } from "./comment-attachment-list";
import {
  COMMENT_ATTACHMENT_ATTRIBUTE,
  COMMENT_ATTACHMENT_NODE,
  readCommentAttachment,
} from "./comment-attachments";

const AttachmentView = ({ node, editor, deleteNode }: NodeViewProps) => (
  <NodeViewWrapper className="not-prose" contentEditable={false}>
    <CommentAttachmentList
      className="max-w-none grid-cols-1"
      files={[node.attrs as StoryAttachment]}
      onRemove={() => {
        // Recovery and submission can change editability without rerendering.
        if (editor.isEditable) deleteNode();
      }}
    />
  </NodeViewWrapper>
);

export const CommentAttachment = Node.create({
  name: COMMENT_ATTACHMENT_NODE,
  group: "block",
  atom: true,
  selectable: true,
  draggable: true,
  addAttributes() {
    return Object.fromEntries(
      [
        "id",
        "filename",
        "url",
        "mimeType",
        "size",
        "createdAt",
        "uploadedBy",
      ].map((name) => [
        name,
        { default: name === "size" ? 0 : "", rendered: false },
      ]),
    );
  },
  parseHTML() {
    return [
      {
        tag: `div[${COMMENT_ATTACHMENT_ATTRIBUTE}]`,
        getAttrs: (element) => readCommentAttachment(element) ?? false,
      },
    ];
  },
  renderHTML({ node }) {
    const file = node.attrs as StoryAttachment;
    const attributes = {
      [COMMENT_ATTACHMENT_ATTRIBUTE]: file.id,
      "data-filename": file.filename,
      "data-url": file.url,
      "data-mime-type": file.mimeType,
      "data-size": file.size,
      "data-created-at": file.createdAt,
      "data-uploaded-by": file.uploadedBy,
    };
    const link: DOMOutputSpec = [
      "a",
      { href: file.url, target: "_blank", rel: "noopener noreferrer" },
      file.filename,
    ];
    if (file.mimeType.startsWith("image/")) {
      return [
        "div",
        attributes,
        ["img", { src: file.url, alt: file.filename }],
        link,
      ];
    }
    return ["div", attributes, link];
  },
  addNodeView() {
    return ReactNodeViewRenderer(AttachmentView, {
      className:
        "my-[3px] inline-block w-full align-top @min-[254px]/comment-composer:mr-[6px] @min-[254px]/comment-composer:w-[calc(50%-6px)] @min-[384px]/comment-composer:w-[calc(33.333333%-6px)] @min-[514px]/comment-composer:w-[calc(25%-6px)]",
    });
  },
});

type AttachmentDropOptions = {
  onFiles: (editor: Editor, files: File[]) => void;
};

export const setCommentAttachmentHandler = (
  editor: Editor,
  onFiles: AttachmentDropOptions["onFiles"],
) => {
  const storage = (
    editor.storage as unknown as {
      commentAttachmentDrop?: AttachmentDropOptions;
    }
  ).commentAttachmentDrop;
  if (storage) storage.onFiles = onFiles;
};

export const CommentAttachmentDrop = Extension.create<
  AttachmentDropOptions,
  AttachmentDropOptions
>({
  name: "commentAttachmentDrop",
  addOptions() {
    return { onFiles: () => undefined };
  },
  addStorage() {
    return { onFiles: this.options.onFiles };
  },
  addProseMirrorPlugins() {
    const editor = this.editor;
    const storage = this.storage;
    const handleFiles = (files: File[], event: Event) => {
      if (files.length === 0 || !editor.isEditable) return false;
      event.preventDefault();
      storage.onFiles(editor, files);
      return true;
    };
    return [
      new Plugin({
        props: {
          handlePaste: (_view, event) =>
            handleFiles(Array.from(event.clipboardData?.files ?? []), event),
          handleDrop: (_view, event) =>
            handleFiles(Array.from(event.dataTransfer?.files ?? []), event),
        },
      }),
    ];
  },
});

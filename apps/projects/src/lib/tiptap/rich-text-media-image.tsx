"use client";

import type { PointerEvent as ReactPointerEvent } from "react";
import { useEffect, useRef } from "react";
import { mergeAttributes, Node, type Editor } from "@tiptap/core";
import Image from "@tiptap/extension-image";
import {
  NodeViewWrapper,
  ReactNodeViewRenderer,
  type NodeViewProps,
} from "@tiptap/react";
import {
  AlignHorizontalCenterIcon,
  AlignLeftIcon,
  AlignRightIcon,
  DeleteIcon,
  LoadingIcon,
  MaximizeIcon,
} from "icons";
import { cn } from "lib";
import {
  clampNumber,
  DEFAULT_RICH_TEXT_IMAGE_ALIGNMENT,
  DEFAULT_RICH_TEXT_IMAGE_WIDTH,
  getRichTextImageStyle,
  getRichTextImageStyleString,
  MIN_RICH_TEXT_IMAGE_WIDTH_PX,
  normalizeRichTextImageAlignment,
  normalizeRichTextImageWidth,
  parseRichTextImageWidthToPixels,
  type RichTextImageAlignment,
} from "./rich-text-media-image-utils";

const IMAGE_ALIGNMENT_OPTIONS = [
  { icon: AlignLeftIcon, label: "Left", value: "left" },
  {
    icon: AlignHorizontalCenterIcon,
    label: "Center",
    value: "center",
  },
  { icon: AlignRightIcon, label: "Right", value: "right" },
] as const;

const selectNode = (editor: Editor, getPos: NodeViewProps["getPos"]) => {
  if (typeof getPos !== "function") return;
  const position = getPos();
  if (typeof position === "number") {
    editor.chain().focus().setNodeSelection(position).run();
  }
};

const MediaUploadOverlay = () => (
  <div className="bg-surface-elevated border-border-strong pointer-events-none absolute inset-0 flex items-center justify-center rounded-xl border">
    <span className="flex items-center gap-2 font-medium">
      <LoadingIcon className="size-4 animate-spin" />
      Uploading media…
    </span>
  </div>
);

const ImageAlignmentControls = ({
  alignment,
  editor,
  updateAttributes,
}: {
  alignment: RichTextImageAlignment;
  editor: Editor;
  updateAttributes: NodeViewProps["updateAttributes"];
}) => (
  <div
    className="border-border-strong bg-surface-elevated absolute -top-12 left-1/2 z-10 flex -translate-x-1/2 items-center gap-1 rounded-xl border-[0.5px] p-1 shadow-xl"
    contentEditable={false}
  >
    {IMAGE_ALIGNMENT_OPTIONS.map(({ icon: Icon, label, value }) => (
      <button
        aria-label={`Align image ${value}`}
        className={cn(
          "hover:bg-state-hover flex items-center gap-1.5 rounded-lg px-2.5 py-1 font-medium outline-none",
          { "bg-state-active": alignment === value },
        )}
        key={value}
        onClick={() => {
          updateAttributes({ align: value });
          editor.commands.focus();
        }}
        onMouseDown={(event) => {
          event.preventDefault();
        }}
        type="button"
      >
        <Icon className="h-4" strokeWidth={2} />
        {label}
      </button>
    ))}
    <span className="bg-border-strong mx-1 h-5 w-px" />
    <button
      aria-label="Delete image"
      className="text-danger hover:bg-state-hover rounded-lg p-2 outline-none"
      onClick={() => editor.chain().focus().deleteSelection().run()}
      onMouseDown={(event) => {
        event.preventDefault();
      }}
      type="button"
    >
      <DeleteIcon className="size-4" />
    </button>
  </div>
);

const ResizableRichTextImageView = ({
  editor,
  getPos,
  node,
  selected,
  updateAttributes,
}: NodeViewProps) => {
  const wrapperRef = useRef<HTMLDivElement | null>(null);
  const resizeCleanupRef = useRef<(() => void) | null>(null);
  const alignment = normalizeRichTextImageAlignment(node.attrs.align);
  const isUploading = node.attrs.isUploading === true;

  useEffect(
    () => () => {
      resizeCleanupRef.current?.();
    },
    [],
  );

  const handleResizeStart = (event: ReactPointerEvent<HTMLButtonElement>) => {
    event.preventDefault();
    event.stopPropagation();
    selectNode(editor, getPos);

    const wrapper = wrapperRef.current;
    if (!wrapper) return;
    const editorElement = editor.view.dom;
    const maxWidth = Math.max(editorElement.clientWidth, wrapper.clientWidth);
    const startWidth = parseRichTextImageWidthToPixels(
      node.attrs.width,
      maxWidth,
    );
    const startX = event.clientX;

    const handlePointerMove = (moveEvent: PointerEvent) => {
      const nextWidth = clampNumber(
        startWidth + moveEvent.clientX - startX,
        MIN_RICH_TEXT_IMAGE_WIDTH_PX,
        maxWidth,
      );
      updateAttributes({ width: `${Math.round(nextWidth)}px` });
    };
    const cleanup = () => {
      window.removeEventListener("pointermove", handlePointerMove);
      window.removeEventListener("pointerup", cleanup);
      window.removeEventListener("pointercancel", cleanup);
      resizeCleanupRef.current = null;
    };

    resizeCleanupRef.current?.();
    resizeCleanupRef.current = cleanup;
    window.addEventListener("pointermove", handlePointerMove);
    window.addEventListener("pointerup", cleanup);
    window.addEventListener("pointercancel", cleanup);
  };

  return (
    <NodeViewWrapper
      className="relative my-8 block max-w-full"
      onClick={() => {
        selectNode(editor, getPos);
      }}
      ref={wrapperRef}
      style={getRichTextImageStyle(node.attrs.width, alignment)}
    >
      {selected && !isUploading ? (
        <ImageAlignmentControls
          alignment={alignment}
          editor={editor}
          updateAttributes={updateAttributes}
        />
      ) : null}
      {/* eslint-disable-next-line @next/next/no-img-element -- TipTap node views require a native image element for editable sizing and selection. */}
      <img
        alt={(node.attrs.alt as string | null) ?? ""}
        className={cn(
          "border-border my-0 block h-auto w-full max-w-full rounded-xl border",
          { "border-primary border-2": selected && !isUploading },
        )}
        draggable={false}
        src={node.attrs.src as string}
        title={(node.attrs.title as string | null) ?? ""}
      />
      {isUploading ? <MediaUploadOverlay /> : null}
      {selected && !isUploading ? (
        <button
          aria-label="Resize image"
          className="border-border-strong bg-surface-elevated absolute -right-2 -bottom-2 flex size-8 cursor-se-resize items-center justify-center rounded-lg border shadow-lg"
          contentEditable={false}
          onPointerDown={handleResizeStart}
          type="button"
        >
          <MaximizeIcon className="size-4" />
        </button>
      ) : null}
    </NodeViewWrapper>
  );
};

const RichTextVideoView = ({
  editor,
  getPos,
  node,
  selected,
}: NodeViewProps) => {
  const isUploading = node.attrs.isUploading === true;
  return (
    <NodeViewWrapper
      className="relative my-5 block w-full max-w-full"
      onClick={() => {
        selectNode(editor, getPos);
      }}
    >
      {/* eslint-disable-next-line jsx-a11y/media-has-caption -- User-uploaded workspace videos do not have a caption-track upload contract. */}
      <video
        className={cn(
          "border-border my-0 aspect-video w-full rounded-xl border bg-black object-contain",
          { "border-primary border-2": selected && !isUploading },
        )}
        controls={!isUploading}
        preload="metadata"
        src={node.attrs.src as string}
      />
      {isUploading ? <MediaUploadOverlay /> : null}
    </NodeViewWrapper>
  );
};

export const RichTextImage = Image.extend({
  addAttributes() {
    return {
      ...this.parent?.(),
      width: {
        default: DEFAULT_RICH_TEXT_IMAGE_WIDTH,
        parseHTML: (element) =>
          normalizeRichTextImageWidth(
            element.getAttribute("data-width") || element.style.width,
          ),
        renderHTML: (attributes) => ({
          "data-width": normalizeRichTextImageWidth(attributes.width),
          style: getRichTextImageStyleString(
            attributes.width,
            attributes.align,
          ),
        }),
      },
      align: {
        default: DEFAULT_RICH_TEXT_IMAGE_ALIGNMENT,
        parseHTML: (element) =>
          normalizeRichTextImageAlignment(element.getAttribute("data-align")),
        renderHTML: (attributes) => ({
          "data-align": normalizeRichTextImageAlignment(attributes.align),
        }),
      },
      attachmentId: {
        default: null,
        parseHTML: (element) => element.getAttribute("data-attachment-id"),
        renderHTML: (attributes) =>
          typeof attributes.attachmentId === "string" &&
          attributes.attachmentId.trim()
            ? { "data-attachment-id": attributes.attachmentId }
            : {},
      },
      uploadId: {
        default: null,
        parseHTML: () => null,
        renderHTML: () => ({}),
      },
      isUploading: {
        default: false,
        parseHTML: () => false,
        renderHTML: () => ({}),
      },
    };
  },
  addNodeView() {
    return ReactNodeViewRenderer(ResizableRichTextImageView);
  },
});

export const RichTextVideo = Node.create({
  name: "documentVideo",
  group: "block",
  atom: true,
  draggable: true,
  selectable: true,
  addAttributes() {
    return {
      src: { default: null },
      attachmentId: {
        default: null,
        parseHTML: (element) => element.getAttribute("data-attachment-id"),
        renderHTML: (attributes) =>
          typeof attributes.attachmentId === "string" &&
          attributes.attachmentId.trim()
            ? { "data-attachment-id": attributes.attachmentId }
            : {},
      },
      uploadId: {
        default: null,
        parseHTML: () => null,
        renderHTML: () => ({}),
      },
      isUploading: {
        default: false,
        parseHTML: () => false,
        renderHTML: () => ({}),
      },
    };
  },
  parseHTML() {
    return [{ tag: "video[data-document-media-video]" }];
  },
  renderHTML({ HTMLAttributes }) {
    return [
      "video",
      mergeAttributes(HTMLAttributes, {
        class:
          "my-5 aspect-video w-full max-w-full rounded-xl border border-border bg-black object-contain",
        controls: "true",
        "data-document-media-video": "true",
        preload: "metadata",
      }),
    ];
  },
  addNodeView() {
    return ReactNodeViewRenderer(RichTextVideoView);
  },
});

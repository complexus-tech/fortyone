"use client";

import type { PointerEvent as ReactPointerEvent } from "react";
import { useEffect, useRef } from "react";
import {
  Extension,
  mergeAttributes,
  Node,
  type Editor,
  type JSONContent,
} from "@tiptap/core";
import Image from "@tiptap/extension-image";
import { DOMSerializer, type Node as ProseMirrorNode } from "@tiptap/pm/model";
import { Plugin } from "@tiptap/pm/state";
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
import type { DocumentMedia } from "./types";

export const DOCUMENT_MEDIA_ACCEPT =
  "image/jpeg,image/png,image/webp,video/mp4";

const MAX_DOCUMENT_MEDIA_SIZE_BYTES = 25 * 1024 * 1024;
const MIN_IMAGE_WIDTH_PX = 160;
const DEFAULT_IMAGE_WIDTH = "100%";
const DEFAULT_IMAGE_ALIGNMENT = "center";
const DOCUMENT_MEDIA_UPLOAD_TRANSACTION = "documentMediaUploadTransaction";
const DOCUMENT_MEDIA_RECONCILE_TRANSACTION =
  "documentMediaReconcileTransaction";
const SUPPORTED_DOCUMENT_MEDIA_TYPES = new Set([
  "image/jpeg",
  "image/png",
  "image/webp",
  "video/mp4",
]);

type ImageAlignment = "left" | "center" | "right";

const IMAGE_ALIGNMENT_OPTIONS = [
  { icon: AlignLeftIcon, label: "Left", value: "left" },
  {
    icon: AlignHorizontalCenterIcon,
    label: "Center",
    value: "center",
  },
  { icon: AlignRightIcon, label: "Right", value: "right" },
] as const;

type UploadDocumentMediaFilesOptions = {
  cleanup: (media: DocumentMedia) => Promise<void>;
  editor: Editor;
  files: File[];
  onError: (file: File, error: unknown) => void;
  position?: number;
  upload: (file: File) => Promise<DocumentMedia>;
};

const normalizeImageAlignment = (value: unknown): ImageAlignment =>
  value === "left" || value === "right" ? value : DEFAULT_IMAGE_ALIGNMENT;

const normalizeImageWidth = (value: unknown) => {
  if (typeof value !== "string") return DEFAULT_IMAGE_WIDTH;
  const width = value.trim();
  if (/^\d+(?:\.\d+)?(?:px|%)$/.test(width)) return width;
  return DEFAULT_IMAGE_WIDTH;
};

const getImageStyle = (width: unknown, alignment: unknown) => {
  const align = normalizeImageAlignment(alignment);
  return {
    display: "block",
    marginLeft: align === "left" ? "0" : "auto",
    marginRight: align === "right" ? "0" : "auto",
    maxWidth: "100%",
    width: normalizeImageWidth(width),
  };
};

const getImageStyleString = (width: unknown, alignment: unknown) => {
  const style = getImageStyle(width, alignment);
  return Object.entries(style)
    .map(([property, value]) => {
      const cssProperty = property.replace(
        /[A-Z]/g,
        (match) => `-${match.toLowerCase()}`,
      );
      return `${cssProperty}:${value}`;
    })
    .join(";");
};

const parseImageWidthToPixels = (width: unknown, maxWidth: number) => {
  const normalized = normalizeImageWidth(width);
  if (normalized.endsWith("%")) {
    return (Number.parseFloat(normalized) / 100) * maxWidth;
  }
  return Number.parseFloat(normalized);
};

const clamp = (value: number, minimum: number, maximum: number) =>
  Math.min(Math.max(value, minimum), maximum);

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
  alignment: ImageAlignment;
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

const ResizableDocumentImageView = ({
  editor,
  getPos,
  node,
  selected,
  updateAttributes,
}: NodeViewProps) => {
  const wrapperRef = useRef<HTMLDivElement | null>(null);
  const resizeCleanupRef = useRef<(() => void) | null>(null);
  const alignment = normalizeImageAlignment(node.attrs.align);
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
    const startWidth = parseImageWidthToPixels(node.attrs.width, maxWidth);
    const startX = event.clientX;

    const handlePointerMove = (moveEvent: PointerEvent) => {
      const nextWidth = clamp(
        startWidth + moveEvent.clientX - startX,
        MIN_IMAGE_WIDTH_PX,
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
      style={getImageStyle(node.attrs.width, alignment)}
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

const DocumentVideoView = ({
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

export const DocumentImage = Image.extend({
  addAttributes() {
    return {
      ...this.parent?.(),
      width: {
        default: DEFAULT_IMAGE_WIDTH,
        parseHTML: (element) =>
          normalizeImageWidth(
            element.getAttribute("data-width") || element.style.width,
          ),
        renderHTML: (attributes) => ({
          "data-width": normalizeImageWidth(attributes.width),
          style: getImageStyleString(attributes.width, attributes.align),
        }),
      },
      align: {
        default: DEFAULT_IMAGE_ALIGNMENT,
        parseHTML: (element) =>
          normalizeImageAlignment(element.getAttribute("data-align")),
        renderHTML: (attributes) => ({
          "data-align": normalizeImageAlignment(attributes.align),
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
    return ReactNodeViewRenderer(ResizableDocumentImageView);
  },
});

export const DocumentVideo = Node.create({
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
    return ReactNodeViewRenderer(DocumentVideoView);
  },
});

const createUploadId = () => crypto.randomUUID();

const resolvedUploads = new WeakMap<Editor, Map<string, DocumentMedia>>();
const failedUploads = new WeakMap<Editor, Set<string>>();

const getResolvedUploads = (editor: Editor) => {
  const existing = resolvedUploads.get(editor);
  if (existing) return existing;
  const created = new Map<string, DocumentMedia>();
  resolvedUploads.set(editor, created);
  return created;
};

const getFailedUploads = (editor: Editor) => {
  const existing = failedUploads.get(editor);
  if (existing) return existing;
  const created = new Set<string>();
  failedUploads.set(editor, created);
  return created;
};

const findMediaPositionByUploadId = (
  editor: Editor,
  uploadId: string,
): number | null => {
  const result: { position: number | null } = { position: null };
  editor.state.doc.descendants((node, currentPosition) => {
    if (node.attrs.uploadId === uploadId) {
      result.position = currentPosition;
      return false;
    }
    return result.position === null;
  });
  return result.position;
};

const updateMediaNodeByUploadId = (
  editor: Editor,
  uploadId: string,
  attributes: Record<string, unknown>,
) => {
  const position = findMediaPositionByUploadId(editor, uploadId);
  if (position === null || editor.isDestroyed) return false;
  const node = editor.state.doc.nodeAt(position);
  if (!node) return false;
  const transaction = editor.state.tr
    .setNodeMarkup(position, undefined, {
      ...node.attrs,
      ...attributes,
    })
    .setMeta("addToHistory", false)
    .setMeta(DOCUMENT_MEDIA_UPLOAD_TRANSACTION, true);
  editor.view.dispatch(transaction);
  return true;
};

const removeMediaNodeByUploadId = (editor: Editor, uploadId: string) => {
  const position = findMediaPositionByUploadId(editor, uploadId);
  if (position === null || editor.isDestroyed) return;
  const node = editor.state.doc.nodeAt(position);
  if (!node) return;
  const transaction = editor.state.tr
    .delete(position, position + node.nodeSize)
    .setMeta("addToHistory", false)
    .setMeta(DOCUMENT_MEDIA_UPLOAD_TRANSACTION, true);
  editor.view.dispatch(transaction);
};

const validateDocumentMediaFile = (file: File) => {
  if (!SUPPORTED_DOCUMENT_MEDIA_TYPES.has(file.type)) {
    throw new Error("Choose a JPEG, PNG, WebP, or MP4 file.");
  }
  if (file.size > MAX_DOCUMENT_MEDIA_SIZE_BYTES) {
    throw new Error("Media files cannot be larger than 25 MB.");
  }
};

const insertUploadingMedia = ({
  editor,
  file,
  position,
  previewUrl,
  uploadId,
}: {
  editor: Editor;
  file: File;
  position?: number;
  previewUrl: string;
  uploadId: string;
}) => {
  const media =
    file.type === "video/mp4"
      ? {
          type: "documentVideo",
          attrs: {
            attachmentId: null,
            isUploading: true,
            src: previewUrl,
            uploadId,
          },
        }
      : {
          type: "image",
          attrs: {
            align: DEFAULT_IMAGE_ALIGNMENT,
            attachmentId: null,
            isUploading: true,
            src: previewUrl,
            uploadId,
            width: DEFAULT_IMAGE_WIDTH,
          },
        };
  const content = [media, { type: "paragraph" }];
  const chain = editor
    .chain()
    .focus()
    .command(({ tr }) => {
      tr.setMeta(DOCUMENT_MEDIA_UPLOAD_TRANSACTION, true);
      return true;
    });
  return typeof position === "number"
    ? chain.insertContentAt(position, content).run()
    : chain.insertContent(content).run();
};

const hasUploadingDocumentMedia = (editor: Editor) => {
  let uploading = false;
  editor.state.doc.descendants((node) => {
    if (node.attrs.isUploading === true) uploading = true;
    return !uploading;
  });
  return uploading;
};

const getUploadingMediaIds = (document: Editor["state"]["doc"]) => {
  const uploadIds = new Set<string>();
  document.descendants((node) => {
    if (
      node.attrs.isUploading === true &&
      typeof node.attrs.uploadId === "string"
    ) {
      uploadIds.add(node.attrs.uploadId);
    }
  });
  return uploadIds;
};

const removeUploadingMediaFromJSON = (
  node: JSONContent,
): JSONContent | null => {
  if (
    (node.type === "image" || node.type === "documentVideo") &&
    node.attrs?.isUploading === true
  ) {
    return null;
  }

  if (!node.content) return node;

  return {
    ...node,
    content: node.content
      .map(removeUploadingMediaFromJSON)
      .filter((child): child is JSONContent => child !== null),
  };
};

const isEmptyDocumentContent = (content: JSONContent[] | undefined) =>
  !content?.length ||
  content.every((node) => node.type === "paragraph" && !node.content?.length);

export const getPersistableDocumentContent = (editor: Editor) => {
  if (!hasUploadingDocumentMedia(editor)) {
    return {
      contentHtml: editor.isEmpty ? "" : editor.getHTML(),
      contentText: editor.isEmpty ? "" : editor.getText(),
    };
  }

  const sanitizedJSON = removeUploadingMediaFromJSON(editor.getJSON());
  const content = sanitizedJSON?.content ?? [];
  if (!sanitizedJSON || isEmptyDocumentContent(content)) {
    return { contentHtml: "", contentText: "" };
  }

  const sanitizedDocument = editor.schema.nodeFromJSON({
    ...sanitizedJSON,
    content,
  });
  const container = window.document.createElement("div");
  container.appendChild(
    DOMSerializer.fromSchema(editor.schema).serializeFragment(
      sanitizedDocument.content,
      { document: window.document },
    ),
  );

  return {
    contentHtml: container.innerHTML,
    contentText: sanitizedDocument.textBetween(
      0,
      sanitizedDocument.content.size,
      "\n\n",
    ),
  };
};

export const uploadDocumentMediaFiles = async ({
  cleanup,
  editor,
  files,
  onError,
  position,
  upload,
}: UploadDocumentMediaFilesOptions) => {
  let nextPosition = position;
  const pendingUploads: Promise<void>[] = [];

  for (const file of files) {
    try {
      validateDocumentMediaFile(file);
    } catch (error) {
      onError(file, error);
      continue;
    }

    const previewUrl = URL.createObjectURL(file);
    const uploadId = createUploadId();
    const inserted = insertUploadingMedia({
      editor,
      file,
      position: nextPosition,
      previewUrl,
      uploadId,
    });
    if (!inserted) {
      URL.revokeObjectURL(previewUrl);
      onError(file, new Error("Could not insert this media file."));
      continue;
    }

    if (typeof nextPosition === "number") {
      const insertedPosition = findMediaPositionByUploadId(editor, uploadId);
      if (insertedPosition !== null) {
        const insertedNode = editor.state.doc.nodeAt(insertedPosition);
        nextPosition = insertedPosition + (insertedNode?.nodeSize ?? 1) + 2;
      }
    }

    pendingUploads.push(
      upload(file)
        .then(async (media) => {
          getResolvedUploads(editor).set(uploadId, media);
          const updated = updateMediaNodeByUploadId(editor, uploadId, {
            attachmentId: media.id,
            isUploading: false,
            src: media.url,
            uploadId: null,
          });
          if (!updated) {
            await cleanup(media);
            getResolvedUploads(editor).delete(uploadId);
            getFailedUploads(editor).add(uploadId);
            if (!editor.isDestroyed) {
              throw new Error("Could not finish inserting this media file.");
            }
          }
        })
        .catch((error: unknown) => {
          if (!getResolvedUploads(editor).has(uploadId)) {
            getFailedUploads(editor).add(uploadId);
          }
          removeMediaNodeByUploadId(editor, uploadId);
          onError(file, error);
        })
        .finally(() => {
          URL.revokeObjectURL(previewUrl);
        }),
    );
  }

  await Promise.all(pendingUploads);
};

type DocumentMediaDropOptions = {
  onFiles: (editor: Editor, files: File[], position?: number) => void;
};

export const DocumentMediaDrop = Extension.create<DocumentMediaDropOptions>({
  name: "documentMediaDrop",
  addOptions() {
    return {
      onFiles: () => undefined,
    };
  },
  addProseMirrorPlugins() {
    const editor = this.editor;
    const onFiles = this.options.onFiles;
    return [
      new Plugin({
        appendTransaction(transactions, _oldState, newState) {
          if (
            transactions.some((transaction) =>
              transaction.getMeta(DOCUMENT_MEDIA_RECONCILE_TRANSACTION),
            )
          ) {
            return null;
          }

          const resolved = getResolvedUploads(editor);
          const failed = getFailedUploads(editor);
          const replacements: {
            media: DocumentMedia;
            node: ProseMirrorNode;
            position: number;
          }[] = [];
          const removals: { nodeSize: number; position: number }[] = [];

          newState.doc.descendants((node, position) => {
            const uploadId = node.attrs.uploadId;
            if (typeof uploadId !== "string") return;
            const media = resolved.get(uploadId);
            if (media) {
              replacements.push({ media, node, position });
            } else if (failed.has(uploadId)) {
              removals.push({ nodeSize: node.nodeSize, position });
            }
          });

          if (replacements.length === 0 && removals.length === 0) return null;

          const transaction = newState.tr
            .setMeta("addToHistory", false)
            .setMeta(DOCUMENT_MEDIA_UPLOAD_TRANSACTION, true)
            .setMeta(DOCUMENT_MEDIA_RECONCILE_TRANSACTION, true);
          replacements.forEach(({ media, node, position }) => {
            transaction.setNodeMarkup(position, undefined, {
              ...node.attrs,
              attachmentId: media.id,
              isUploading: false,
              src: media.url,
              uploadId: null,
            });
          });
          removals
            .sort((left, right) => right.position - left.position)
            .forEach(({ nodeSize, position }) => {
              transaction.delete(position, position + nodeSize);
            });
          return transaction;
        },
        filterTransaction(transaction, state) {
          if (
            !transaction.docChanged ||
            transaction.getMeta(DOCUMENT_MEDIA_UPLOAD_TRANSACTION)
          ) {
            return true;
          }

          const uploadingBefore = getUploadingMediaIds(state.doc);
          if (uploadingBefore.size === 0) return true;
          const uploadingAfter = getUploadingMediaIds(transaction.doc);
          return [...uploadingBefore].every((uploadId) =>
            uploadingAfter.has(uploadId),
          );
        },
        props: {
          handleDrop(view, event) {
            const files = Array.from(event.dataTransfer?.files ?? []);
            if (files.length === 0) return false;
            event.preventDefault();
            const position = view.posAtCoords({
              left: event.clientX,
              top: event.clientY,
            })?.pos;
            onFiles(editor, files, position);
            return true;
          },
          handlePaste(_view, event) {
            const files = Array.from(event.clipboardData?.files ?? []);
            if (files.length === 0) return false;
            event.preventDefault();
            onFiles(editor, files);
            return true;
          },
        },
      }),
    ];
  },
});

"use client";

import { Extension, type Editor, type JSONContent } from "@tiptap/core";
import { DOMSerializer, type Node as ProseMirrorNode } from "@tiptap/pm/model";
import { Plugin } from "@tiptap/pm/state";
import {
  DEFAULT_RICH_TEXT_IMAGE_ALIGNMENT,
  DEFAULT_RICH_TEXT_IMAGE_WIDTH,
} from "./rich-text-media-image-utils";

export { RichTextImage, RichTextVideo } from "./rich-text-media-image";

export type RichTextMedia = {
  createdAt: string;
  filename: string;
  id: string;
  mimeType: string;
  size: number;
  uploadedBy: string;
  url: string;
};

export const RICH_TEXT_MEDIA_ACCEPT =
  "image/jpeg,image/png,image/webp,video/mp4";

const MAX_DOCUMENT_MEDIA_SIZE_BYTES = 25 * 1024 * 1024;
const RICH_TEXT_MEDIA_UPLOAD_TRANSACTION = "documentMediaUploadTransaction";
const RICH_TEXT_MEDIA_RECONCILE_TRANSACTION =
  "documentMediaReconcileTransaction";
const SUPPORTED_DOCUMENT_MEDIA_TYPES = new Set([
  "image/jpeg",
  "image/png",
  "image/webp",
  "video/mp4",
]);

type UploadRichTextMediaFilesOptions = {
  cleanup: (media: RichTextMedia) => Promise<void>;
  editor: Editor;
  files: File[];
  onError: (file: File, error: unknown) => void;
  position?: number;
  upload: (file: File) => Promise<RichTextMedia>;
};

const createUploadId = () => crypto.randomUUID();

const resolvedUploads = new WeakMap<Editor, Map<string, RichTextMedia>>();
const failedUploads = new WeakMap<Editor, Set<string>>();

const getResolvedUploads = (editor: Editor) => {
  const existing = resolvedUploads.get(editor);
  if (existing) return existing;
  const created = new Map<string, RichTextMedia>();
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
    .setMeta(RICH_TEXT_MEDIA_UPLOAD_TRANSACTION, true);
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
    .setMeta(RICH_TEXT_MEDIA_UPLOAD_TRANSACTION, true);
  editor.view.dispatch(transaction);
};

const validateRichTextMediaFile = (file: File) => {
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
            align: DEFAULT_RICH_TEXT_IMAGE_ALIGNMENT,
            attachmentId: null,
            isUploading: true,
            src: previewUrl,
            uploadId,
            width: DEFAULT_RICH_TEXT_IMAGE_WIDTH,
          },
        };
  const content = [media, { type: "paragraph" }];
  const chain = editor
    .chain()
    .focus()
    .command(({ tr }) => {
      tr.setMeta(RICH_TEXT_MEDIA_UPLOAD_TRANSACTION, true);
      return true;
    });
  return typeof position === "number"
    ? chain.insertContentAt(position, content).run()
    : chain.insertContent(content).run();
};

export const hasPendingRichTextMedia = (editor: Editor) => {
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

export const clearRichTextContent = (editor: Editor) =>
  editor
    .chain()
    .command(({ tr }) => {
      tr.setMeta(RICH_TEXT_MEDIA_UPLOAD_TRANSACTION, true);
      return true;
    })
    .clearContent()
    .run();

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

export const getPersistableRichTextContent = (editor: Editor) => {
  if (!hasPendingRichTextMedia(editor)) {
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

export const uploadRichTextMediaFiles = async ({
  cleanup,
  editor,
  files,
  onError,
  position,
  upload,
}: UploadRichTextMediaFilesOptions) => {
  let nextPosition = position;
  const pendingUploads: Promise<void>[] = [];

  for (const file of files) {
    try {
      validateRichTextMediaFile(file);
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

type RichTextMediaDropOptions = {
  onFiles: (editor: Editor, files: File[], position?: number) => void;
};

export const RichTextMediaDrop = Extension.create<RichTextMediaDropOptions>({
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
              transaction.getMeta(RICH_TEXT_MEDIA_RECONCILE_TRANSACTION),
            )
          ) {
            return null;
          }

          const resolved = getResolvedUploads(editor);
          const failed = getFailedUploads(editor);
          const replacements: {
            media: RichTextMedia;
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
            .setMeta(RICH_TEXT_MEDIA_UPLOAD_TRANSACTION, true)
            .setMeta(RICH_TEXT_MEDIA_RECONCILE_TRANSACTION, true);
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
            transaction.getMeta(RICH_TEXT_MEDIA_UPLOAD_TRANSACTION)
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

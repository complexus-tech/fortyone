import { BlurImage, Box, Button } from "ui";
import { CloseIcon, LoadingIcon, ProductTourIcon as PlayIcon } from "icons";
import type { StoryAttachment } from "../types";
import { CommentFileIcon } from "./comment-file-icon";
import {
  formatAttachmentSize,
  getAttachmentTypeLabel,
} from "./attachment-file-info";

export const CommentAttachmentCard = ({
  file,
  canPreview,
  isUploading,
  onOpen,
  onRemove,
}: {
  file: StoryAttachment;
  canPreview: boolean;
  isUploading: boolean;
  onOpen: () => void;
  onRemove?: () => void;
}) => {
  const isImage = file.mimeType.startsWith("image/");
  const isVideo = file.mimeType.startsWith("video/");
  return (
    <Box className="group border-border bg-surface relative h-[72px] min-w-0 overflow-hidden rounded-lg border">
      <button
        aria-label={`${canPreview ? "Preview" : "Download"} ${file.filename}`}
        className="focus-visible:ring-primary flex h-full w-full min-w-0 items-center text-left outline-none focus-visible:ring-2 focus-visible:ring-inset disabled:opacity-60"
        disabled={isUploading}
        onClick={onOpen}
        title={`${file.filename} · ${getAttachmentTypeLabel(file)} · ${formatAttachmentSize(file.size)}`}
        type="button"
      >
        {file.url && isImage ? (
          <BlurImage
            alt={file.filename}
            className="bg-surface-muted/50 h-full w-full"
            imageClassName="object-contain"
            src={file.url}
            unoptimized
          />
        ) : null}
        {file.url && isVideo ? (
          <video
            aria-label={`${file.filename} thumbnail`}
            className="bg-surface-muted/50 h-full w-full object-contain"
            controls={false}
            muted
            playsInline
            preload="metadata"
            src={file.url}
          />
        ) : null}
        {!file.url || (!isImage && !isVideo) ? (
          <span className="flex w-full min-w-0 flex-col items-center gap-1 px-2 py-2">
            <CommentFileIcon
              className="size-[28px] shrink-0"
              filename={file.filename}
              mimeType={file.mimeType}
            />
            <span className="block w-full min-w-0 truncate text-center text-[12px] leading-[16px]">
              {file.filename}
            </span>
          </span>
        ) : null}
      </button>
      {file.url && isVideo ? (
        <span
          aria-hidden
          className="pointer-events-none absolute bottom-2 left-2 rounded-full bg-black/65 p-1 text-white"
        >
          <PlayIcon className="size-5 text-white" strokeWidth={1.75} />
        </span>
      ) : null}
      {isUploading ? (
        <LoadingIcon
          aria-label="Uploading attachment"
          className="absolute top-2 right-2 size-4 animate-spin"
        />
      ) : null}
      {onRemove ? (
        <Button
          aria-label={`Remove ${file.filename}`}
          asIcon
          className="border-border bg-surface/95 absolute top-1.5 right-1.5 border shadow-sm [.ProseMirror[contenteditable=false]_&]:hidden"
          color="tertiary"
          onClick={(event) => {
            event.preventDefault();
            event.stopPropagation();
            onRemove();
          }}
          onMouseDownCapture={(event) => {
            event.stopPropagation();
          }}
          onPointerDownCapture={(event) => {
            event.stopPropagation();
          }}
          rounded="full"
          size="xs"
          variant="naked"
        >
          <CloseIcon className="size-3.5" />
        </Button>
      ) : null}
    </Box>
  );
};

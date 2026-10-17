"use client";

import { useState } from "react";
import { Video02Icon } from "icons";
import { cn } from "lib";
import { ThumbnailActions } from "@/components/ui/attachment-preview/thumbnail-actions";
import { CachedAttachmentThumbnail } from "@/components/ui/attachment-preview/cached-thumbnail";
import { PdfThumbnail } from "@/components/ui/attachment-preview/pdf-thumbnail";
import { getPreviewMediaKind } from "@/components/ui/attachment-preview/types";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { CommentFileIcon } from "@/modules/story/components/comment-file-icon";
import { getAttachmentTypeLabel } from "@/modules/story/components/attachment-file-info";
import type { ChatAttachment, ChatMessage } from "./types";
import {
  formatFileSize,
  getChatAttachmentThumbnailUrl,
  getChatPdfPreviewUrl,
} from "./attachment-url";
import { ChatAudioAttachment } from "./chat-audio-attachment";
import { useChatAttachmentPreview } from "./attachment-preview-context";

const PREVIEW_BUTTON_CLASS =
  "border-border bg-surface dark:bg-[color-mix(in_oklab,var(--color-background)_65%,var(--color-surface))] hover:border-border-strong focus-visible:ring-primary/40 overflow-hidden rounded-lg border text-left outline-none transition-colors focus-visible:ring-2";

type AttachmentPreview = (
  attachmentId: string,
  trigger: HTMLButtonElement,
) => void;

const ImageAttachment = ({
  attachment,
  url,
  single,
  remaining,
  onPreview,
}: {
  attachment: ChatAttachment;
  url: string;
  single: boolean;
  remaining: number;
  onPreview: AttachmentPreview;
}) => {
  const [imageSize, setImageSize] = useState<{
    url: string;
    width: number;
    height: number;
  } | null>(null);
  const dimensions = imageSize?.url === url ? imageSize : null;
  const imageWidth = dimensions
    ? Math.min(
        480,
        dimensions.width,
        (360 * dimensions.width) / dimensions.height,
      )
    : 480;

  return (
    <div
      className={cn(
        "group/attachment relative max-w-full min-w-0",
        single && "w-fit",
      )}
    >
      <button
        aria-label={
          remaining
            ? `Preview ${attachment.fileName} and ${remaining} more images`
            : `Preview ${attachment.fileName}`
        }
        className={cn(
          PREVIEW_BUTTON_CLASS,
          "relative isolate flex min-w-0 items-center justify-center p-0",
          single ? "max-h-[360px] max-w-[480px]" : "aspect-[4/3] w-full",
        )}
        onClick={(event) => {
          onPreview(attachment.id, event.currentTarget);
        }}
        style={
          single
            ? {
                width: imageWidth,
                maxWidth: "100%",
                aspectRatio: dimensions
                  ? `${dimensions.width} / ${dimensions.height}`
                  : "4 / 3",
              }
            : undefined
        }
        type="button"
      >
        <CachedAttachmentThumbnail
          className="absolute inset-0 aspect-auto h-full w-full"
          fileName={attachment.fileName}
          imageClassName={single ? "object-contain" : "object-cover"}
          onImageLoad={(width, height) => {
            if (single && width && height) setImageSize({ url, width, height });
          }}
          url={url}
        />
        {remaining ? (
          <span
            aria-hidden
            className="bg-background/70 text-foreground pointer-events-none absolute inset-0 flex items-center justify-center rounded-[inherit] text-xl font-semibold backdrop-blur-[2px]"
          >
            +{remaining}
          </span>
        ) : null}
      </button>
      <div className="bg-surface absolute top-1.5 right-1.5 rounded-md opacity-0 transition-opacity group-hover/attachment:opacity-100 focus-within:opacity-100">
        <ThumbnailActions url={url} />
      </div>
    </div>
  );
};

const FileAttachment = ({
  attachment,
  expanded,
  paired,
  url,
  thumbnailUrl,
  onPreview,
}: {
  attachment: ChatAttachment;
  expanded: boolean;
  paired: boolean;
  url: string;
  thumbnailUrl: string;
  onPreview: AttachmentPreview;
}) => (
  <div className="group/attachment relative w-fit max-w-full">
    <button
      aria-label={`Preview ${attachment.fileName}`}
      className={cn(
        PREVIEW_BUTTON_CLASS,
        "w-[160px] max-w-full align-top",
        expanded && !paired && "w-[360px]",
        expanded && paired && "w-[300px]",
      )}
      onClick={(event) => {
        onPreview(attachment.id, event.currentTarget);
      }}
      type="button"
    >
      <span className="flex min-w-0 items-center gap-2 px-2.5 py-2.5">
        <CommentFileIcon
          className="size-8"
          filename={attachment.fileName}
          mimeType={attachment.mediaType}
        />
        <span className="min-w-0 flex-1">
          <span className="block truncate font-medium">
            {attachment.fileName}
          </span>
          <span className="text-text-muted mt-0.5 block truncate">
            {getAttachmentTypeLabel({
              mimeType: attachment.mediaType,
              filename: attachment.fileName,
            })}
            {" · "}
            {formatFileSize(attachment.sizeBytes)}
          </span>
        </span>
      </span>
      {expanded ? (
        <PdfThumbnail
          fileName={attachment.fileName}
          thumbnailUrl={thumbnailUrl}
          url={url}
        />
      ) : null}
    </button>
    {expanded ? (
      <div className="bg-surface absolute top-1.5 right-1.5 rounded-md opacity-0 transition-opacity group-hover/attachment:opacity-100 focus-within:opacity-100">
        <ThumbnailActions url={thumbnailUrl} />
      </div>
    ) : null}
  </div>
);

const VideoAttachment = ({
  attachment,
  onPreview,
}: {
  attachment: ChatAttachment;
  onPreview: AttachmentPreview;
}) => (
  <button
    aria-label={`Play ${attachment.fileName}`}
    className={cn(
      PREVIEW_BUTTON_CLASS,
      "relative flex aspect-video w-[360px] max-w-full flex-col items-center justify-center gap-3 p-4",
    )}
    onClick={(event) => {
      onPreview(attachment.id, event.currentTarget);
    }}
    type="button"
  >
    <span className="bg-surface-muted text-text-muted flex size-12 items-center justify-center rounded-full">
      <Video02Icon aria-hidden className="size-6" />
    </span>
    <span className="max-w-full truncate font-medium">
      {attachment.fileName}
    </span>
    <span className="text-text-muted">
      Video · {formatFileSize(attachment.sizeBytes)}
    </span>
  </button>
);

export const MessageAttachments = ({ message }: { message: ChatMessage }) => {
  const { workspaceSlug } = useWorkspacePath();
  const { openMessage } = useChatAttachmentPreview();
  if (message.deletedAt || !message.attachments.length) return null;

  const images = message.attachments.filter(
    (attachment) => getPreviewMediaKind(attachment) === "image",
  );
  const files = message.attachments.filter(
    (attachment) => getPreviewMediaKind(attachment) !== "image",
  );
  const expandedPdfs =
    message.attachments.length <= 2 &&
    message.attachments.every(
      (attachment) => getPreviewMediaKind(attachment) === "pdf",
    );
  const attachmentUrl = (attachment: ChatAttachment) =>
    getChatAttachmentThumbnailUrl(
      workspaceSlug,
      message.conversationId,
      attachment.id,
    );
  const onPreview: AttachmentPreview = (attachmentId, trigger) => {
    openMessage(message, attachmentId, trigger);
  };

  return (
    <div aria-label="Message attachments" className="mt-2 min-w-0 space-y-2">
      {images.length ? (
        <div
          className={cn(
            images.length > 1
              ? "grid w-full grid-cols-[repeat(auto-fit,minmax(min(100%,9.5rem),1fr))] gap-2"
              : "max-w-[480px]",
          )}
          style={
            images.length > 1
              ? {
                  maxWidth:
                    Math.min(images.length, 4) * 160 +
                    (Math.min(images.length, 4) - 1) * 8,
                }
              : undefined
          }
        >
          {images.slice(0, 4).map((attachment, index) => (
            <ImageAttachment
              attachment={attachment}
              key={attachment.id}
              onPreview={onPreview}
              remaining={index === 3 ? Math.max(0, images.length - 4) : 0}
              single={images.length === 1}
              url={attachmentUrl(attachment)}
            />
          ))}
        </div>
      ) : null}
      {files.length ? (
        <div className="flex max-w-[664px] flex-wrap items-start gap-2">
          {files.map((attachment) => {
            const type = getPreviewMediaKind(attachment);
            if (type === "video") {
              return (
                <VideoAttachment
                  attachment={attachment}
                  key={attachment.id}
                  onPreview={onPreview}
                />
              );
            }
            if (type === "audio") {
              return (
                <ChatAudioAttachment
                  attachment={attachment}
                  conversationId={message.conversationId}
                  key={attachment.id}
                  messageId={message.id}
                />
              );
            }
            return (
              <FileAttachment
                attachment={attachment}
                expanded={expandedPdfs}
                key={attachment.id}
                onPreview={onPreview}
                paired={files.length === 2}
                thumbnailUrl={attachmentUrl(attachment)}
                url={getChatPdfPreviewUrl(
                  workspaceSlug,
                  message.conversationId,
                  attachment.id,
                )}
              />
            );
          })}
        </div>
      ) : null}
    </div>
  );
};

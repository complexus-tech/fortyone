"use client";

import type { ReactNode } from "react";
import { lazy, Suspense, useRef } from "react";
import {
  ChevronLeftIcon,
  ChevronRightIcon,
  CloseIcon,
  DocsIcon,
  DownloadIcon,
} from "icons";
import { Button, buttonVariants, Dialog, Tooltip } from "ui";
import { ImagePreview } from "./image-preview";
import { PreviewError, PreviewLoading } from "./preview-state";
import type { PreviewAttachment } from "./types";
import {
  getAdjacentPreviewId,
  getPreviewAttachments,
  getPreviewMediaKind,
  getSafePreviewUrl,
  isPreviewNavigationTarget,
} from "./types";

const LazyPdfPreview = lazy(async () => {
  const { PdfPreview } = await import("./pdf-preview");
  return { default: PdfPreview };
});

const formatSize = (bytes: number) =>
  bytes < 1_048_576
    ? `${Math.max(1, Math.round(bytes / 1024))} KB`
    : `${(bytes / 1_048_576).toFixed(1)} MB`;

const AttachmentContent = ({
  attachment,
}: {
  attachment: PreviewAttachment;
}) => {
  const kind = getPreviewMediaKind(attachment);
  if (!getSafePreviewUrl(attachment.url)) {
    return (
      <PreviewError message="This file is no longer available. Close the preview and refresh the conversation." />
    );
  }
  if (kind === "image") return <ImagePreview attachment={attachment} />;
  if (kind === "pdf")
    return (
      <Suspense fallback={<PreviewLoading label="Loading PDF viewer…" />}>
        <LazyPdfPreview attachment={attachment} />
      </Suspense>
    );
  if (kind === "video")
    return (
      <div
        className="bg-surface-muted/40 flex min-h-0 flex-1 items-center justify-center p-4"
        data-preview-local-keys
      >
        {/* Video clips may not contain spoken audio; source captions remain a separate attachment capability. */}
        {/* eslint-disable-next-line jsx-a11y/media-has-caption -- User-uploaded video has no caption-track contract yet. */}
        <video
          aria-label={attachment.fileName}
          className="max-h-full max-w-full"
          controls
          playsInline
          preload="metadata"
          src={attachment.url}
        />
      </div>
    );
  return (
    <div className="flex min-h-48 flex-1 flex-col items-center justify-center gap-3 px-6 text-center">
      <DocsIcon aria-hidden className="text-text-muted size-12" />
      <p className="max-w-full truncate font-medium">{attachment.fileName}</p>
      <p className="text-text-muted max-w-md">
        A preview is not available for this file type. Download it to open it in
        its app.
      </p>
      <a
        className={buttonVariants({
          color: "tertiary",
          size: "sm",
          variant: "outline",
        })}
        download={attachment.fileName}
        href={attachment.url}
        rel="noreferrer noopener"
        target="_blank"
      >
        <DownloadIcon aria-hidden className="size-4" />
        Download file
      </a>
    </div>
  );
};

export type AttachmentPreviewDialogProps = {
  open: boolean;
  attachments: PreviewAttachment[];
  selectedId: string | null;
  onSelect: (id: string) => void;
  onClose: () => void;
  onCloseAutoFocus?: (event: Event) => void;
  contextLabel?: string;
  footer?: ReactNode;
  loading?: boolean;
  error?: string;
  onRetry?: () => void;
};

export const AttachmentPreviewDialog = ({
  open,
  attachments,
  selectedId,
  onSelect,
  onClose,
  onCloseAutoFocus,
  contextLabel,
  footer,
  loading = false,
  error,
  onRetry,
}: AttachmentPreviewDialogProps) => {
  const contentRef = useRef<HTMLDivElement>(null);
  const files = getPreviewAttachments(attachments);
  const activeIndex = files.findIndex(
    (attachment) => attachment.id === selectedId,
  );
  const selected = activeIndex >= 0 ? files[activeIndex] : undefined;
  const isAudioSelection = attachments.some(
    (attachment) =>
      attachment.id === selectedId &&
      getPreviewMediaKind(attachment) === "audio",
  );
  const showNavigation =
    files.length > 1 && Boolean(selected) && !loading && !error;
  const previousId = getAdjacentPreviewId(files, selectedId, -1);
  const nextId = getAdjacentPreviewId(files, selectedId, 1);
  const downloadUrl = selected ? getSafePreviewUrl(selected.url) : null;
  let body: ReactNode = (
    <PreviewError
      message="This attachment is no longer available. Close the preview and refresh the conversation."
      onRetry={onRetry}
    />
  );
  if (loading) body = <PreviewLoading label="Loading attachments…" />;
  else if (error) body = <PreviewError message={error} onRetry={onRetry} />;
  else if (selected)
    body = (
      <AttachmentContent
        attachment={selected}
        key={`${selected.id}:${selected.url}`}
      />
    );

  return (
    <Dialog
      onOpenChange={(nextOpen) => {
        if (!nextOpen) onClose();
      }}
      open={Boolean(open && !isAudioSelection)}
    >
      <Dialog.Content
        className="my-auto flex h-[min(88dvh,800px)] flex-col md:my-auto"
        hideClose
        onCloseAutoFocus={onCloseAutoFocus}
        onKeyDown={(event) => {
          if (
            event.defaultPrevented ||
            event.altKey ||
            event.ctrlKey ||
            event.metaKey ||
            event.nativeEvent.isComposing ||
            !showNavigation ||
            isPreviewNavigationTarget(event.target) ||
            window.getSelection()?.isCollapsed === false
          )
            return;
          const direction = event.key === "ArrowLeft" ? -1 : 1;
          if (event.key !== "ArrowLeft" && event.key !== "ArrowRight") return;
          const next = getAdjacentPreviewId(files, selectedId, direction);
          event.preventDefault();
          if (next) onSelect(next);
        }}
        onOpenAutoFocus={(event) => {
          event.preventDefault();
          contentRef.current?.focus();
        }}
        ref={contentRef}
        size="lg"
        tabIndex={-1}
      >
        <Dialog.Header className="border-border flex min-h-16 shrink-0 flex-wrap items-center justify-between gap-2 border-b px-4 py-3">
          <div className="min-w-0 flex-1">
            <Dialog.Title className="truncate leading-5">
              {selected?.fileName ?? "Attachment preview"}
            </Dialog.Title>
            <Dialog.Description className="truncate px-0 pt-1">
              {contextLabel ?? "Shared attachment"}
              {selected ? ` · ${formatSize(selected.sizeBytes)}` : ""}
            </Dialog.Description>
          </div>
          <div className="flex shrink-0 items-center gap-1">
            {showNavigation ? (
              <div
                aria-label="Attachment navigation"
                className="mr-2 flex items-center gap-1"
                role="group"
              >
                <Tooltip title="Previous attachment">
                  <Button
                    aria-label="Previous attachment"
                    asIcon
                    color="tertiary"
                    disabled={!previousId}
                    onClick={() => {
                      if (previousId) onSelect(previousId);
                    }}
                    size="sm"
                    type="button"
                    variant="naked"
                  >
                    <ChevronLeftIcon className="size-4" />
                  </Button>
                </Tooltip>
                <span
                  aria-atomic="true"
                  aria-live="polite"
                  className="min-w-12 text-center tabular-nums"
                >
                  {activeIndex + 1} of {files.length}
                </span>
                <Tooltip title="Next attachment">
                  <Button
                    aria-label="Next attachment"
                    asIcon
                    color="tertiary"
                    disabled={!nextId}
                    onClick={() => {
                      if (nextId) onSelect(nextId);
                    }}
                    size="sm"
                    type="button"
                    variant="naked"
                  >
                    <ChevronRightIcon className="size-4" />
                  </Button>
                </Tooltip>
              </div>
            ) : null}
            {downloadUrl && selected && !loading && !error ? (
              <Tooltip title="Download file">
                <a
                  aria-label={`Download ${selected.fileName}`}
                  className={buttonVariants({
                    asIcon: true,
                    color: "tertiary",
                    size: "sm",
                    variant: "naked",
                  })}
                  download={selected.fileName}
                  href={downloadUrl}
                  rel="noreferrer noopener"
                  target="_blank"
                >
                  <DownloadIcon aria-hidden className="size-4" />
                </a>
              </Tooltip>
            ) : null}
            <Tooltip title="Close preview">
              <Button
                aria-label="Close attachment preview"
                asIcon
                color="tertiary"
                onClick={onClose}
                size="sm"
                type="button"
                variant="naked"
              >
                <CloseIcon className="size-4" />
              </Button>
            </Tooltip>
          </div>
        </Dialog.Header>
        {open && !isAudioSelection ? body : null}
        {footer ? (
          <Dialog.Footer className="border-border shrink-0 items-center justify-end gap-2 border-t px-4 py-3">
            {footer}
          </Dialog.Footer>
        ) : null}
      </Dialog.Content>
    </Dialog>
  );
};

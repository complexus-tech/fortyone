"use client";

import type { ReactNode } from "react";
import { useEffect, useState } from "react";
import { CloseIcon, DocsIcon } from "icons";
import { Box, Button, Flex, Text } from "ui";
import { cn } from "lib";

type FeedbackAttachmentPreviewLayout = "page" | "widget";

const IMAGE_FILE_EXTENSION = /\.(?:gif|jpe?g|png|webp)$/i;
const VIDEO_FILE_EXTENSION = /\.mp4$/i;

const formatFileSize = (bytes: number) => {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
};

const FeedbackAttachmentPreview = ({
  file,
  onRemove,
}: {
  file: File;
  onRemove: () => void;
}) => {
  const isImage =
    file.type.startsWith("image/") || IMAGE_FILE_EXTENSION.test(file.name);
  const isVideo =
    file.type.startsWith("video/") || VIDEO_FILE_EXTENSION.test(file.name);
  const [previewUrl, setPreviewUrl] = useState("");

  useEffect(() => {
    if ((!isImage && !isVideo) || typeof URL.createObjectURL !== "function") {
      setPreviewUrl("");
      return;
    }

    const objectUrl = URL.createObjectURL(file);
    setPreviewUrl(objectUrl);
    return () => {
      URL.revokeObjectURL(objectUrl);
    };
  }, [file, isImage, isVideo]);

  let mediaPreview: ReactNode = (
    <Flex align="center" className="h-full" justify="center">
      <DocsIcon className="text-text-muted h-6" />
    </Flex>
  );
  if (previewUrl) {
    mediaPreview = isImage ? (
      /* eslint-disable-next-line @next/next/no-img-element -- Local object URLs cannot be optimized by next/image. */
      <img
        alt={file.name}
        className="h-full w-full object-cover transition-transform duration-300 group-hover:scale-105"
        src={previewUrl}
      />
    ) : (
      <video
        aria-label={file.name}
        className="h-full w-full object-cover"
        muted
        playsInline
        preload="metadata"
        src={previewUrl}
      />
    );
  }

  if (isImage || isVideo) {
    return (
      <Box className="group border-border bg-surface-muted relative h-24 min-w-0 overflow-hidden rounded-xl border">
        {mediaPreview}
        <Box className="absolute inset-x-0 bottom-0 bg-gradient-to-t from-black/80 to-transparent px-2.5 pt-7 pb-2">
          <Text className="truncate text-xs text-white" fontWeight="medium">
            {file.name}
          </Text>
        </Box>
        <Button
          aria-label={`Remove ${file.name}`}
          asIcon
          className="absolute top-2 right-2 border-white/20 bg-black/70 text-white shadow-sm backdrop-blur-sm hover:bg-black/90"
          color="black"
          onClick={onRemove}
          rounded="full"
          size="xs"
        >
          <CloseIcon className="h-3.5 text-white" strokeWidth={3} />
        </Button>
      </Box>
    );
  }

  return (
    <Box className="group border-border bg-surface-muted/50 relative h-24 min-w-0 overflow-hidden rounded-xl border">
      <Flex align="center" className="h-full pb-7" justify="center">
        <Box className="bg-surface rounded-lg p-2">
          <DocsIcon className="text-text-muted h-5" />
        </Box>
      </Flex>
      <Box className="border-border/60 bg-surface/90 absolute inset-x-0 bottom-0 border-t px-2 py-1.5 backdrop-blur-sm">
        <Text className="truncate text-[11px]" fontWeight="medium">
          {file.name}
        </Text>
        <Text className="text-[10px]" color="muted">
          {formatFileSize(file.size)}
        </Text>
      </Box>
      <Button
        aria-label={`Remove ${file.name}`}
        asIcon
        className="bg-surface/85 absolute top-2 right-2 shadow-sm backdrop-blur-sm"
        color="tertiary"
        onClick={onRemove}
        rounded="full"
        size="xs"
        variant="naked"
      >
        <CloseIcon className="h-3.5" strokeWidth={3} />
      </Button>
    </Box>
  );
};

export const FeedbackAttachmentPreviews = ({
  files,
  layout,
  onRemove,
}: {
  files: File[];
  layout: FeedbackAttachmentPreviewLayout;
  onRemove: (file: File) => void;
}) => {
  if (files.length === 0) return null;

  return (
    <div
      className={cn(
        "mt-4 grid gap-2",
        layout === "page"
          ? "grid-cols-2 sm:grid-cols-3 lg:grid-cols-6"
          : "grid-cols-3",
      )}
    >
      {files.map((file) => (
        <FeedbackAttachmentPreview
          file={file}
          key={`${file.name}:${file.size}:${file.lastModified}`}
          onRemove={() => {
            onRemove(file);
          }}
        />
      ))}
    </div>
  );
};

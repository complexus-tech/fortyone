"use client";

import type { ReactNode } from "react";
import { useState } from "react";
import { BlurImage, Box, Button, Dialog, Flex, Menu, Text, Wrapper } from "ui";
import {
  DocsIcon,
  DownloadIcon,
  CloseIcon,
  MoreHorizontalIcon,
  NewTabIcon,
  DeleteIcon,
  LoadingIcon,
  ArrowLeftIcon,
  ArrowRightIcon,
} from "icons";
import MediaThemeSutro from "player.style/sutro/react";
import { cn } from "lib";
import { ConfirmDialog } from "@/components/ui";
import { useIsAdminOrOwner } from "@/hooks/owner";
import type { StoryAttachment } from "../types";
import { getAdjacentAttachmentIndex } from "./attachment-preview-navigation";

export const ObjectViewer = ({
  data,
  type,
  className = "",
}: {
  data: string;
  type: string;
  className?: string;
}) => {
  return (
    <object
      aria-label="file preview"
      className={cn("block h-full w-full", className)}
      data={data}
      type={type}
    />
  );
};

interface StoryAttachmentPreviewProps {
  file: StoryAttachment;
  files?: StoryAttachment[];
  className?: string;
  children?: ReactNode;
  onDownload?: () => void;
  onDelete?: () => void;
  onDeleteFile?: (file: StoryAttachment) => void;
  isInChat?: boolean;
}

export const StoryAttachmentPreview = ({
  file,
  files,
  className,
  children,
  onDownload,
  onDelete,
  onDeleteFile,
  isInChat,
}: StoryAttachmentPreviewProps) => {
  const [isOpen, setIsOpen] = useState(false);
  const [fileToDelete, setFileToDelete] = useState<StoryAttachment | null>(
    null,
  );
  const previewFiles = files?.length ? files : [file];
  const fileIndex = Math.max(
    0,
    previewFiles.findIndex((previewFile) => previewFile.id === file.id),
  );
  const [previewIndex, setPreviewIndex] = useState(fileIndex);
  const normalizedPreviewIndex =
    previewFiles.length > 0 ? previewIndex % previewFiles.length : 0;
  const activeFile = previewFiles[normalizedPreviewIndex] ?? file;
  const isImage = file.mimeType.startsWith("image/");
  const isVideo = file.mimeType.startsWith("video");
  const isPdf = file.mimeType.includes("pdf");
  const isActiveImage = activeFile.mimeType.startsWith("image/");
  const isActiveVideo = activeFile.mimeType.startsWith("video");
  const isActivePdf = activeFile.mimeType.includes("pdf");
  const isUploading = file.id.includes("temp-");
  const { isAdminOrOwner } = useIsAdminOrOwner(file.uploadedBy);
  const { isAdminOrOwner: canManageActiveFile } = useIsAdminOrOwner(
    activeFile.uploadedBy,
  );
  const canNavigate = previewFiles.length > 1;

  const openPreview = () => {
    if (isUploading) return;
    setPreviewIndex(fileIndex);
    setIsOpen(true);
  };

  const showPreviousFile = () => {
    setPreviewIndex((currentIndex) =>
      getAdjacentAttachmentIndex(currentIndex, previewFiles.length, -1),
    );
  };

  const showNextFile = () => {
    setPreviewIndex((currentIndex) =>
      getAdjacentAttachmentIndex(currentIndex, previewFiles.length, 1),
    );
  };

  const formatFileSize = (bytes: number) => {
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(2)} KB`;
    return `${(bytes / (1024 * 1024)).toFixed(2)} MB`;
  };

  let thumbnail: ReactNode = children;

  if (!thumbnail) {
    if (isImage || isVideo) {
      thumbnail = (
        <Box
          className="group border-border bg-surface-muted ring-accent relative h-24 overflow-hidden rounded-xl border hover:ring-2 md:h-28 2xl:h-36 dark:shadow-none"
          onClick={openPreview}
        >
          {isImage ? (
            <BlurImage
              alt={file.filename}
              className="h-full w-full object-cover transition-transform duration-300 group-hover:scale-105"
              src={file.url}
            />
          ) : (
            <video
              className="h-full w-full object-cover"
              controls={false}
              muted
              src={file.url}
            />
          )}
          {isUploading ? (
            <Box className="bg-dark/50 absolute inset-0 flex items-center justify-center">
              <LoadingIcon className="h-6 animate-spin" />
            </Box>
          ) : null}

          {isInChat && onDelete ? (
            <Button
              aria-label={`Remove ${file.filename}`}
              asIcon
              className="absolute top-2 right-2 border-white/20 bg-black/75 shadow-sm backdrop-blur-sm hover:bg-black/90 focus-visible:bg-black/90 focus-visible:ring-2 focus-visible:ring-white/80"
              color="black"
              onClick={(event) => {
                event.stopPropagation();
                onDelete();
              }}
              rounded="full"
              size="xs"
            >
              <CloseIcon className="h-4 text-white" strokeWidth={3} />
            </Button>
          ) : null}
        </Box>
      );
    } else {
      thumbnail = (
        <Wrapper className="ring-accent px-3 py-2 transition-shadow duration-300 hover:ring-2 md:px-4 md:py-2.5">
          <Flex align="center" className="gap-3 md:gap-6" justify="between">
            <Flex
              align="center"
              className="flex-1"
              gap={3}
              onClick={openPreview}
            >
              <Box className="bg-surface-muted rounded-lg">
                {isUploading ? (
                  <LoadingIcon className="h-5 animate-spin md:h-6" />
                ) : (
                  <DocsIcon className="h-5 md:h-6" />
                )}
              </Box>
              <Box>
                <Text className="mb-0.5 line-clamp-1 first-letter:uppercase">
                  {isUploading ? "Uploading..." : file.filename}
                </Text>
                <Text className="text-[0.95rem]" color="muted">
                  {file.size > 0 ? formatFileSize(file.size) : "PDF file"}
                </Text>
              </Box>
            </Flex>
            {isInChat && onDelete ? (
              <Button
                aria-label={`Remove ${file.filename}`}
                asIcon
                color="tertiary"
                onClick={onDelete}
                rounded="full"
                size="sm"
              >
                <CloseIcon className="h-4 text-current" strokeWidth={3} />
              </Button>
            ) : null}
            {!isInChat && (
              <Flex align="center" gap={1}>
                {isPdf ? (
                  <Button
                    asIcon
                    color="tertiary"
                    disabled={isUploading}
                    onClick={() => {
                      openPreview();
                    }}
                    variant="naked"
                  >
                    <NewTabIcon className="h-5" />
                  </Button>
                ) : null}
                <Button
                  asIcon
                  color="tertiary"
                  disabled={isUploading}
                  onClick={onDownload}
                  variant="naked"
                >
                  <DownloadIcon className="h-5" />
                </Button>
                {isAdminOrOwner ? (
                  <Menu>
                    <Menu.Button>
                      <Button
                        asIcon
                        color="tertiary"
                        disabled={isUploading}
                        variant="naked"
                      >
                        <MoreHorizontalIcon className="h-5" />
                      </Button>
                    </Menu.Button>
                    <Menu.Items align="end" className="w-36">
                      <Menu.Group>
                        <Menu.Item
                          onClick={() => {
                            setFileToDelete(file);
                          }}
                        >
                          <DeleteIcon /> Delete...
                        </Menu.Item>
                      </Menu.Group>
                    </Menu.Items>
                  </Menu>
                ) : null}
              </Flex>
            )}
          </Flex>
        </Wrapper>
      );
    }
  }

  return (
    <>
      <Box className={cn("cursor-pointer", className)}>{thumbnail}</Box>
      <Dialog onOpenChange={setIsOpen} open={isOpen}>
        <Dialog.Content
          className={cn(
            "relative my-auto rounded-2xl border-0 md:mt-auto md:mb-auto dark:border-[0.5px]",
          )}
          hideClose
          onKeyDown={(event) => {
            if (!canNavigate) return;
            if (event.key === "ArrowLeft") {
              event.preventDefault();
              showPreviousFile();
            }
            if (event.key === "ArrowRight") {
              event.preventDefault();
              showNextFile();
            }
          }}
          size="lg"
        >
          {canNavigate ? (
            <>
              <Button
                aria-label="View previous attachment"
                asIcon
                className="absolute top-1/2 left-4 z-20 -translate-y-1/2 border-white/20 bg-black/75 text-white shadow-lg backdrop-blur-sm hover:bg-black/90 focus-visible:bg-black/90 focus-visible:ring-2 focus-visible:ring-white/80"
                color="black"
                onClick={showPreviousFile}
                rounded="full"
              >
                <ArrowLeftIcon className="h-5 text-white dark:text-white" />
              </Button>
              <Button
                aria-label="View next attachment"
                asIcon
                className="absolute top-1/2 right-4 z-20 -translate-y-1/2 border-white/20 bg-black/75 text-white shadow-lg backdrop-blur-sm hover:bg-black/90 focus-visible:bg-black/90 focus-visible:ring-2 focus-visible:ring-white/80"
                color="black"
                onClick={showNextFile}
                rounded="full"
              >
                <ArrowRightIcon className="h-5 text-white dark:text-white" />
              </Button>
            </>
          ) : null}
          <Dialog.Header className="border-border-strong flex items-center justify-between gap-3 border-b-[0.5px] px-3">
            <Flex align="center" className="min-w-0" gap={2}>
              <Dialog.Title className="truncate">
                {activeFile.filename}
              </Dialog.Title>
              {canNavigate ? (
                <Text
                  className="shrink-0 leading-none"
                  color="muted"
                  fontWeight="medium"
                >
                  {normalizedPreviewIndex + 1} of {previewFiles.length}
                </Text>
              ) : null}
            </Flex>
            <Flex
              align="center"
              className="pointer-events-auto"
              justify="between"
            >
              <Flex align="center" gap={3}>
                {canManageActiveFile && !isInChat ? (
                  <Button
                    asIcon
                    color="tertiary"
                    leftIcon={<DeleteIcon className="h-4.5" />}
                    onClick={() => {
                      setFileToDelete(activeFile);
                    }}
                    size="sm"
                  >
                    <span className="sr-only">Delete</span>
                  </Button>
                ) : null}
                {!isInChat ? (
                  <>
                    <Button
                      asIcon
                      color="tertiary"
                      href={activeFile.url}
                      leftIcon={<NewTabIcon className="h-4.5" />}
                      size="sm"
                      target="_blank"
                    >
                      <span className="sr-only">Open in new tab</span>
                    </Button>
                    <Button
                      asIcon
                      color="tertiary"
                      href={activeFile.url}
                      leftIcon={<DownloadIcon className="h-4.5" />}
                      size="sm"
                      target="_blank"
                    >
                      <span className="sr-only">Download</span>
                    </Button>
                  </>
                ) : null}

                <Button
                  asIcon
                  color="tertiary"
                  leftIcon={<CloseIcon />}
                  onClick={() => {
                    setIsOpen(false);
                  }}
                  size="sm"
                >
                  <span className="sr-only">Close</span>
                </Button>
              </Flex>
            </Flex>
          </Dialog.Header>
          {isActiveImage ? (
            <Box className="flex h-[50dvh] justify-center overflow-y-auto px-2 md:h-[60dvh]">
              <BlurImage
                alt={activeFile.filename}
                className="h-full w-full"
                imageClassName="object-contain"
                src={activeFile.url}
              />
            </Box>
          ) : null}

          {isActiveVideo ? (
            <MediaThemeSutro
              className={cn(
                "aspect-video h-[55dvh] w-full overflow-hidden rounded-lg",
                className,
              )}
              title={activeFile.filename}
            >
              <video
                className="h-full w-full"
                muted
                playsInline
                slot="media"
                src={activeFile.url}
              />
            </MediaThemeSutro>
          ) : null}

          {isActivePdf ? (
            <ObjectViewer
              className="min-h-[80dvh] overflow-hidden rounded-lg"
              data={activeFile.url}
              type="application/pdf"
            />
          ) : null}
        </Dialog.Content>
      </Dialog>
      <ConfirmDialog
        confirmText="Yes, delete"
        description="Are you sure you want to delete this attachment? You cannot undo this action."
        isOpen={Boolean(fileToDelete)}
        onClose={() => {
          setFileToDelete(null);
        }}
        onConfirm={() => {
          if (fileToDelete && onDeleteFile) {
            onDeleteFile(fileToDelete);
          } else if (onDelete) {
            onDelete();
          }
          setFileToDelete(null);
        }}
        title="Delete attachment"
      />
    </>
  );
};

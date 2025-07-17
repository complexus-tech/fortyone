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
} from "icons";
import MediaThemeSutro from "player.style/sutro/react";
import { cn } from "lib";
import { ConfirmDialog } from "@/components/ui";
import { useIsAdminOrOwner } from "@/hooks/owner";
import type { StoryAttachment } from "../types";

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
  className?: string;
  children?: ReactNode;
  onDownload?: () => void;
  onDelete?: () => void;
  isInChat?: boolean;
}

export const StoryAttachmentPreview = ({
  file,
  className,
  children,
  onDownload,
  onDelete,
  isInChat,
}: StoryAttachmentPreviewProps) => {
  const [isOpen, setIsOpen] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const isImage = file.mimeType.includes("image");
  const isVideo = file.mimeType.startsWith("video");
  const isPdf = file.mimeType.includes("pdf");
  const isUploading = file.id.includes("temp-");
  const { isAdminOrOwner } = useIsAdminOrOwner(file.uploadedBy);

  const formatFileSize = (bytes: number) => {
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(2)} KB`;
    return `${(bytes / (1024 * 1024)).toFixed(2)} MB`;
  };

  const renderThumbnail = () => {
    if (children) return children;

    if (isImage || isVideo) {
      return (
        <Box
          className="group relative h-24 overflow-hidden rounded-xl border border-gray-50 bg-gray-50/70 shadow-lg shadow-gray-100 ring-gray-200 transition-all duration-300 hover:ring hover:grayscale dark:border-dark-200 dark:bg-dark-200/50 dark:shadow-none dark:ring-dark-50 md:h-28 2xl:h-36"
          onClick={() => {
            if (isUploading) return;
            setIsOpen(true);
          }}
        >
          {isImage ? (
            <BlurImage
              alt={file.filename}
              className="h-full w-full object-cover transition-all duration-300 group-hover:scale-105"
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
            <Box className="absolute inset-0 flex items-center justify-center bg-dark/50">
              <LoadingIcon className="h-6 animate-spin" />
            </Box>
          ) : null}

          {isInChat && onDelete ? (
            <Button
              asIcon
              className="absolute right-2 top-2"
              color="invert"
              onClick={onDelete}
              rounded="full"
              size="xs"
            >
              <CloseIcon
                className="h-4 text-white dark:text-dark"
                strokeWidth={3}
              />
            </Button>
          ) : null}
        </Box>
      );
    }

    return (
      <Wrapper className="px-3 py-2 ring-gray-200 ring-offset-1 transition-all duration-300 hover:ring dark:ring-dark-50 dark:ring-offset-dark md:px-4 md:py-2.5">
        <Flex align="center" className="gap-3 md:gap-6" justify="between">
          <Flex
            align="center"
            className="flex-1"
            gap={3}
            onClick={() => {
              if (isUploading) return;
              setIsOpen(true);
            }}
          >
            <Box className="rounded-[0.6rem] bg-gray-100/50 p-2 dark:bg-dark-200/80">
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
              asIcon
              color="invert"
              onClick={onDelete}
              rounded="full"
              size="sm"
            >
              <CloseIcon
                className="h-4 text-white dark:text-dark"
                strokeWidth={3}
              />
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
                    setIsOpen(true);
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
                          setIsDeleting(true);
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
  };

  return (
    <>
      <Box className={cn("cursor-pointer", className)}>{renderThumbnail()}</Box>
      <Dialog onOpenChange={setIsOpen} open={isOpen}>
        <Dialog.Content
          className={cn("relative my-auto px-2 pt-2 md:mb-auto md:mt-auto", {
            "bg-dark dark:bg-dark": isImage,
          })}
          hideClose
          size="lg"
        >
          <Dialog.Header className="sr-only">
            <Dialog.Title className="mb-0 px-6">{file.filename}</Dialog.Title>
          </Dialog.Header>
          {isImage ? (
            <Box className="flex h-[50dvh] items-center justify-center overflow-y-auto rounded-[0.6rem] md:h-[70dvh]">
              <BlurImage
                alt={file.filename}
                className="h-full bg-dark dark:bg-dark"
                imageClassName="object-contain"
                src={file.url}
              />
            </Box>
          ) : null}

          {isVideo ? (
            <MediaThemeSutro
              className={cn(
                "aspect-video h-[55dvh] w-full overflow-hidden rounded-[0.6rem]",
                className,
              )}
              title={file.filename}
            >
              <video
                className="h-full w-full"
                muted
                playsInline
                slot="media"
                src={file.url}
              />
            </MediaThemeSutro>
          ) : null}

          {isPdf ? (
            <ObjectViewer
              className="min-h-[80dvh] overflow-hidden rounded-[0.6rem]"
              data={file.url}
              type="application/pdf"
            />
          ) : null}

          {isImage || isVideo ? (
            <Box className="dark pointer-events-none absolute left-0 right-0 top-0 z-[3] h-20 bg-gradient-to-b from-dark/80 px-6 py-5">
              <Flex
                align="center"
                className="pointer-events-auto"
                justify="between"
              >
                <Text className="text-white" fontSize="sm">
                  {file.filename}
                </Text>
                <Flex align="center" gap={3}>
                  {isAdminOrOwner && !isInChat ? (
                    <Button
                      asIcon
                      color="tertiary"
                      leftIcon={<DeleteIcon className="h-4" />}
                      onClick={() => {
                        setIsDeleting(true);
                      }}
                      size="xs"
                    >
                      <span className="sr-only">Delete</span>
                    </Button>
                  ) : null}
                  {!isInChat ? (
                    <>
                      <Button
                        asIcon
                        color="tertiary"
                        href={file.url}
                        leftIcon={<NewTabIcon className="h-4" />}
                        size="xs"
                        target="_blank"
                      >
                        <span className="sr-only">Open in new tab</span>
                      </Button>
                      <Button
                        asIcon
                        color="tertiary"
                        href={file.url}
                        leftIcon={<DownloadIcon className="h-4" />}
                        onClick={onDownload}
                        size="xs"
                        target="_blank"
                      >
                        <span className="sr-only">Download</span>
                      </Button>
                    </>
                  ) : null}

                  <Button
                    asIcon
                    color="tertiary"
                    leftIcon={<CloseIcon className="h-4" />}
                    onClick={() => {
                      setIsOpen(false);
                    }}
                    rounded="full"
                    size="xs"
                  >
                    <span className="sr-only">Close</span>
                  </Button>
                </Flex>
              </Flex>
            </Box>
          ) : null}
        </Dialog.Content>
      </Dialog>
      <ConfirmDialog
        confirmText="Yes, delete"
        description="Are you sure you want to delete this attachment? You cannot undo this action."
        isOpen={isDeleting}
        onClose={() => {
          setIsDeleting(false);
        }}
        onConfirm={() => {
          if (onDelete) {
            onDelete();
          }
        }}
        title="Delete attachment"
      />
    </>
  );
};

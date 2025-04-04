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
} from "icons";
import MediaThemeSutro from "player.style/sutro/react";
import { cn } from "lib";
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
}

export const StoryAttachmentPreview = ({
  file,
  className,
  children,
  onDownload,
}: StoryAttachmentPreviewProps) => {
  const [isOpen, setIsOpen] = useState(false);
  const isImage = file.mimeType.includes("image");
  const isVideo = file.mimeType.startsWith("video/");
  const isPdf = file.mimeType === "application/pdf";

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
          className="group relative h-28 overflow-hidden rounded-xl border border-gray-50 bg-gray-50/70 shadow-lg shadow-gray-100 ring-gray-200 transition-all duration-300 hover:ring hover:grayscale dark:border-dark-200 dark:bg-dark-200/50 dark:shadow-none dark:ring-dark-50"
          onClick={() => {
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
        </Box>
      );
    }

    return (
      <Wrapper className="px-4 py-2.5 ring-gray-200 ring-offset-1 transition-all duration-300 hover:ring dark:ring-dark-50 dark:ring-offset-dark">
        <Flex align="center" gap={6} justify="between">
          <Flex
            align="center"
            className="flex-1"
            gap={3}
            onClick={() => {
              setIsOpen(true);
            }}
          >
            <Box className="rounded-lg bg-gray-100/50 p-2 dark:bg-dark-200/80">
              <DocsIcon className="h-6" />
            </Box>
            <Box>
              <Text className="mb-0.5 line-clamp-1 first-letter:uppercase">
                {file.filename}
              </Text>
              <Text className="text-[0.95rem]" color="muted">
                {formatFileSize(file.size)}
              </Text>
            </Box>
          </Flex>
          <Flex align="center" gap={1}>
            {isPdf ? (
              <Button
                asIcon
                color="tertiary"
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
              onClick={onDownload}
              variant="naked"
            >
              <DownloadIcon className="h-5" />
            </Button>
            <Menu>
              <Menu.Button>
                <Button asIcon color="tertiary" variant="naked">
                  <MoreHorizontalIcon className="h-5" />
                </Button>
              </Menu.Button>
              <Menu.Items align="end" className="w-36">
                <Menu.Group>
                  <Menu.Item>
                    <DeleteIcon /> Delete...
                  </Menu.Item>
                </Menu.Group>
              </Menu.Items>
            </Menu>
          </Flex>
        </Flex>
      </Wrapper>
    );
  };

  return (
    <>
      <Box className={cn("cursor-pointer", className)}>{renderThumbnail()}</Box>
      <Dialog onOpenChange={setIsOpen} open={isOpen}>
        <Dialog.Content
          className={cn("relative my-auto px-2 pt-2", {
            "bg-dark dark:bg-dark": isImage,
          })}
          hideClose
          size="lg"
        >
          <Dialog.Header className="sr-only">
            <Dialog.Title className="mb-0 px-6">{file.filename}</Dialog.Title>
          </Dialog.Header>
          {isImage ? (
            <Box className="flex h-[70vh] items-center justify-center overflow-y-auto rounded-lg">
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
                "aspect-video h-[55vh] w-full overflow-hidden rounded-lg",
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
              className="min-h-[80vh] overflow-hidden rounded-lg"
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
                  <Button
                    asIcon
                    color="tertiary"
                    leftIcon={<DeleteIcon className="h-4" />}
                    size="xs"
                  >
                    <span className="sr-only">Delete</span>
                  </Button>
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
    </>
  );
};

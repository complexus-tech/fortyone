"use client";

import { Attachment } from "@/types";
import { useCallback, useState } from "react";
import { useDropzone } from "react-dropzone";
import { Button, Flex, Text, Box, DropZone } from "ui";
import { ArrowDownIcon, CloseIcon, PlusIcon, SettingsIcon } from "icons";
import { AttachmentPreview } from "@/components/ui/preview";
import Link from "next/link";

const formatFileSize = (bytes: number): string => {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  if (bytes < 1024 * 1024 * 1024)
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
  return `${(bytes / (1024 * 1024 * 1024)).toFixed(1)} GB`;
};

export const FilePicker = () => {
  const [images, setImages] = useState<Attachment[]>([]);

  const onDrop = useCallback(
    (acceptedFiles: Attachment[]) => {
      if (acceptedFiles.length > 0) {
        setImages((prev) => [
          ...prev,
          ...acceptedFiles.map((file) =>
            Object.assign(file, {
              preview: URL.createObjectURL(file),
            }),
          ),
        ]);
      }
    },
    [setImages],
  );

  const { getRootProps, getInputProps, isDragActive } = useDropzone({
    onDrop,
    accept: {
      "image/*": [],
    },
  });
  return (
    <Box>
      {images.length > 0 ? (
        <>
          <Box className="rounded-xl border border-gray-100 text-left dark:border-dark-100">
            <Box className="max-h-[500px] overflow-y-auto">
              {images.map((image, idx) => (
                <Flex
                  key={idx}
                  className="border-b border-gray-100 px-3 py-3 dark:border-dark-100"
                  justify="between"
                  align="center"
                  gap={6}
                >
                  <Flex gap={2}>
                    <Box className="w-16 shrink-0">
                      <AttachmentPreview
                        key={idx}
                        url={image.preview!}
                        name={image.name}
                        className="rounded-lg"
                      />
                    </Box>
                    <Box>
                      <Text className="line-clamp-1">
                        {image.name}
                        Lorem ipsum dolor, sit amet consectetur adipisicing.
                      </Text>
                      <Text color="muted" className="opacity-80">
                        {formatFileSize(image.size)}
                      </Text>
                    </Box>
                  </Flex>
                  <Flex gap={2} className="shrink-0">
                    <Flex gap={2} align="center">
                      <Text color="muted" className="opacity-80">
                        Output
                      </Text>
                      <Button
                        className="text-sm font-semibold"
                        size="sm"
                        color="tertiary"
                        rightIcon={<ArrowDownIcon className="h-3 w-auto" />}
                      >
                        JPG
                      </Button>
                    </Flex>
                    <Button
                      size="sm"
                      asIcon
                      color="tertiary"
                      variant="naked"
                      leftIcon={
                        <SettingsIcon
                          strokeWidth={2.3}
                          className="h-5 w-auto"
                        />
                      }
                    />
                    <Button
                      size="sm"
                      asIcon
                      color="tertiary"
                      variant="naked"
                      leftIcon={
                        <CloseIcon strokeWidth={2.3} className="h-5 w-auto" />
                      }
                    />
                  </Flex>
                </Flex>
              ))}
            </Box>
            <Flex className="p-3" justify="between" align="center" gap={4}>
              <Button
                color="tertiary"
                className="gap-1"
                rightIcon={
                  <ArrowDownIcon strokeWidth={2.3} className="h-3.5 w-auto" />
                }
              >
                Add more images
              </Button>
              <Button>Convert now</Button>
            </Flex>
          </Box>
        </>
      ) : (
        <DropZone>
          <DropZone.Root
            className="block h-auto rounded-xl border-[1.5px] border-dashed border-gray-200 px-4 py-12 dark:border-dark-100 dark:bg-black"
            isDragActive={isDragActive}
            rootProps={getRootProps()}
          >
            <DropZone.Input inputProps={getInputProps()} />
            <DropZone.Body isDragActive={isDragActive}>
              <Box>
                <Flex direction="column" align="center" gap={4}>
                  <Button rounded="lg" size="lg">
                    Choose Files
                  </Button>
                  <Text color="muted">
                    Max file size 10MB.{" "}
                    <Link
                      href="/signup"
                      className="font-semibold text-primary underline"
                    >
                      Sign up
                    </Link>{" "}
                    for more
                  </Text>
                </Flex>
              </Box>
            </DropZone.Body>
          </DropZone.Root>
        </DropZone>
      )}
    </Box>
  );
};

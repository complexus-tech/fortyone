import { useDropzone } from "react-dropzone";
import { Box, Button, DropZone, Flex, Text, Tooltip, Wrapper } from "ui";
import {
  AttachmentIcon,
  DocsIcon,
  DownloadIcon,
  MoreHorizontalIcon,
  PlusIcon,
} from "icons";
import { useCallback, useState } from "react";
import { cn } from "lib";

interface CustomFile extends File {
  preview?: string;
}

export const Attachments = ({ className }: { className?: string }) => {
  const f1 = new File(["test"], "test.png", { type: "image/png" });
  const f2 = new File(["test"], "test.png", { type: "image/png" });
  const f3 = new File(["test"], "test.png", { type: "image/png" });
  const f4 = new File(["test"], "test.png", { type: "image/png" });
  const f5 = new File(["test"], "test.png", { type: "image/png" });

  const [files, setFiles] = useState<CustomFile[]>([f1, f2, f3, f4, f5]);

  const onDrop = useCallback((acceptedFiles: CustomFile[]) => {
    if (acceptedFiles.length > 0) {
      setFiles((prevFiles) => [
        ...prevFiles,
        ...acceptedFiles.map((file) =>
          Object.assign(file, {
            preview: URL.createObjectURL(file),
          }),
        ),
      ]);
    }
  }, []);

  const { getRootProps, getInputProps, isDragActive } = useDropzone({
    onDrop,
  });

  return (
    <Box className={className} suppressHydrationWarning>
      <Flex align="center" className="mb-2" justify="between">
        <Text as="h4" className="flex items-center gap-1" fontWeight="medium">
          <AttachmentIcon className="h-5 w-auto" />
          Attachments
        </Text>
        <Tooltip title="Add attachment">
          <Button
            asIcon
            color="tertiary"
            leftIcon={<PlusIcon />}
            size="sm"
            variant="naked"
          >
            <span className="sr-only">Add attachment</span>
          </Button>
        </Tooltip>
      </Flex>
      {files.length === 0 && (
        <DropZone>
          <DropZone.Root isDragActive={isDragActive} rootProps={getRootProps()}>
            <DropZone.Input
              inputProps={getInputProps({
                multiple: true,
              })}
            />
            <DropZone.Body isDragActive={isDragActive} />
          </DropZone.Root>
        </DropZone>
      )}

      <Box
        className={cn("grid grid-cols-5 gap-3", {
          "mt-3": files.length > 0,
        })}
      >
        {files
          .filter((file) => file.type.includes("image"))
          .map((file, index) => (
            <Box
              className="relative h-28 overflow-hidden rounded-xl bg-gray-50/70 dark:bg-dark-200/50"
              key={index}
            >
              <Box className="absolute right-2 top-2">
                <Text className="text-xs font-medium text-white">X</Text>
              </Box>
              <Box className="absolute bottom-2 left-2">
                <Text className="text-xs font-medium text-white">
                  {file.name}
                </Text>
              </Box>
            </Box>
          ))}
      </Box>
      <Box
        className={cn("grid gap-2", {
          "mt-3": files.length > 0,
        })}
      >
        {files
          .filter((file) => file.type.includes("image"))
          .map((file, index) => (
            <Wrapper className="px-4 py-2.5" key={index}>
              <Flex align="center" gap={6} justify="between">
                <Flex align="center" gap={3}>
                  <Box className="rounded-lg bg-gray-100/50 p-2 dark:bg-dark-200/80">
                    <DocsIcon className="h-6" />
                  </Box>
                  <Box>
                    <Text className="mb-0.5 line-clamp-1 first-letter:uppercase">
                      {file.name}
                    </Text>
                    <Text className="text-[0.95rem]" color="muted">
                      7.53MB
                    </Text>
                  </Box>
                </Flex>
                <Flex align="center" gap={1}>
                  <Button asIcon color="tertiary" variant="naked">
                    <DownloadIcon className="h-5" />
                  </Button>
                  <Button asIcon color="tertiary" variant="naked">
                    <MoreHorizontalIcon className="h-5" />
                  </Button>
                </Flex>
              </Flex>
            </Wrapper>
          ))}
      </Box>
    </Box>
  );
};

import { useDropzone } from "react-dropzone";
import { Box, Button, DropZone, Flex, Text, Tooltip } from "ui";
import { AttachmentIcon, PlusIcon } from "icons";
import { useCallback } from "react";
import { useStoryAttachments } from "../hooks/story-attachments";
import { AttachmentsSkeleton } from "./attachments-skeleton";
import { StoryAttachmentPreview } from "./story-attachment-preview";

interface CustomFile extends File {
  preview?: string;
}

export const Attachments = ({
  className,
  storyId,
}: {
  className?: string;
  storyId: string;
}) => {
  const { data: attachments = [], isPending } = useStoryAttachments(storyId);

  const onDrop = useCallback((acceptedFiles: CustomFile[]) => {
    if (acceptedFiles.length > 0) {
      // upload files here
    }
  }, []);

  const { getRootProps, getInputProps, isDragActive } = useDropzone({
    onDrop,
  });

  const imagesAndVideos = attachments.filter(
    (file) =>
      file.mimeType.includes("image") || file.mimeType.includes("video"),
  );
  const otherFiles = attachments.filter(
    (file) =>
      !file.mimeType.includes("image") && !file.mimeType.includes("video"),
  );

  if (isPending) {
    return <AttachmentsSkeleton />;
  }

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
      {attachments.length === 0 && (
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
      {imagesAndVideos.length > 0 && (
        <Box className="mt-3 grid grid-cols-5 gap-3">
          {imagesAndVideos.map((file) => (
            <StoryAttachmentPreview
              file={file}
              key={file.id}
              onDownload={() => window.open(file.url, "_blank")}
            />
          ))}
        </Box>
      )}
      {otherFiles.length > 0 && (
        <Box className="mt-3 grid gap-2">
          {otherFiles.map((file) => (
            <StoryAttachmentPreview
              file={file}
              key={file.id}
              onDownload={() => window.open(file.url, "_blank")}
            />
          ))}
        </Box>
      )}
    </Box>
  );
};

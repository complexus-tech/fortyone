import { useDropzone } from "react-dropzone";
import { Box, Button, DropZone, Flex, Text, Tooltip } from "ui";
import { AttachmentIcon, PlusIcon } from "icons";
import { toast } from "sonner";
import { useSubscriptionFeatures } from "@/lib/hooks/subscription-features";
import { useStoryAttachments } from "../hooks/story-attachments";
import { useUploadAttachmentMutation } from "../hooks/upload-attachment-mutation";
import { useDeleteAttachmentMutation } from "../hooks/delete-attachment-mutation";
import { AttachmentsSkeleton } from "./attachments-skeleton";
import { StoryAttachmentPreview } from "./story-attachment-preview";

interface CustomFile extends File {
  preview?: string;
}

const formatFileSize = (bytes: number) => {
  return `${(bytes / (1024 * 1024)).toFixed(2)} MB`;
};

export const Attachments = ({
  className,
  storyId,
}: {
  className?: string;
  storyId: string;
}) => {
  const { getLimit } = useSubscriptionFeatures();
  const { data: attachments = [], isPending } = useStoryAttachments(storyId);
  const uploadMutation = useUploadAttachmentMutation(storyId);
  const deleteMutation = useDeleteAttachmentMutation(storyId);
  const maxFileUploads = getLimit("maxFileUploads");

  const onDrop = (acceptedFiles: CustomFile[]) => {
    if (acceptedFiles.length > 0) {
      if (acceptedFiles.length > 1) {
        toast.warning("Only one file can be uploaded at a time", {
          description: "Please select only one file",
        });
        return;
      }
      const file = acceptedFiles[0];
      const MAX_FILE_SIZE =
        maxFileUploads === "10MB" ? 10 * 1024 * 1024 : 256 * 1024 * 1024;
      if (file.size > MAX_FILE_SIZE) {
        toast.warning(
          `File size(${formatFileSize(file.size)}) exceeds the 10MB limit`,
          {
            description: `${file.name} is too large.`,
          },
        );
        return;
      }
      uploadMutation.mutate(file);
    }
  };

  const {
    getRootProps,
    getInputProps,
    isDragActive,
    open: openFilePicker,
  } = useDropzone({
    onDrop,
    multiple: false,
    maxFiles: 1,
    accept: {
      "image/*": [".png", ".jpg", ".jpeg", ".webp"],
      "video/*": [".mp4"],
      "application/pdf": [".pdf"],
    },
    noClick: attachments.length > 0,
    noKeyboard: attachments.length > 0,
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

  const handleDelete = (attachmentId: string) => {
    deleteMutation.mutate(attachmentId);
  };

  return (
    <Box className={className} suppressHydrationWarning>
      <Flex align="center" className="mt-4 mb-2" justify="between">
        <Text as="h4" className="flex items-center gap-1" fontWeight="semibold">
          <AttachmentIcon className="h-5 w-auto" />
          Attachments
        </Text>
        {attachments.length > 0 && (
          <Tooltip title="Add attachment">
            <Button
              asIcon
              color="tertiary"
              leftIcon={<PlusIcon />}
              onClick={openFilePicker}
              size="sm"
              variant="naked"
            >
              <span className="sr-only">Add attachment</span>
            </Button>
          </Tooltip>
        )}
      </Flex>
      {attachments.length === 0 && (
        <DropZone>
          <DropZone.Root isDragActive={isDragActive} rootProps={getRootProps()}>
            <DropZone.Input
              inputProps={getInputProps({
                multiple: true,
              })}
            />
            <DropZone.Body
              isDragActive={isDragActive}
              message="Drag 'n' drop your file here, or click to select the file."
            />
          </DropZone.Root>
        </DropZone>
      )}
      {imagesAndVideos.length > 0 && (
        <Box className="mt-3 grid grid-cols-3 gap-3 md:grid-cols-5">
          {imagesAndVideos.map((file) => (
            <StoryAttachmentPreview
              file={file}
              key={file.id}
              onDelete={() => {
                handleDelete(file.id);
              }}
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
              onDelete={() => {
                handleDelete(file.id);
              }}
              onDownload={() => window.open(file.url, "_blank")}
            />
          ))}
        </Box>
      )}
    </Box>
  );
};

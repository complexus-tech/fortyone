import type { FileRejection } from "react-dropzone";
import { useDropzone } from "react-dropzone";
import { Box, Button, DropZone, Flex, Text, Tooltip } from "ui";
import { AttachmentIcon, PlusIcon } from "icons";
import { toast } from "sonner";
import { cn } from "lib";
import { useSubscriptionFeatures } from "@/lib/hooks/subscription-features";
import { useStoryAttachments } from "../hooks/story-attachments";
import { useUploadAttachmentMutation } from "../hooks/upload-attachment-mutation";
import { useDeleteAttachmentMutation } from "../hooks/delete-attachment-mutation";
import {
  FREE_ATTACHMENT_SIZE_LIMIT,
  getAttachmentRejectionMessage,
  MAX_ATTACHMENT_BATCH_FILES,
  PAID_ATTACHMENT_SIZE_LIMIT,
  uploadAttachmentsConcurrently,
} from "./attachment-upload";
import { AttachmentsSkeleton } from "./attachments-skeleton";
import { StoryAttachmentPreview } from "./story-attachment-preview";

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
  const maxFileSize =
    maxFileUploads === "10MB"
      ? FREE_ATTACHMENT_SIZE_LIMIT
      : PAID_ATTACHMENT_SIZE_LIMIT;

  const onDrop = (acceptedFiles: File[], rejectedFiles: FileRejection[]) => {
    const hasTooManyFiles = rejectedFiles.some((rejection) =>
      rejection.errors.some((error) => error.code === "too-many-files"),
    );
    if (hasTooManyFiles) {
      toast.error(
        `You can upload up to ${MAX_ATTACHMENT_BATCH_FILES} files at once`,
        {
          description: `Select ${MAX_ATTACHMENT_BATCH_FILES} or fewer files and try again.`,
        },
      );
    } else {
      rejectedFiles.forEach((rejection) => {
        toast.error(`Could not upload ${rejection.file.name}`, {
          description: getAttachmentRejectionMessage(rejection, maxFileSize),
        });
      });
    }

    if (acceptedFiles.length > 0) {
      void uploadAttachmentsConcurrently(
        acceptedFiles,
        uploadMutation.mutateAsync,
      );
    }
  };

  const {
    getRootProps,
    getInputProps,
    isDragActive,
    open: openFilePicker,
  } = useDropzone({
    onDrop,
    multiple: true,
    maxFiles: MAX_ATTACHMENT_BATCH_FILES,
    maxSize: maxFileSize,
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
      file.mimeType.startsWith("image/") || file.mimeType.startsWith("video/"),
  );
  const otherFiles = attachments.filter(
    (file) =>
      !file.mimeType.startsWith("image/") &&
      !file.mimeType.startsWith("video/"),
  );
  const uploadedAttachments = [...imagesAndVideos, ...otherFiles].filter(
    (file) => !file.id.startsWith("temp-"),
  );

  if (isPending) {
    return <AttachmentsSkeleton />;
  }

  const hasAttachments = attachments.length > 0;

  const handleDelete = (attachmentId: string) => {
    deleteMutation.mutate(attachmentId);
  };

  return (
    <Box className={className} suppressHydrationWarning>
      <Flex
        align="center"
        className={cn(
          "border-border",
          hasAttachments && "border-b-[0.5px] pb-2",
        )}
        justify="between"
      >
        <Text as="h4" className="flex items-center gap-1" fontWeight="semibold">
          <AttachmentIcon className="h-5 w-auto" />
          Attachments
        </Text>
        {hasAttachments ? (
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
        ) : null}
      </Flex>
      {!hasAttachments && (
        <Box>
          <DropZone>
            <DropZone.Root
              className={cn("dark:bg-surface/80", {
                "dark:bg-surface-muted/80": isDragActive,
              })}
              isDragActive={isDragActive}
              rootProps={getRootProps()}
            >
              <DropZone.Input inputProps={getInputProps()} />
              <DropZone.Body
                isDragActive={isDragActive}
                message={`Drag and drop up to ${MAX_ATTACHMENT_BATCH_FILES} files here, or click to select them.`}
              />
            </DropZone.Root>
          </DropZone>
        </Box>
      )}
      {hasAttachments ? <input {...getInputProps()} /> : null}
      {imagesAndVideos.length > 0 && (
        <Box className="mt-3 grid grid-cols-3 gap-3 md:grid-cols-4 lg:grid-cols-6">
          {imagesAndVideos.map((file) => (
            <StoryAttachmentPreview
              file={file}
              files={uploadedAttachments}
              key={file.id}
              onDelete={() => {
                handleDelete(file.id);
              }}
              onDeleteFile={(attachment) => {
                handleDelete(attachment.id);
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
              files={uploadedAttachments}
              key={file.id}
              onDelete={() => {
                handleDelete(file.id);
              }}
              onDeleteFile={(attachment) => {
                handleDelete(attachment.id);
              }}
              onDownload={() => window.open(file.url, "_blank")}
            />
          ))}
        </Box>
      )}
    </Box>
  );
};

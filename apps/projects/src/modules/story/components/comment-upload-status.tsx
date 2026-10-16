import { AttachmentIcon, CloseIcon, LoadingIcon } from "icons";
import { Box, Button, Flex, Text } from "ui";
import type { AttachmentUpload } from "../hooks/use-comment-attachments";

export const CommentUploadStatus = ({
  uploads,
  onRetry,
  onRemove,
}: {
  uploads: AttachmentUpload[];
  onRetry: (id: string) => void;
  onRemove: (id: string) => void;
}) => {
  if (uploads.length === 0) return null;
  return (
    <Box aria-live="polite" className="space-y-2">
      {uploads.map((upload) => (
        <Flex
          align="center"
          className="border-border/60 rounded-lg border px-3 py-2"
          gap={2}
          key={upload.id}
        >
          {upload.status === "uploading" ? (
            <LoadingIcon className="size-4 shrink-0 animate-spin" />
          ) : (
            <AttachmentIcon className="size-4 shrink-0" />
          )}
          <Box className="min-w-0 flex-1">
            <Text className="truncate" fontSize="sm">
              {upload.file.name}
            </Text>
            <Text color="muted" fontSize="sm">
              {upload.error ?? "Uploading…"}
            </Text>
          </Box>
          {upload.status === "failed" ? (
            <>
              <Button
                color="tertiary"
                onClick={() => {
                  onRetry(upload.id);
                }}
                size="sm"
                variant="naked"
              >
                Retry
              </Button>
              <Button
                aria-label={`Remove ${upload.file.name}`}
                asIcon
                color="tertiary"
                onClick={() => {
                  onRemove(upload.id);
                }}
                size="sm"
                variant="naked"
              >
                <CloseIcon className="size-4" />
              </Button>
            </>
          ) : null}
        </Flex>
      ))}
    </Box>
  );
};

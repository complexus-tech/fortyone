import { Box } from "ui";
import { toast } from "sonner";
import { cn } from "lib";
import type { StoryAttachment } from "../types";
import { StoryAttachmentPreview } from "./story-attachment-preview";
import { downloadCommentAttachment } from "./comment-attachment-download";

const downloadAttachment = (file: StoryAttachment) => {
  void downloadCommentAttachment(file).catch((error: unknown) => {
    toast.error("Could not download attachment", {
      description: error instanceof Error ? error.message : "Try again.",
    });
  });
};

export const CommentAttachmentList = ({
  files,
  onRemove,
  className,
}: {
  files: StoryAttachment[];
  onRemove?: (id: string) => void;
  className?: string;
}) => (
  <Box
    className={cn(
      "not-prose grid max-w-[600px] grid-cols-[repeat(auto-fill,minmax(min(100%,124px),1fr))] gap-[6px]",
      className,
    )}
  >
    {files.map((file) => (
      <StoryAttachmentPreview
        className="min-w-0"
        file={file}
        files={files}
        isInChat
        key={file.id}
        onDelete={
          onRemove
            ? () => {
                onRemove(file.id);
              }
            : undefined
        }
        onDownload={downloadAttachment}
        unoptimizedImages
        variant="comment"
      />
    ))}
  </Box>
);

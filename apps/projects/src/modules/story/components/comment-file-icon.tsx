import { FileTypeIcon } from "@/components/ui/file-type-icon";
import { getCommentFileType } from "./comment-file-type";

export const CommentFileIcon = ({
  filename,
  mimeType,
  className,
}: {
  filename?: string;
  mimeType: string;
  className?: string;
}) => (
  <FileTypeIcon
    className={className}
    kind={getCommentFileType(mimeType, filename)}
  />
);

import { CachedAttachmentThumbnail } from "@/components/ui/attachment-preview/cached-thumbnail";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { getChatAttachmentThumbnailUrl } from "./attachment-url";
import { Video02Icon } from "icons";
import { cn } from "lib";
import { FileTypeIcon } from "@/components/ui/file-type-icon";
import { CommentFileIcon } from "@/modules/story/components/comment-file-icon";
import { getAttachmentTypeLabel } from "@/modules/story/components/attachment-file-info";
import type { ChatResource } from "./types";
import { formatFileSize } from "./attachment-url";

export const resourceTitle = (resource: ChatResource) =>
  resource.kind === "file"
    ? resource.fileName || resource.title
    : resource.title || resource.url;

export const resourceMetadata = (resource: ChatResource) => {
  if (resource.kind === "file") {
    const type = getAttachmentTypeLabel({
      mimeType: resource.mediaType,
      filename: resource.fileName,
    });
    return `${type} · ${formatFileSize(resource.sizeBytes)}`;
  }
  try {
    return new URL(resource.url).hostname;
  } catch {
    return resource.url;
  }
};

export const ResourceIcon = ({
  resource,
  className,
}: {
  resource: ChatResource;
  className?: string;
}) =>
  resource.kind === "link" ? (
    <FileTypeIcon className={className} kind="link" />
  ) : (
    <CommentFileIcon
      className={cn("size-10", className)}
      filename={resource.fileName}
      mimeType={resource.mediaType}
    />
  );

export const ResourceSharedBy = ({ resource }: { resource: ChatResource }) => (
  <p className="text-text-muted truncate">
    Shared by {resource.author.fullName} on{" "}
    <time dateTime={resource.createdAt}>
      {new Date(resource.createdAt).toLocaleDateString(undefined, {
        month: "short",
        day: "numeric",
        ...(new Date(resource.createdAt).getFullYear() !==
        new Date().getFullYear()
          ? { year: "numeric" as const }
          : {}),
      })}
    </time>
  </p>
);

export const ResourceThumbnail = ({
  resource,
  url,
}: {
  resource: ChatResource;
  url: string | null;
}) => {
  const { workspaceSlug } = useWorkspacePath();
  const isImage = resource.mediaType.toLowerCase().startsWith("image/");
  return (
    <span className="border-border bg-surface-muted relative flex aspect-[4/3] w-full items-center justify-center overflow-hidden rounded-lg border">
      {isImage && url && resource.kind === "file" ? (
        <CachedAttachmentThumbnail
          className="absolute inset-0 aspect-auto h-full w-full"
          fileName={resource.fileName}
          url={getChatAttachmentThumbnailUrl(
            workspaceSlug,
            resource.conversationId,
            resource.attachmentId,
          )}
        />
      ) : (
        <ResourceIcon className="size-9" resource={resource} />
      )}
      {!isImage ? (
        <span className="bg-background/80 text-foreground absolute right-2 bottom-2 flex items-center gap-1 rounded-md px-2 py-1 backdrop-blur-sm">
          <Video02Icon className="size-3.5" /> Video
        </span>
      ) : null}
    </span>
  );
};

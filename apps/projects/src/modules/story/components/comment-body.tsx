import { useMemo, useSyncExternalStore } from "react";
import { Box } from "ui";
import { splitCommentAttachments } from "./comment-attachments";
import { CommentAttachmentList } from "./comment-attachment-list";
import { renderCommentHTML, sanitizeCommentHTML } from "./comment-content";

const subscribe = () => () => undefined;
const clientSnapshot = () => true;
const serverSnapshot = () => false;

export const CommentBody = ({
  html,
  className,
}: {
  html: string;
  className?: string;
}) => {
  const hydrated = useSyncExternalStore(
    subscribe,
    clientSnapshot,
    serverSnapshot,
  );
  const content = useMemo(() => {
    if (!hydrated) return { html: sanitizeCommentHTML(html), attachments: [] };
    const content = splitCommentAttachments(html);
    return { ...content, html: renderCommentHTML(content.html) };
  }, [html, hydrated]);
  return (
    <Box className={className}>
      <Box
        className="story-comment-content prose prose-stone dark:prose-invert prose-headings:font-semibold prose-a:text-primary prose-pre:bg-surface-muted prose-pre:text-foreground max-w-full leading-6"
        html={content.html}
      />
      {content.attachments.length > 0 ? (
        <Box className="mt-3">
          <CommentAttachmentList files={content.attachments} />
        </Box>
      ) : null}
    </Box>
  );
};

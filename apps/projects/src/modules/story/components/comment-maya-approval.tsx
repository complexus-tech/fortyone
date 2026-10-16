import { Box, Button, Text } from "ui";
import type { CommentMayaRun } from "../types/comment-maya";
import { useCommentMayaAction } from "../hooks/comment-maya";

/** The review belongs to Maya's persisted reply, alongside the conversation. */
export const CommentMayaApproval = ({
  run,
  storyId,
  currentUserId,
}: {
  run: CommentMayaRun;
  storyId: string;
  currentUserId?: string;
}) => {
  const action = useCommentMayaAction(storyId);
  if (run.status !== "awaiting_approval") return null;
  const hasReview = Boolean(run.proposalReview?.trim());
  return (
    <Box className="mt-2 ml-9">
      {hasReview ? (
        <details className="text-sm">
          <summary className="text-text-muted hover:text-foreground focus-visible:ring-primary w-fit cursor-pointer rounded py-1 outline-none focus-visible:ring-2">
            Review proposed changes
          </summary>
          <div
            aria-label="Proposed changes"
            className="border-border/50 bg-surface-muted/30 focus-visible:ring-primary mt-1 max-h-64 overflow-y-auto rounded-lg border px-3 py-2 outline-none focus-visible:ring-2"
            role="region"
            // eslint-disable-next-line jsx-a11y/no-noninteractive-tabindex -- The bounded review must support keyboard scrolling.
            tabIndex={0}
          >
            <Text className="text-sm leading-relaxed break-words whitespace-pre-wrap">
              {run.proposalReview}
            </Text>
          </div>
        </details>
      ) : null}
      {run.requesterId === currentUserId ? (
        <Button
          className="mt-1 px-2"
          color="tertiary"
          disabled={action.isPending || !hasReview}
          onClick={() => {
            action.mutate({ run, action: "approve" });
          }}
          size="sm"
          variant="outline"
        >
          Approve and apply
        </Button>
      ) : null}
    </Box>
  );
};

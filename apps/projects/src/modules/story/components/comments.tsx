import { Box, Button, Skeleton } from "ui";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { useCommentFeed } from "../hooks/use-comment-feed";
import { useCommentMayaAction } from "../hooks/comment-maya";
import { useMayaPendingReplies } from "../hooks/use-maya-pending-replies";
import { CommentRow } from "./comment-row";
import { CommentLatestControl } from "./comment-latest-control";
import { CommentFeedNotices } from "./comment-feed-notices";

const CommentList = ({
  isDialog,
  storyId,
  teamId,
}: {
  isDialog?: boolean;
  storyId: string;
  teamId: string;
}) => {
  const {
    commentsQuery,
    stateQuery,
    threadQuery,
    targetId,
    highlightedId,
    revealKey,
    loadedTarget,
    comments,
    containerRef,
    stateById,
    hasUnread,
    latestCommentId,
    jumpToLatest,
    currentUserId,
    mayaQuery,
  } = useCommentFeed(storyId);
  const mayaAction = useCommentMayaAction(storyId);
  const mayaPendingByThreadId = useMayaPendingReplies(
    comments,
    mayaQuery.data?.runs,
    mayaQuery.isError || mayaQuery.fetchStatus === "paused",
  );

  return (
    <Box ref={containerRef}>
      {hasUnread && latestCommentId ? (
        <CommentLatestControl
          commentId={latestCommentId}
          containerRef={containerRef}
          onJump={jumpToLatest}
        />
      ) : null}
      <CommentFeedNotices
        commentsQuery={commentsQuery}
        detailsQuery={stateQuery}
        targetQuery={targetId && !loadedTarget ? threadQuery : undefined}
      />
      {comments.map((comment) => (
        <CommentRow
          checkingMaya={mayaQuery.isFetching || commentsQuery.isRefetching}
          comment={comment}
          currentUserId={currentUserId}
          highlightedId={highlightedId}
          isDialog={isDialog}
          key={comment.id}
          lastReadAt={
            stateQuery.data
              ? stateQuery.data.threadReads?.[comment.id] ?? null
              : undefined
          }
          mayaPending={mayaPendingByThreadId.get(comment.id)}
          mayaRuns={mayaQuery.data?.runs}
          onCheckMaya={() => {
            void mayaQuery.refetch();
            void commentsQuery.refetch();
          }}
          onRetryMaya={(run) => {
            mayaAction.mutate({ run, action: "retry" });
          }}
          readObservedAt={stateQuery.data?.observedAt}
          retryingMayaRunId={
            mayaAction.isPending ? mayaAction.variables.run.id : undefined
          }
          revealKey={revealKey}
          stateById={stateById}
          storyId={storyId}
          targetId={targetId}
          teamId={teamId}
        />
      ))}
      {commentsQuery.isPending || commentsQuery.isFetchingNextPage ? (
        <Box className="mt-4 space-y-3">
          {[0, 1].map((index) => (
            <Box className="flex gap-3" key={index}>
              <Skeleton className="h-8 w-8 rounded-full" />
              <Box className="flex-1 space-y-2">
                <Skeleton className="h-4 w-32" />
                <Skeleton className="h-12 w-full" />
              </Box>
            </Box>
          ))}
        </Box>
      ) : null}
      {commentsQuery.hasNextPage ? (
        <Button
          className="ml-6 px-3 text-[0.95rem]"
          color="tertiary"
          disabled={commentsQuery.isFetchingNextPage}
          onClick={() => {
            void commentsQuery.fetchNextPage();
          }}
          size="sm"
          variant="naked"
        >
          {commentsQuery.isFetchingNextPage
            ? "Loading..."
            : "Load more comments"}
        </Button>
      ) : null}
    </Box>
  );
};

export const Comments = (props: {
  isDialog?: boolean;
  storyId: string;
  teamId: string;
}) => {
  const { workspaceSlug } = useWorkspacePath();
  const { data: session } = useSession();
  return (
    <CommentList
      key={`${workspaceSlug}:${session?.user.id}:${props.storyId}`}
      {...props}
    />
  );
};

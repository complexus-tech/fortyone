import { Box, Button, Flex, Skeleton, Text } from "ui";
import { ArrowUpIcon } from "icons";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { useCommentFeed } from "../hooks/use-comment-feed";
import { isUnreadComment } from "../lib/comment-navigation";
import { CommentRow } from "./comment-row";

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
    revealKey,
    loadedTarget,
    comments,
    containerRef,
    lastReadAt,
    stateById,
    hasUnread,
    jumpToLatest,
    currentUserId,
  } = useCommentFeed(storyId);

  return (
    <Box ref={containerRef}>
      {hasUnread ? (
        <Flex align="center" className="mb-3 ml-10" justify="end">
          <Button
            className="px-2 text-xs"
            color="tertiary"
            leftIcon={<ArrowUpIcon className="size-3.5" />}
            onClick={jumpToLatest}
            size="sm"
            variant="naked"
          >
            Jump to latest
          </Button>
        </Flex>
      ) : null}
      {targetId && !loadedTarget && threadQuery.isPending ? (
        <Text className="mb-3 ml-10 text-sm" color="muted">
          Finding comment…
        </Text>
      ) : null}
      {targetId && !loadedTarget && threadQuery.isError ? (
        <Text className="mb-3 ml-10 text-sm" color="muted">
          This comment is no longer available or you do not have access.
        </Text>
      ) : null}
      {stateQuery.isError ? (
        <Flex align="center" className="mb-3 ml-10" gap={2}>
          <Text className="text-sm" color="muted">
            Some comment details could not load.
          </Text>
          <Button
            color="tertiary"
            onClick={() => {
              void stateQuery.refetch();
            }}
            size="sm"
            variant="naked"
          >
            Retry
          </Button>
        </Flex>
      ) : null}
      {commentsQuery.isError ? (
        <Flex align="center" className="mb-3 ml-10" gap={2}>
          <Text className="text-sm" color="muted">
            Could not load comments.
          </Text>
          <Button
            color="tertiary"
            onClick={() => {
              void commentsQuery.refetch();
            }}
            size="sm"
            variant="naked"
          >
            Retry
          </Button>
        </Flex>
      ) : null}
      {comments.map((comment, index) => (
        <CommentRow
          comment={comment}
          currentUserId={currentUserId}
          isDialog={isDialog}
          key={comment.id}
          lastReadAt={lastReadAt}
          revealKey={revealKey}
          showUnreadDivider={
            isUnreadComment(comment, lastReadAt, currentUserId) &&
            (index === 0 ||
              !isUnreadComment(comments[index - 1], lastReadAt, currentUserId))
          }
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

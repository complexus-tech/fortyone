import { Button, Flex, Text } from "ui";
import { isCommentUnavailableError } from "../lib/comment-errors";

type RetryableQuery = {
  isError: boolean;
  refetch: () => unknown;
};

export const CommentFeedNotices = ({
  targetQuery,
  detailsQuery,
  commentsQuery,
}: {
  targetQuery?: RetryableQuery & {
    isPending: boolean;
    isFetching?: boolean;
    error?: unknown;
  };
  detailsQuery: RetryableQuery;
  commentsQuery: RetryableQuery;
}) => (
  <>
    {targetQuery?.isPending ? (
      <Text className="mb-3 ml-10 text-sm" color="muted">
        Finding comment…
      </Text>
    ) : null}
    {targetQuery?.isError && isCommentUnavailableError(targetQuery.error) ? (
      <Text className="mb-3 ml-10 text-sm" color="muted">
        This comment is no longer available or you do not have access.
      </Text>
    ) : null}
    {targetQuery?.isError && !isCommentUnavailableError(targetQuery.error) ? (
      <Flex align="center" className="mb-3 ml-10" gap={2}>
        <Text color="muted">Could not refresh this comment.</Text>
        <Button
          color="tertiary"
          disabled={targetQuery.isFetching}
          onClick={() => {
            void targetQuery.refetch();
          }}
          size="sm"
          variant="naked"
        >
          {targetQuery.isFetching ? "Retrying…" : "Retry"}
        </Button>
      </Flex>
    ) : null}
    {detailsQuery.isError ? (
      <Flex align="center" className="mb-3 ml-10" gap={2}>
        <Text className="text-sm" color="muted">
          Some comment details could not load.
        </Text>
        <Button
          color="tertiary"
          onClick={() => {
            void detailsQuery.refetch();
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
  </>
);

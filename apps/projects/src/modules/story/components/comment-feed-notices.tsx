import { Button, Flex, Text } from "ui";

type RetryableQuery = {
  isError: boolean;
  refetch: () => unknown;
};

export const CommentFeedNotices = ({
  targetQuery,
  detailsQuery,
  commentsQuery,
}: {
  targetQuery?: { isPending: boolean; isError: boolean };
  detailsQuery: RetryableQuery;
  commentsQuery: RetryableQuery;
}) => (
  <>
    {targetQuery?.isPending ? (
      <Text className="mb-3 ml-10 text-sm" color="muted">
        Finding comment…
      </Text>
    ) : null}
    {targetQuery?.isError ? (
      <Text className="mb-3 ml-10 text-sm" color="muted">
        This comment is no longer available or you do not have access.
      </Text>
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

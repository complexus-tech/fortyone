import { ArrowUpIcon, BellIcon, CopyIcon, ShareIcon } from "icons";
import { Avatar, Box, Button, Flex, Text } from "ui";
import { PublicPortalShell } from "./portal-shell";
import { RequestStatusPill } from "./request-card";
import type { PublicPortal, PublicPortalViewer, PublicRequest } from "./types";
import { getBoard } from "./utils";

export const PublicPortalRequestDetailPage = ({
  portal,
  request,
  viewer,
}: {
  portal: PublicPortal;
  request: PublicRequest;
  viewer?: PublicPortalViewer | null;
}) => {
  const board = getBoard(portal, request.boardId);

  return (
    <PublicPortalShell activeTab="requests" portal={portal} viewer={viewer}>
      <Box className="mx-auto grid w-full max-w-[78rem] gap-7 px-4 py-8 md:grid-cols-[minmax(0,1fr)_19rem] md:px-6">
        <main>
          <Flex align="start" gap={3}>
            <Avatar name={request.authorName} size="md" />
            <Box className="min-w-0 flex-1">
              <Text className="text-[1.05rem]" fontWeight="semibold">
                {request.authorName}
              </Text>
              <Text color="muted">{request.createdAtLabel}</Text>
            </Box>
          </Flex>

          <Text as="h1" className="mt-8 text-2xl" fontWeight="semibold">
            {request.title}
          </Text>
          <Text
            className="mt-4 max-w-3xl text-[1.02rem] leading-7"
            color="muted"
          >
            {request.description}
          </Text>

          <Flex align="center" className="mt-6" gap={3}>
            <RequestStatusPill status={request.status} />
            <Flex align="center" className="text-text-muted" gap={1}>
              <span>{request.commentCount} comments</span>
            </Flex>
            <Button
              className="ml-auto h-8 rounded-xl border-none px-3"
              color="tertiary"
              leftIcon={<ArrowUpIcon className="h-3.5" />}
              size="sm"
            >
              {request.voteCount}
            </Button>
          </Flex>

          <Box className="border-border/70 mt-8 border-t-[0.5px] pt-8">
            <Box className="border-border bg-surface rounded-3xl border-[0.5px] p-4">
              <Text className="text-[1rem]" color="muted">
                Add a comment...
              </Text>
              <Flex align="end" className="mt-16" justify="between">
                <Flex align="center" className="text-text-muted" gap={3}>
                  <span className="size-4 rounded-full border border-current" />
                  <span className="size-4 rounded-sm border border-current" />
                </Flex>
                <Button color="tertiary" size="sm">
                  Comment
                </Button>
              </Flex>
            </Box>

            {request.comments.length > 0 ? (
              <Box className="mt-6 space-y-5">
                {request.comments.map((comment) => (
                  <Flex align="start" gap={3} key={comment.id}>
                    <Avatar name={comment.authorName} size="sm" />
                    <Box>
                      <Flex align="center" gap={2}>
                        <Text fontWeight="semibold">{comment.authorName}</Text>
                        <Text color="muted">{comment.createdAtLabel}</Text>
                      </Flex>
                      <Text className="mt-1 leading-6" color="muted">
                        {comment.body}
                      </Text>
                    </Box>
                  </Flex>
                ))}
              </Box>
            ) : null}
          </Box>
        </main>

        <aside className="space-y-8">
          <Button
            className="h-12 w-full justify-center text-[1rem]"
            color="primary"
            leftIcon={<ArrowUpIcon className="h-4" />}
            rounded="xl"
          >
            Vote for this request
          </Button>
          <Box className="border-border bg-surface rounded-3xl border-[0.5px] p-5">
            <Flex className="py-2" justify="between">
              <Text color="muted">Status</Text>
              <RequestStatusPill status={request.status} />
            </Flex>
            {board ? (
              <Flex className="py-2" justify="between">
                <Text color="muted">Board</Text>
                <Flex align="center" gap={2}>
                  <span
                    className={`${board.colorClassName} size-2 rounded-full`}
                  />
                  <Text>{board.name}</Text>
                </Flex>
              </Flex>
            ) : null}
            <Flex className="py-2" justify="between">
              <Text color="muted">Posted</Text>
              <Text>{request.createdAtLabel}</Text>
            </Flex>
          </Box>

          <Box>
            <Text className="mb-4" fontWeight="semibold">
              Actions
            </Text>
            <Flex className="text-text-muted gap-4" direction="column">
              <Flex align="center" gap={2}>
                <BellIcon className="h-4" />
                <span>Follow this post</span>
              </Flex>
              <Flex align="center" gap={2}>
                <CopyIcon className="h-4" />
                <span>Copy link</span>
              </Flex>
              <Flex align="center" gap={2}>
                <ShareIcon className="h-4" />
                <span>Share</span>
              </Flex>
            </Flex>
          </Box>
        </aside>
      </Box>
    </PublicPortalShell>
  );
};

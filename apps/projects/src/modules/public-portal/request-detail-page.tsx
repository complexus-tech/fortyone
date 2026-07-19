import { Avatar, Box, Flex, Text } from "ui";
import { PublicPortalShell } from "./portal-shell";
import { RequestStatusPill } from "./request-card";
import type { PublicPortal, PublicPortalViewer, PublicRequest } from "./types";
import { getBoard } from "./utils";
import { FeedbackVoteButton } from "./feedback-controls";
import { getPublicAvatarColor } from "./avatar-color";
import { FeedbackDiscussion } from "./feedback-comments";
import { PublicPortalActions } from "./sidebar";

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
    <PublicPortalShell activeTab="feedback" portal={portal} viewer={viewer}>
      <Box className="mx-auto grid w-full max-w-[78rem] gap-7 px-4 py-8 md:grid-cols-[minmax(0,1fr)_19rem] md:px-6">
        <main>
          <Flex align="start" gap={3}>
            <Avatar
              name={request.authorName}
              size="md"
              src={request.authorAvatar}
              style={{
                backgroundColor: getPublicAvatarColor(request.authorName),
              }}
            />
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
            <Box className="ml-auto">
              <FeedbackVoteButton
                portal={portal}
                request={request}
                showDownvote
              />
            </Box>
          </Flex>

          <Box className="border-border/70 mt-8 border-t-[0.5px] pt-8">
            <FeedbackDiscussion
              portal={portal}
              request={request}
              viewer={viewer}
            />
          </Box>
        </main>

        <aside className="space-y-8">
          <Box className="border-border bg-surface rounded-xl border-[0.5px] p-5">
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
          <PublicPortalActions portal={portal} />
        </aside>
      </Box>
    </PublicPortalShell>
  );
};

import { BellIcon, CopyIcon, ShareIcon } from "icons";
import { Avatar, Box, Flex, Text } from "ui";
import type { Team } from "@/modules/teams/types";
import { PublicPortalShell } from "./portal-shell";
import { RequestStatusPill } from "./request-card";
import type { PublicPortal, PublicPortalViewer, PublicRequest } from "./types";
import { getBoard } from "./utils";
import {
  CreateStoryFromFeedbackButton,
  FeedbackCommentComposer,
  FeedbackVoteButton,
} from "./feedback-controls";
import { getPublicAvatarColor } from "./avatar-color";

const EMPTY_TEAMS: Team[] = [];

export const PublicPortalRequestDetailPage = ({
  portal,
  request,
  viewer,
  teams = EMPTY_TEAMS,
}: {
  portal: PublicPortal;
  request: PublicRequest;
  viewer?: PublicPortalViewer | null;
  teams?: Team[];
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
              <FeedbackVoteButton portal={portal} request={request} />
            </Box>
          </Flex>

          <Box className="border-border/70 mt-8 border-t-[0.5px] pt-8">
            <FeedbackCommentComposer portal={portal} request={request} />

            {request.comments.length > 0 ? (
              <Box className="mt-6 space-y-5">
                {request.comments.map((comment) => (
                  <Flex align="start" gap={3} key={comment.id}>
                    <Avatar
                      name={comment.authorName}
                      size="sm"
                      src={comment.authorAvatar}
                      style={{
                        backgroundColor: getPublicAvatarColor(
                          comment.authorName,
                        ),
                      }}
                    />
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
          <FeedbackVoteButton portal={portal} request={request} />
          <CreateStoryFromFeedbackButton
            portal={portal}
            request={request}
            teams={teams}
          />
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

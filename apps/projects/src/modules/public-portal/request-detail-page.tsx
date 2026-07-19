"use client";

import { Avatar, Box, Flex, Text } from "ui";
import { PublicPortalShell } from "./portal-shell";
import { RequestStatusPill } from "./request-card";
import type { PublicPortal, PublicPortalViewer, PublicRequest } from "./types";
import { getBoard, getRequestCallbackUrl } from "./utils";
import { FeedbackVoteButton } from "./feedback-controls";
import { getPublicAvatarColor } from "./avatar-color";
import { FeedbackDiscussion } from "./feedback-comments";
import { PublicPortalActions } from "./sidebar";
import { usePublicFeedbackDetail } from "./client-query";

export const PublicPortalRequestDetailPage = ({
  portal,
  request,
  viewer,
}: {
  portal: PublicPortal;
  request: PublicRequest;
  viewer?: PublicPortalViewer | null;
}) => {
  const { data: activeRequest } = usePublicFeedbackDetail({ portal, request });
  const board = getBoard(portal, activeRequest.boardId);

  return (
    <PublicPortalShell
      activeTab="feedback"
      loginCallbackUrl={getRequestCallbackUrl(portal, activeRequest)}
      portal={portal}
      viewer={viewer}
    >
      <Box className="mx-auto grid w-full max-w-[78rem] gap-7 px-4 py-8 md:grid-cols-[minmax(0,1fr)_19rem] md:px-6">
        <main>
          <Flex align="start" gap={3}>
            <Avatar
              name={activeRequest.authorName}
              size="md"
              src={activeRequest.authorAvatar}
              style={{
                backgroundColor: getPublicAvatarColor(activeRequest.authorName),
              }}
            />
            <Box className="min-w-0 flex-1">
              <Text className="text-[1.05rem]" fontWeight="semibold">
                {activeRequest.authorName}
              </Text>
              <Text color="muted">{activeRequest.createdAtLabel}</Text>
            </Box>
          </Flex>

          <Text as="h1" className="mt-8 text-2xl" fontWeight="semibold">
            {activeRequest.title}
          </Text>
          <Text
            className="mt-4 max-w-3xl text-[1.02rem] leading-7"
            color="muted"
          >
            {activeRequest.description}
          </Text>

          <Flex align="center" className="mt-6" gap={3}>
            <RequestStatusPill status={activeRequest.status} />
            <Flex align="center" className="text-text-muted" gap={1}>
              <span>{activeRequest.commentCount} comments</span>
            </Flex>
            <Box className="ml-auto">
              <FeedbackVoteButton
                portal={portal}
                request={activeRequest}
                showDownvote
              />
            </Box>
          </Flex>

          <Box className="border-border/70 mt-8 border-t-[0.5px] pt-8">
            <FeedbackDiscussion
              portal={portal}
              request={activeRequest}
              viewer={viewer}
            />
          </Box>
        </main>

        <aside className="space-y-8">
          <Box className="border-border bg-surface rounded-xl border-[0.5px] p-5">
            <Flex className="py-2" justify="between">
              <Text color="muted">Status</Text>
              <RequestStatusPill status={activeRequest.status} />
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
              <Text>{activeRequest.createdAtLabel}</Text>
            </Flex>
          </Box>
          <PublicPortalActions portal={portal} />
        </aside>
      </Box>
    </PublicPortalShell>
  );
};

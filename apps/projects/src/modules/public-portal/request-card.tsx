import Link from "next/link";
import { CommentIcon } from "icons";
import { Avatar, Box, Flex, Text } from "ui";
import { cn } from "lib";
import { requestStatusMeta } from "./status";
import type { PublicPortal, PublicRequest } from "./types";
import { getBoard, getRequestPath } from "./utils";
import { getPublicAvatarColor } from "./avatar-color";
import { FeedbackVoteButton } from "./feedback-controls";

export const RequestStatusPill = ({
  status,
}: {
  status: PublicRequest["status"];
}) => {
  const meta = requestStatusMeta[status];
  return (
    <span
      className={cn(
        "inline-flex h-7 items-center gap-2 rounded-lg border px-2.5 text-[0.92rem] font-medium",
        meta.badgeClassName,
      )}
    >
      <span className={cn("size-2 rounded-full", meta.dotClassName)} />
      {meta.label}
    </span>
  );
};

export const PublicRequestCard = ({
  portal,
  request,
}: {
  portal: PublicPortal;
  request: PublicRequest;
}) => {
  const board = getBoard(portal, request.boardId);

  return (
    <Box className="hover:bg-state-hover/25 group transition-colors">
      <Box className="border-border/70 border-b-[0.5px] py-5">
        <Flex align="start" className="gap-3">
          <Link
            className="flex min-w-0 flex-1 gap-4"
            href={getRequestPath(portal, request)}
          >
            <Avatar
              className="mt-0.5"
              name={request.authorName}
              size="sm"
              src={request.authorAvatar}
              style={{
                backgroundColor: getPublicAvatarColor(request.authorName),
              }}
            />
            <Box className="min-w-0 flex-1">
              <Flex align="center" className="min-w-0 flex-wrap gap-1.5">
                <Text fontWeight="semibold">{request.authorName}</Text>
                {board ? (
                  <>
                    <Text color="muted">in</Text>
                    <Text color="muted">{board.name}</Text>
                  </>
                ) : null}
              </Flex>
              <Text
                className="mt-1.5 line-clamp-1 text-[1.08rem] group-hover:opacity-90"
                fontWeight="semibold"
              >
                {request.title}
              </Text>
              {request.description ? (
                <Text className="mt-1.5 line-clamp-2" color="muted">
                  {request.description}
                </Text>
              ) : null}
              <Flex align="center" className="mt-4 gap-2">
                <RequestStatusPill status={request.status} />
                {request.commentCount > 0 ? (
                  <Flex
                    align="center"
                    aria-label={`${request.commentCount} comments`}
                    className="text-text-muted gap-1"
                  >
                    <CommentIcon className="h-4" />
                    <span>{request.commentCount}</span>
                  </Flex>
                ) : null}
              </Flex>
            </Box>
          </Link>
          <FeedbackVoteButton compact portal={portal} request={request} />
        </Flex>
      </Box>
    </Box>
  );
};

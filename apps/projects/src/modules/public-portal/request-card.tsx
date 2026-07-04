import Link from "next/link";
import { ArrowUpIcon, CommentIcon } from "icons";
import { Avatar, Box, Button, Flex, Text } from "ui";
import { cn } from "lib";
import { requestStatusMeta } from "./status";
import type { PublicPortal, PublicRequest } from "./types";
import { getBoard, getRequestPath } from "./utils";
import { getPublicAvatarColor } from "./avatar-color";

export const RequestStatusPill = ({
  status,
}: {
  status: PublicRequest["status"];
}) => {
  const meta = requestStatusMeta[status];
  return (
    <span
      className={cn(
        "inline-flex h-7 items-center gap-2 rounded-full border px-2.5 text-[0.92rem] font-medium",
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
    <Link className="group block" href={getRequestPath(portal, request)}>
      <Box className="border-border/70 border-b-[0.5px] py-5 transition">
        <Flex align="start" className="gap-4">
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
              <Text className="mt-1.5 line-clamp-2 max-w-2xl" color="muted">
                {request.description}
              </Text>
            ) : null}
            <Flex align="center" className="mt-4 gap-2">
              <RequestStatusPill status={request.status} />
              <Flex align="center" className="text-text-muted gap-1.5">
                <CommentIcon className="h-4" />
                <span>{request.commentCount}</span>
              </Flex>
            </Flex>
          </Box>
          <Button
            className="mt-1 h-7 shrink-0 gap-0 rounded-xl px-2"
            color="tertiary"
            leftIcon={<ArrowUpIcon className="-mr-1 h-3.5 text-current" />}
            size="xs"
            variant="naked"
          >
            {request.voteCount}
          </Button>
        </Flex>
      </Box>
    </Link>
  );
};

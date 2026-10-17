import { Avatar, Box, Button, Flex, Text } from "ui";
import { cn } from "lib";
import { useMayaAssignee } from "@/lib/hooks/members";
import type { CommentMayaPendingStatus } from "../types/comment-maya";
import { CommentAuthor } from "./comment-author";

const LABELS: Record<
  CommentMayaPendingStatus,
  { body: string; compact: string }
> = {
  thinking: { body: "Thinking", compact: "Maya is thinking" },
  working: {
    body: "Still working on this",
    compact: "Maya is still working on this",
  },
  delayed: {
    body: "This is taking longer than usual.",
    compact: "Maya is taking longer than usual.",
  },
  offline: {
    body: "Couldn’t check for a response.",
    compact: "Couldn’t check for Maya’s response.",
  },
  failed: {
    body: "Couldn’t finish this.",
    compact: "Maya couldn’t finish this.",
  },
};

export const CommentMayaPendingLabel = ({
  status,
  compact = false,
  className,
}: {
  status: CommentMayaPendingStatus;
  compact?: boolean;
  className?: string;
}) => {
  const active = status === "thinking" || status === "working";
  return (
    <span
      aria-atomic
      aria-live="polite"
      className={cn(
        "text-text-muted inline-flex items-center text-[0.95rem] leading-6",
        className,
      )}
      role="status"
    >
      <span className="sr-only">{LABELS[status].compact}</span>
      <span aria-hidden className="inline-flex items-baseline gap-[3px]">
        {compact ? LABELS[status].compact : LABELS[status].body}
        {active ? (
          <span className="comment-maya-dots inline-flex items-end gap-[2px]">
            <span />
            <span />
            <span />
          </span>
        ) : null}
      </span>
    </span>
  );
};

export const CommentMayaPending = ({
  status,
  isDialog,
  onCheckAgain,
  onRetry,
  retryDisabled,
  checkAgainDisabled,
}: {
  status: CommentMayaPendingStatus;
  isDialog?: boolean;
  onCheckAgain?: () => void;
  onRetry?: () => void;
  retryDisabled?: boolean;
  checkAgainDisabled?: boolean;
}) => {
  const { data: maya } = useMayaAssignee();
  return (
    <Box className="ml-9 pt-1">
      <Flex align="center" className="flex-wrap" gap={1}>
        {maya ? (
          <CommentAuthor
            isDialog={isDialog}
            user={{
              ...maya,
              fullName: maya.fullName || "Maya",
              username: maya.username || "maya",
              isSystem: true,
            }}
            userId={maya.id}
          />
        ) : (
          <Flex gap={1}>
            <Box
              className={cn(
                "bg-surface relative top-px flex aspect-square items-center rounded-full p-[0.3rem]",
                isDialog && "dark:bg-surface-elevated/95",
              )}
            >
              <Avatar name="Maya" size="xs" />
            </Box>
            <Text className="relative top-0.5 ml-1 text-black dark:text-white">
              maya
            </Text>
          </Flex>
        )}
      </Flex>
      <Box className="mt-0.5 ml-9">
        <CommentMayaPendingLabel status={status} />
        {(status === "delayed" || status === "offline") && onCheckAgain ? (
          <Button
            className="ml-2 px-0 text-[0.95rem]"
            color="tertiary"
            disabled={checkAgainDisabled}
            onClick={onCheckAgain}
            size="sm"
            variant="naked"
          >
            Check again
          </Button>
        ) : null}
        {status === "failed" && onRetry ? (
          <Button
            className="ml-2 px-0 text-[0.95rem]"
            color="tertiary"
            disabled={retryDisabled}
            onClick={onRetry}
            size="sm"
            variant="naked"
          >
            Retry
          </Button>
        ) : null}
      </Box>
    </Box>
  );
};

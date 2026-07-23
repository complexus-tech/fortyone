"use client";

import { useState } from "react";
import Link from "next/link";
import { usePathname, useRouter, useSearchParams } from "next/navigation";
import { cn } from "lib";
import {
  CheckIcon,
  CloseIcon,
  DeleteIcon,
  NotificationsCheckIcon,
  NotificationsUnreadIcon,
  RequestsIcon,
  UndoIcon,
} from "icons";
import { Avatar, Box, Button, ContextMenu, Flex, Text, TimeAgo } from "ui";
import { ConfirmDialog, Dot } from "@/components/ui";
import { LIST_ITEM_ATTENTION_BORDER } from "@/components/ui/list-item-attention";
import { useWorkspacePath } from "@/hooks";
import { useUserRole } from "@/hooks/role";
import { openDialogAfterMenuClose } from "@/utils/menu-dialog-state";
import { usePlanTeamFeedback } from "./hooks/use-plan-feedback";
import { useSetTeamFeedbackReadState } from "./hooks/use-read-state";
import {
  useRestoreTeamFeedback,
  useTrashTeamFeedback,
} from "./hooks/use-trash";
import { useUpdateTeamFeedbackStatus } from "./hooks/use-update-status";
import { FeedbackStatus } from "./status";
import type { TeamFeedbackItem } from "./types";

export const TeamFeedbackCard = ({
  feedback,
  index,
}: {
  feedback: TeamFeedbackItem;
  index: number;
}) => {
  const pathname = usePathname();
  const router = useRouter();
  const searchParams = useSearchParams();
  const { withWorkspace } = useWorkspacePath();
  const { userRole } = useUserRole();
  const planFeedback = usePlanTeamFeedback();
  const restoreFeedback = useRestoreTeamFeedback();
  const setReadState = useSetTeamFeedbackReadState();
  const trashFeedback = useTrashTeamFeedback();
  const updateStatus = useUpdateTeamFeedbackStatus();
  const [isClosing, setIsClosing] = useState(false);
  const [isTrashing, setIsTrashing] = useState(false);
  const isTrashed = Boolean(feedback.deletedAt);
  const isActive = !isTrashed && pathname.includes(feedback.id);
  const isLinked = feedback.storyLinks.some((link) => link.isPrimary);
  const isUnread = !isTrashed && !feedback.readAt;
  const canManageTrash = userRole === "admin";
  const canPlan = !isLinked && feedback.status !== "closed";
  const status = searchParams.get("status");
  const search = searchParams.get("search");
  const feedbackHref = withWorkspace(
    `/teams/${feedback.board.teamId}/feedback/${feedback.id}`,
  );
  const feedbackParams = new URLSearchParams();
  if (status) feedbackParams.set("status", status);
  if (search) feedbackParams.set("search", search);
  const feedbackQuery = feedbackParams.toString();
  const href = feedbackQuery
    ? `${feedbackHref}?${feedbackQuery}`
    : feedbackHref;

  const handlePlan = () => {
    planFeedback.mutate(
      {
        feedbackId: feedback.id,
        payload: { teamId: feedback.board.teamId },
      },
      {
        onSuccess: (response) => {
          if (!response.error?.message && response.data?.storyId) {
            router.push(withWorkspace(`/story/${response.data.storyId}`));
          }
        },
      },
    );
  };

  const handleRestore = () => {
    restoreFeedback.mutate(feedback.id);
  };

  const card = (
    <Box
      className={cn(
        "border-border block border-b-[0.5px] px-5 py-[0.655rem] transition md:px-4",
        {
          "bg-surface-muted": isActive,
          "cursor-default opacity-80": isTrashed,
          "hover:bg-surface-muted cursor-pointer": !isTrashed,
          [LIST_ITEM_ATTENTION_BORDER]: isUnread,
        },
      )}
    >
      <Flex align="center" className="mb-2" gap={2} justify="between">
        <Text
          className="line-clamp-1 flex-1 font-medium"
          color={isUnread ? undefined : "muted"}
        >
          {feedback.title}
        </Text>
        <Text className="shrink-0 text-[0.95rem]" color="muted">
          {isTrashed && feedback.restoreUntil ? (
            <>
              Deletes <TimeAgo timestamp={feedback.restoreUntil} />
            </>
          ) : (
            <TimeAgo timestamp={feedback.createdAt} />
          )}
        </Text>
      </Flex>
      <Flex align="center" gap={3} justify="between">
        <Flex align="center" className="min-w-0 flex-1" gap={2}>
          <Avatar
            className="shrink-0"
            name={feedback.authorName}
            size="xs"
            src={feedback.authorAvatar}
          />
          <Dot className="size-3" color={feedback.board.color} />
          <Text className="line-clamp-1" color="muted">
            in {feedback.board.name}
          </Text>
        </Flex>
        {isTrashed ? (
          <Button
            color="tertiary"
            disabled={!canManageTrash || restoreFeedback.isPending}
            leftIcon={<UndoIcon />}
            loading={restoreFeedback.isPending}
            onClick={handleRestore}
            size="xs"
            variant="outline"
          >
            Restore
          </Button>
        ) : (
          <FeedbackStatus status={feedback.status} />
        )}
      </Flex>
    </Box>
  );

  return (
    <ContextMenu>
      <ContextMenu.Trigger>
        {isTrashed ? (
          card
        ) : (
          <Link
            className="block"
            href={href}
            prefetch={index <= 10 ? true : null}
          >
            {card}
          </Link>
        )}
      </ContextMenu.Trigger>
      <ContextMenu.Items>
        {isTrashed ? (
          <ContextMenu.Group>
            <ContextMenu.Item
              disabled={!canManageTrash || restoreFeedback.isPending}
              onSelect={handleRestore}
            >
              <UndoIcon />
              Restore feedback
            </ContextMenu.Item>
          </ContextMenu.Group>
        ) : (
          <>
            <ContextMenu.Group>
              {isUnread ? (
                <ContextMenu.Item
                  disabled={setReadState.isPending}
                  onSelect={() => {
                    setReadState.mutate({
                      feedbackId: feedback.id,
                      isRead: true,
                    });
                  }}
                >
                  <NotificationsCheckIcon />
                  Mark as read
                </ContextMenu.Item>
              ) : (
                <ContextMenu.Item
                  disabled={setReadState.isPending}
                  onSelect={() => {
                    setReadState.mutate({
                      feedbackId: feedback.id,
                      isRead: false,
                    });
                  }}
                >
                  <NotificationsUnreadIcon />
                  Mark as unread
                </ContextMenu.Item>
              )}
              <ContextMenu.Item
                disabled={!canPlan || planFeedback.isPending}
                onSelect={handlePlan}
              >
                <CheckIcon className="text-icon" />
                Plan feedback
              </ContextMenu.Item>
              <ContextMenu.Item
                disabled={
                  isLinked ||
                  feedback.status === "reviewing" ||
                  feedback.status === "closed"
                }
                onSelect={() => {
                  updateStatus.mutate({
                    feedbackId: feedback.id,
                    payload: { status: "reviewing", roadmapSummary: null },
                  });
                }}
              >
                <RequestsIcon />
                Mark as reviewing
              </ContextMenu.Item>
              <ContextMenu.Item
                className="text-danger"
                disabled={isLinked || feedback.status === "closed"}
                onSelect={() => {
                  openDialogAfterMenuClose(setIsClosing);
                }}
              >
                <CloseIcon className="text-danger" />
                Close feedback...
              </ContextMenu.Item>
            </ContextMenu.Group>
            {canManageTrash ? (
              <>
                <ContextMenu.Separator className="my-2" />
                <ContextMenu.Group>
                  <ContextMenu.Item
                    className="text-danger"
                    disabled={isLinked || trashFeedback.isPending}
                    onSelect={() => {
                      openDialogAfterMenuClose(setIsTrashing);
                    }}
                  >
                    <DeleteIcon className="text-danger" />
                    Move to trash...
                  </ContextMenu.Item>
                </ContextMenu.Group>
              </>
            ) : null}
          </>
        )}
      </ContextMenu.Items>
      <ConfirmDialog
        confirmText="Move to trash"
        description="This feedback will be hidden from team lists and the public portal. You can restore it from Trash for 30 days."
        isLoading={trashFeedback.isPending}
        isOpen={isTrashing}
        loadingText="Moving..."
        onCancel={() => {
          setIsTrashing(false);
        }}
        onClose={() => {
          setIsTrashing(false);
        }}
        onConfirm={() => {
          trashFeedback.mutate(feedback.id, {
            onSuccess: () => {
              setIsTrashing(false);
            },
          });
        }}
        title="Move this feedback to trash?"
      />
      <ConfirmDialog
        confirmText="Close feedback"
        description="Closing removes this item from the team's active feedback queue. It will remain available in the feedback portal."
        isLoading={updateStatus.isPending}
        isOpen={isClosing}
        loadingText="Closing..."
        onCancel={() => {
          setIsClosing(false);
        }}
        onClose={() => {
          setIsClosing(false);
        }}
        onConfirm={() => {
          updateStatus.mutate(
            {
              feedbackId: feedback.id,
              payload: { status: "closed", roadmapSummary: null },
            },
            {
              onSuccess: (response) => {
                if (!response.error?.message) {
                  setIsClosing(false);
                }
              },
            },
          );
        }}
        title="Close this feedback?"
      />
    </ContextMenu>
  );
};

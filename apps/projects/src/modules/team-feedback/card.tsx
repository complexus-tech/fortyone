"use client";

import { useState } from "react";
import Link from "next/link";
import { usePathname, useRouter, useSearchParams } from "next/navigation";
import { cn } from "lib";
import { CheckIcon, CloseIcon, RequestsIcon } from "icons";
import { Box, ContextMenu, Flex, Text, TimeAgo } from "ui";
import { ConfirmDialog } from "@/components/ui";
import { useWorkspacePath } from "@/hooks";
import { usePlanTeamFeedback } from "./hooks/use-plan-feedback";
import { useUpdateTeamFeedbackStatus } from "./hooks/use-update-status";
import { FeedbackStatusPill } from "./status";
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
  const planFeedback = usePlanTeamFeedback();
  const updateStatus = useUpdateTeamFeedbackStatus();
  const [isClosing, setIsClosing] = useState(false);
  const isActive = pathname.includes(feedback.id);
  const isLinked = feedback.storyLinks.some((link) => link.isPrimary);
  const canPlan = !isLinked && feedback.status !== "closed";
  const status = searchParams.get("status");
  const feedbackHref = withWorkspace(
    `/teams/${feedback.board.teamId}/feedback/${feedback.id}`,
  );
  const href = status
    ? `${feedbackHref}?status=${encodeURIComponent(status)}`
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

  return (
    <ContextMenu>
      <ContextMenu.Trigger>
        <Link
          className="block"
          href={href}
          prefetch={index <= 10 ? true : null}
        >
          <Box
            className={cn(
              "border-border hover:bg-surface-muted block cursor-pointer border-b-[0.5px] px-5 py-[0.655rem] transition md:px-4",
              { "bg-surface-muted": isActive },
            )}
          >
            <Flex align="center" className="mb-2" gap={2} justify="between">
              <Text className="line-clamp-1 flex-1 font-medium">
                {feedback.title}
              </Text>
              <Text className="shrink-0 text-[0.95rem]" color="muted">
                <TimeAgo timestamp={feedback.createdAt} />
              </Text>
            </Flex>
            <Flex align="center" gap={3} justify="between">
              <Flex align="center" className="min-w-0 flex-1" gap={2}>
                <span
                  className="size-2.5 shrink-0 rounded-full"
                  style={{ backgroundColor: feedback.board.color }}
                />
                <Text className="line-clamp-1" color="muted">
                  {feedback.authorName} in {feedback.board.name}
                </Text>
              </Flex>
              <FeedbackStatusPill compact status={feedback.status} />
            </Flex>
          </Box>
        </Link>
      </ContextMenu.Trigger>
      <ContextMenu.Items>
        <ContextMenu.Group>
          <ContextMenu.Item
            disabled={!canPlan || planFeedback.isPending}
            onSelect={handlePlan}
          >
            <CheckIcon />
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
              setIsClosing(true);
            }}
          >
            <CloseIcon className="text-danger" />
            Close feedback...
          </ContextMenu.Item>
        </ContextMenu.Group>
      </ContextMenu.Items>
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

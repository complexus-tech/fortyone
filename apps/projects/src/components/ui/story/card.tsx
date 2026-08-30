"use client";
import Link from "next/link";
import { Avatar, Badge, Box, Button, Flex, Text } from "ui";
import { useDraggable } from "@dnd-kit/core";
import { cn } from "lib";
import { useQueryClient } from "@tanstack/react-query";
import { memo, useCallback, useEffect, useMemo, useRef } from "react";
import { useSession } from "@/lib/auth/client";
import type { Story as StoryProps } from "@/modules/stories/types";
import type { DetailedStory } from "@/modules/story/types";
import { getStoryPath } from "@/shared/routing/story";
import { useMediaQuery, useUserRole, useWorkspacePath } from "@/hooks";
import { storyKeys } from "@/modules/stories/constants";
import { getStory } from "@/modules/story/queries/get-story";
import { getStoryAttachments } from "@/modules/story/queries/get-attachments";
import { linkKeys } from "@/constants/keys";
import { getLinks } from "@/lib/queries/links/get-links";
import { useAutomationPreferences } from "@/lib/hooks/users/preferences";
import { useFigmaHandoffStatuses } from "@/lib/hooks/figma";
import { FigmaIcon } from "@/modules/settings/workspace/integrations/figma/icon";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { useBoard } from "../board-context";
import { MemberTooltip } from "../member-tooltip";
import { PriorityIcon } from "../priority-icon";
import { StoryContextMenu } from "./context-menu";
import { AssigneesMenu } from "./assignees-menu";
import { StoryProperties } from "./properties";

type StoryCardProps = {
  story: StoryProps;
  className?: string;
  handleStoryClick: (storyId: string) => void;
};

type StoryCardDragBindings = Pick<
  ReturnType<typeof useDraggable>,
  | "attributes"
  | "listeners"
  | "setActivatorNodeRef"
  | "setNodeRef"
  | "isDragging"
>;

type StoryCardContentProps = StoryCardProps &
  StoryCardDragBindings & {
    canPrefetch: () => boolean;
  };

const StoryCardContent = memo(function StoryCardContent({
  story,
  className,
  handleStoryClick,
  attributes,
  canPrefetch,
  isDragging,
  listeners,
  setActivatorNodeRef,
  setNodeRef,
}: StoryCardContentProps) {
  const { data: session } = useSession();
  const { data: preferences } = useAutomationPreferences();
  const openStoryInDialog = preferences?.openStoryInDialog;
  const isDesktop = useMediaQuery("(min-width: 768px)");
  const { userRole } = useUserRole();
  const { workspaceSlug, withWorkspace } = useWorkspacePath();
  const queryClient = useQueryClient();
  const { data: figmaHandoffStatuses } = useFigmaHandoffStatuses();

  const teamCode = story.team?.code;
  const storyReference = teamCode
    ? `${teamCode}-${story.sequenceId}`
    : String(story.sequenceId);
  const selectedAssignee = story.assignee;
  const figmaHandoffStatus = figmaHandoffStatuses?.[story.id];

  const { isColumnVisible, updateStory } = useBoard();
  const handleUpdate = useCallback(
    (data: Partial<DetailedStory>) => {
      updateStory(story.id, data);
    },
    [story.id, updateStory],
  );

  return (
    <Box
      onMouseEnter={() => {
        if (!canPrefetch()) return;

        if (session) {
          const ctx = { session, workspaceSlug };
          queryClient.prefetchQuery({
            queryKey: storyKeys.detail(workspaceSlug, story.id),
            queryFn: () => getStory(story.id, ctx),
            staleTime: DURATION_FROM_MILLISECONDS.MINUTE,
          });
          queryClient.prefetchQuery({
            queryKey: storyKeys.attachments(workspaceSlug, story.id),
            queryFn: () => getStoryAttachments(story.id, ctx),
            staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 5,
          });
          queryClient.prefetchQuery({
            queryKey: linkKeys.story(story.id),
            queryFn: () => getLinks(story.id, ctx),
            staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
          });
        }
      }}
    >
      <StoryContextMenu story={story}>
        <div ref={setNodeRef}>
          <Box
            className={cn(
              "border-border shadow-shadow hover:bg-surface-elevated dark:bg-surface w-[340px] rounded-xl border-[0.5px] bg-white px-4 pb-4 shadow-lg transition-colors duration-200 ease-linear select-none [contain-intrinsic-size:auto_9rem] [content-visibility:auto]",
              {
                "bg-surface-muted opacity-70": isDragging,
                "pointer-events-none opacity-40": story.id.startsWith("123"),
              },
              className,
            )}
          >
            <div
              className={cn("cursor-pointer pt-3 pb-1.5", {
                "cursor-grabbing": isDragging,
              })}
              ref={setActivatorNodeRef}
              {...attributes}
              {...listeners}
            >
              <Link
                className="flex justify-between gap-2"
                href={withWorkspace(
                  getStoryPath({
                    id: story.id,
                    sequenceId: story.sequenceId,
                    teamCode,
                  }),
                )}
                onClick={(e) => {
                  if (isDesktop && openStoryInDialog) {
                    e.preventDefault();
                    if (!isDragging) {
                      handleStoryClick(story.id);
                    }
                  }
                }}
              >
                <Text className="line-clamp-3 text-[1.1rem] leading-[1.4rem]">
                  {story.title}
                </Text>
                {isColumnVisible("ID") && (
                  <Text
                    className="shrink-0 text-[0.95rem] leading-[1.4rem] uppercase"
                    color="muted"
                  >
                    {storyReference}
                  </Text>
                )}
              </Link>
            </div>
            <Flex align="center" className="mt-1 gap-1.5" wrap>
              {figmaHandoffStatus ? (
                <Badge
                  className="gap-1"
                  color={
                    figmaHandoffStatus === "COMPLETED" ? "success" : "secondary"
                  }
                  title={
                    figmaHandoffStatus === "COMPLETED"
                      ? "Completed in Figma"
                      : "Ready for development in Figma"
                  }
                >
                  <FigmaIcon className="h-3 w-auto" />
                  {figmaHandoffStatus === "COMPLETED"
                    ? "Completed"
                    : "Dev ready"}
                </Badge>
              ) : null}
              {isColumnVisible("Assignee") && (
                <AssigneesMenu>
                  <MemberTooltip member={selectedAssignee}>
                    <span>
                      <AssigneesMenu.Trigger>
                        <Button
                          asIcon={story.collaboratorCount === 0}
                          className="gap-1 px-1"
                          color="tertiary"
                          disabled={userRole === "guest"}
                          size="xs"
                          type="button"
                          variant="outline"
                        >
                          <Avatar
                            name={
                              selectedAssignee?.fullName ||
                              selectedAssignee?.username
                            }
                            rounded="md"
                            size="xs"
                            src={selectedAssignee?.avatarUrl}
                          />
                          {story.collaboratorCount > 0 ? (
                            <span
                              className="text-text-muted pr-0.5 text-xs"
                              title={`${story.collaboratorCount} collaborator${story.collaboratorCount === 1 ? "" : "s"}`}
                            >
                              +{story.collaboratorCount}
                            </span>
                          ) : null}
                        </Button>
                      </AssigneesMenu.Trigger>
                    </span>
                  </MemberTooltip>
                  <AssigneesMenu.Items
                    assigneeId={story.assigneeId}
                    onAssigneeSelected={(assigneeId) => {
                      handleUpdate({ assigneeId });
                    }}
                    teamId={story.teamId}
                  />
                </AssigneesMenu>
              )}
              <StoryProperties
                {...story}
                asKanban
                handleUpdate={handleUpdate}
                teamCode={teamCode}
              />
            </Flex>
          </Box>
        </div>
      </StoryContextMenu>
    </Box>
  );
});

export const StoryCardPreview = ({
  className,
  story,
}: Pick<StoryCardProps, "className" | "story">) => {
  const teamCode = story.team?.code;
  const storyReference = teamCode
    ? `${teamCode}-${story.sequenceId}`
    : String(story.sequenceId);
  const selectedAssignee = story.assignee;

  return (
    <Box
      className={cn(
        "border-border shadow-shadow dark:bg-surface h-full min-h-28 w-[340px] rounded-xl border bg-white px-4 py-3 shadow-lg select-none",
        className,
      )}
    >
      <Flex align="start" gap={2} justify="between">
        <Text className="line-clamp-3 text-[1.1rem] leading-[1.4rem]">
          {story.title}
        </Text>
        <Text
          className="shrink-0 text-[0.95rem] leading-[1.4rem] uppercase"
          color="muted"
        >
          {storyReference}
        </Text>
      </Flex>
      <Flex align="center" className="mt-3 gap-1.5" wrap>
        <Avatar
          name={selectedAssignee?.fullName || selectedAssignee?.username}
          rounded="md"
          size="xs"
          src={selectedAssignee?.avatarUrl}
        />
        <Flex
          align="center"
          className="border-border text-text-secondary h-6 gap-1 rounded-md border px-2 text-xs"
        >
          <PriorityIcon className="h-4" priority={story.priority} />
          {story.priority}
        </Flex>
        {story.estimateValue ? (
          <Badge color="secondary">{story.estimateValue}</Badge>
        ) : null}
      </Flex>
    </Box>
  );
};

export const StoryCard = (props: StoryCardProps) => {
  const { story } = props;
  const dragData = useMemo(() => ({ story }), [story]);
  const {
    active,
    attributes,
    isDragging,
    listeners,
    setActivatorNodeRef,
    setNodeRef,
  } = useDraggable({
    data: dragData,
    id: story.id,
  });
  const isDragActiveRef = useRef(Boolean(active));
  useEffect(() => {
    isDragActiveRef.current = Boolean(active);
  }, [active]);
  const canPrefetch = useCallback(() => !isDragActiveRef.current, []);

  return (
    <StoryCardContent
      {...props}
      attributes={attributes}
      canPrefetch={canPrefetch}
      isDragging={isDragging}
      listeners={listeners}
      setActivatorNodeRef={setActivatorNodeRef}
      setNodeRef={setNodeRef}
    />
  );
};

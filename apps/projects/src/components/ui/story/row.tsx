"use client";
import Link from "next/link";
import { Flex, Text, Tooltip, Avatar, Checkbox, Box } from "ui";
import { useDraggable } from "@dnd-kit/core";
import { cn } from "lib";
import { useQueryClient } from "@tanstack/react-query";
import { ArrowRight2Icon, StoryIcon, SubStoryIcon } from "icons";
import { useState } from "react";
import type { KeyboardEvent } from "react";
import { useRouter } from "next/navigation";
import { useSession } from "@/lib/auth/client";
import type {
  Story as StoryProps,
  StoryPriority,
} from "@/modules/stories/types";
import type { DetailedStory } from "@/modules/story/types";
import type { UserSummary } from "@/types";
import { getStoryPath } from "@/modules/story/utils/story-url";
import { useUpdateStoryMutation } from "@/modules/story/hooks/update-mutation";
import {
  useMediaQuery,
  useTerminology,
  useUserRole,
  useWorkspacePath,
} from "@/hooks";
import { storyKeys } from "@/modules/stories/constants";
import { getStory } from "@/modules/story/queries/get-story";
import { getStoryAttachments } from "@/modules/story/queries/get-attachments";
import { linkKeys } from "@/constants/keys";
import { getLinks } from "@/lib/queries/links/get-links";
import { useAutomationPreferences } from "@/lib/hooks/users/preferences";
import { useBoard } from "../board-context";
import { MemberTooltip } from "../member-tooltip";
import { PriorityIcon } from "../priority-icon";
import { RowWrapper } from "../row-wrapper";
import { StoryStatusIcon } from "../story-status-icon";
import { AssigneesMenu } from "./assignees-menu";
import { StoryContextMenu } from "./context-menu";
import { DragHandle } from "./drag-handle";
import { StoryProperties } from "./properties";

export const StoryRowPreview = ({
  assignee,
  onSelect,
  priority,
  reference,
  statusId,
  statusName,
  title,
}: {
  assignee?: UserSummary | null;
  onSelect: () => void;
  priority: StoryPriority;
  reference: string;
  statusId?: string | null;
  statusName?: string;
  title: string;
}) => {
  const handleKeyDown = (event: KeyboardEvent<HTMLDivElement>) => {
    if (event.key !== "Enter" && event.key !== " ") return;

    event.preventDefault();
    onSelect();
  };

  return (
    <RowWrapper
      className="cursor-pointer gap-4 border-b-0 px-6 md:px-6"
      onClick={onSelect}
      onKeyDown={handleKeyDown}
      role="button"
    >
      <Flex align="center" className="min-w-0 flex-1 select-none" gap={2}>
        <Text
          className="flex min-w-[6ch] shrink-0 items-center gap-1 truncate text-[0.95rem]"
          color="muted"
        >
          {reference}
        </Text>
        <Text
          className="line-clamp-1 min-w-0 hover:opacity-90"
          fontWeight="medium"
        >
          {title}
        </Text>
      </Flex>
      <Flex align="center" className="text-text-muted shrink-0" gap={3}>
        <MemberTooltip member={assignee}>
          <span className="flex items-center">
            <Avatar
              name={assignee?.fullName || assignee?.username}
              size="sm"
              src={assignee?.avatarUrl}
            />
          </span>
        </MemberTooltip>
        <span className="flex items-center gap-2 text-sm sm:min-w-24">
          {statusId ? (
            <StoryStatusIcon statusId={statusId} />
          ) : (
            <span className="bg-text-muted size-3 rounded-sm opacity-50" />
          )}
          <span className="hidden truncate sm:inline">
            {statusName ?? "No status"}
          </span>
        </span>
        <span className="flex items-center gap-2 text-sm sm:min-w-24">
          <PriorityIcon priority={priority} />
          <span className="hidden sm:inline">{priority}</span>
        </span>
      </Flex>
    </RowWrapper>
  );
};

export const StoryRow = ({
  story,
  isSubStory = false,
  isInSearch = false,
  handleStoryClick,
  className,
}: {
  story: StoryProps;
  isSubStory?: boolean;
  isInSearch?: boolean;
  className?: string;
  handleStoryClick: (storyId: string) => void;
}) => {
  const router = useRouter();
  const { data: session } = useSession();
  const [isExpanded, setIsExpanded] = useState(false);
  const queryClient = useQueryClient();
  const { userRole } = useUserRole();
  const { getTermDisplay } = useTerminology();
  const { workspaceSlug, withWorkspace } = useWorkspacePath();
  const isDesktop = useMediaQuery("(min-width: 768px)");
  const { attributes, listeners, setNodeRef, isDragging } = useDraggable({
    id: story.id,
  });
  const { selectedStories, setSelectedStories, isColumnVisible } = useBoard();
  const { data: preferences } = useAutomationPreferences();
  const openStoryInDialog = preferences?.openStoryInDialog;

  const teamCode = story.team?.code;
  const storyReference = teamCode
    ? `${teamCode}-${story.sequenceId}`
    : String(story.sequenceId);
  const selectedAssignee = story.assignee;

  const { mutate } = useUpdateStoryMutation();

  const handleUpdate = (data: Partial<DetailedStory>) => {
    if ("id" in data) {
      mutate({
        storyId: data.id!,
        payload: data,
      });
    } else {
      mutate({
        storyId: story.id,
        payload: data,
      });
    }
  };

  return (
    <Box
      onMouseEnter={() => {
        if (session) {
          const ctx = { session, workspaceSlug };
          queryClient.prefetchQuery({
            queryKey: storyKeys.detail(workspaceSlug, story.id),
            queryFn: () => getStory(story.id, ctx),
          });
          queryClient.prefetchQuery({
            queryKey: storyKeys.attachments(workspaceSlug, story.id),
            queryFn: () => getStoryAttachments(story.id, ctx),
          });
          queryClient.prefetchQuery({
            queryKey: linkKeys.story(story.id),
            queryFn: () => getLinks(story.id, ctx),
          });
        }
        router.prefetch(
          withWorkspace(
            getStoryPath({
              id: story.id,
              sequenceId: story.sequenceId,
              teamCode,
            }),
          ),
        );
      }}
    >
      <div ref={setNodeRef}>
        <StoryContextMenu story={story}>
          <RowWrapper
            className={cn(
              "@container gap-4",
              {
                "bg-surface-muted opacity-70": isDragging,
                "pointer-events-none opacity-40": story.id.startsWith("123"),
                "bg-surface-muted pl-10 md:pl-18": isSubStory,
              },
              className,
            )}
          >
            <Flex
              align="center"
              className="relative min-w-0 flex-1 select-none"
              gap={2}
            >
              {isInSearch ? <StoryIcon className="h-[1.1rem]" /> : null}
              {isSubStory || isInSearch ? null : (
                <DragHandle {...listeners} {...attributes} />
              )}
              <Checkbox
                checked={selectedStories.includes(story.id)}
                className="shrink-0 rounded md:absolute md:-left-[1.6rem]"
                disabled={userRole === "guest"}
                onCheckedChange={(checked) => {
                  setSelectedStories(
                    checked
                      ? [...selectedStories, story.id]
                      : selectedStories.filter(
                          (storyId) => storyId !== story.id,
                        ),
                  );
                }}
              />
              {isColumnVisible("ID") && (
                <Tooltip
                  title={`${getTermDisplay("storyTerm", { capitalize: true })} ID: ${storyReference}`}
                >
                  <Text
                    className={cn(
                      "flex min-w-[6ch] shrink-0 items-center gap-1 truncate text-[0.95rem] transition-colors",
                      {
                        "cursor-pointer dark:hover:text-white/80":
                          story.subStories.length > 0,
                      },
                    )}
                    color="muted"
                    onClick={() => {
                      setIsExpanded(!isExpanded);
                    }}
                    role="button"
                    tabIndex={0}
                  >
                    {storyReference}
                    {story.subStories.length > 0 && (
                      <ArrowRight2Icon
                        className={cn("h-4 shrink-0 transition-transform", {
                          "rotate-90": isExpanded,
                        })}
                        strokeWidth={3}
                      />
                    )}
                  </Text>
                </Tooltip>
              )}

              <Link
                className="flex min-w-0 flex-1 items-center gap-1.5"
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
                    handleStoryClick(story.id);
                  }
                }}
              >
                {isSubStory ? <SubStoryIcon className="shrink-0" /> : null}
                <Text
                  className="line-clamp-1 min-w-0 hover:opacity-90"
                  fontWeight="medium"
                >
                  {story.title}
                </Text>
              </Link>
            </Flex>
            <Flex align="center" className="shrink-0" gap={3}>
              <StoryProperties
                {...story}
                handleUpdate={handleUpdate}
                isExpanded={isExpanded}
                setIsExpanded={setIsExpanded}
                teamCode={teamCode}
              />
              {isColumnVisible("Assignee") && (
                <AssigneesMenu>
                  <MemberTooltip member={selectedAssignee}>
                    <span>
                      <AssigneesMenu.Trigger>
                        <button
                          className="flex items-center gap-1"
                          disabled={userRole === "guest"}
                          type="button"
                        >
                          <Avatar
                            name={
                              selectedAssignee?.fullName ||
                              selectedAssignee?.username
                            }
                            size="sm"
                            src={selectedAssignee?.avatarUrl}
                          />
                          {story.collaboratorCount > 0 ? (
                            <span
                              className="text-text-muted text-xs"
                              title={`${story.collaboratorCount} collaborator${story.collaboratorCount === 1 ? "" : "s"}`}
                            >
                              +{story.collaboratorCount}
                            </span>
                          ) : null}
                        </button>
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
            </Flex>
          </RowWrapper>
        </StoryContextMenu>
      </div>
      {isExpanded && story.subStories.length > 0 ? (
        <>
          {story.subStories.map((subStory) => (
            <StoryRow
              handleStoryClick={handleStoryClick}
              isSubStory
              key={subStory.id}
              story={{ ...subStory, subStories: [], labels: [] }}
            />
          ))}
        </>
      ) : null}
    </Box>
  );
};

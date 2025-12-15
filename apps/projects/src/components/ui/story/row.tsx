"use client";
import Link from "next/link";
import { Flex, Text, Tooltip, Avatar, Checkbox, Box, Button } from "ui";
import { useDraggable } from "@dnd-kit/core";
import { cn } from "lib";
import { useQueryClient } from "@tanstack/react-query";
import { ArrowRight2Icon, StoryIcon, SubStoryIcon } from "icons";
import { useState } from "react";
import { useRouter } from "next/navigation";
import { useSession } from "next-auth/react";
import type { Story as StoryProps } from "@/modules/stories/types";
import { slugify } from "@/utils";
import type { DetailedStory } from "@/modules/story/types";
import { useUpdateStoryMutation } from "@/modules/story/hooks/update-mutation";
import { useTeams } from "@/modules/teams/hooks/teams";
import { useUserRole, useMediaQuery } from "@/hooks";
import { useTeamMembers } from "@/lib/hooks/team-members";
import { storyKeys } from "@/modules/stories/constants";
import { getStory } from "@/modules/story/queries/get-story";
import { getStoryAttachments } from "@/modules/story/queries/get-attachments";
import { linkKeys } from "@/constants/keys";
import { getLinks } from "@/lib/queries/links/get-links";
import { RowWrapper } from "../row-wrapper";
import { useBoard } from "../board-context";
import { AssigneesMenu } from "./assignees-menu";
import { StoryContextMenu } from "./context-menu";
import { DragHandle } from "./drag-handle";
import { StoryProperties } from "./properties";
import { useAutomationPreferences } from "@/lib/hooks/users/preferences";

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
  const { data: teams = [] } = useTeams();
  const { data: members = [] } = useTeamMembers(story.teamId);
  const { userRole } = useUserRole();
  const isDesktop = useMediaQuery("(min-width: 768px)");
  const { attributes, listeners, setNodeRef, isDragging } = useDraggable({
    id: story.id,
  });
  const { selectedStories, setSelectedStories, isColumnVisible } = useBoard();
  const { data: preferences } = useAutomationPreferences();
  const openStoryInDialog = preferences?.openStoryInDialog;

  const teamCode = teams.find((team) => team.id === story.teamId)?.code;

  const selectedAssignee = members.find(
    (member) => member.id === story.assigneeId,
  );

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
        queryClient.prefetchQuery({
          queryKey: storyKeys.detail(story.id),
          queryFn: () => getStory(story.id, session!),
        });
        queryClient.prefetchQuery({
          queryKey: storyKeys.attachments(story.id),
          queryFn: () => getStoryAttachments(story.id, session!),
        });
        queryClient.prefetchQuery({
          queryKey: linkKeys.story(story.id),
          queryFn: () => getLinks(story.id, session!),
        });
        router.prefetch(`/story/${story.id}/${slugify(story.title)}`);
      }}
    >
      <div ref={setNodeRef}>
        <StoryContextMenu story={story}>
          <RowWrapper
            className={cn(
              "gap-4",
              {
                "bg-gray-50 opacity-70 dark:bg-dark-50/40 dark:opacity-50":
                  isDragging,
                "pointer-events-none opacity-40": story.id.startsWith("123"),
                "bg-gray-50/50 pl-10 dark:bg-dark-200/50 md:pl-[4.5rem]":
                  isSubStory,
              },
              className,
            )}
          >
            <Flex
              align="center"
              className="relative shrink select-none"
              gap={2}
            >
              {isInSearch ? <StoryIcon className="h-[1.1rem]" /> : null}
              {isSubStory || isInSearch ? null : (
                <DragHandle {...listeners} {...attributes} />
              )}
              <Checkbox
                checked={selectedStories.includes(story.id)}
                className="shrink-0 rounded-[0.35rem] md:absolute md:-left-[1.6rem]"
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
                <Tooltip title={`Story ID: ${teamCode}-${story.sequenceId}`}>
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
                    {teamCode}-{story.sequenceId}
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
                className="flex items-center gap-1.5"
                href={`/story/${story.id}/${slugify(story.title)}`}
                onClick={(e) => {
                  if (isDesktop && openStoryInDialog) {
                    e.preventDefault();
                    handleStoryClick(story.id);
                  }
                }}
              >
                {isSubStory ? <SubStoryIcon className="shrink-0" /> : null}
                <Text
                  className="line-clamp-1 hover:opacity-90"
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
                  <Tooltip
                    className="mr-2 py-2.5"
                    title={
                      selectedAssignee ? (
                        <Box>
                          <Flex gap={2}>
                            <Avatar
                              className="mt-0.5"
                              name={selectedAssignee.fullName}
                              src={selectedAssignee.avatarUrl}
                            />
                            <Box>
                              <Link
                                className="mb-2 flex gap-1"
                                href={`/profile/${selectedAssignee.id}`}
                              >
                                <Text fontSize="md" fontWeight="medium">
                                  {selectedAssignee.fullName}
                                </Text>
                                <Text color="muted" fontSize="md">
                                  ({selectedAssignee.username})
                                </Text>
                              </Link>
                              <Button
                                className="mb-0.5 ml-px px-2"
                                color="tertiary"
                                href={`/profile/${selectedAssignee.id}`}
                                size="xs"
                              >
                                Go to profile
                              </Button>
                            </Box>
                          </Flex>
                        </Box>
                      ) : null
                    }
                  >
                    <span>
                      <AssigneesMenu.Trigger>
                        <button
                          className="flex"
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
                        </button>
                      </AssigneesMenu.Trigger>
                    </span>
                  </Tooltip>
                  <AssigneesMenu.Items
                    assigneeId={selectedAssignee?.id}
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

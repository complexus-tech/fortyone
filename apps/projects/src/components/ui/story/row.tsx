"use client";
import Link from "next/link";
import { Flex, Text, Tooltip, Avatar, Checkbox, Box, Button } from "ui";
import { useDraggable } from "@dnd-kit/core";
import { cn } from "lib";
import type { Story as StoryProps } from "@/modules/stories/types";
import { slugify } from "@/utils";
import type { DetailedStory } from "@/modules/story/types";
import { useUpdateStoryMutation } from "@/modules/story/hooks/update-mutation";
import { useTeams } from "@/modules/teams/hooks/teams";
import { useMembers } from "@/lib/hooks/members";
import { useUserRole } from "@/hooks";
import { RowWrapper } from "../row-wrapper";
import { useBoard } from "../board-context";
import { AssigneesMenu } from "./assignees-menu";
import { StoryContextMenu } from "./context-menu";
import { DragHandle } from "./drag-handle";
import { StoryProperties } from "./properties";

export const StoryRow = ({ story }: { story: StoryProps }) => {
  const { data: teams = [] } = useTeams();
  const { data: members = [] } = useMembers();
  const { userRole } = useUserRole();
  const { attributes, listeners, setNodeRef, isDragging } = useDraggable({
    id: story.id,
  });
  const { selectedStories, setSelectedStories, isColumnVisible } = useBoard();

  const teamCode = teams.find((team) => team.id === story.teamId)?.code;

  const selectedAssignee = members.find(
    (member) => member.id === story.assigneeId,
  );

  const { mutate } = useUpdateStoryMutation();

  const handleUpdate = (data: Partial<DetailedStory>) => {
    mutate({
      storyId: story.id,
      payload: data,
    });
  };

  return (
    <div ref={setNodeRef}>
      <StoryContextMenu story={story}>
        <RowWrapper
          className={cn("gap-4", {
            "bg-gray-50 opacity-70 dark:bg-dark-50/40 dark:opacity-50":
              isDragging,
          })}
        >
          <Flex align="center" className="relative shrink select-none" gap={2}>
            <DragHandle {...listeners} {...attributes} />
            <Checkbox
              checked={selectedStories.includes(story.id)}
              className="absolute -left-[1.6rem] rounded-[0.35rem]"
              disabled={userRole === "guest"}
              onCheckedChange={(checked) => {
                setSelectedStories(
                  checked
                    ? [...selectedStories, story.id]
                    : selectedStories.filter((storyId) => storyId !== story.id),
                );
              }}
            />
            {isColumnVisible("ID") && (
              <Tooltip title={`Story ID: ${teamCode}-${story.sequenceId}`}>
                <Text
                  className="min-w-[6ch] truncate text-[0.99rem]"
                  color="muted"
                >
                  {teamCode}-{story.sequenceId}
                </Text>
              </Tooltip>
            )}
            <Link href={`/story/${story.id}/${slugify(story.title)}`}>
              <Text
                className="line-clamp-1 hover:opacity-90"
                fontWeight="medium"
              >
                {story.title}
              </Text>
            </Link>
          </Flex>
          <Flex align="center" className="shrink-0" gap={3}>
            <StoryProperties {...story} handleUpdate={handleUpdate} />
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
                          color="tertiary"
                          name={selectedAssignee?.fullName}
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
                />
              </AssigneesMenu>
            )}
          </Flex>
        </RowWrapper>
      </StoryContextMenu>
    </div>
  );
};

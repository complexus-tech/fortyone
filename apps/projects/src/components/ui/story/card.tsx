"use client";
import Link from "next/link";
import { Box, Flex, Button, Text, Avatar, Tooltip } from "ui";
import { useDraggable } from "@dnd-kit/core";
import { cn } from "lib";
import { useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { useSession } from "next-auth/react";
import type { Story as StoryProps } from "@/modules/stories/types";
import { slugify } from "@/utils";
import { useTeams } from "@/modules/teams/hooks/teams";
import type { DetailedStory } from "@/modules/story/types";
import { useUpdateStoryMutation } from "@/modules/story/hooks/update-mutation";
import { useMembers } from "@/lib/hooks/members";
import { useUserRole } from "@/hooks";
import { storyKeys } from "@/modules/stories/constants";
import { getStory } from "@/modules/story/queries/get-story";
import { getStoryAttachments } from "@/modules/story/queries/get-attachments";
import { linkKeys } from "@/constants/keys";
import { getLinks } from "@/lib/queries/links/get-links";
import { useBoard } from "../board-context";
import { StoryContextMenu } from "./context-menu";
import { AssigneesMenu } from "./assignees-menu";
import { StoryProperties } from "./properties";

export const StoryCard = ({
  story,
  className,
}: {
  story: StoryProps;
  className?: string;
}) => {
  const router = useRouter();
  const { data: session } = useSession();
  const { data: teams = [] } = useTeams();
  const { data: members = [] } = useMembers();
  const { userRole } = useUserRole();
  const queryClient = useQueryClient();

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

  const { attributes, listeners, setNodeRef, isDragging } = useDraggable({
    id: story.id,
  });

  const { isColumnVisible } = useBoard();
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
      <StoryContextMenu story={story}>
        <Box
          className={cn(
            "w-[340px] select-none rounded-[0.6rem] border-[0.5px] border-gray-100 bg-white px-4 pb-4 shadow-lg shadow-gray-100/50 backdrop-blur transition duration-200 ease-linear hover:bg-white/50 dark:border-dark-50 dark:bg-dark-200/60 dark:shadow-none dark:hover:bg-dark-200/90",
            {
              "bg-gray-50 opacity-70 dark:bg-dark-50/40 dark:opacity-50":
                isDragging,
              "pointer-events-none opacity-40": story.id === "123",
            },
            className,
          )}
        >
          <div
            className={cn("cursor-pointer pb-2 pt-3", {
              "cursor-grabbing": isDragging,
            })}
            ref={setNodeRef}
            {...attributes}
            {...listeners}
          >
            <Link
              className="flex justify-between gap-2"
              href={`/story/${story.id}/${slugify(story.title)}`}
            >
              <Text
                className="line-clamp-3 text-[1.1rem] leading-[1.6rem]"
                fontWeight="medium"
              >
                {story.title}
              </Text>
              {isColumnVisible("ID") && (
                <Text
                  className="shrink-0 text-[0.95rem] uppercase leading-[1.6rem]"
                  color="muted"
                  fontWeight="medium"
                >
                  {teamCode}-{story.sequenceId}
                </Text>
              )}
            </Link>
          </div>
          <Flex align="center" className="mt-2 gap-1.5" wrap>
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
                            name={
                              selectedAssignee.fullName ||
                              selectedAssignee.username
                            }
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
                      <Button
                        asIcon
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
                      </Button>
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
            <StoryProperties
              {...story}
              asKanban
              handleUpdate={handleUpdate}
              teamCode={teamCode}
            />
          </Flex>
        </Box>
      </StoryContextMenu>
    </Box>
  );
};

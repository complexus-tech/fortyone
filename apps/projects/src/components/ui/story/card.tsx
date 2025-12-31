"use client";
import Link from "next/link";
import { Box, Flex, Button, Text, Avatar } from "ui";
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
import { useMediaQuery, useUserRole } from "@/hooks";
import { storyKeys } from "@/modules/stories/constants";
import { getStory } from "@/modules/story/queries/get-story";
import { getStoryAttachments } from "@/modules/story/queries/get-attachments";
import { linkKeys } from "@/constants/keys";
import { getLinks } from "@/lib/queries/links/get-links";
import { useBoard } from "../board-context";
import { StoryContextMenu } from "./context-menu";
import { AssigneesMenu } from "./assignees-menu";
import { StoryProperties } from "./properties";
import { useAutomationPreferences } from "@/lib/hooks/users/preferences";
import { MemberTooltip } from "../member-tooltip";

export const StoryCard = ({
  story,
  className,
  handleStoryClick,
}: {
  story: StoryProps;
  className?: string;
  handleStoryClick: (storyId: string) => void;
}) => {
  const router = useRouter();
  const { data: session } = useSession();
  const { data: teams = [] } = useTeams();
  const { data: members = [] } = useMembers();
  const { data: preferences } = useAutomationPreferences();
  const openStoryInDialog = preferences?.openStoryInDialog;
  const isDesktop = useMediaQuery("(min-width: 768px)");
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
            "border-border dark:bg-dark-200/80 dark:hover:bg-dark-100/70 w-[340px] select-none rounded-[0.9rem] border-[0.5px] bg-white px-4 pb-4 shadow-lg shadow-gray-100/50 backdrop-blur transition duration-200 ease-linear hover:bg-white/50 dark:shadow-none",
            {
              "dark:bg-dark-50/40 bg-gray-50 opacity-70 dark:opacity-50":
                isDragging,
              "pointer-events-none opacity-40": story.id.startsWith("123"),
            },
            className,
          )}
        >
          <div
            className={cn("cursor-pointer pb-1.5 pt-3", {
              "cursor-grabbing": isDragging,
            })}
            ref={setNodeRef}
            {...attributes}
            {...listeners}
          >
            <Link
              className="flex justify-between gap-2"
              href={`/story/${story.id}/${slugify(story.title)}`}
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
                  className="shrink-0 text-[0.95rem] uppercase leading-[1.4rem]"
                  color="muted"
                >
                  {teamCode}-{story.sequenceId}
                </Text>
              )}
            </Link>
          </div>
          <Flex align="center" className="mt-1 gap-1.5" wrap>
            {isColumnVisible("Assignee") && (
              <AssigneesMenu>
                <MemberTooltip member={selectedAssignee}>
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
                </MemberTooltip>
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
